package reports

/***********************************************************************
   Copyright 2018 Information Trust Institute

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
***********************************************************************/

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	change "github.com/iti/pbconf/lib/pbchange"
	database "github.com/iti/pbconf/lib/pbdatabase"
	logging "github.com/iti/pbconf/lib/pblogger"
)

type PbReportQuery struct {
	Name      string `validate:"nonzero"`
	Query     string `validate:"nonzero"`
	TimeStamp string
	Period    string
}

type PbReportQueryHttp struct {
	PbReportQuery
	Author        string
	CommitMessage string
}

var KeyTime = "time"
var KeyModule = "module"
var KeyLevel = "level"
var KeyMsg = "message"
var FIELD_SEPARATOR = "\t"

func ParseQuery(b io.Reader) ([]op, error) {
	return Parse(b)
}

func RunReport(db database.AppDatabase, log logging.Logger, b io.Reader) (string, error) {
	var buffer string
	var err error
	parsed_stmt, err := ParseQuery(b)
	if err != nil {
		return "", err
	}
	stmt := parsed_stmt[0]
	switch stmt.op {
	case "selectdb":
		buffer, err = db_reporter(db, log, stmt)
	case "selectcme":
		buffer, err = cme_reporter(log, stmt)
	case "selectlog":
		buffer, err = log_reporter(log, stmt)
	default:
		return "", nil
	}
	if err != nil {
		log.Debug("RunReport: Error: %s", err.Error())
	}
	return buffer, err
}

func db_reporter(db database.AppDatabase, log logging.Logger, stmt op) (string, error) {
	var buffer bytes.Buffer
	query_str := fmt.Sprintf("SELECT %s FROM %s", stmt.fieldlist, stmt.source)
	if stmt.clauses.clause != nil { //construct the where clause string
		where_clause := ""
		for i, cond := range stmt.clauses.clause {
			where_clause = where_clause + cond.op1 + cond.operator + cond.op2
			if i < len(stmt.clauses.clause)-1 {
				where_clause = where_clause + " " + stmt.clauses.logical_cond[i] + " "
			}
		}
		query_str = query_str + " WHERE " + where_clause
	}
	log.Debug("Query to db:%s", query_str)
	rows, err := db.Query(query_str)
	if err != nil {
		log.Debug("1. Error: %s", err.Error())
		return "", err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	nrCols := len(columns)
	values := make([]interface{}, nrCols)
	valuePtrs := make([]interface{}, nrCols)
	//output the column names
	for i, colName := range columns {
		buffer.WriteString(colName)
		if i < len(columns)-1 {
			buffer.WriteString(FIELD_SEPARATOR)
		}
	}
	buffer.WriteString("\n")
	for rows.Next() {
		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Debug("2. Error: %s", err.Error())
			return "", err
		}

		for i, _ := range columns {
			val := values[i]
			switch valType := val.(type) {
			case []byte:
				buffer.WriteString(string(valType))
			case int32, int64:
				buffer.WriteString(fmt.Sprintf("%v", valType))
			default:
				fmt.Printf("%%%%%%%%%% got: %v", valType)
			}
			if i < len(columns)-1 {
				buffer.WriteString(FIELD_SEPARATOR)
			}
		}
		buffer.WriteString("\n")
	}
	if err := rows.Err(); err != nil {
		log.Debug("3. Error: %s", err.Error())
		return "", err
	}
	return buffer.String(), nil
}

