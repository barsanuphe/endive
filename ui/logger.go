package ui

import (
	"fmt"
	"os"

	"github.com/op/go-logging"
	"github.com/ttacon/chalk"
	"launchpad.net/go-xdg"
)

var format = logging.MustStringFormatter(
	`%{time:15:04:05.000} | %{level:.1s} | %{shortfunc} ▶ %{message}`,
)

// InitLogger is the main logger, ensure the log file is in the correct XDG directory.
func (ui *UI) InitLogger(xdgPath string) (err error) {
	logPath, err := xdg.Data.Find(xdgPath)
	if err != nil {
		logPath, err = xdg.Data.Ensure(xdgPath)
		if err != nil {
			return
		}
	}
	return ui.getLogger(logPath)
}

// getLogger returns a global logger
func (ui *UI) getLogger(name string) (err error) {
	ui.logger = logging.MustGetLogger(name)
	fileName := name
	ui.logFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		return
	}
	// file log: everything
	fileLog := logging.NewLogBackend(ui.logFile, "", 0)
	fileLogFormatter := logging.NewBackendFormatter(fileLog, format)
	logging.SetBackend(fileLogFormatter)
	ui.Debug("Logger set up.")
	return
}

// CloseLog correctly ends logging.
func (ui *UI) CloseLog() {
	if ui.logFile != nil {
		ui.logFile.Close()
	}
}

// Error message logging.
func (ui *UI) Error(msg string) {
	fmt.Println(ui.RedBold("ERROR: " + msg))
	if ui.logger != nil {
		ui.logger.Error(msg)
	}
}

// Errorf message logging
func (ui *UI) Errorf(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(ui.RedBold("ERROR: " + msg))
	if ui.logger != nil {
		ui.logger.Error(msg)
	}
}

// Warning message logging
func (ui *UI) Warning(msg string) {
	fmt.Println(ui.Red("WARNING: " + msg))
	if ui.logger != nil {
		ui.logger.Warning(msg)
	}
}

// Warningf message logging.
func (ui *UI) Warningf(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(ui.Red("WARNING: " + msg))
	if ui.logger != nil {
		ui.logger.Warning(msg)
	}
}

// Debug message logging
func (ui *UI) Debug(msg string) {
	if ui.logger != nil {
		ui.logger.Debug(msg)
	}
}

// Debugf message logging
func (ui *UI) Debugf(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	if ui.logger != nil {
		ui.logger.Debug(msg)
	}
}

// Info message logging
func (ui *UI) Info(msg string) {
	fmt.Println(msg)
	if ui.logger != nil {
		ui.logger.Info(msg)
	}
}

// Infof message logging
func (ui *UI) Infof(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(msg)
	if ui.logger != nil {
		ui.logger.Info(msg)
	}
}

// BlueBold outputs a string in blue bold.
func (ui *UI) BlueBold(in string) string {
	return chalk.Bold.TextStyle(chalk.Blue.Color(in))
}

// GreenBold outputs a string in green bold.
func (ui *UI) GreenBold(in string) string {
	return chalk.Bold.TextStyle(chalk.Green.Color(in))
}

// Green outputs a string in green.
func (ui *UI) Green(in string) string {
	return chalk.Green.Color(in)
}

// RedBold outputs a string in red bold.
func (ui *UI) RedBold(in string) string {
	return chalk.Bold.TextStyle(chalk.Red.Color(in))
}

// Red outputs a string in red.
func (ui *UI) Red(in string) string {
	return chalk.Red.Color(in)
}

// Yellow outputs a string in yellow.
func (ui *UI) Yellow(in string) string {
	return chalk.Yellow.Color(in)
}

// YellowBold outputs a string in yellow.
func (ui *UI) YellowBold(in string) string {
	return chalk.Bold.TextStyle(chalk.Yellow.Color(in))
}

// Choice message logging
func (ui *UI) Choice(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Print(ui.BlueBold(msg))
}

// Title message logging
func (ui *UI) Title(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(ui.GreenBold(msg))
}

// SubTitle message logging
func (ui *UI) SubTitle(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(ui.Green(" + " + msg))
}

// SubPart message logging
func (ui *UI) SubPart(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Println(ui.Green("\n ──┤") + ui.GreenBold(msg) + ui.Green("├──"))
}
