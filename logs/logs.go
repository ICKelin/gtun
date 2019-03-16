package logs

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

func Debug(format string, v ...interface{}) {
	print("[D]", format, v...)
}

func Info(format string, v ...interface{}) {
	print("[I]", format, v...)
}

func Warn(format string, v ...interface{}) {
	print("[W]", format, v...)
}

func Error(format string, v ...interface{}) {
	print("[E]", format, v...)
}

func Fatal(format string, v ...interface{}) {
	print("[FATAL]", format, v...)
	os.Exit(1)
}

func print(level, format string, v ...interface{}) {
	_, path, line, _ := runtime.Caller(2)
	sp := strings.Split(path, "/")

	file := sp[len(sp)-1]
	log.Printf(fmt.Sprintf("%s:%d %s %s\n", file, line, level, format), v...)
}