func cme_reporter(log logging.Logger, stmt op) (string, error) {
	var buffer bytes.Buffer
	fields := strings.Split(stmt.fieldlist, ",")

	engine, err := change.GetCMEngine(nil)
	if err != nil {
		log.Debug("Could not get an instance of Change management Engine.")
		return "", err
	}
	cmType := change.StringToCMType(strings.ToUpper(stmt.source))

	if fields[0] == "DIFF" {
		diffIds := make([]string, 0, len(stmt.clauses.clause))
		for _, cond := range stmt.clauses.clause {
			diffIds = append(diffIds, cond.op2)
		}
		diff_reader, err := engine.Diff(cmType, "", diffIds...)
		if err != nil {
			return "", err
		}
		_, err = buffer.ReadFrom(diff_reader)
		return buffer.String(), err
	}

	var limit, ll_limit int
	if stmt.limit != "" {
		limit, err = strconv.Atoi(stmt.limit)
		if err != nil {
			log.Debug("cme_reporter: Got error converting limit to int:%s", err.Error())
			return "", err
		}
		ll_limit = limit
		if stmt.clauses.clause != nil {
			ll_limit = 0 //if there is a WHERE clause, dont limit what is returned prematurely
		}
	}
	logLines, err := engine.Log(cmType, "", ll_limit)
	if err != nil {
		log.Debug("Error getting log lines from cme engine:%s", err.Error())
		return "", err
	}
	//if WHERE clause exists, check that all operand1 exist in struct fields
	if stmt.clauses.clause != nil {
		struct_fields := make(map[string]bool)
		ll := reflect.ValueOf(logLines[0]) //we want to check where clause operand1 against one instance of loglines, and skip checking every line
		//put all the fields of LogLine struct into a map
		for i := 0; i < ll.NumField(); i++ {
			struct_fields[ll.Type().Field(i).Name] = true
		}
		for _, cond := range stmt.clauses.clause {
			if !(struct_fields[cond.op1]) {
				return "", errors.New("Illegal condition in the WHERE clause, operand 1 must be one of the fields of the LogLine struct")
			}
			if cond.op1 != "Time" && (cond.operator == "<" || cond.operator == ">" || cond.operator == "<=" || cond.operator == ">=") {
				return "", errors.New("Illegal condition in the WHERE clause, operator is not defined for this field.")
			}
		}
	}
	//output the headers
	dataFields := make([]string, 0, len(fields)) //make a slice atleast 1 long(* or actual fields)
	if len(fields) == 1 && fields[0] == "*" {
		ll := reflect.ValueOf(logLines[0])
		for i := 0; i < ll.NumField(); i++ {
			fieldName := ll.Type().Field(i).Name
			buffer.WriteString(fieldName)
			dataFields = append(dataFields, fieldName)
			if i < ll.NumField()-1 {
				buffer.WriteString(FIELD_SEPARATOR)
			}
		}
		buffer.WriteString("\n")
	} else {
		for i, name := range fields {
			buffer.WriteString(name)
			dataFields = append(dataFields, name)
			if i < len(fields)-1 {
				buffer.WriteString(FIELD_SEPARATOR)
			}
		}
		buffer.WriteString("\n")
	}
	//now output data
	ll_limit = limit
	for i, _ := range logLines {
		line := reflect.ValueOf(logLines[i])
		//should this line be skipped
		skip_line, err := cme_conditional(logLines[i], stmt.clauses)
		if err != nil {
			return "", err
		}
		if skip_line == true {
			continue
		} else if limit > 0 {
			if ll_limit > 0 {
				ll_limit -= 1
			} else {
				continue
			}
		}
		//print specified fields
		for p, _ := range dataFields {
			field := strings.Trim(dataFields[p], " ") //make sure field doesnt start or have spaces
			structValue := reflect.Indirect(line).FieldByName(field)
			if field == "Time" {
				buffer.WriteString(structValue.Interface().(time.Time).String())
			} else {
				buffer.WriteString(structValue.String())
			}
			if p < len(dataFields)-1 {
				buffer.WriteString(FIELD_SEPARATOR)
			}
		}
		buffer.WriteString("\n")
	}
	return buffer.String(), nil
}

func cme_conditional(line change.LogLine, clauses cmpd_clause) (bool, error) {
	skip_lines := make([]bool, len(clauses.clause))
	var err error
	for i, rel_clause := range clauses.clause {
		skip_lines[i], err = cme_single_conditional(line, rel_clause)
		if err != nil {
			return true, err
		}
	}
	//now all relational conditions have been evaluated
	skip_stmt := false
	if len(skip_lines) > 0 {
		skip_stmt = skip_lines[0] //where conditions could cause us to skip writing current line to output report
	}
	for i, op := range clauses.logical_cond {
		logical_op := strings.ToUpper(op)
		if logical_op == "AND" {
			skip_stmt = skip_stmt || skip_lines[i+1]
		} else if logical_op == "OR" {
			skip_stmt = skip_stmt && skip_lines[i+1]
		}
	}
	return skip_stmt, nil
}

