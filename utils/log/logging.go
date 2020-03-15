package log

import (
	"io"
	"log"
	"os"
)

var infoLogger = getInfoLogger()
var LogPath = "./ucp.log"

func getInfoLogger() *log.Logger {
	var Logger *log.Logger
	_, err := os.OpenFile(LogPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	mm := io.MultiWriter(os.Stdout)
	Logger = log.New(mm, "", log.Ldate|log.Ltime|log.Lshortfile)
	return Logger
}

func Println(v ...interface{}) {
	infoLogger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

func Fatalln(v ...interface{}) {
	infoLogger.Println(v...)
}

func Fatalf(format string, v ...interface{}) {
	infoLogger.Fatalf(format, v...)
}
