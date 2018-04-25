package pblogger

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
	"errors"
	//"io"
	"os"
	"strings"
	"sync"

	logging "github.com/iti/go-logging"
	config "github.com/iti/pbconf/lib/pbconfig"
)

/*

Level 0 – Emergency; System is unstable; Node failure imminent.
Level 1 – Alert; Immediate action required; Node failure likely
Level 2 – Critical; An unrecoverable failure condition has occurred; Node will shut down.
Level 3 – Error; A recoverable failure condition has occurred.  Node will continue to operate, however functionality may be limited.
Level 4 – Warning; An operationally significant, unexpected event has occurred.  Node will continue to operate, however functionality may be limited.
Level 5 – Notice; An operationally significant, but expected or normal, event has occurred.  Node will continue to operate normally.
Level 6 – Informational; Non-critical informational messages.
Level 7 – Debug; Information useful to developers and of little value to operators.

*/

type ILogger interface {
	Panic(...interface{})            // Level 0
	Fatal(...interface{})            // Level 1
	Critical(string, ...interface{}) // Level 2
	Error(string, ...interface{})    // Level 3
	Warning(string, ...interface{})  // Level 4
	Notice(string, ...interface{})   // Level 5
	Info(string, ...interface{})     // Level 6
	Debug(string, ...interface{})    // Level 7
}

type Logger struct {
	logging.Logger

	// The Ring Backend, so upstream process can get them
	rb *RingBackend
}

var globalLogFile string

type Level logging.Level

const (
	defaultFormat = "%{time} %{module} (%{level}): %{message}"
)

var levelLock sync.Mutex

func InitLogger(cmdLineLevel string, cfg *config.Config, prefix string) error {
	var level string

	if cmdLineLevel == "" {
		level = cfg.Global.LogLevel
	} else {
		level = cmdLineLevel
	}

	backends := make([]logging.Backend, 0)

	if level == "" {
		level = "INFO"
	}

	// Initial output to stdout
	stdoutbe := logging.NewLogBackend(os.Stdout, prefix, 0)
	lvlStdoutBe := logging.AddModuleLevel(stdoutbe)

	// Set Initial log level
	ll, err := logging.LogLevel(strings.ToUpper(level))
	if err != nil {
		return errors.New("Log Level \"" + level + "\" not recognized")
	}
	lvlStdoutBe.SetLevel(ll, "")

	backends = append(backends, lvlStdoutBe)

	// Initial Ring Buffer backend
	rBe := NewRingBackend(ll, cfg)
	rBeFormatter := logging.NewBackendFormatter(rBe, logging.MustStringFormatter(defaultFormat))
	backends = append(backends, rBeFormatter)

	// If LogFile is defined, log to a file too
	if cfg.Global.LogFile != "" {
		f, err := os.OpenFile(cfg.Global.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err == nil {
			logBe := logging.NewLogBackend(f, prefix, 0)
			lvlLogBe := logging.AddModuleLevel(logBe)
			lvlLogBe.SetLevel(ll, "")
			backends = append(backends, lvlLogBe)
		}
	}
	globalLogFile = cfg.Global.LogFile
	// Set up Alarm Handler
	all, _ := logging.LogLevel(strings.ToUpper(cfg.Global.AlarmThreshold))
	alarmBE := NewAlarmBackend(all, cfg)
	backends = append(backends, alarmBE)

	logging.SetBackend(backends...)
	logging.SetFormatter(logging.MustStringFormatter(defaultFormat))

	return nil
}

func GetLogger(name string) (Logger, error) {
	l, err := logging.GetLogger(name)
	return Logger{*l, GetRing()}, err
}

func (l *Logger) GetGlobalLogFile() string {
	return globalLogFile
}
func LogLevel(level string) (Level, error) {
	lvl, err := logging.LogLevel(level)
	return Level(lvl), err
}
func (level Level) String() string {
	p := logging.Level(level)
	return p.String()
}
func SetLevel(level, name string) error {
	ll, err := logging.LogLevel(strings.ToUpper(level))
	if err != nil {
		return errors.New("Log Level \"" + level + "\" not recognized")
	}

	logging.SetLevel(ll, name)

	return nil
}
func GetLevel(moduleName string) string {
	lvl := logging.GetLevel(moduleName)
	return lvl.String()
}
func SetFormat(format string) error {
	f, err := logging.NewStringFormatter(format)
	if err != nil {
		return err
	}

	logging.SetFormatter(f)
	return nil
}

func (l *Logger) Log(level string, format string, args ...interface{}) {
	switch level {
	case "PANIC", "Panic", "panic":
		l.Panicf(format, args...)
	case "FATAL", "Fatal", "fatal":
		l.Fatalf(format, args...)
	case "ERROR", "Error", "error":
		l.Errorf(format, args...)
	case "WARNING", "Warning", "warning":
		l.Warningf(format, args...)
	case "NOTICE", "Notice", "notice":
		l.Noticef(format, args...)
	case "INFO", "Info", "info":
		l.Infof(format, args...)
	case "DEBUG", "Debug", "debug":
		l.Debugf(format, args...)
	}
}

func (l *Logger) Panic(format string, args ...interface{}) {
	l.Panicf(format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.Fatalf(format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.Warningf(format, args...)
}

func (l *Logger) Notice(format string, args ...interface{}) {
	l.Noticef(format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.Debugf(format, args...)
}

func (l *Logger) GetRing() *RingBackend {
	return l.rb
}