func cme_single_conditional(line change.LogLine, cond relational_cond) (bool, error) {
	var opTime, llTime time.Time
	var err error
	if cond.op1 == "Time" {
		opTime, err = time.Parse(time.RFC3339, cond.op2)
		if err != nil {
			return true, err
		}
		llTime = line.Time
	}
	ll := reflect.ValueOf(line)
	val_ll := reflect.Indirect(ll).FieldByName(cond.op1).String()
	switch cond.operator {
	case "=":
		if cond.op1 == "Time" && !opTime.Equal(llTime) {
			return true, nil
		} else if cond.op1 != "Time" && cond.op2 != val_ll {
			return true, nil
		}
	case "<>":
		if cond.op1 == "Time" && opTime.Equal(llTime) {
			return true, nil
		} else if cond.op1 != "Time" && cond.op2 == val_ll {
			return true, nil
		}
	case "<":
		if cond.op1 == "Time" && (llTime.After(opTime) || llTime.Equal(opTime)) {
			return true, nil
		}
	case ">":
		if cond.op1 == "Time" && (llTime.Before(opTime) || llTime.Equal(opTime)) {
			return true, nil
		}
	case "<=":
		if cond.op1 == "Time" && llTime.After(opTime) {
			return true, nil
		}
	case ">=":
		if cond.op1 == "Time" && llTime.Before(opTime) {
			return true, nil
		}
	case "CONTAINS":
		if strings.Contains(val_ll, cond.op2) == false {
			return true, nil
		}
	}
	return false, nil
}

