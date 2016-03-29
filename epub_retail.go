package main

type EpubRetail struct {
	Epub
}

// SetReadOnly a retail epub ebook.
func (e *EpubRetail) SetReadOnly() (err error) {
	return
}

// Check the retail epub integrity.
func (e *EpubRetail) Check() (hasNotChanged bool, err error) {
	// TODO
	return
}
