package main

// Library manages Epubs
type Library struct {
	Root              string
	ConfigurationFile string

	Epubs Epubs
}

func (l *Library) ImportRetail() (err error) {
	return err
}

func (l *Library) ImportNonRetail() (err error) {
	return err
}



