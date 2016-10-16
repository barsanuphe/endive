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
	"github.com/urfave/cli"

	en "github.com/barsanuphe/endive/endive"
)

func generateCLI(e *Endive) (app *cli.App) {
	var firstNBooks, lastNBooks int
	var sortBy string

	limitFlags := []cli.Flag{
		cli.IntFlag{
			Name:        "first,f",
			Usage:       "Only display the first `N` books",
			Value:       -1,
			Destination: &firstNBooks,
		},
		cli.IntFlag{
			Name:        "last,l",
			Usage:       "Only display the last `N` books",
			Value:       -1,
			Destination: &lastNBooks,
		},
		cli.StringFlag{
			Name:        "sort,s",
			Usage:       "sort results by `CRITERIA` among: id, author, title, year.",
			Destination: &sortBy,
		},
	}

	app = cli.NewApp()
	app.Name = "endive"
	app.Usage = "Organize your epub collection."
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		{
			Name:     "config",
			Category: "configuration",
			Aliases:  []string{"c"},
			Usage:    "options for configuration",
			Action: func(c *cli.Context) {
				e.UI.Display(e.Config.String())
			},
		},
		{
			Name:     "import",
			Category: "library",
			Aliases:  []string{"i"},
			Usage:    "options for importing epubs",
			Subcommands: []cli.Command{
				{
					Name:    "retail",
					Aliases: []string{"r"},
					Usage:   "import retail epubs",
					Action: func(c *cli.Context) {
						importEpubs(e, c, true)
					},
				},
				{
					Name:    "list-retail",
					Aliases: []string{"lsr"},
					Usage:   "list importable retail epubs",
					Action: func(c *cli.Context) {
						listImportableEpubs(e, c, true)
					},
				},
				{
					Name:    "nonretail",
					Aliases: []string{"nr"},
					Usage:   "import non-retail epubs",
					Action: func(c *cli.Context) {
						importEpubs(e, c, false)
					},
				},
				{
					Name:    "list-nonretail",
					Aliases: []string{"lsnr"},
					Usage:   "list importable nonretail epubs",
					Action: func(c *cli.Context) {
						listImportableEpubs(e, c, false)
					},
				},
			},
		},
		{
			Name:     "export",
			Category: "library",
			Aliases:  []string{"x"},
			Usage:    "export to E-Reader",
			Action: func(c *cli.Context) {
				exportFilter(c, e)
			},
			Subcommands: []cli.Command{
				{
					Name:    "all",
					Aliases: []string{"books"},
					Usage:   "export everything.",
					Action: func(c *cli.Context) {
						exportAll(e)
					},
				},
			},
		},
		{
			Name:     "check",
			Category: "library",
			Aliases:  []string{"fsck"},
			Usage:    "check library",
			Action: func(c *cli.Context) {
				err := e.Library.Check()
				if err != nil {
					e.UI.Error("Check found errors! " + err.Error())
				} else {
					e.UI.Info("No errors found.")
				}

			},
		},
		{
			Name:     "refresh",
			Category: "library",
			Aliases:  []string{"r"},
			Usage:    "refresh library",
			Action: func(c *cli.Context) {
				if c.NArg() != 0 {
					e.UI.Display("refresh subcommand does not require arguments.")
					return
				}
				e.UI.Display("Refreshing library...")
				renamed, err := e.Refresh()
				if err != nil {
					panic(err)
				}
				e.UI.Display("Refresh done, renamed " + strconv.Itoa(renamed) + " epubs.")
			},
		},
		{
			Name:     "index",
			Category: "library",
			Usage:    "manipulate index",
			Subcommands: []cli.Command{
				{
					Name:    "rebuild",
					Aliases: []string{"r"},
					Usage:   "rebuild index from scratch",
					Action: func(c *cli.Context) {
						if err := e.Library.RebuildIndex(); err != nil {
							e.UI.Error(err.Error())
						}
					},
				},
				{
					Name:    "check",
					Aliases: []string{"c", "fsck"},
					Usage:   "check all books are in the index, add them otherwise",
					Action: func(c *cli.Context) {
						if err := e.Library.CheckIndex(); err != nil {
							e.UI.Error(err.Error())
						}
					},
				},
			},
		},
		{
			Name:     "info",
			Category: "book",
			Aliases:  []string{"information"},
			Usage:    "get info about a specific book",
			Action: func(c *cli.Context) {
				showInfo(c, e)
			},
		},
		{
			Name:     "metadata",
			Category: "book",
			Aliases:  []string{"md"},
			Usage:    "edit book metadata",
			Subcommands: []cli.Command{
				{
					Name:    "refresh",
					Aliases: []string{"r"},
					Usage:   "reload metadata from epub and online sources (overwrites previous changes).",
					Action: func(c *cli.Context) {
						refreshMetadata(c, e)
					},
				},
				{
					Name:      "edit",
					Aliases:   []string{"modify", "e"},
					Usage:     "edit metadata field using book ID: metadata edit ID field values",
					ArgsUsage: "ID [field [value]]",
					Action: func(c *cli.Context) {
						editMetadata(c, e)
					},
				},
			},
		},
		{
			Name:      "progress",
			Category:  "book",
			Aliases:   []string{"p"},
			Usage:     "modify reading progress for a given book",
			ArgsUsage: "ID [unread/shortlisted/reading/read [rating [review]]]",
			Action: func(c *cli.Context) {
				setProgress(c, e)
			},
		},
		{
			Name:     "list",
			Category: "search",
			Aliases:  []string{"ls"},
			Usage:    "list epubs in the collection with specific filters",
			Subcommands: []cli.Command{
				{
					Name:    "books",
					Aliases: []string{"b"},
					Usage:   "list all books: endive list books [sortBy CRITERIA]",
					Flags:   limitFlags,
					Action: func(c *cli.Context) {
						displayBooks(e.UI, e.Library.Collection, firstNBooks, lastNBooks, sortBy)
					},
				},
				{
					Name:    "untagged",
					Aliases: []string{"u"},
					Usage:   "list untagged epubs.",
					Flags:   limitFlags,
					Action: func(c *cli.Context) {
						displayBooks(e.UI, e.Library.Collection.Untagged(), firstNBooks, lastNBooks, sortBy)
					},
				},
				{
					Name:    "incomplete",
					Aliases: []string{"i"},
					Usage:   "list books with incomplete epubs.",
					Flags:   limitFlags,
					Action: func(c *cli.Context) {
						displayBooks(e.UI, e.Library.Collection.Incomplete(), firstNBooks, lastNBooks, sortBy)
					},
				},
				{
					Name:    "tags",
					Aliases: []string{"t"},
					Usage:   "list tags",
					Action: func(c *cli.Context) {
						listTags(c, e)
					},
				},
				{
					Name:    "series",
					Aliases: []string{"s"},
					Usage:   "list series.",
					Action: func(c *cli.Context) {
						listSeries(c, e)
					},
				},
				{
					Name:    "authors",
					Aliases: []string{"a"},
					Usage:   "list authors.",
					Action: func(c *cli.Context) {
						authors := e.Library.Collection.Authors()
						e.UI.Display(en.TabulateMap(authors, "Author", "# of Books"))
					},
				},
				{
					Name:    "publishers",
					Aliases: []string{"p"},
					Usage:   "list publishers.",
					Action: func(c *cli.Context) {
						publishers := e.Library.Collection.Publishers()
						e.UI.Display(en.TabulateMap(publishers, "Publisher", "# of Books"))
					},
				},
				{
					Name:    "nonretail",
					Aliases: []string{"nrt"},
					Usage:   "list books that only have non-retail versions.",
					Flags:   limitFlags,
					Action: func(c *cli.Context) {
						displayBooks(e.UI, e.Library.Collection.NonRetailOnly(), firstNBooks, lastNBooks, sortBy)
					},
				},
				{
					Name:    "retail",
					Aliases: []string{"rt"},
					Usage:   "list books that have retail versions.",
					Flags:   limitFlags,
					Action: func(c *cli.Context) {
						displayBooks(e.UI, e.Library.Collection.Retail(), firstNBooks, lastNBooks, sortBy)
					},
				},
			},
		},
		{
			Name:        "search",
			Category:    "search",
			Aliases:     []string{"s", "find"},
			Usage:       "search the library for specific books",
			Description: "A list of strings can be given as input to search for books. \n   It is also possible to restrict a value to a specific field: `field:value`.",
			ArgsUsage:   "arg1 [args2] [field:value] [+field2:value]",
			Flags:       limitFlags,
			Action: func(c *cli.Context) {
				search(c, e, firstNBooks, lastNBooks, sortBy)
			},
		},
	}
	return
}

func main() {
	fmt.Println(chalk.Bold.TextStyle("\n# # # E N D I V E # # #\n"))

	// create main Endive struct
	endive, err := NewEndive()
	if err != nil {
		endive.UI.Error("Could not create Endive: " + err.Error())
		// if error other than usage elsewhere, remove lock.
		if err != en.ErrorCannotLockDB {
			en.RemoveLock()
		}
		os.Exit(-1)
	}
	defer endive.UI.CloseLog()
	defer en.RemoveLock()
	defer endive.Library.Close()

	// handle interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		endive.UI.Error("Interrupt!")
		endive.UI.Error("Stopping everything, saving what can be.")
		endive.Library.Close()
		en.RemoveLock()
		endive.UI.CloseLog()
		os.Exit(1)
	}()

	// generate CLI interface and run it
	app := generateCLI(endive)
	app.Run(os.Args)
}
