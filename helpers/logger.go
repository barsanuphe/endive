package helpers

import (
	"fmt"
	"os"

	"github.com/op/go-logging"
)

var format = logging.MustStringFormatter(
	`%{time:15:04:05.000} | %{level:.1s} | %{shortfunc} ▶ %{message}`,
)

func GetLogger(name string) (log *logging.Logger, logFile *os.File) {
	log = logging.MustGetLogger(name)

	// TODO set log file in XDG dir
	fileName := name
	logFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error opening file: %v", err)
		panic(err)
	}

	// stdout logger
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	// show everything but DEBUG on stdout
	stdoutLeveled := logging.AddModuleLevel(stdout)
	stdoutLeveled.SetLevel(logging.INFO, "")

	// file log: everything
	fileLog := logging.NewLogBackend(logFile, "", 0)
	fileLogFormatter := logging.NewBackendFormatter(fileLog, format)

	// Set the backends to be used.
	logging.SetBackend(stdoutLeveled, fileLogFormatter)
	return
}
