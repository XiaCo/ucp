package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

var InfoLogger *log.Logger

func Init(logPath string) {
	InfoLogger = setInfoLogger(logPath)
}

func setInfoLogger(logPath string) *log.Logger {
	var wtr io.Writer
	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		wtr = io.MultiWriter(os.Stdout, f)
	} else {
		wtr = os.Stdout
	}
	return log.New(wtr, "", log.Ldate|log.Ltime|log.Lshortfile)
	//log.SetOutput(wtr)
	//log.SetPrefix("")
	//log.SetFlags(log.Ldate|log.Ltime|log.Lshortfile)
}

func Println(v ...interface{}) {
	InfoLogger.Output(2, fmt.Sprintln(v...))
}

func Printf(format string, v ...interface{}) {
	InfoLogger.Output(2, fmt.Sprintf(format, v...))
}

func Fatalln(v ...interface{}) {
	InfoLogger.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}