func log_reporter(log logging.Logger, stmt op) (string, error) {
	fields := strings.Split(stmt.fieldlist, ",")
	for i, field := range fields { //convert all fields to lower case
		fields[i] = strings.ToLower(field)
	}
	conditions := stmt.clauses.clause
	for i, cond := range conditions { //convert all operand1 to lower case
		conditions[i].op1 = strings.ToLower(cond.op1)
	}

	useRingBuffer := false
	rfc3339Milli := "2006-01-02T15:04:05.999Z07:00"
	records := log.GetRing().GetValues()
	if conditions != nil { //if conditions contain time as operand1, we may be able to get away with using the ring buffer
		//first check all conditions contain allowed keywords only
		keywords := map[string]bool{KeyTime: true, KeyModule: true, KeyLevel: true, KeyMsg: true}
		for _, cond := range conditions {
			if !(keywords[cond.op1]) {
				return "", errors.New("Illegal condition in the WHERE clause, operand 1 must be time, module, level, or message")
			}
		}
		//if there is an OR condition, we cannot use ring buffer
		logicalOrExists := false
		for _, logical_op := range stmt.clauses.logical_cond {
			if logical_op == "OR" {
				logicalOrExists = true
			}
		}
		firstRecord, lastRecord := records[0], records[len(records)-1]
		timeConditions := filterConditions(conditions, KeyTime)
		if logicalOrExists == false && len(timeConditions) == 1 {
			opTime, err := time.Parse(time.RFC3339, timeConditions[0].op2)
			if err != nil {
				return "", err
			}
			switch {
			case timeConditions[0].operator == "=" && firstRecord.Time.Before(opTime) && opTime.Before(lastRecord.Time),
				timeConditions[0].operator == ">" && (firstRecord.Time.Before(opTime) || firstRecord.Time.Equal(opTime)),
				timeConditions[0].operator == ">=" && firstRecord.Time.Before(opTime):
				useRingBuffer = true
			}
		} else if logicalOrExists == false && len(timeConditions) == 2 {
			op1 := timeConditions[0].operator
			op1Time, err := time.Parse(time.RFC3339, timeConditions[0].op2)
			if err != nil {
				return "", err
			}
			op2 := timeConditions[1].operator
			op2Time, err := time.Parse(time.RFC3339, timeConditions[1].op2)
			if err != nil {
				return "", err
			}
			if ((op1 == "<" && (op1Time.Before(lastRecord.Time) || op1Time.Equal(lastRecord.Time))) ||
				(op1 == ">" && (op1Time.After(firstRecord.Time) || op1Time.Equal(firstRecord.Time))) ||
				(op1 == "<=" && op1Time.Before(lastRecord.Time)) ||
				(op1 == ">=" && op1Time.After(firstRecord.Time))) &&
				((op2 == "<" && (op2Time.Before(lastRecord.Time) || op2Time.Equal(lastRecord.Time))) ||
					(op2 == ">" && (op2Time.After(firstRecord.Time) || op2Time.Equal(firstRecord.Time))) ||
					(op2 == "<=" && op2Time.Before(lastRecord.Time)) ||
					(op2 == ">=" && op2Time.After(firstRecord.Time))) {
				useRingBuffer = true
			}
		}
	}
	//OUTPUT THE DATA HEADERS FOR BOTH RING BUFFER AND LOG FILE
	var buffer bytes.Buffer
	if len(fields) == 1 && fields[0] == "*" {
		buffer.WriteString(strings.Title(KeyTime) + FIELD_SEPARATOR + strings.Title(KeyModule) + FIELD_SEPARATOR + strings.Title(KeyLevel) + FIELD_SEPARATOR + strings.Title(KeyMsg))
	} else {
		for i, field := range fields {
			if field == KeyModule || field == KeyTime || field == KeyLevel || field == KeyMsg {
				buffer.WriteString(strings.Title(field))
				if i < len(fields)-1 {
					buffer.WriteString(FIELD_SEPARATOR)
				}
			}
		}
	}
	buffer.WriteString("\n")

	if useRingBuffer {
		//output data
		for _, record := range records {
			skip_stmt := false //where conditions could cause us to skip writing current line to output report
			loglevel := logging.Level(record.Level)
			moduleStr := record.Module
			loglineTime := record.Time
			logmsg := record.Message()
			var err error
			for _, cond := range conditions {
				switch cond.operator {
				case "=":
					skip_stmt, err = processEqualOperator(cond, loglevel, moduleStr)
				case "<>":
					skip_stmt, err = processNotEqualOperator(cond, loglevel, moduleStr)
				case "<":
					skip_stmt, err = processLessThanOperator(cond, loglevel, loglineTime)
				case ">":
					skip_stmt, err = processGreaterThanOperator(cond, loglevel, loglineTime)
				case "<=":
					skip_stmt, err = processLessThanEqualOperator(cond, loglevel, loglineTime)
				case ">=":
					skip_stmt, err = processGreaterThanEqualOperator(cond, loglevel, loglineTime)
				}
				if err != nil {
					return "", err
				}
				if skip_stmt == true {
					break
				}
			}
			if skip_stmt == true {
				continue
			}
			timeStr := loglineTime.Format(rfc3339Milli)
			levelStr := loglevel.String()
			for i, field := range fields {
				if field == KeyModule {
					buffer.WriteString(moduleStr)
				} else if field == KeyTime {
					buffer.WriteString(timeStr)
				} else if field == KeyLevel {
					buffer.WriteString(levelStr)
				} else if field == KeyMsg {
					buffer.WriteString(logmsg)
				} else {
					buffer.WriteString(timeStr + FIELD_SEPARATOR + moduleStr + FIELD_SEPARATOR + levelStr + FIELD_SEPARATOR + logmsg)
				}
				if i < len(fields)-1 {
					buffer.WriteString(FIELD_SEPARATOR)
				}
			}
			buffer.WriteString("\n")
		}
		return buffer.String(), nil
	}

	logFile := log.GetGlobalLogFile()
	err := logFileReporter(&buffer, logFile, fields, stmt.clauses)
	return buffer.String(), err
}

