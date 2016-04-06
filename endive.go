package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/ttacon/chalk"
)

func main() {
	fmt.Println(chalk.Bold.TextStyle("\n# # # E N D I V E # # #\n"))

	// get library
	l, err := OpenLibrary()
	if err != nil {
		panic(err)
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
						aliases, err := l.ConfigurationFile.ListAuthorAliases()
						if err != nil {
							panic(err)
						}
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
			Name:    "import",
			Aliases: []string{"i"},
			Usage:   "options for importing epubs",
			Subcommands: []cli.Command{
				{
					Name:    "retail",
					Aliases: []string{"r"},
					Usage:   "import retail epubs",
					Action: func(c *cli.Context) {
						// import
						fmt.Println("Importing retail epubs...")
						err := l.ImportRetail()
						if err != nil {
							panic(err)
						}

					},
				},
				{
					Name:    "nonretail",
					Aliases: []string{"n"},
					Usage:   "import non-retail epubs",
					Action: func(c *cli.Context) {
						fmt.Println("Importing non-retail epubs...")
						err := l.ImportNonRetail()
						if err != nil {
							panic(err)
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
			Name:    "search",
			Aliases: []string{"c"},
			Usage:   "search the epub collection",
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
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "list epubs in the collection",
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
	}
	app.Run(os.Args)
}
