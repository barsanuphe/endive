package main

type Epubs []Epub

// ListNonRetailOnly among known epubs.
func (e *Epubs) ListNonRetailOnly() (nonretail Epubs, err error) {
	return
}

// ListRetailOnly among known epubs.
func (e *Epubs) ListRetailOnly() (retail Epubs, err error) {
	return
}

// ListAuthors among known epubs.
func (e *Epubs) ListAuthors() (authors []string, err error) {
	return
}

// ListTags associated with known epubs.
func (e *Epubs) ListTags() (tags []string, err error) {
	return
}

// ListUntagged among known epubs.
func (e *Epubs) ListUntagged() (untagged Epubs, err error) {
	return
}

// ListWithTag among known epubs.
func (e *Epubs) ListWithTag(tag string) (tagged Epubs, err error) {
	return
}