func logFileReporter(buffer *bytes.Buffer, logFile string, fields []string, clauses cmpd_clause) error {
	//file stuff
	conditions := clauses.clause
	if logFile == "" {
		fmt.Println("No logging to the Global LogFile found. Set it in the configuration file.")
		return errors.New("No logging to the Global LogFile found. Set it in the configuration file.")
	}
	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Got error while opening log file %s\n", logFile)
		return errors.New(fmt.Sprintf("Got error while opening log file %s\n", logFile))
	}
	defer file.Close()

	time_re := `([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.?[0-9]{0,3}-[0-9]{2}:[0-9]{2})`
	module_re := `([a-zA-Z\s:]+)`
	level_re := `\((\w+)\):`
	msg_re := `(.*)`
	re := regexp.MustCompile(time_re + `\s+` + module_re + `\s+` + level_re + `\s+` + msg_re)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		//convert the current log line into tokens
		matchStr := re.FindStringSubmatch(line)
		if matchStr == nil || len(matchStr) < 5 {
			fmt.Println("Could not parse the log line into tokens")
			fmt.Printf("Got matchs:%v\n", matchStr)
			return errors.New(fmt.Sprintf("Not able to parse logline:%s into tokens", line))
		}
		timeStr := strings.TrimSpace(matchStr[1])
		loglineTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			fmt.Printf("Could not convert timestamp %s to time in RFC 3339 format\n", timeStr)
			return err
		}
		moduleStr := matchStr[2]
		levelStr := matchStr[3]
		loglevel, err := logging.LogLevel(levelStr)
		if err != nil {
			fmt.Printf("Could not convert level %s in scanned line to log level\n", levelStr)
			return err
		}
		logmsg := matchStr[4]
		// end of converting the logline into tokens
		skip_lines := make([]bool, len(conditions))
		if conditions != nil {
			for i, cond := range conditions {
				skip_lines[i], err = log_single_conditional(cond, loglevel, moduleStr, loglineTime)
				// if err != nil or we are skipping the rest of the log line
				if err != nil {
					return err
				}
			}
		}
		skip_stmt := false
		if len(skip_lines) > 0 {
			skip_stmt = skip_lines[0] //where conditions could cause us to skip writing current line to output report
		}
		for i, op := range clauses.logical_cond {
			logical_op := strings.ToUpper(op)
			if logical_op == "AND" {
				skip_stmt = skip_stmt || skip_lines[i+1]
			} else if logical_op == "OR" {
				skip_stmt = skip_stmt && skip_lines[i+1]
			}
		}
		if skip_stmt == true {
			continue
		}
		for i, field := range fields {
			if field == KeyModule {
				buffer.WriteString(moduleStr)
			} else if field == KeyTime {
				buffer.WriteString(timeStr)
			} else if field == KeyLevel {
				buffer.WriteString(levelStr)
			} else if field == KeyMsg {
				buffer.WriteString(logmsg)
			} else {
				buffer.WriteString(timeStr + FIELD_SEPARATOR + moduleStr + FIELD_SEPARATOR + levelStr + FIELD_SEPARATOR + logmsg)
			}
			if i < len(fields)-1 {
				buffer.WriteString(FIELD_SEPARATOR)
			}
		}
		buffer.WriteString("\n")
	}
	if err = scanner.Err(); err != nil {
		fmt.Println("Scanner error while reading file Error: ", err)
		return err
	}
	return nil
}

// evaluates a single X>y relational expression
func log_single_conditional(cond relational_cond, loglevel logging.Level, moduleStr string, loglineTime time.Time) (bool, error) {
	switch cond.operator {
	case "=":
		return processEqualOperator(cond, loglevel, moduleStr)
	case "<>":
		return processNotEqualOperator(cond, loglevel, moduleStr)
	case "<":
		return processLessThanOperator(cond, loglevel, loglineTime)
	case ">":
		return processGreaterThanOperator(cond, loglevel, loglineTime)
	case "<=":
		return processLessThanEqualOperator(cond, loglevel, loglineTime)
	case ">=":
		return processGreaterThanEqualOperator(cond, loglevel, loglineTime)
	}
	return false, nil
}

func filterConditions(conditions []relational_cond, key string) []relational_cond {
	filteredConditions := make([]relational_cond, 0)
	for _, cond := range conditions {
		if cond.op1 == key {
			filteredConditions = append(filteredConditions, cond)
		}
	}
	return filteredConditions
}

func processEqualOperator(cond relational_cond, loglineLevel logging.Level, moduleStr string) (bool, error) {
	if cond.op1 == KeyMsg || cond.op1 == KeyTime {
		return true, errors.New("WHERE clause: message/time not implemented for = operator")
	}
	if cond.op1 == KeyLevel {
		level, err := logging.LogLevel(cond.op2)
		if err != nil {
			return true, err
		}
		if loglineLevel != level {
			return true, nil
		}
	}
	if cond.op1 == KeyModule && moduleStr != cond.op2 {
		return true, nil
	}
	return false, nil
}

func processNotEqualOperator(cond relational_cond, loglineLevel logging.Level, moduleStr string) (bool, error) {
	if cond.op1 == KeyMsg || cond.op1 == KeyTime {
		return true, errors.New("WHERE clause: message/time not implemented for <> operator")
	}
	if cond.op1 == KeyLevel {
		level, err := logging.LogLevel(cond.op2)
		if err != nil {
			return true, err
		}
		if loglineLevel == level {
			return true, nil
		}
	}
	if cond.op1 == KeyModule && moduleStr == cond.op2 {
		return true, nil
	}
	return false, nil
}

