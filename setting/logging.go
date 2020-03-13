package setting

import (
	"io"
	"log"
	"os"
)

var InfoLogger *log.Logger
var LogPath = "./ucp.log"

func GetInfoLogger() *log.Logger {
	if InfoLogger != nil {
		return InfoLogger
	}
	_, err := os.OpenFile(LogPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	mm := io.MultiWriter(os.Stdout)
	InfoLogger = log.New(mm, "", log.Ldate|log.Ltime|log.Lshortfile)
	return InfoLogger
}
