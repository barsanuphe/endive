# Endive

[![GoDoc](https://godoc.org/github.com/barsanuphe/endive?status.svg)](https://godoc.org/github.com/barsanuphe/endive)
[![Build Status](https://travis-ci.org/barsanuphe/endive.svg?branch=master)](https://travis-ci.org/barsanuphe/endive)
[![Report Card](http://goreportcard.com/badge/barsanuphe/endive)](http://goreportcard.com/report/barsanuphe/endive)
[![Coverage Status](https://coveralls.io/repos/github/barsanuphe/endive/badge.svg?branch=master)](https://coveralls.io/github/barsanuphe/endive?branch=master)

In very early development.
Right now, it will probably just delete your books and insult you.

## What it is

Endive is an epub collection organizer.

It can import epub ebooks, rename them from metadata, and sort them in a folder
structure derived from metadata.

Endive just adds its own lightweight JSON database to keep track of other
metadata, such as:
- if the epub is retail or not
- its reading status
- when it was read
- if it needs to be replaced with a better version
- and other things, including metadata which should be in the epub itself (for
example if a retail epub is missing the publication date, it can be stored in
the database instead).

Endive is designed to make sure the contents of retail epubs are untouched.
Non-retail epubs can be trumped by importing retail versions.

## How to install

If you have a working Go installation, just run:

    $ go get -u github.com/barsanuphe/endive

## How to run

You probably should not run it for now, if you value your files.

    $ endive help

## Third party libraries

Endive uses:

|                 | Library       |
| --------------- |:-------------:|
| Epub parser     | [github.com/barsanuphe/epubgo](https://github.com/barsanuphe/epubgo), forked from [github.com/meskio/epubgo](https://github.com/meskio/epubgo) |
| JSON search     | [github.com/blevesearch/bleve](https://github.com/blevesearch/bleve) |
| CLI             | [github.com/codegangsta/cli](https://github.com/codegangsta/cli)     |
| Color output    | [github.com/ttacon/chalk](https://github.com/ttacon/chalk)           |
| Tables output   | [github.com/bndr/gotabulate](https://github.com/bndr/gotabulate)     |
| XDG directories | [launchpad.net/go-xdg](https://launchpad.net/go-xdg)                 |
| YAML Parser     | [github.com/spf13/viper](https://github.com/spf13/viper)             |
