/*
Endive is a tool to keep your epub library in great shape.

It can rename and organize your library from the epub metadata, and can keep
track of retail and non-retail versions.

It is in a very early development, with basically nothing actually working,
and chances of files disappearing if invoked.

*/
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"errors"

	b "github.com/barsanuphe/endive/book"
	l "github.com/barsanuphe/endive/library"
	"github.com/codegangsta/cli"
	"github.com/ttacon/chalk"
)

func checkArgsWithID(l l.Library, args []string) (book *b.Book, other []string, err error) {
	if len(args) < 1 {
		err = errors.New("Not enough arguments")
		return
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return
	}
	// get book from ID
	book, err = l.FindByID(id)
	if err != nil {
		return
	}
	other = args[1:]
	return
}

func main() {
	fmt.Println(chalk.Bold.TextStyle("\n# # # E N D I V E # # #\n"))

	// get library
	lb, err := l.OpenLibrary()
	if err != nil {
		fmt.Println("Error loading configuration. Check it.")
		fmt.Println(err.Error())
		return
	}
	defer lb.Save()

	app := cli.NewApp()
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
						// print config
						fmt.Println(lb.ConfigurationFile.String())
					},
				},
				{
					Name:  "aliases",
					Usage: "show aliases defined in configuration",
					Action: func(c *cli.Context) {
						// print aliases
						aliases := lb.ConfigurationFile.ListAuthorAliases()
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
				// TODO get port as argument
				// TODO
				fmt.Println("Serving...")
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
						// import
						fmt.Println("Importing retail epubs...")
						if len(lb.ConfigurationFile.RetailSource) == 0 {
							fmt.Println("No retail source found in configuration file!")
						} else {
							err := lb.ImportRetail()
							if err != nil {
								panic(err)
							}
						}
					},
				},
				{
					Name:    "nonretail",
					Aliases: []string{"n"},
					Usage:   "import non-retail epubs",
					Action: func(c *cli.Context) {
						fmt.Println("Importing non-retail epubs...")
						if len(lb.ConfigurationFile.NonRetailSource) == 0 {
							fmt.Println("No non-retail source found in configuration file!")
						} else {
							err := lb.ImportNonRetail()
							if err != nil {
								panic(err)
							}
						}
					},
				},
			},
		},
		{
			Name:    "export",
			Aliases: []string{"x"},
			Usage:   "export to E-Reader",
			Action: func(c *cli.Context) {
				// TODO
				fmt.Println("Exporting...")
			},
		},
		{
			Name:    "check",
			Aliases: []string{"fsck"},
			Usage:   "check library",
			Action: func(c *cli.Context) {
				fmt.Println("Checking...")
				// TODO check all retail epubs for changes
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
			Name:     "search",
			Category: "searching",
			Aliases:  []string{"c"},
			Usage:    "search the epub collection",
			Action: func(c *cli.Context) {
				if c.NArg() == 0 {
					fmt.Println("No query found!")
				} else {
					query := strings.Join(c.Args(), " ")
					fmt.Println("Searching for '" + query + "'...")
					results, err := lb.RunQuery(query)
					if err != nil {
						panic(err)
					}
					fmt.Println(results)
				}
			},
		},
		{
			Name:     "list",
			Category: "searching",
			Aliases:  []string{"ls"},
			Usage:    "list epubs in the collection",
			Subcommands: []cli.Command{
				{
					Name:    "untagged",
					Aliases: []string{"u"},
					Usage:   "list untagged epubs.",
					Action: func(c *cli.Context) {
						// TODO
						fmt.Println("Listing untagged epubs...")
					},
				},
				{
					Name:    "tags",
					Aliases: []string{"t"},
					Usage:   "list tags",
					Action: func(c *cli.Context) {
						// TODO
						fmt.Println("Listing tags...")
					},
				},
				{
					Name:    "series",
					Aliases: []string{"c"},
					Usage:   "list series.",
					Action: func(c *cli.Context) {
						// TODO
						fmt.Println("Listing series...")
					},
				},
				{
					Name:    "authors",
					Aliases: []string{"a"},
					Usage:   "list authors.",
					Action: func(c *cli.Context) {
						// TODO
						fmt.Println("Listing authors...")
					},
				},
			},
		},
		{
			Name:     "tag",
			Category: "tags",
			Usage:    "manage tags in the collection",
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a"},
					Usage:   "add tag(s) to book.",
					Action: func(c *cli.Context) {
						// TODO
						// params: ID, tag list
						fmt.Println("Adding tag to book ID#...")
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"r"},
					Usage:   "remove tag(s) from book.",
					Action: func(c *cli.Context) {
						// TODO
						fmt.Println("Removing tag from book ID#...")
					},
				},
				{
					Name:    "list",
					Aliases: []string{"c"},
					Usage:   "list tags for a book.",
					Action: func(c *cli.Context) {
						// TODO
						fmt.Println("Listing tags for book ID#...")
					},
				},
			},
		},
		{
			Name:     "series",
			Category: "series",
			Usage:    "manage series in the collection",
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a"},
					Usage:   "add series to book with: ID seriesname seriesindex",
					Action: func(c *cli.Context) {
						book, seriesInfo, err := checkArgsWithID(lb, c.Args())
						if err != nil || len(seriesInfo) != 2 {
							fmt.Println("Error parsing arguments: " + err.Error())
							return
						}
						seriesIndex, err := strconv.ParseFloat(seriesInfo[1], 32)
						if err != nil {
							fmt.Println("Index must be a float.")
							return
						}
						fmt.Printf("Adding single series to book %s...\n", book.ShortString())
						// add series
						if book.Series.Add(seriesInfo[0], float32(seriesIndex)) {
							fmt.Printf("Series %s #%f added to %s\n", seriesInfo[0], seriesIndex, book.ShortString())
						}
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"r"},
					Usage:   "remove series from book.",
					Action: func(c *cli.Context) {
						book, series, err := checkArgsWithID(lb, c.Args())
						if err != nil {
							fmt.Println("Error parsing arguments: " + err.Error())
							return
						}
						fmt.Printf("Removing series from book %s...\n", book.ShortString())
						// remove series
						if book.Series.Remove(series...) {
							fmt.Printf("Series %s removed from %s\n", strings.Join(series, ", "), book.ShortString())
						}
					},
				},
				{
					Name:    "list",
					Aliases: []string{"c"},
					Usage:   "list series for a book.",
					Action: func(c *cli.Context) {
						book, _, err := checkArgsWithID(lb, c.Args())
						if err != nil {
							fmt.Println("Error parsing arguments: " + err.Error())
							return
						}
						fmt.Printf("Listing series from book %s...\n", book.ShortString())
						fmt.Println(book.Series.String())
					},
				},
			},
		},
	}
	app.Run(os.Args)
}
