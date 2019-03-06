package logs

import (
	"fmt"
	"log"
	"os"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func Debug(format string, v ...interface{}) {
	print("[DEBUG]", format, v...)
}

func Info(format string, v ...interface{}) {
	print("[INF]", format, v...)
}

func Warn(format string, v ...interface{}) {
	print("[WARN]", format, v...)
}

func Error(format string, v ...interface{}) {
	print("[ERROR]", format, v...)
}

func Fatal(format string, v ...interface{}) {
	print("[FATAL]", format, v...)
	os.Exit(1)
}

func print(level, format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("%s %s\n", level, format), v...)
}
