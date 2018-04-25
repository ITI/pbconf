package pbdatabase

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
	"database/sql"
	"fmt"
	logging "github.com/iti/pbconf/lib/pblogger"
	_ "github.com/mattn/go-sqlite3"
)

type AppDatabase struct {
	log logging.Logger
	*sql.DB
}

// Open returns a handle to the database connection, so that it can be closed by the main function
func Open(dbFile, loglevel string) AppDatabase {
	l, _ := logging.GetLogger("Database access")
	logging.SetLevel(loglevel, "Database access")
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		l.Fatal(err.Error())
	}
	return AppDatabase{l, db}
}

func (dbHandle *AppDatabase) LoadSchema() {
	sqlTables := []string{
		`CREATE TABLE IF NOT EXISTS Users(
		id INTEGER PRIMARY KEY,
	    name TEXT,
	    passwordHash TEXT,
	    email TEXT,
	    role TEXT
	);
	`,
		`CREATE TABLE IF NOT EXISTS Nodes(
		id INTEGER PRIMARY KEY,
	    name TEXT UNIQUE
	);
	`,
		`CREATE TABLE IF NOT EXISTS ConfigItems(
		id INTEGER PRIMARY KEY,
		key TEXT,
		value TEXT
	);
	`,
		`CREATE TABLE IF NOT EXISTS ConfigLines(
		node INTEGER,
		configItem INTEGER,
		FOREIGN KEY(node) REFERENCES Nodes(id),
		FOREIGN KEY(configItem) REFERENCES ConfigItems(id)
	);
	`,
		`CREATE TABLE IF NOT EXISTS DeviceConfigItems(
	    id INTEGER PRIMARY KEY,
	    key TEXT,
	    value TEXT
	);
	`,
		`CREATE TABLE IF NOT EXISTS Devices(
	    id INTEGER PRIMARY KEY,
	    name TEXT UNIQUE,
	    parentNode INTEGER,
	    FOREIGN KEY(parentNode) REFERENCES Nodes(id)
	);
	`,
		`CREATE TABLE IF NOT EXISTS DeviceConfigLines(
	    device INTEGER,
	    deviceConfigItem INTEGER,
	    FOREIGN KEY(device) REFERENCES Devices(id),
	    FOREIGN KEY(deviceConfigItem) REFERENCES DeviceConfigItems(id)
	);
	`,
		`CREATE TABLE IF NOT EXISTS Log(
		tableName TEXT UNIQUE,
		lastModified DATETIME
		);
	`}

	for _, sqlStmt := range sqlTables {
		_, err := dbHandle.Exec(sqlStmt)
		if err != nil {
			dbHandle.log.Info("%q: %s\n", err, sqlStmt)
			return
		}
	}
	dbHandle.createLogTableTriggers()
}

func (dbHandle *AppDatabase) createLogTableTriggers() {
	dbTables := []string{"Nodes", "Devices"}
	for _, table := range dbTables {
		sqlStmt := fmt.Sprintf("CREATE TRIGGER IF NOT EXISTS %sTrigger AFTER INSERT ON %s BEGIN INSERT OR REPLACE INTO Log(tablename, LastModified) Values('%s', datetime('now')); END;", table, table, table)
		_, err := dbHandle.Exec(sqlStmt)
		if err != nil {
			dbHandle.log.Debug("Create %sTableTrigger error: %s", table, err.Error())
		}
	}
	table := "Devices"
	sqlStmt := fmt.Sprintf("CREATE TRIGGER IF NOT EXISTS %sDelTrigger AFTER DELETE ON %s BEGIN INSERT OR REPLACE INTO Log(tablename, LastModified) Values('%s', datetime('now')); END;", table, table, table)
	_, err := dbHandle.Exec(sqlStmt)
	if err != nil {
		dbHandle.log.Debug("Create %sTableTrigger error: %s", table, err.Error())
	}
}

func (dbHandle *AppDatabase) LoadRootNode(root string) {
	node := PbNode{Name: root}
	exists, err := node.ExistsByName(*dbHandle)
	if err != nil {
		dbHandle.log.Error(fmt.Sprintf("Error checking for root node existence: %s", err.Error()))
		return
	}
	if !exists {
		err = node.Create(*dbHandle)
		if err != nil {
			dbHandle.log.Error(fmt.Sprintf("Error creating root node: %s", err.Error()))
			return
		}
	}
}

func (dbHandle *AppDatabase) RowExists(stmt, errtag string, args ...interface{}) (bool, error) {
	var result int
	err := dbHandle.QueryRow(stmt, args...).Scan(&result)
	if err != nil {
		dbHandle.log.Debug("%s error: %s", errtag, err.Error())
		return false, err
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}
