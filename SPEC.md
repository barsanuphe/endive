# What is endive?

endive manages epub files.

# Features

**TBC** tagged requirements can be done with calibre/sigil, may not be
implemented.

## Configuration

- [x] the library layout and epub filename can be defined by a configuration
file, optionnally using metadata, including: author ($a), title ($t), year of
publication ($y), language ($l), isbn ($i).
- [x] the configuration file allows defining author aliases, which are used for
renaming the epubs and in the database.
- [x] the configuration file can point to a list of directories to be used as
sources for retail epubs, and another entry can point to a similar list for
non-retail epubs.
- [x] the configuration file is located in the proper XDG configuration
directory.
- [x] the database file is located in the library root.
- [x] the configuration file can hold a Goodreads API key, to get additional
metadata.
- [ ] the configuration file can contain a wishlist (author/title).

## Importing epubs

- [x] endive can scrape special directories for epubs to import.
- [x] endive must log the hash and filename of all imported files, in a
dedicated database in the relevant XDG data directory.
- [x] already imported epubs (hash already logged) must be ignored during
import.
- [x] duplicates other than retail/non-retail versions of the same work are not
allowed. epubs are duplicates if they have the same author and title, and/or
ISBN.
- [x] epubs must be copied into the library, which contains the epubs and the
database.
- [x] retail epubs have a forced "[retail]" suffix
- [x] if a newly imported retail version of an ebook had a non-retail
counterpart already in the library, it trumps and replaces it, ie the non-retail
version is deleted.
- [x] if a newly imported epub (retail or not) has a duplicate in the library
that was tagged as needing replacement, it trumps and replaces it.
- [ ] if an imported epub is on the configuration wishlist, endive must remove
it from the wishlist.

## Library

### Epub metadata

- [x] endive must read epub metadata, including: author, title, year of
publication, publisher, language, description, ISBN.
- [x] if an ISBN10 number is found, convert to ISBN13.
- [x] endive can get additional metadata from Goodreads, in case the epub
metadata is incomplete.

### Database

- [x] endive can keep track of progression: unread, reading, read, in shortlist.
- [x] one or several series can be associated with an epub.
- [x] epubs can be flagged as needing replacement.
- [x] retail epubs are identified as such.
- [x] endive must calculate and store the sha256 hash of every epub.
- [x] the hash of retail epubs can be checked to detect unwanted modifications.
- [x] tags can be added to epubs.
- [x] the database must be easily exportable and searchable (JSON).
- [x] the database can contain the date when the epub was read.
- [x] when a metadata field is defined in both the epub metadata and the
database, endive must use the database version.
- [ ] the user can store in the database whether a physical copy of the book is
also available.
- [x] all metadata fields can be edited by the CLI.
-

### Organization

- [x] a given book can have retail and non-retail versions side by side.
- [x] if an epub has a retail version, the non-retail version is assumed to be
derived from the retail, ie their metadata are the same.
- [x] other duplicates are not allowed.
- [x] the only allowed ebook format is epub.
- [x] library organization can be refreshed by the user, upon modification of
the configuration files or of epub metadata.
- [x] the library cannot contain an empty directory after refresh.

### Search

- [x] epubs without retail versions can be listed.
- [x] the library can be searched with the following creteria:
    author, title, series, progress, retail, tags, description
- [x] search can be limited to a specific number of results (first or last
    books matching filter).

### User interface

- [x] all features can be accessed with a CLI.
- [ ] all features can be served over http.

## E-reader synchronization

- [x] endive can synchronize selected epubs with a USB-mounted KOBO e-reader.
- [ ] **TBC** endive can synchronize KOBO collections, especially regarding
reading progress.

