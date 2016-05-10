/*
Endive is a tool to keep your epub library in great shape.

It can rename and organize your library from the epub metadata, and can keep
track of retail and non-retail versions.

It is in a very early development: things can crash and files disappear.

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
	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
	l "github.com/barsanuphe/endive/library"

	"github.com/codegangsta/cli"
	"github.com/ttacon/chalk"
)

func checkPort(c *cli.Context) (port int, err error) {
	if len(c.Args()) < 1 {
		err = errors.New("Not enough arguments")
		return
	}
	port, err = strconv.Atoi(c.Args()[0])
	if err != nil {
		err = errors.New("Argument must be a valid port number")
		return
	}
	return
}

func checkArgsWithID(l *l.Library, args []string) (book *b.Book, other []string, err error) {
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

func markRead(lb *l.Library, c *cli.Context) {
	book, args, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	if len(args) >= 1 {
		// TODO check rating format
		book.Rating = args[0]
	}
	if len(args) == 2 {
		book.Review = args[1]
	}
	book.SetReadDateToday()
	book.SetProgress("read")
}

func showInfo(lb *l.Library, c *cli.Context) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	relativePath, err := filepath.Rel(lb.Config.LibraryRoot, book.FullPath())
	if err != nil {
		panic(err)
	}

	var rows [][]string
	rows = append(rows, []string{"ID", strconv.Itoa(book.ID)})
	rows = append(rows, []string{"Filename", relativePath})
	rows = append(rows, []string{"Author", book.Metadata.Author()})
	rows = append(rows, []string{"Title", book.Metadata.Title()})
	rows = append(rows, []string{"Publication Year", book.Metadata.Year})
	if book.Metadata.ISBN != "" {
		rows = append(rows, []string{"ISBN", book.Metadata.ISBN})
	}
	if len(book.Metadata.Tags) != 0 {
		rows = append(rows, []string{"Tags", book.Metadata.Tags.String()})
	}
	if len(book.Metadata.Series) != 0 {
		rows = append(rows, []string{"Series", book.Metadata.Series.String()})
	}
	available := ""
	if book.HasRetail() {
		available += "retail "
	}
	if book.HasNonRetail() {
		available += "non-retail"
	}
	rows = append(rows, []string{"Available versions", available})
	rows = append(rows, []string{"Progress", book.Progress})
	if book.ReadDate != "" {
		rows = append(rows, []string{"Read Date", book.ReadDate})
	}
	if book.Rating != "" {
		rows = append(rows, []string{"Rating", book.Rating})
	}
	if book.Review != "" {
		rows = append(rows, []string{"Review", book.Review})
	}
	fmt.Println(h.TabulateRows(rows, "Info", "Book"))
}

func listTags(lb *l.Library, c *cli.Context) (err error) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		// list all tags
		tags := lb.ListTags()
		fmt.Println(h.TabulateMap(tags, "Tag", "# of Books"))

	} else {
		// if ID, list tags of ID
		var rows [][]string
		rows = append(rows, []string{book.ShortString(), book.Metadata.Tags.String()})
		fmt.Println(h.TabulateRows(rows, "Book", "Tags"))
	}
	return
}

func addTags(lb *l.Library, c *cli.Context) {
	book, tags, err := checkArgsWithID(lb, c.Args())
	if err != nil || len(tags) == 0 {
		fmt.Println("Error parsing arguments")
		return
	}
	if book.Metadata.Tags.AddFromNames(tags...) {
		fmt.Printf("Tags added to %s\n", book.ShortString())
	}
}

func removeTags(lb *l.Library, c *cli.Context) {
	book, tags, err := checkArgsWithID(lb, c.Args())
	if err != nil || len(tags) == 0 {
		fmt.Println("Error parsing arguments")
		return
	}
	if book.Metadata.Tags.RemoveFromNames(tags...) {
		fmt.Printf("Tags removed from %s\n", book.ShortString())
	}
}

func listSeries(lb *l.Library, c *cli.Context) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		// list all series
		series := lb.ListSeries()
		fmt.Println(h.TabulateMap(series, "Series", "# of Books"))
	} else {
		// if ID, list series of ID
		var rows [][]string
		rows = append(rows, []string{book.ShortString(), book.Metadata.Series.String()})
		fmt.Println(h.TabulateRows(rows, "Book", "Series"))
	}
	return
}

func removeSeries(lb *l.Library, c *cli.Context) {
	book, series, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	// remove series
	if book.Metadata.Series.Remove(series...) {
		fmt.Printf("Series %s removed from %s\n", strings.Join(series, ", "), book.ShortString())
	}
}

func addSeries(lb *l.Library, c *cli.Context) {
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
	if book.Metadata.Series.Add(seriesInfo[0], float32(seriesIndex)) {
		fmt.Printf("Series %s #%f added to %s\n", seriesInfo[0], seriesIndex, book.ShortString())
	}
}

func search(lb *l.Library, c *cli.Context) {
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

func importEpubs(lb *l.Library, c *cli.Context, isRetail bool) {
	if len(c.Args()) >= 1 {
		// check valid path
		validPaths := []string{}
		validHashes := []string{}
		for _, path := range c.Args() {
			validPath, err := h.FileExists(path)
			if err == nil && filepath.Ext(validPath) == ".epub" {
				validPaths = append(validPaths, validPath)
				validHash, err := h.CalculateSHA256(path)
				if err != nil {
					return
				}
				validHashes = append(validHashes, validHash)
			}
		}
		// import valid paths
		err := lb.ImportEpubs(validPaths, validHashes, isRetail)
		if err != nil {
			return
		}
	} else {
		var err error
		if isRetail {
			if len(lb.Config.RetailSource) == 0 {
				fmt.Println("No retail source found in configuration file!")
				return
			}
			fmt.Println("Importing retail epubs...")
			err = lb.ImportRetail()
		} else {
			if len(lb.Config.NonRetailSource) == 0 {
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
}

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
				// TODO export with search ?
				fmt.Println("Exporting selection to E-Reader...")
			},
			Subcommands: []cli.Command{
				{
					Name:    "all",
					Aliases: []string{"books"},
					Usage:   "export everything.",
					Action: func(c *cli.Context) {
						fmt.Println("Exporting everything to E-Reader...")
						// TODO
					},
				},
				{
					Name:    "shortlisted",
					Aliases: []string{"unread", "reading", "read"},
					Usage:   "export books according to their reading progress",
					Action: func(c *cli.Context) {
						usedAlias := c.Parent().Args().First()
						fmt.Printf("Exporting selection (progress %s) to E-Reader...\n", usedAlias)
						// TODO
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
						// TODO
						// ask confirmation
					},
				},
				{
					Name:    "edit",
					Aliases: []string{"modify"},
					Usage:   "edit metadata field using book ID: metadata edit ID field values",
					Action: func(c *cli.Context) {
						// TODO
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
		{
			Name:     "tag",
			Aliases:  []string{"tags"},
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
					Aliases: []string{"ls"},
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
					Usage:   "add (or modify index of) a series to a book with: ID seriesname seriesindex",
					Action: func(c *cli.Context) {
						addSeries(lb, c)
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"r"},
					Usage:   "remove series from book with: ID seriesname.",
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

	err := h.GetEndiveLogger(cfg.XdgLogPath)
	defer h.LogFile.Close()

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
