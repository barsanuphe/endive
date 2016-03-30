# Endive

In early development.

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

## How to install

If you have a working Go installation, just run:

    $ go get -u github.com/barsanuphe/endive

## Third party libraries

Endive uses:

|               | Library       |
| ------------- |:-------------:|
| Epub parsing  | [github.com/barsanuphe/epubgo](https://github.com/barsanuphe/epubgo), forked from [github.com/meskio/epubgo](https://github.com/meskio/epubgo) |
| JSON search   | [github.com/blevesearch/bleve](https://github.com/blevesearch/bleve) |
| CLI           | [github.com/codegangsta/cli](https://github.com/codegangsta/cli)     |
| Color output  | [github.com/ttacon/chalk](https://github.com/ttacon/chalk)           |
