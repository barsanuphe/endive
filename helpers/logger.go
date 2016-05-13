package helpers

import (
	"fmt"
	"os"

	"github.com/op/go-logging"
	"github.com/ttacon/chalk"
	"launchpad.net/go-xdg"
)

// Logger provides a logger to both stdout and a log file (for debug).
var logger *logging.Logger

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

// CloseEndiveLogFile correctly ends logging.
func CloseEndiveLogFile() {
	if LogFile != nil {
		LogFile.Close()
	}
}

// GetLogger returns a global logger
func GetLogger(name string) (err error) {
	logger = logging.MustGetLogger(name)
	fileName := name
	LogFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		return
	}
	// file log: everything
	fileLog := logging.NewLogBackend(LogFile, "", 0)
	fileLogFormatter := logging.NewBackendFormatter(fileLog, format)
	logging.SetBackend(fileLogFormatter)
	Debug("Logger set up.")
	return
}

// BlueBold outputs a string in blue bold.
func BlueBold(in string) string {
	return chalk.Bold.TextStyle(chalk.Blue.Color(in))
}

// GreenBold outputs a string in green bold.
func GreenBold(in string) string {
	return chalk.Bold.TextStyle(chalk.Green.Color(in))
}

// RedBold outputs a string in red bold.
func RedBold(in string) string {
	return chalk.Bold.TextStyle(chalk.Red.Color(in))
}

// Red outputs a string in red.
func Red(in string) string {
	return chalk.Red.Color(in)
}

// Debug message logging
func Debug(msg string) {
	if logger != nil {
		logger.Debug(msg)
	}
}

// Debugf message logging
func Debugf(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	if logger != nil {
		logger.Debug(msg)
	}
}

// Warning message logging.
func Warning(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(Red("WARNING: " + msg))
	if logger != nil {
		logger.Warning(msg)
	}
}

// Error message logging.
func Error(msg string) {
	fmt.Println(RedBold("ERROR: " + msg))
	if logger != nil {
		logger.Error(msg)
	}
}

// Errorf message logging
func Errorf(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Printf(RedBold("ERROR: " + msg))
	if logger != nil {
		logger.Error(msg)
	}
}

// Info message logging
func Info(msg string) {
	fmt.Println(msg)
	if logger != nil {
		logger.Info(msg)
	}
}

// Infof message logging
func Infof(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(msg)
	if logger != nil {
		logger.Info(msg)
	}
}

// Choice message logging
func Choice(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(BlueBold(msg))
}

// Title message logging
func Title(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(GreenBold(msg))
}
