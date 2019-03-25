// Package logging provides logging facility for the project.
package logging

import (
	"fmt"
	"log"
	"log/syslog"
)

// logger is a global instance of logger.
var logger *Logger

// Get layzily returns logger instance. There is only one logger for entire
// project. In other words the function implements singleton pattern.
func Get() *Logger {
	if logger == nil {
		var err error
		if logger, err = NewLogger(); err != nil {
			log.Fatalf("failed to instantiate logger: %s", err)
		}
	}
	return logger
}

// Log is a wrapper over Logger.Infof method for providing logging facilities
// to third-party libraries.
func Log(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Logger is a wrapper over built-in SysLog writer. It provides API similar to
// Logger type in standard library.
type Logger struct {
	collector *syslog.Writer
}

func NewLogger() (*Logger, error) {
	var priority = syslog.LOG_USER
	var tag = "nvim-bnf"

	if ptr, err := syslog.New(priority, tag); err != nil {
		return nil, err
	} else {
		return &Logger{ptr}, nil
	}
}

func (l *Logger) Close() error {
	return l.collector.Close()
}

func (l *Logger) Debugf(format string, args ...interface{}) (int, error) {
	var msg = fmt.Sprintf(format, args...)
	return len(msg), l.collector.Debug(msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) (int, error) {
	var msg = fmt.Sprintf(format, args...)
	return len(msg), l.collector.Err(msg)
}

func (l *Logger) Infof(format string, args ...interface{}) (int, error) {
	var msg = fmt.Sprintf(format, args...)
	return len(msg), l.collector.Info(msg)
}

func (l *Logger) Noticef(format string, args ...interface{}) (int, error) {
	var msg = fmt.Sprintf(format, args...)
	return len(msg), l.collector.Notice(msg)
}

func (l *Logger) Warnf(format string, args ...interface{}) (int, error) {
	var msg = fmt.Sprintf(format, args...)
	return len(msg), l.collector.Warning(msg)
}
