# Endive

[![GoDoc](https://godoc.org/github.com/barsanuphe/endive?status.svg)](https://godoc.org/github.com/barsanuphe/endive)
[![Build Status](https://travis-ci.org/barsanuphe/endive.svg?branch=master)](https://travis-ci.org/barsanuphe/endive)
[![codecov](https://codecov.io/gh/barsanuphe/endive/branch/master/graph/badge.svg)](https://codecov.io/gh/barsanuphe/endive)
[![GPLv3](https://img.shields.io/badge/license-GPLv3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0.en.html)
[![Go Report Card](https://goreportcard.com/badge/github.com/barsanuphe/endive)](https://goreportcard.com/report/github.com/barsanuphe/endive)

In early development.

Right now, it will probably just delete your precious books and insult you.


## What it is

**Endive** is an epub collection organizer.

It can import epub ebooks, rename them from metadata, and sort them in a folder
structure derived from metadata.

**Endive** just adds its own lightweight database to keep track of other
metadata, such as:
- if the epub is retail or not
- its reading status
- when it was read
- if it needs to be replaced with a better version
- additional information from Goodreads to compensate for the fact that in
general, the embedded metadata is awful and incomplete.

**Endive** does not modify the contents of the epubs it manages.
In fact, it can check and report if retail epubs have changed since they were
imported.

## Table of Contents

- [Prerequisites](#Prerequisites)
- [Installation](#installation)
- [Testing](#testing)
- [Configuration](#configuration)
- [Usage](#usage)
- [Third party libraries](#third-party-libraries)

## Prerequisites

Retail epubs have awful metadata.
Publishers do not spare the embedded OPF file much love.
This means: missing information, wildly different ways to include vital
information such as ISBN number (if it is included at all, or correctly
included).

To get as faithful as possible information about epub files, **Endive** relies
on getting information from Goodreads.
Since this means using the Goodreads API, you must have an account and request
[an API Key](https://www.goodreads.com/api/keys)
([terms and conditions](https://www.goodreads.com/api/terms)).

See the configuration instructions to find out what to do with that key.

## Installation

If you have a [working Go installation](https://golang.org/doc/install), just run:

    $ go get -u github.com/barsanuphe/endive
    $ go install ...endive

Note that **endive** uses *less* do display information, so it will only work
where *less* is available.

## Testing

Testing requires the `GR_API_KEY` environment variable to be set with your very
own Goodreads API key.

    $ export GR_API_KEY=XXXXXXXXX
    $ go test ./...

## Configuration

**Endive** uses a YAML configuration file, in the configuration XDG directory
(which should be `/home/user/.config/endive/`):

    # library location and database filename
    library_root: /home/user/endive
    database_filename: endive.json

    # $a author, $y year, $t title, $i isbn, $l language, $r retail status,
	# $c fiction/nonfiction, $g main genre, $p reading progress, $s series
    epub_filename_format: $a/$a ($y) $t

    # list of directories that will be scraped for ebooks,
    # and automatically flagged as retail or non-retail
    nonretail_source:
        - /home/user/nonretail
        - /home/user/nonretail_source2
    retail_source:
        - /home/user/retail

    # see prerequisites
    goodreads_api_key: XXXXXXXXXXXXXX

    # associate main alias to alternative aliases
    # only the main alias will be used by endive
    author_aliases:
        Alexandre Dumas:
            - Alexandre Dumas Père
        China Miéville:
            - China Mieville
            - China Miévile
        Richard K. Morgan:
            - Richard Morgan
        Robert Silverberg:
            - Robert K. Silverberg
        Jared Diamond:
            - Diamond, Jared
    tag_aliases:
        science-fiction:
            - sci-fi
            - sf
            - sciencefiction
    publisher_aliases:
        Tor:
            - Tom Doherty Associates


## Usage

You probably should not run it for now, if you value your files.

Import from retail sources:

    $ endive import retail

Import specific non-retail epub:

    $ endive import nonretail book.epub

List epubs in english written by (Charles) Stross:

    $ endive search language:en +author:stross

    # # # E N D I V E # # #

    Searching for 'language:en +author:stross'...
    -------  -------------------  -------------------------  ---------  ----------------------------------------------------------------------------
     ID       Author               Title                      Year       Filename
    -------  -------------------  -------------------------  ---------  ----------------------------------------------------------------------------
     2        Charles Stross       The Apocalypse Codex       2012       Charles Stross/Charles Stross (2012) The Apocalypse Codex [retail].epub

     24       Charles Stross       The Hidden Family          2005       Charles Stross/Charles Stross (2005) The Hidden Family [retail].epub
    -------  -------------------  -------------------------  ---------  ----------------------------------------------------------------------------

Note that search is powered by [bleve](https://github.com/blevesearch/bleve),
and therefore uses its
[syntax](http://www.blevesearch.com/docs/Query-String-Query/).

Available fields are: `author`, `title`, `year`, `language`, `tags`, `series`,
and probably a few more.

Same search, ordered by year:

    $ endive search language:en +author:stross sortby year

    # # # E N D I V E # # #

    Searching for 'language:en +author:stross'...
    -------  -------------------  -------------------------  ---------  ----------------------------------------------------------------------------
     ID       Author               Title                      Year       Filename
    -------  -------------------  -------------------------  ---------  ----------------------------------------------------------------------------
     24       Charles Stross       The Hidden Family          2005       Charles Stross/Charles Stross (2005) The Hidden Family [retail].epub

     2        Charles Stross       The Apocalypse Codex       2012       Charles Stross/Charles Stross (2012) The Apocalypse Codex [retail].epub
    -------  -------------------  -------------------------  ---------  ----------------------------------------------------------------------------

Results can be sorted by: id, author, title, series, genre, category, rating,
averagerating, year.

Show info about a book with a specific *ID*:

    $ endive info *ID*

Edit title for book with specific *ID*:

    $ endive metadata edit *ID* title "New Title"

Refresh library after configuration changes:

    $ endive refresh

List all books:

    $ endive list books

List all books that do not have a retail version:

    $ endive list nonretail

Sorting all books by author:

    $ endive list books sortby author

Sorting all books by year, limit to first 10 results:

    $ endive list books sortby year first 10

For other commands, see:

    $ endive help


## Third party libraries

**Endive** uses:

|                 | Library       |
| --------------- |:-------------:|
| Epub parser     | [github.com/barsanuphe/epubgo](https://github.com/barsanuphe/epubgo), forked from [github.com/meskio/epubgo](https://github.com/meskio/epubgo)             |
| Search          | [github.com/blevesearch/bleve](https://github.com/blevesearch/bleve) |
| CLI             | [github.com/codegangsta/cli](https://github.com/codegangsta/cli)     |
| Color output    | [github.com/ttacon/chalk](https://github.com/ttacon/chalk)           |
| Tables output   | [github.com/barsanuphe/gotabulate](https://github.com/barsanuphe/gotabulate), forked from [github.com/bndr/gotabulate](https://github.com/bndr/gotabulate) |
| XDG directories | [launchpad.net/go-xdg](https://launchpad.net/go-xdg)                 |
| YAML Parser     | [github.com/spf13/viper](https://github.com/spf13/viper)             |
| ISBN validator  | [github.com/moraes/isbn](https://github.com/moraes/isbn)             |
| Spinner         | [github.com/tj/go-spin](https://github.com/tj/go-spin)               |
| Diff            | [github.com/kylelemons/godebug/pretty](https://github.com/kylelemons/godebug/pretty)               |
