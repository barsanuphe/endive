# What is endive?

endive manages epub files.

# Features

**TBC** tagged requirements can be done with calibre/sigil, may not be
implemented.

## Configuration

- [ ] the library layout and epub filename can be defined by a configuration
file, optionnally using metadata, including: author ($a), title ($t), year of
publication ($y), language ($l).
- [ ] the configuration file allows defining author aliases, which are used for
renaming the epubs and in the database.
- [ ] a directory for retail epubs source and another for non-retail epubs can
be defined in the configuration file.
- [ ] the configuration file is located in the proper XDG configuration
directory.
- [ ] the database file is located in the library root.
- [ ] the configuration file can contain a wishlist (author/title).

## Importing epubs

- [ ] endive can scrape special directories for epubs to import.
- [ ] endive must log the hash and filename of all imported files.
- [ ] already imported epubs (hash already logged) must be ignored during
import.
- [ ] epubs must be copied into the library, which contains the epubs and the
database.
- [ ] retail epubs are to be given read-only permissions.
- [ ] if an imported epub is on the configuration wishlist, endive must remove
it from the wishlist.

## Library

### Epub metadata

- [x] endive must read epub metadata, including: author, title, year of
publication, language.

### Database

- [ ] endive can keep track of progression: unread, reading, read, in shortlist.
- [x] one or several series can be associated with an epub.
- [ ] epubs can be flagged as needing replacement.
- [ ] retail epubs are identified as such.
- [x] endive must calculate and store the sha256 hash of every epub.
- [ ] the hash of retail epubs can be checked to detect unwanted modifications.
- [x] tags can be added to epubs.
- [x] the database must be easily exportable (JSON).
- [ ] the database can contain the date when the epub was read.
- [ ] when a metadata field is defined in both the epub metadata and the
database, endive must use the database version.
- [ ] the user can store in the database whether a physical copy of the book is
also available.

### Organization

- [ ] a given book can have retail and non-retail versions side by side.
- [ ] other duplicates are not allowed.
- [ ] the only allowed ebook format is epub.
- [ ] library organization can be refreshed by the user, upon modification of
the configuration files or of epub metadata.

### Modifiying epubs

- [ ] **TBC** non-retail epub covers can be updated.
- [ ] **TBC** non-retail epub metadata can be updated.

### Search

- [ ] epubs without retail versions can be listed.
- [ ] the library can be searched with the following creteria:
    author name, title, series, progress, retail, tags
- [ ] **TBC** all searches can be outputed as json.

### User interface

- [ ] all features can be accessed with a CLI.
- [ ] all features can be served over http.
- [ ] **TBC** all features can be accessed by a GUI.

## E-reader synchronization

- [ ] endive can synchronize selected epubs with a USB-mounted KOBO e-reader.
- [ ] **TBC** endive can synchronize KOBO collections, especially regarding
reading progress.

