/*
Endive is a tool to keep your epub library in great shape.

It can rename and organize your library from the epub metadata, and can keep
track of retail and non-retail versions.

It is in a very early development, with basically nothing actually working,
and chances of files disappearing if invoked.

*/
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	b "github.com/barsanuphe/endive/book"
	l "github.com/barsanuphe/endive/library"
	"github.com/bndr/gotabulate"
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

func showInfo(lb l.Library, c *cli.Context) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	relativePath, err := filepath.Rel(lb.ConfigurationFile.LibraryRoot, book.GetMainFilename())
	if err != nil {
		panic(err)
	}

	var rows [][]string
	rows = append(rows, []string{"ID", strconv.Itoa(book.ID)})
	rows = append(rows, []string{"Filename", relativePath})
	rows = append(rows, []string{"Author", book.Metadata.GetFirstValue("creator")})
	rows = append(rows, []string{"Title", book.Metadata.GetFirstValue("title")})
	rows = append(rows, []string{"Publication Year", book.Metadata.GetFirstValue("year")})
	if len(book.Tags) != 0 {
		rows = append(rows, []string{"Tags", strings.Join(book.Tags, " / ")})
	}
	if len(book.Series) != 0 {
		rows = append(rows, []string{"Series", book.Series.String()})
	}
	available := ""
	if book.HasRetail() {
		available += "retail "
	}
	if book.HasNonRetail() {
		available += "non-retail"
	}
	rows = append(rows, []string{"Available versions", available})

	t := gotabulate.Create(rows)
	t.SetHeaders([]string{"Info", "Book"})
	t.SetEmptyString("N/A")
	t.SetAlign("left")
	fmt.Println(t.Render("simple"))
}

func listTags(lb l.Library, c *cli.Context) (err error) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	fmt.Println(book.ShortString())
	fmt.Println(strings.Join(book.Tags, " / "))
	return
}

func addTags(lb l.Library, c *cli.Context) {
	book, tags, err := checkArgsWithID(lb, c.Args())
	if err != nil || len(tags) == 0 {
		fmt.Println("Error parsing arguments")
		return
	}
	if book.AddTags(tags...) {
		fmt.Printf("Tags added to %s\n", book.ShortString())
	}
}

func removeTags(lb l.Library, c *cli.Context) {
	book, tags, err := checkArgsWithID(lb, c.Args())
	if err != nil || len(tags) == 0 {
		fmt.Println("Error parsing arguments")
		return
	}
	if book.RemoveTags(tags...) {
		fmt.Printf("Tags removed from %s\n", book.ShortString())
	}
}

func listSeries(lb l.Library, c *cli.Context) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	fmt.Println(book.ShortString())
	fmt.Println(book.Series.String())
}

func removeSeries(lb l.Library, c *cli.Context) {
	book, series, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	// remove series
	if book.Series.Remove(series...) {
		fmt.Printf("Series %s removed from %s\n", strings.Join(series, ", "), book.ShortString())
	}
}

func addSeries(lb l.Library, c *cli.Context) {
	book, seriesInfo, err := checkArgsWithID(lb, c.Args())
	if err != nil || len(seriesInfo) != 2 {
		fmt.Println("Error parsing arguments")
		return
	}
	seriesIndex, err := strconv.ParseFloat(seriesInfo[1], 32)
	if err != nil {
		fmt.Println("Index must be a float.")
		return
	}
	// add series
	if book.Series.Add(seriesInfo[0], float32(seriesIndex)) {
		fmt.Printf("Series %s #%f added to %s\n", seriesInfo[0], seriesIndex, book.ShortString())
	}
}

func search(lb l.Library, c *cli.Context) {
	if c.NArg() == 0 {
		fmt.Println("No query found!")
	} else {
		query := strings.Join(c.Args(), " ")
		fmt.Println("Searching for '" + query + "'...")
		results, err := lb.RunQuery(query)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		fmt.Println(results)
	}
}

func importEpubs(lb l.Library, c *cli.Context, isRetail bool) {
	var err error
	if isRetail {
		if len(lb.ConfigurationFile.RetailSource) == 0 {
			fmt.Println("No retail source found in configuration file!")
			return
		}
		fmt.Println("Importing retail epubs...")
		err = lb.ImportRetail()
	} else {
		if len(lb.ConfigurationFile.NonRetailSource) == 0 {
			fmt.Println("No non-retail source found in configuration file!")
			return
		}
		fmt.Println("Importing non-retail epubs...")
		err = lb.ImportNonRetail()
	}
	if err != nil {
		panic(err)
	}
}

func generateCLI(lb l.Library) (app *cli.App) {
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
						fmt.Println(lb.ConfigurationFile.String())
					},
				},
				{
					Name:  "aliases",
					Usage: "show aliases defined in configuration",
					Action: func(c *cli.Context) {
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
				// TODO
				fmt.Println("Exporting...")
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
						addTags(lb, c)
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"r"},
					Usage:   "remove tag(s) from book.",
					Action: func(c *cli.Context) {
						removeTags(lb, c)
					},
				},
				{
					Name:    "list",
					Aliases: []string{"c"},
					Usage:   "list tags for a book.",
					Action: func(c *cli.Context) {
						listTags(lb, c)
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
						addSeries(lb, c)
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"r"},
					Usage:   "remove series from book.",
					Action: func(c *cli.Context) {
						removeSeries(lb, c)
					},
				},
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Usage:   "list series for a book.",
					Action: func(c *cli.Context) {
						listSeries(lb, c)
					},
				},
			},
		},
	}
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

	// generate CLI interface and run it
	app := generateCLI(lb)
	app.Run(os.Args)
}
