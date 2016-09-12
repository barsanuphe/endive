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
	"path/filepath"
	"strconv"
	"syscall"

	b "github.com/barsanuphe/endive/book"
	e "github.com/barsanuphe/endive/endive"
	i "github.com/barsanuphe/endive/index"
	l "github.com/barsanuphe/endive/library"
	u "github.com/barsanuphe/endive/ui"

	"github.com/codegangsta/cli"
	"github.com/ttacon/chalk"
	"launchpad.net/go-xdg"
)

func generateCLI(lb *l.Library, ui e.UserInterface) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "E N D I V E"
	app.Usage = "Organize your epub collection."
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "options for configuration",
			Action: func(c *cli.Context) {
				ui.Display(lb.Config.String())
			},
		},
		{
			Name:     "import",
			Category: "importing",
			Aliases:  []string{"i"},
			Usage:    "options for importing epubs",
			Subcommands: []cli.Command{
				{
					Name:    "retail",
					Aliases: []string{"r"},
					Usage:   "import retail epubs",
					Action: func(c *cli.Context) {
						importEpubs(lb, c, ui, true)
					},
				},
				{
					Name:    "nonretail",
					Aliases: []string{"nr"},
					Usage:   "import non-retail epubs",
					Action: func(c *cli.Context) {
						importEpubs(lb, c, ui, false)
					},
				},
			},
		},
		{
			Name:    "export",
			Aliases: []string{"x"},
			Usage:   "export to E-Reader",
			Action: func(c *cli.Context) {
				exportFilter(lb, c, ui)
			},
			Subcommands: []cli.Command{
				{
					Name:    "all",
					Aliases: []string{"books"},
					Usage:   "export everything.",
					Action: func(c *cli.Context) {
						exportAll(lb, c, ui)
					},
				},
			},
		},
		{
			Name:    "check",
			Aliases: []string{"fsck"},
			Usage:   "check library",
			Action: func(c *cli.Context) {
				err := lb.Check()
				if err != nil {
					ui.Error("Check found errors! " + err.Error())
				} else {
					ui.Info("No errors found.")
				}

			},
		},
		{
			Name:    "metadata",
			Aliases: []string{"md"},
			Usage:   "edit book metadata",
			Subcommands: []cli.Command{
				{
					Name:    "refresh",
					Aliases: []string{"r"},
					Usage:   "reload metadata from epub and online sources (overwrites previous changes).",
					Action: func(c *cli.Context) {
						refreshMetadata(lb, c, ui)
					},
				},
				{
					Name:    "edit",
					Aliases: []string{"modify", "e"},
					Usage:   "edit metadata field using book ID: metadata edit ID field values",
					Action: func(c *cli.Context) {
						editMetadata(lb, c, ui)
					},
				},
			},
		},
		{
			Name:  "index",
			Usage: "manipulate index",
			Subcommands: []cli.Command{
				{
					Name:    "rebuild",
					Aliases: []string{"r"},
					Usage:   "rebuild index from scratch",
					Action: func(c *cli.Context) {
						if err := lb.RebuildIndex(); err != nil {
							ui.Error(err.Error())
						}
					},
				},
			},
		},
		{
			Name:    "refresh",
			Aliases: []string{"r"},
			Usage:   "refresh library",
			Action: func(c *cli.Context) {
				if c.NArg() != 0 {
					ui.Display("refresh subcommand does not require arguments.")
					return
				}
				ui.Display("Refreshing library...")
				renamed, err := lb.Refresh()
				if err != nil {
					panic(err)
				}
				ui.Display("Refresh done, renamed " + strconv.Itoa(renamed) + " epubs.")
			},
		},
		{
			Name:    "read",
			Aliases: []string{"rd"},
			Usage:   "mark as read: read ID [rating [review]]",
			Action: func(c *cli.Context) {
				markRead(lb, c, ui)
			},
		},
		{
			Name:     "info",
			Category: "information",
			Aliases:  []string{"information"},
			Usage:    "get info about a specific book",
			Action: func(c *cli.Context) {
				showInfo(lb, c, ui)
			},
		},
		{
			Name:     "search",
			Category: "searching",
			Aliases:  []string{"s", "find"},
			Usage:    "search the epub collection",
			Action: func(c *cli.Context) {
				search(lb, c, ui)
			},
		},
		{
			Name:     "list",
			Category: "searching",
			Aliases:  []string{"ls"},
			Usage:    "list epubs in the collection",
			Subcommands: []cli.Command{
				{
					Name:    "books",
					Aliases: []string{"b"},
					Usage:   "list all books: endive list books [sortBy CRITERIA]",
					Action: func(c *cli.Context) {
						books := make([]b.Book, len(lb.Books), len(lb.Books))
						copy(books, lb.Books)
						displayBooks(lb, c, ui, books)
					},
				},
				{
					Name:    "untagged",
					Aliases: []string{"u"},
					Usage:   "list untagged epubs.",
					Action: func(c *cli.Context) {
						displayBooks(lb, c, ui, lb.ListUntagged())
					},
				},
				{
					Name:    "incomplete",
					Aliases: []string{"i"},
					Usage:   "list books with incomplete epubs.",
					Action: func(c *cli.Context) {
						displayBooks(lb, c, ui, lb.ListIncomplete())
					},
				},
				{
					Name:    "tags",
					Aliases: []string{"t"},
					Usage:   "list tags",
					Action: func(c *cli.Context) {
						listTags(lb, c, ui)
					},
				},
				{
					Name:    "series",
					Aliases: []string{"s"},
					Usage:   "list series.",
					Action: func(c *cli.Context) {
						listSeries(lb, c, ui)
					},
				},
				{
					Name:    "authors",
					Aliases: []string{"a"},
					Usage:   "list authors.",
					Action: func(c *cli.Context) {
						authors := lb.ListAuthors()
						ui.Display(e.TabulateMap(authors, "Author", "# of Books"))
					},
				},
				{
					Name:    "publishers",
					Aliases: []string{"p"},
					Usage:   "list publishers.",
					Action: func(c *cli.Context) {
						publishers := lb.ListPublishers()
						ui.Display(e.TabulateMap(publishers, "Publisher", "# of Books"))
					},
				},
				{
					Name:    "nonretail",
					Aliases: []string{"nrt"},
					Usage:   "list books that only have non-retail versions.",
					Action: func(c *cli.Context) {
						displayBooks(lb, c, ui, lb.ListNonRetailOnly())
					},
				},
				{
					Name:    "retail",
					Aliases: []string{"rt"},
					Usage:   "list books that have retail versions.",
					Action: func(c *cli.Context) {
						displayBooks(lb, c, ui, lb.ListRetail())
					},
				},
			},
		},
	}
	return
}

