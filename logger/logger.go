// Package logger implements a simple logging package.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Instantiate different log levels
var (
	Trace   *log.Logger // Just about anything
	Info    *log.Logger // Important information
	Warning *log.Logger // Be concerned
	Error   *log.Logger // Critical problem
	Panic   *log.Logger // When encounter panic
	Fatal   *log.Logger // Failure
	file    *os.File
	err     error
)

// init creates and open a log file got logging,
// base on current date time a new logfile will be created (log_YYYY_MM_DD.log).
func init() {
	t := time.Now()
	fileName := fmt.Sprintf("log_%d_%02d_%02d.log", t.Year(), int(t.Month()), t.Day())
	file, err = os.OpenFile("log/"+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	Trace = log.New(io.MultiWriter(file, os.Stderr), "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(io.MultiWriter(file, os.Stderr), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(file, os.Stderr), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(file, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Panic = log.New(io.MultiWriter(file, os.Stderr), "PANIC: ", log.Ldate|log.Ltime|log.Lshortfile)
	Fatal = log.New(io.MultiWriter(file, os.Stderr), "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// CloseLogger closes the logfile when interrupted signal is detected.
func CloseLogger() {
	err := file.Close()
	if err != nil {
		return
	}
}
