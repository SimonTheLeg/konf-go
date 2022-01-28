package log

import (
	"io"
	"log"
	"os"
)

var infoL *log.Logger
var warnL *log.Logger

func InitLogger(info, warn io.Writer) {
	infoL = log.New(info, "INFO: ", 0)
	warnL = log.New(warn, "WARN: ", 0)
}

func Info(format string, v ...interface{}) {
	infoL.Printf(format, v...)
}

func Warn(format string, v ...interface{}) {
	warnL.Printf(format, v...)
}

func init() {
	InitLogger(os.Stderr, os.Stderr)
}
