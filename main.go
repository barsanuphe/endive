/*
Endive is a tool to keep your epub library in great shape.

It can rename and organize your library from the epub metadata, and can keep
track of retail and non-retail versions.

It is in a very early development: things can crash and files disappear.

*/
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/docopt/docopt-go"
	"github.com/ttacon/chalk"

	b "github.com/barsanuphe/endive/book"
	en "github.com/barsanuphe/endive/endive"
)

const (
	incorrectInput      = "Incorrect input, check endive -h for complete help."
	incorrectIDValue    = "Incorrect ID: %s"
	noBookFound         = "Book with ID %d cannot be found"
	numberOfBooksHeader = "# of Books"
	incorrectFlag       = "--first and --last only support integer values"
	invalidLimit        = -1
	endiveVersion       = "Endive -- CLI Epub collection manager -- v1.0."
	endiveUsage         = `
Endive.
This is an epub collection manager.

The main commands are:
	config		Display current configuration
	collection	Do some maintenance on the collection
	import, i	Import epubs to the collection
	export, x	Export epubs to ereader
	info		Display information
	edit		Edit metadata
	progress, p	Set book reading progress
	list, ls	List books
	search, s	Search for specific books

Searching / Exporting:
	A list of strings can be given as input to search for books.
	It is also possible to restrict a value to a specific field: field:value.
	Valid fields are:
		author, title, year, language, series, tag, publisher, category,
		type, genre, description, exported, progress, review.
	Examples:
		'author:XX title:YY' will give results satifsying any of the two conditions.
		'author:XX +title:YY' will give results satifsying both conditions.
		'author:XX -title:YY' will give results satifsying the first condition excluding the second.

Usage:
	endive config
	endive collection (check|refresh|rebuild-index|check-index)
	endive (import|i) ((retail|r)|(nonretail|nr)) [--list] [<epub>...]
	endive (export|x) (all|<search-criteria>...)
	endive info [tags|series|authors|publishers] [<ID>]
	endive edit <ID> [<field> [--refresh|<value>]]
	endive (progress|p) <ID> (unread|read|reading|shortlisted) [<rating> [<review>]]
	endive (list|ls) [--incomplete|--nonretail|--retail] [--first=N] [--last=N] [--sort=SORT]
	endive (search|s) <search-criteria>... [--first=N] [--last=N] [--sort=SORT]
	endive -h | --help
	endive --version

Options:
	-h --help            Show this screen.
	--version            Show version.
	--list               List importable epubs only.
	-f N --first=N       Filter only the n first books
	-l N --last=N        Filter only the n last books
	-s SORT --sort=SORT  Sort results [default: id]
	--untagged           Filter books with no tags
	--incomplete         Filter books with incomplete metadata
	--retail             Only show retail books
	--nonretail          Only show non-retail books
	--refresh            Refresh value from GR.`
)

