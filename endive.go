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

	"github.com/codegangsta/cli"
	"github.com/ttacon/chalk"
)

func main() {
	fmt.Println(chalk.Bold.TextStyle("\n# # # E N D I V E # # #\n"))

	// get library
	l, err := OpenLibrary()
	if err != nil {
		fmt.Println("Error loading configuration. Check it.")
		fmt.Println(err.Error())
		return
	}
	defer l.Save()

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
						fmt.Println(l.ConfigurationFile.String())
					},
				},
				{
					Name:  "aliases",
					Usage: "show aliases defined in configuration",
					Action: func(c *cli.Context) {
						// print aliases
						aliases := l.ConfigurationFile.ListAuthorAliases()
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
						if len(l.ConfigurationFile.RetailSource) == 0 {
							fmt.Println("No retail source found in configuration file!")
						} else {
							err := l.ImportRetail()
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
						if len(l.ConfigurationFile.NonRetailSource) == 0 {
							fmt.Println("No non-retail source found in configuration file!")
						} else {
							err := l.ImportNonRetail()
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
				renamed, err := l.Refresh()
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
					results, err := l.RunQuery(query)
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
					Usage:   "add series to book.",
					Action: func(c *cli.Context) {
						// TODO
						// params: ID, series:index
						fmt.Println("Adding series to book ID#...")
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"r"},
					Usage:   "remove series from book.",
					Action: func(c *cli.Context) {
						// TODO
						fmt.Println("Removing series from book ID#...")
					},
				},
				{
					Name:    "list",
					Aliases: []string{"c"},
					Usage:   "list series for a book.",
					Action: func(c *cli.Context) {
						// TODO
						fmt.Println("Listing series for book ID#...")
					},
				},
			},
		},
	}
	app.Run(os.Args)
}
