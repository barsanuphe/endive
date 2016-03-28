package main

// Series can track a series and an epub's position.
type Series struct {
	Name  string
	Index int
}

// Epub can manipulate an epub file.
type Epub struct {
	Filename        string
	RelativePath    string
	NewFilename     string
	NewRelativePath string
	Hash            string
	IsRetail        bool
	Progress        int
	Series          []Series
	Author          string
	Title           string
	PublicationYear int
	ReadDate        string // month
	Tags            []string
}
