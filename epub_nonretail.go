package main

// deals with non retail epubs
type EpubNonRetail struct {
	Epub
}

// WriteMetadata from endive DB to epub file.
func (e *EpubNonRetail) WriteMetadata() (err error) {
	// TODO: use epubgo to write new meta fields, except for author, title, year
	return
}
