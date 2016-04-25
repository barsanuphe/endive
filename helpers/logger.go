package helpers

import (
	"fmt"
	"os"

	"github.com/op/go-logging"
	"launchpad.net/go-xdg"
)

// Logger provides a logger to both stdout and a log file (for debug).
var Logger *logging.Logger

// LogFile is the pointer to the log file, to be closed by the main function.
var LogFile *os.File

var format = logging.MustStringFormatter(
	`%{time:15:04:05.000} | %{level:.1s} | %{shortfunc} ▶ %{message}`,
)

// GetEndiveLogger is the main logger, ensure the log file is in the correct XDG directory.
func GetEndiveLogger(xdgPath string) (err error) {
	logPath, err := xdg.Data.Find(xdgPath)
	if err != nil {
		logPath, err = xdg.Data.Ensure(xdgPath)
		if err != nil {
			return
		}
	}
	return GetLogger(logPath)
}

// GetLogger returns a global logger
func GetLogger(name string) (err error) {
	Logger = logging.MustGetLogger(name)

	// TODO set log file in XDG dir
	fileName := name
	LogFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		return
	}

	// stdout logger
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	// show everything but DEBUG on stdout
	stdoutLeveled := logging.AddModuleLevel(stdout)
	stdoutLeveled.SetLevel(logging.INFO, "")

	// file log: everything
	fileLog := logging.NewLogBackend(LogFile, "", 0)
	fileLogFormatter := logging.NewBackendFormatter(fileLog, format)

	// Set the backends to be used.
	logging.SetBackend(stdoutLeveled, fileLogFormatter)
	Logger.Debug("Logger set up.")
	return
}
