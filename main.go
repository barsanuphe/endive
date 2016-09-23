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

	"github.com/codegangsta/cli"
	"github.com/ttacon/chalk"

	en "github.com/barsanuphe/endive/endive"
)

func generateCLI(e *Endive) (app *cli.App) {
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
				e.UI.Display(e.Config.String())
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
						importEpubs(e, c, true)
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
			},
		},
		{
			Name:    "export",
			Aliases: []string{"x"},
			Usage:   "export to E-Reader",
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
			Name:    "check",
			Aliases: []string{"fsck"},
			Usage:   "check library",
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
			Name:    "metadata",
			Aliases: []string{"md"},
			Usage:   "edit book metadata",
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
					Name:    "edit",
					Aliases: []string{"modify", "e"},
					Usage:   "edit metadata field using book ID: metadata edit ID field values",
					Action: func(c *cli.Context) {
						editMetadata(c, e)
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
			Name:    "refresh",
			Aliases: []string{"r"},
			Usage:   "refresh library",
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
			Name:    "read",
			Aliases: []string{"rd"},
			Usage:   "mark as read: read ID [rating [review]]",
			Action: func(c *cli.Context) {
				markRead(c, e)
			},
		},
		{
			Name:     "info",
			Category: "information",
			Aliases:  []string{"information"},
			Usage:    "get info about a specific book",
			Action: func(c *cli.Context) {
				showInfo(c, e)
			},
		},
		{
			Name:     "search",
			Category: "searching",
			Aliases:  []string{"s", "find"},
			Usage:    "search the epub collection",
			Action: func(c *cli.Context) {
				search(c, e)
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
						displayBooks(c, e.UI, e.Library.Collection)
					},
				},
				{
					Name:    "untagged",
					Aliases: []string{"u"},
					Usage:   "list untagged epubs.",
					Action: func(c *cli.Context) {
						displayBooks(c, e.UI, e.Library.Collection.Untagged())
					},
				},
				{
					Name:    "incomplete",
					Aliases: []string{"i"},
					Usage:   "list books with incomplete epubs.",
					Action: func(c *cli.Context) {
						displayBooks(c, e.UI, e.Library.Collection.Incomplete())
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
					Action: func(c *cli.Context) {
						displayBooks(c, e.UI, e.Library.Collection.NonRetailOnly())
					},
				},
				{
					Name:    "retail",
					Aliases: []string{"rt"},
					Usage:   "list books that have retail versions.",
					Action: func(c *cli.Context) {
						displayBooks(c, e.UI, e.Library.Collection.Retail())
					},
				},
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
