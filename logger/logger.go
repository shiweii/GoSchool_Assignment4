package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var (
	Trace   *log.Logger // Just about anything
	Info    *log.Logger // Important information
	Warning *log.Logger // Be concerned
	Error   *log.Logger // Critical problem)
	Fatal   *log.Logger
	file    *os.File
	err     error
)

func init() {
	t := time.Now()
	fileName := fmt.Sprintf("log_%d_%02d_%02d.log", t.Year(), int(t.Month()), t.Day())
	file, err = os.OpenFile("log/"+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	Trace = log.New(os.Stdout, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(io.MultiWriter(file, os.Stderr), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(file, os.Stderr), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(file, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Fatal = log.New(io.MultiWriter(file, os.Stderr), "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func CloseLogger() {
	err := file.Close()
	if err != nil {
		return
	}
}