func main() {
	var ui e.UserInterface
	ui = u.UI{}
	fmt.Println(chalk.Bold.TextStyle("\n# # # E N D I V E # # #\n"))

	err := ui.InitLogger(e.XdgLogPath)
	defer ui.CloseLog()

	// get library
	lb, err := OpenLibrary(ui)

	if err != nil {
		ui.Error("Error opening library.")
		ui.Error(err.Error())
		// if error other than usage elsewhere, remove lock.
		if err != e.ErrorCannotLockDB {
			e.RemoveLock()
		}
		return
	}
	defer lb.Close()

	// handle interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		ui.Error("Interrupt!")
		ui.Error("Stopping everything, saving what can be.")
		lb.Close()
		os.Exit(1)
	}()

	// generate CLI interface and run it
	app := generateCLI(lb, ui)
	app.Run(os.Args)
}

const xdgIndexPath string = e.Endive + "/" + e.Endive + ".index"

// getIndexPath gets the default index path
func getIndexPath() (path string) {
	path, err := xdg.Cache.Find(xdgIndexPath)
	if err != nil {
		if os.IsNotExist(err) {
			path = filepath.Join(xdg.Cache.Dirs()[0], xdgIndexPath)
		} else {
			panic(err)
		}
	}
	return
}

// OpenLibrary constucts a valid new Library
func OpenLibrary(ui e.UserInterface) (lib *l.Library, err error) {
	// config
	configPath, err := e.GetConfigPath()
	if err != nil {
		return
	}
	config := e.Config{Filename: configPath}
	// config load
	ui.Debugf("Loading Config %s.\n", config.Filename)
	err = config.Load()
	if err != nil {
		if err == e.WarningGoodReadsAPIKeyMissing {
			ui.Warning(err.Error())
		} else {
			ui.Error(err.Error())
		}
		return
	}
	// check config
	ui.Debug("Checking Config...")
	err = config.Check()
	switch err {
	case e.ErrorLibraryRootDoesNotExist:
		return
	case e.WarningNonRetailSourceDoesNotExist, e.WarningRetailSourceDoesNotExist:
		ui.Warning(err.Error())
	}
	// check lock
	err = e.SetLock()
	if err != nil {
		return
	}

	// known hashes
	hashesPath, err := e.GetKnownHashesPath()
	if err != nil {
		return
	}
	// load known hashes file
	hashes := e.KnownHashes{Filename: hashesPath}
	err = hashes.Load()
	if err != nil {
		return
	}

	// index
	index := &i.Index{}
	index.SetPath(getIndexPath())

	lib = &l.Library{Config: config, KnownHashes: hashes, Index: index, UI: ui}
	lib.DatabaseFile = config.DatabaseFile
	err = lib.Load()
	if err != nil {
		return
	}

	return lib, err
}
