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
	"strconv"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
	l "github.com/barsanuphe/endive/library"

	"github.com/codegangsta/cli"
	"github.com/ttacon/chalk"
)

func generateCLI(lb *l.Library) (app *cli.App) {
	app = cli.NewApp()
	app.Name = "E N D I V E"
	app.Usage = "Organize your epub collection."
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "options for configuration",
			Subcommands: []cli.Command{
				{
					Name:    "show",
					Aliases: []string{"ls"},
					Usage:   "show configuration",
					Action: func(c *cli.Context) {
						fmt.Println(lb.Config.String())
					},
				},
				{
					Name:  "aliases",
					Usage: "show aliases defined in configuration",
					Action: func(c *cli.Context) {
						aliases := lb.Config.ListAuthorAliases()
						fmt.Println(aliases)
					},
				},
			},
		},
		{
			Name:    "serve",
			Aliases: []string{"s"},
			Usage:   "serve over http",
			Action: func(c *cli.Context) {
				port, err := checkPort(c)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Printf("Serving on port %d...\n", port)
				// TODO
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
						importEpubs(lb, c, true)
					},
				},
				{
					Name:    "nonretail",
					Aliases: []string{"n"},
					Usage:   "import non-retail epubs",
					Action: func(c *cli.Context) {
						importEpubs(lb, c, false)
					},
				},
			},
		},
		{
			Name:    "export",
			Aliases: []string{"x"},
			Usage:   "export to E-Reader",
			Action: func(c *cli.Context) {
				exportFilter(lb, c)
			},
			Subcommands: []cli.Command{
				{
					Name:    "all",
					Aliases: []string{"books"},
					Usage:   "export everything.",
					Action: func(c *cli.Context) {
						exportAll(lb, c)
					},
				},
			},
		},
		{
			Name:    "check",
			Aliases: []string{"fsck"},
			Usage:   "check library",
			Action: func(c *cli.Context) {
				fmt.Printf("Checking...")
				err := lb.Check()
				if err != nil {
					fmt.Printf(" KO!\n")
					panic(err)
				}
				fmt.Printf(" OK\n")
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
						refreshMetadata(lb, c)
					},
				},
				{
					Name:    "edit",
					Aliases: []string{"modify"},
					Usage:   "edit metadata field using book ID: metadata edit ID field values",
					Action: func(c *cli.Context) {
						editMetadata(lb, c)
					},
				},
			},
		},
		{
			Name:    "refresh",
			Aliases: []string{"r"},
			Usage:   "refresh library",
			Action: func(c *cli.Context) {
				fmt.Println("Refreshing library...")
				renamed, err := lb.Refresh()
				if err != nil {
					panic(err)
				}
				fmt.Println("Refresh done, renamed " + strconv.Itoa(renamed) + " epubs.")
			},
		},
		{
			Name:    "read",
			Aliases: []string{"rd"},
			Usage:   "mark as read: read ID [rating [review]]",
			Action: func(c *cli.Context) {
				markRead(lb, c)
			},
		},
		{
			Name:     "info",
			Category: "information",
			Aliases:  []string{"information"},
			Usage:    "get info about a specific book",
			Action: func(c *cli.Context) {
				showInfo(lb, c)
			},
		},
		{
			Name:     "search",
			Category: "searching",
			Aliases:  []string{"c"},
			Usage:    "search the epub collection",
			Action: func(c *cli.Context) {
				search(lb, c)
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
					Usage:   "list all books.",
					Action: func(c *cli.Context) {
						fmt.Println(lb.TabulateList(lb.Books))
					},
				},
				{
					Name:    "untagged",
					Aliases: []string{"u"},
					Usage:   "list untagged epubs.",
					Action: func(c *cli.Context) {
						list := lb.ListUntagged()
						fmt.Println(lb.TabulateList(list))
					},
				},
				{
					Name:    "tags",
					Aliases: []string{"t"},
					Usage:   "list tags",
					Action: func(c *cli.Context) {
						listTags(lb, c)
					},
				},
				{
					Name:    "series",
					Aliases: []string{"s"},
					Usage:   "list series.",
					Action: func(c *cli.Context) {
						listSeries(lb, c)
					},
				},
				{
					Name:    "authors",
					Aliases: []string{"a"},
					Usage:   "list authors.",
					Action: func(c *cli.Context) {
						authors := lb.ListAuthors()
						fmt.Println(h.TabulateMap(authors, "Author", "# of Books"))
					},
				},
				{
					Name:    "nonretail",
					Aliases: []string{"nrt"},
					Usage:   "list books that only have non-retail versions.",
					Action: func(c *cli.Context) {
						list := lb.ListNonRetailOnly()
						fmt.Println(lb.TabulateList(list))
					},
				},
				{
					Name:    "retail",
					Aliases: []string{"rt"},
					Usage:   "list books that only have retail versions.",
					Action: func(c *cli.Context) {
						list := lb.ListRetail()
						fmt.Println(lb.TabulateList(list))
					},
				},
			},
		},
	}
	return
}

func main() {
	fmt.Println(chalk.Bold.TextStyle("\n# # # E N D I V E # # #\n"))

	err := h.GetEndiveLogger(cfg.XdgLogPath)
	defer h.CloseEndiveLogFile()

	// get library
	lb, err := l.OpenLibrary()
	if err != nil {
		h.Error("Error opening library.")
		h.Error(err.Error())
		return
	}
	defer lb.Close()

	// generate CLI interface and run it
	app := generateCLI(lb)
	app.Run(os.Args)
}
