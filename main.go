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

	"github.com/ttacon/chalk"

	en "github.com/barsanuphe/endive/endive"
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

	cli := CLI{}
	if err := cli.parseArgs(e, os.Args[1:]); err != nil {
		fmt.Println(incorrectInput)
		e.UI.Error(err.Error())
		return
	}
	if cli.builtInCommand {
		// nothing to do
		return
	}
	// now dealing with commands
	if cli.showConfig {
		e.UI.Display(e.Config.String())
	} else if cli.checkCollection {
		if err := e.Library.Check(); err != nil {
			e.UI.Error("Check found modified files since import! " + err.Error())
		} else {
			e.UI.Info("All epubs checked successfully.")
		}
	} else if cli.refreshCollection {
		e.UI.Display("Refreshing library...")
		if renamed, err := e.Refresh(); err == nil {
			e.UI.Display("Refresh done, renamed " + strconv.Itoa(renamed) + " epubs.")
		} else {
			e.UI.Error("Could not refresh collection.")
		}
	} else if cli.rebuildIndex {
		if err := e.Library.RebuildIndex(); err != nil {
			e.UI.Error(err.Error())
		}
	} else if cli.checkIndex {
		if err := e.Library.CheckIndex(); err != nil {
			e.UI.Error(err.Error())
		}
	} else if cli.importEpubs {
		if cli.listImport {
			listImportableEpubs(e, cli.importRetail)
		} else {
			importEpubs(e, cli.epubs, cli.importRetail)
		}
	} else if cli.export {
		if len(cli.searchTerms) == 0 {
			exportCollection(e, cli.collection)
		} else {
			exportFilter(e, cli.searchTerms)
		}
	} else if cli.info != "" {
		switch cli.info {
		case infoGeneral:
			showInfo(e, nil)
		case infoBook:
			showInfo(e, cli.books[0])
		default:
			e.UI.Display(en.TabulateMap(cli.collectionMap, cli.info, numberOfBooksHeader))
		}
	} else if cli.review {
		reviewBook(e, cli.books[0], cli.rating, cli.reviewText)
	} else if cli.edit {
		if cli.field != "" {
			editMetadata(e, cli.books, cli.field)
		} else {
			editMetadata(e, cli.books)
		}
	} else if cli.reset {
		if cli.field != "" {
			refreshMetadata(e, cli.books, cli.field)
		} else {
			refreshMetadata(e, cli.books)
		}
	} else if cli.set {
		if cli.field != "" {
			editMetadata(e, cli.books, cli.field, cli.value)
		} else {
			setProgress(e, cli.books, cli.progress)
		}

	} else if cli.search {
		search(e, cli.searchTerms, cli.firstN, cli.lastN, cli.sortBy)
	} else if cli.list {
		displayBooks(e.UI, cli.collection, cli.firstN, cli.lastN, cli.sortBy)
	}
}
