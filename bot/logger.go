package bot

import "fmt"

type debugLogger struct{}

func (log *debugLogger) Debug(format string, args ...interface{}) { fmt.Printf("[DBG] " + format + "\n", args...) }
func (log *debugLogger) Info(format string, args ...interface{})  { fmt.Printf("[INF] " + format + "\n", args...) }
func (log *debugLogger) Warn(format string, args ...interface{})  { fmt.Printf("[WRN] " + format + "\n", args...) }
func (log *debugLogger) Error(format string, args ...interface{}) { fmt.Printf("[ERR] " + format + "\n", args...) }
