package log

import (
	"io"
	"log"
	"os"
)

var infoL *log.Logger
var warnL *log.Logger

// InitLogger initializes a new logger
// Initialization must be done, before logging funcs can be called
func InitLogger(info, warn io.Writer) {
	infoL = log.New(info, "INFO: ", 0)
	warnL = log.New(warn, "WARN: ", 0)
}

// Info prints the supplied format string using the Info logger
func Info(format string, v ...interface{}) {
	infoL.Printf(format, v...)
}

// Warn prints the supplied format string using the Warn logger
func Warn(format string, v ...interface{}) {
	warnL.Printf(format, v...)
}

func init() {
	InitLogger(os.Stderr, os.Stderr)
}
