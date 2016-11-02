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
	"syscall"

	"github.com/docopt/docopt-go"
	"github.com/ttacon/chalk"

	en "github.com/barsanuphe/endive/endive"
)

const (
	metadataRefreshDescription = "Refreshes metadata by reading again from Epub and retrieving metadata from Goodreads.\n" +
		"   Note: the values currently in the database will be lost.\n\n" +
		"   If no further argument is given all the metadata fields will be refreshed.\n" +
		"   If a valid field name is given, only it will be refreshed."
	configDescription = "Displays the contents of the configuration file in a table."
	searchDescription = "A list of strings can be given as input to search for books.\n" +
		"   It is also possible to restrict a value to a specific field: field:value.\n" +
		"   Valid fields are: author, title, year, language, series, tag, publisher, category, type, genre, description, exported, progress, review.\n\n" +
		"   Examples: \n\n" +
		"       'author:XX title:YY' will give results satifsying any of the two conditions.\n" +
		"       'author:XX +title:YY' will give results satifsying both conditions.\n" +
		"       'author:XX -title:YY' will give results satifsying the first condition excluding the second.\n"

	IncorrectIDValue = "Incorrect ID: %s"
	IncorrectFlag = "--first and --last only support integer values"
	InvalidID        = -1
	EndiveVersion    = "Endive -- CLI Epub collection manager -- v1.0."
	EndiveUsage      = `
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

Usage:
	endive config
	endive collection (check|refresh|rebuild-index|check-index)
	endive (import|i) ((retail|r)|(nonretail|nr)) [--list] [<epub>...]
	endive (export|x) (all|<search-criteria>...)
	endive info [tags|series|authors|publishers] [<ID>]
	endive edit <ID> [<field> [--refresh|<value>]]
	endive (progress|p) <ID> (unread|read|reading|shortlisted) [<rating> [<review>]]
	endive (list|ls) [--untagged|--incomplete|--nonretail|--retail] [--first=N] [--last=N] [--sort=SORT]
	endive (search|s) <search-criteria>... [--first=N] [--last=N] [--sort=SORT]
	endive -h | --help
	endive --version

Options:
	-h --help            Show this screen.
	--version            Show version.
	--list               List importable epubs only.
	-f N --first=N       Filter only the n first books
	-l N --last=N        Filter only the n last books
	-s SORT --sort=SORT  Sort results [default: ID]
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

	args, err := docopt.Parse(EndiveUsage, nil, true, EndiveVersion, false, false)
	if err != nil {
		fmt.Println("ERR" + err.Error())
		return
	}
	if len(args) == 0 {
		// builtin command, nothing to do.
		return
	}

	// checking if ID was given and validating
	id := InvalidID
	if args["<ID>"] != nil {
		id, err = strconv.Atoi(args["<ID>"].(string))
		if err != nil {
			e.UI.Errorf(IncorrectIDValue, args["<ID>"].(string))
			return
		}
	}
	// checking other common flags
	firstNBooks := -1
	lastNBooks := -1
	if args["--first"] != nil {
		firstNBooks, err = strconv.Atoi(args["--first"].(string))
		if err != nil {
			e.UI.Error(IncorrectFlag)
			return
		}
	}
	if args["--last"] != nil {
		lastNBooks, err = strconv.Atoi(args["--last"].(string))
		if err != nil {
			e.UI.Error(IncorrectFlag)
			return
		}
	}
	sortBy := args["--sort"].(string)

	// now dealing with commands
	if args["config"].(bool) {
		e.UI.Display(e.Config.String())
	}

	if args["collection"].(bool) {
		if args["check"].(bool) {
			if err := e.Library.Check(); err != nil {
				e.UI.Error("Check found errors! " + err.Error())
			} else {
				e.UI.Info("No errors found.")
			}
		} else if args["refresh"].(bool) {
			e.UI.Display("Refreshing library...")
			if renamed, err := e.Refresh(); err == nil {
				e.UI.Display("Refresh done, renamed " + strconv.Itoa(renamed) + " epubs.")
			} else {
				e.UI.Error("Could not refresh collection")
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
		listOnly := args["--list"].(bool)
		epubs := args["<epub>"].([]string)
		retail := args["retail"].(bool) || args["r"].(bool)
		if listOnly {
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
			listTags(e, id)
		} else if args["series"].(bool) {
			listSeries(e, id)
		} else if args["authors"].(bool) {
			authors := e.Library.Collection.Authors()
			e.UI.Display(en.TabulateMap(authors, "Author", "# of Books"))
		} else if args["publishers"].(bool) {
			publishers := e.Library.Collection.Publishers()
			e.UI.Display(en.TabulateMap(publishers, "Publisher", "# of Books"))
		} else {
			showInfo(e, id)
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
			refreshMetadata(e, id, editArgs...)
		} else {
			editMetadata(e, id, editArgs...)
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
		setProgress(e, id, progress, rating, review)
	}

	if args["search"].(bool) || args["s"].(bool) {
		if criteria := args["<search-criteria>"].([]string); len(criteria) != 0 {
			search(e, criteria, firstNBooks, lastNBooks, sortBy)
		}
	}

	if args["list"].(bool) || args["ls"].(bool) {
		if args["--untagged"].(bool) {
			displayBooks(e.UI, e.Library.Collection.Untagged(), firstNBooks, lastNBooks, sortBy)
		} else if args["--incomplete"].(bool) {
			displayBooks(e.UI, e.Library.Collection.Incomplete(), firstNBooks, lastNBooks, sortBy)
		} else if args["--retail"].(bool) {
			displayBooks(e.UI, e.Library.Collection.Retail(), firstNBooks, lastNBooks, sortBy)
		} else if args["--nonretail"].(bool) {
			displayBooks(e.UI, e.Library.Collection.NonRetailOnly(), firstNBooks, lastNBooks, sortBy)
		} else {
			displayBooks(e.UI, e.Library.Collection, firstNBooks, lastNBooks, sortBy)
		}
	}
}