func main() {
	fmt.Println(chalk.Bold.TextStyle("\n# # # E N D I V E # # #\n"))

	// create main Endive struct
	e, err := NewEndive()
	if err != nil {
		e.UI.Error("Could not create Endive: " + err.Error())
		// if error other than usage elsewhere, remove lock.
		if err != en.ErrorCannotLockDB {
			en.RemoveLock()
		}
		os.Exit(-1)
	}
	defer e.UI.CloseLog()
	defer en.RemoveLock()
	defer e.Library.Close()

	// handle interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		e.UI.Error("Interrupt!")
		e.UI.Error("Stopping everything, saving what can be.")
		e.Library.Close()
		en.RemoveLock()
		e.UI.CloseLog()
		os.Exit(1)
	}()

	// parse arguments and options
	args, err := docopt.Parse(endiveUsage, nil, true, endiveVersion, false, false)
	if err != nil {
		fmt.Println(incorrectInput)
		return
	}
	if len(args) == 0 {
		// builtin command, nothing to do.
		return
	}

	// checking if ID was given, getting relevant *Book
	var book *b.Book
	if args["<ID>"] != nil {
		id, err := strconv.Atoi(args["<ID>"].(string))
		if err != nil {
			e.UI.Errorf(incorrectIDValue, args["<ID>"].(string))
			return
		}
		bk, err := e.Library.Collection.FindByID(id)
		if err != nil {
			e.UI.Errorf(noBookFound, id)
			return
		}
		book = bk.(*b.Book)
	}

	// checking other common flags
	firstNBooks := invalidLimit
	if args["--first"] != nil {
		firstNBooks, err = strconv.Atoi(args["--first"].(string))
		if err != nil {
			e.UI.Error(incorrectFlag)
			return
		}
	}
	lastNBooks := invalidLimit
	if args["--last"] != nil {
		lastNBooks, err = strconv.Atoi(args["--last"].(string))
		if err != nil {
			e.UI.Error(incorrectFlag)
			return
		}
	}
	sortBy := strings.ToLower(args["--sort"].(string))

	// now dealing with commands
	if args["config"].(bool) {
		e.UI.Display(e.Config.String())
	}

	if args["collection"].(bool) {
		if args["check"].(bool) {
			if err := e.Library.Check(); err != nil {
				e.UI.Error("Check found modified files since import! " + err.Error())
			} else {
				e.UI.Info("All epubs checked successfully.")
			}
		} else if args["refresh"].(bool) {
			e.UI.Display("Refreshing library...")
			if renamed, err := e.Refresh(); err == nil {
				e.UI.Display("Refresh done, renamed " + strconv.Itoa(renamed) + " epubs.")
			} else {
				e.UI.Error("Could not refresh collection.")
			}
		} else if args["rebuild-index"].(bool) {
			if err := e.Library.RebuildIndex(); err != nil {
				e.UI.Error(err.Error())
			}
		} else if args["check-index"].(bool) {
			if err := e.Library.CheckIndex(); err != nil {
				e.UI.Error(err.Error())
			}
		}
	}

	if args["import"].(bool) || args["i"].(bool) {
		epubs := args["<epub>"].([]string)
		retail := args["retail"].(bool) || args["r"].(bool)
		if args["--list"].(bool) {
			listImportableEpubs(e, retail)
		} else {
			importEpubs(e, epubs, retail)
		}
	}

	if args["export"].(bool) || args["x"].(bool) {
		if args["all"].(bool) {
			exportAll(e)
		} else {
			exportFilter(e, args["<search-criteria>"].([]string))
		}
	}

	if args["info"].(bool) {
		if args["tags"].(bool) {
			e.UI.Display(en.TabulateMap(e.Library.Collection.Tags(), "Tag", numberOfBooksHeader))
		} else if args["series"].(bool) {
			e.UI.Display(en.TabulateMap(e.Library.Collection.Series(), "Series", numberOfBooksHeader))
		} else if args["authors"].(bool) {
			e.UI.Display(en.TabulateMap(e.Library.Collection.Authors(), "Author", numberOfBooksHeader))
		} else if args["publishers"].(bool) {
			e.UI.Display(en.TabulateMap(e.Library.Collection.Publishers(), "Publisher", numberOfBooksHeader))
		} else {
			showInfo(e, book)
		}
	}

	if args["edit"].(bool) {
		editArgs := []string{}
		if args["<field>"] != nil {
			editArgs = append(editArgs, args["<field>"].(string))
			if args["<value>"] != nil {
				editArgs = append(editArgs, args["<value>"].(string))
			}
		}
		if args["--refresh"].(bool) {
			refreshMetadata(e, book, editArgs...)
		} else {
			editMetadata(e, book, editArgs...)
		}
	}

	if args["progress"].(bool) || args["p"].(bool) {
		// TODO validate progress with Book package
		var progress, rating, review string
		if args["unread"].(bool) {
			progress = "unread"
		} else if args["read"].(bool) {
			progress = "read"
			if args["<rating>"] != nil {
				rating = args["<rating>"].(string)
			}
			if args["<review>"] != nil {
				review = args["<review>"].(string)
			}
		} else if args["reading"].(bool) {
			progress = "reading"
		} else if args["shortlisted"].(bool) {
			progress = "shortlisted"
		}
		setProgress(e, book, progress, rating, review)
	}

	if args["search"].(bool) || args["s"].(bool) {
		if criteria := args["<search-criteria>"].([]string); len(criteria) != 0 {
			search(e, criteria, firstNBooks, lastNBooks, sortBy)
		}
	}

	if args["list"].(bool) || args["ls"].(bool) {
		collection := e.Library.Collection
		if args["--incomplete"].(bool) {
			collection = collection.Incomplete()
		} else if args["--retail"].(bool) {
			collection = collection.Retail()
		} else if args["--nonretail"].(bool) {
			collection = collection.NonRetailOnly()
		}
		displayBooks(e.UI, collection, firstNBooks, lastNBooks, sortBy)
	}
}
