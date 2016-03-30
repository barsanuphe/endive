# What is endive?

endive manages epub files.

# Features

**TBC** tagged requirements can be done with calibre/sigil, may not be
implemented.

## Configuration

- [ ] the library layout can be defined by a configuration file.
- [ ] the configuration file allows defining author aliases.
- [ ] a directory for retail epubs source and another for non-retail epubs can be
defined in the configuration file.
- [ ] configuration files follow XDG recommendations.

## Importing epubs

- [ ] endive can scrape special directories for epubs to import.
- [ ] already imported epubs must be ignored during import.
- [ ] epubs must be imported into the library, which contains the epubs and the database.
- [ ] retail epubs are to be given read-only permissions.

## Library

### Epub metadata

- [X] endive must read epub metadata, including: author, title, year of publication, language.

### Database

- [ ] endive can keep track of progression: unread, reading, read(book), read(ebook).
- [ ] endive can associate one or several series with an epub.
- [ ] epubs can be flagged as needing replacement.
- [ ] retail epubs are identified as such.
- [X] tags can be added to epubs.
- [X] the database must be easily exportable (JSON).

### Organization

- [ ] a given book can have a retail and non-retail version side by side.
- [ ] other duplicates are not allowed.
- [ ] library organization can be refreshed by the user, upon modification of the
configuration files or of epub metadata.

### Modifiying epubs

- [ ] **TBC** non-retail epub covers can be updated.
- [ ] **TBC** non-retail epub metadata can be updated.

### Search

- [ ] epubs without retail versions can be listed.
- [ ] the library can be searched with the following creteria:
    author name, title, series, progress, retail, tags
- [ ] all searches can be outputed as json.

### User interface

- [ ] all features can be accessed with a CLI.
- [ ] all features can be served over http.
- [ ] **TBC** all features can be accessed by a GUI.

## E-reader synchronization

- [ ] endive can synchronize selected epubs with a USB-mounted KOBO e-reader.
- [ ] **TBC** endive can synchronize KOBO collections, especially regarding reading
progress.