func processLessThanOperator(cond relational_cond, loglineLevel logging.Level, loglineTime time.Time) (bool, error) {
	if cond.op1 == KeyMsg || cond.op1 == KeyModule {
		return true, errors.New("WHERE clause: message/module not implemented for < operator")
	}

	if cond.op1 == KeyTime {
		duration, err := time.ParseDuration(cond.op2)
		if err == nil { //it is duration
			if time.Since(loglineTime) >= duration {
				return true, nil
			}
		} else { //if not duration, must be timestamp
			timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(cond.op2))
			if err == nil { //was recognized as a timestamp
				if loglineTime.After(timestamp) || loglineTime.Equal(timestamp) {
					return true, nil
				}
			} else {
				return true, err //was not recognized as duration or timestamp
			}
		}
	} else if cond.op1 == KeyLevel {
		level, err := logging.LogLevel(cond.op2)
		if err != nil {
			return true, err
		} else if loglineLevel <= level {
			return true, nil
		}
	}
	return false, nil
}

func processGreaterThanOperator(cond relational_cond, loglineLevel logging.Level, loglineTime time.Time) (bool, error) {
	if cond.op1 == KeyMsg || cond.op1 == KeyModule {
		return true, errors.New("WHERE clause: message/module not implemented for > operator")
	}

	if cond.op1 == KeyTime {
		duration, err := time.ParseDuration(cond.op2)
		if err == nil { //it is duration
			if time.Since(loglineTime) <= duration {
				return true, nil
			}
		} else { //if not duration, must be timestamp
			timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(cond.op2))
			if err == nil { //was recognized as a timestamp
				if loglineTime.Before(timestamp) || loglineTime.Equal(timestamp) {
					return true, nil
				}
			} else {
				return true, err //was not recognized as duration or timestamp
			}
		}
	} else if cond.op1 == KeyLevel {
		level, err := logging.LogLevel(cond.op2)
		if err != nil {
			return true, err
		} else if loglineLevel >= level {
			return true, nil
		}
	}
	return false, nil
}

func processLessThanEqualOperator(cond relational_cond, loglineLevel logging.Level, loglineTime time.Time) (bool, error) {
	if cond.op1 == KeyMsg || cond.op1 == KeyModule {
		return true, errors.New("WHERE clause: message/module not implemented for <= operator")
	}

	if cond.op1 == KeyTime {
		duration, err := time.ParseDuration(cond.op2)
		if err == nil { //it is duration
			if time.Since(loglineTime) > duration {
				return true, nil
			}
		} else { //if not duration, must be timestamp
			timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(cond.op2))
			if err == nil { //was recognized as a timestamp
				if loglineTime.After(timestamp) {
					return true, nil
				}
			} else {
				return true, err //was not recognized as duration or timestamp
			}
		}
	} else if cond.op1 == KeyLevel {
		level, err := logging.LogLevel(cond.op2)
		if err != nil {
			return true, err
		} else if loglineLevel < level {
			return true, nil
		}
	}
	return false, nil
}

func processGreaterThanEqualOperator(cond relational_cond, loglineLevel logging.Level, loglineTime time.Time) (bool, error) {
	if cond.op1 == KeyMsg || cond.op1 == KeyModule {
		return true, errors.New("WHERE clause: message/module not implemented for >= operator")
	}

	if cond.op1 == KeyTime {
		duration, err := time.ParseDuration(cond.op2)
		if err == nil { //it is duration
			if time.Since(loglineTime) < duration {
				return true, nil
			}
		} else { //if not duration, must be timestamp
			timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(cond.op2))
			if err == nil { //was recognized as a timestamp
				if loglineTime.Before(timestamp) {
					return true, nil
				}
			} else {
				return true, err //was not recognized as duration or timestamp
			}
		}
	} else if cond.op1 == KeyLevel {
		level, err := logging.LogLevel(cond.op2)
		if err != nil {
			return true, err
		} else if loglineLevel > level {
			return true, nil
		}
	}
	return false, nil
}
