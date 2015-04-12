package bot

import "fmt"
import "os"

const (
	LOG_LEVEL_DEBUG   = 1
	LOG_LEVEL_INFO    = 2
	LOG_LEVEL_WARNING = 3
	LOG_LEVEL_ERROR   = 4
)

type Logger interface {
	SetLevel(int)

	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warning(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Fatal(string, ...interface{})
}

type logger struct {
	level int
}

func NewLogger(level int) Logger {
	return &logger{level}
}

func (self *logger) SetLevel(level int)   {
	self.level = level
}

func (self *logger) Debug(format string, args ...interface{}) {
	self.printLine(LOG_LEVEL_DEBUG, "[dbg] " + format, args...)
}

func (self *logger) Info(format string, args ...interface{}) {
	self.printLine(LOG_LEVEL_INFO, "[inf] " + format, args...)
}

func (self *logger) Warning(format string, args ...interface{}) {
	self.printLine(LOG_LEVEL_WARNING, "[wrn] " + format, args...)
}

// satisfy the goirc library's Logger interface
func (self *logger) Warn(format string, args ...interface{}) {
	self.Warning(format, args...)
}

func (self *logger) Error(format string, args ...interface{}) {
	self.printLine(LOG_LEVEL_ERROR, "[err] " + format, args...)
}

func (self *logger) Fatal(format string, args ...interface{}) {
	self.printLine(LOG_LEVEL_ERROR, "[ftl] " + format, args...)
	os.Exit(1)
}

func (self *logger) printLine(level int, format string, args ...interface{}) {
	if level >= self.level {
		fmt.Printf(format + "\n", args...)
	}
}
