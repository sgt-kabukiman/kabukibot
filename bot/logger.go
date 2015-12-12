package bot

import (
	"fmt"
	"time"
)
import "os"

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

type Logger interface {
	SetLevel(int)

	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warning(string, ...interface{})
	Error(string, ...interface{})
	Fatal(string, ...interface{})
}

type logger struct {
	level int
}

func NewLogger(level int) Logger {
	return &logger{level}
}

func (self *logger) SetLevel(level int) {
	self.level = level
}

func (self *logger) Debug(format string, args ...interface{}) {
	self.printLine(LogLevelDebug, "[d] "+format, args...)
}

func (self *logger) Info(format string, args ...interface{}) {
	self.printLine(LogLevelInfo, "[i] "+format, args...)
}

func (self *logger) Warning(format string, args ...interface{}) {
	self.printLine(LogLevelWarning, "[W] "+format, args...)
}

func (self *logger) Error(format string, args ...interface{}) {
	self.printLine(LogLevelError, "[!] "+format, args...)
}

func (self *logger) Fatal(format string, args ...interface{}) {
	self.printLine(LogLevelError, "[F] "+format, args...)
	os.Exit(1)
}

func (self *logger) printLine(level int, format string, args ...interface{}) {
	if level >= self.level {
		line := fmt.Sprintf("[%s] %s\n", time.Now().Format("Mon Jan 2 2006 15:04:05 MST"), format)

		fmt.Printf(line, args...)
	}
}
