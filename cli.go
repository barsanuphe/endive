package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	docopt "github.com/docopt/docopt-go"

	en "github.com/barsanuphe/endive/endive"
	b "github.com/barsanuphe/endive/book"
)

const (
	incorrectInput      = "Incorrect input, check endive -h for complete help."
	incorrectIDValue    = "Incorrect ID."
	noBookFound         = "Book with ID %d cannot be found"
	numberOfBooksHeader = "# of Books"
	incorrectFlag       = "--first and --last only support integer values"
	invalidLimit        = -1
	endiveVersion       = "Endive -- CLI Epub collection manager -- v1.0."
	endiveUsage         = `
Endive.
This is an epub collection manager.

The main commands are:
	config		Display current configuration
	collection	Do some maintenance on the collection
	import, i	Import epubs to the collection
	export, x	Export epubs to ereader
	info		Display information
	edit		Edit metadata
	progress, p	Set book reading progress
	list, ls	List books
	search, s	Search for specific books

Searching / Exporting:
	A list of strings can be given as input to search for books.
	It is also possible to restrict a value to a specific field: field:value.
	Valid fields are:
		author, title, year, language, series, tag, publisher, category,
		type, genre, description, exported, progress, review.
	Examples:
		'author:XX title:YY' will give results satifsying any of the two conditions.
		'author:XX +title:YY' will give results satifsying both conditions.
		'author:XX -title:YY' will give results satifsying the first condition excluding the second.

Usage:
	endive config
	endive collection (check|refresh|rebuild-index|check-index)
	endive (import|i) ((retail|r)|(nonretail|nr)) [--list] [<epub>...]
	endive (export|x) (all|(id <ID>...)|<search-criteria>...)
	endive info [tags|series|authors|publishers] [<ID>]
	endive (list|ls) [--incomplete|--nonretail|--retail] [--first=N] [--last=N] [--sort=SORT]
	endive (search|s) <search-criteria>... [--first=N] [--last=N] [--sort=SORT]
	endive review <ID> <rating> [<review>]
	endive set (unread|read|reading|shortlisted|(field <field_name> <value>)) <ID>...
	endive edit [(field <field_name>)] <ID>...
	endive reset [(field <field_name>)] <ID>...
	endive -h | --help
	endive --version

Options:
	-h --help            Show this screen.
	--version            Show version.
	--list               List importable epubs only.
	-f N --first=N       Filter only the n first books.
	-l N --last=N        Filter only the n last books.
	-s SORT --sort=SORT  Sort results [default: id].
	--incomplete         Filter books with incomplete metadata.
	--retail             Only show retail books.
	--nonretail          Only show non-retail books.`
)

type CLI struct {
	builtInCommand bool
	// argument values
	books         []*b.Book
	collection    en.Collection
	collectionMap map[string]int
	epubs         []string
	searchTerms   []string
	field         string
	value         string
	// flags
	lastN  int
	firstN int
	sortBy string
	// config
	showConfig bool
	// collection
	checkCollection   bool
	checkIndex        bool
	refreshCollection bool
	rebuildIndex      bool
	// import
	importRetail bool
	importEpubs  bool
	listImport   bool
	// export
	export bool
	// info
	info string
	// search
	search bool
	// review
	review     bool
	rating     string
	reviewText string
	// list
	list bool
	// edit, set, reset
	edit     bool
	set      bool
	reset    bool
	progress string
}

func (o *CLI) parseArgs(e *Endive, osArgs []string) error {
	// parse arguments and options
	args, err := docopt.Parse(endiveUsage, osArgs, true, endiveVersion, false, false)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		// builtin command, nothing to do.
		o.builtInCommand = true
		return nil
	}

	// init
	o.collection = e.Library.Collection
	// checking if IDs were given, getting relevant *Book-s
	if args["<ID>"] != nil {
		idsString := []string{}
		// test if string or []string
		idS, ok := args["<ID>"].(string)
		if ok {
			idsString = append(idsString, idS)
		} else {
			idsString, ok = args["<ID>"].([]string)
			if !ok {
				return errors.New(incorrectInput)
			}
		}
		// if [<ID>], idsString can be an empty slice
		if len(idsString) != 0 {
			// convert to int
			ids := []int{}
			for _, i := range idsString {
				id, err := strconv.Atoi(i)
				if err != nil {
					return errors.New(incorrectInput)
				}
				ids = append(ids, id)
			}
			if len(ids) == 0 {
				return errors.New(incorrectIDValue)
			}
			// get the relevant Books
			for _, id := range ids {
				bk, err := e.Library.Collection.FindByID(id)
				if err != nil {
					return fmt.Errorf(noBookFound, id)
				}
				o.books = append(o.books, bk.(*b.Book))
			}
			// get the equivalent collection
			o.collection = o.collection.WithID(ids...)
			if len(o.books) == 0 {
				return errors.New("No valid book found.")
			}
		}
	}

	// checking other common flags
	o.firstN = invalidLimit
	if args["--first"] != nil {
		o.firstN, err = strconv.Atoi(args["--first"].(string))
		if err != nil {
			return errors.New(incorrectFlag)
		}
	}
	o.lastN = invalidLimit
	if args["--last"] != nil {
		o.lastN, err = strconv.Atoi(args["--last"].(string))
		if err != nil {
			return errors.New(incorrectFlag)
		}
	}
	o.sortBy = strings.ToLower(args["--sort"].(string))

	// commands
	o.showConfig = args["config"].(bool)

	if args["collection"].(bool) {
		o.checkCollection = args["check"].(bool)
		o.rebuildIndex = args["rebuild-index"].(bool)
		o.refreshCollection = args["refresh"].(bool)
		o.checkIndex = args["check-index"].(bool)
	}

	if args["import"].(bool) || args["i"].(bool) {
		o.importEpubs = true
		// if not retail, non-retail.
		o.importRetail = args["retail"].(bool) || args["r"].(bool)
		o.listImport = args["--list"].(bool)
		o.epubs = args["<epub>"].([]string)
		// cheking they are existing epubs
		for _, epub := range o.epubs {
			//  assert epub exists
			if _, err := en.FileExists(epub); err != nil {
				return errors.New("Epub does not exist!")
			}
			// assert it's an epub
			if !strings.HasSuffix(strings.ToLower(epub), en.EpubExtension) {
				return errors.New(epub + " is not an epub")
			}
		}
	}

	o.search = args["search"].(bool) || args["s"].(bool)
	if args["<search-criteria>"] != nil {
		o.searchTerms = args["<search-criteria>"].([]string)
		if o.search && len(o.searchTerms) == 0 {
			return errors.New("No search terms found.")
		}
	}

	if args["export"].(bool) || args["x"].(bool) {
		o.export = true
		if args["all"].(bool) {
			o.collection = e.Library.Collection
		}
		// if ids: o.collection is already set
		// if search: same for o.searchTerms
	}

	if args["info"].(bool) {
		if args["tags"].(bool) {
			o.info = "Tags"
			o.collectionMap = o.collection.Tags()
		} else if args["series"].(bool) {
			o.info = "Series"
			o.collectionMap = o.collection.Series()
		} else if args["authors"].(bool) {
			o.info = "Authors"
			o.collectionMap = o.collection.Authors()
		} else if args["publishers"].(bool) {
			o.info = "Publishers"
			o.collectionMap = o.collection.Publishers()
		} else {
			if len(o.books) != 0 {
				o.info = "Book"
			} else {
				o.info = "General"
			}
		}
	}

	o.review = args["review"].(bool)
	if args["<rating>"] != nil {
		o.rating = args["<rating>"].(string)
		// checking rating is between 0 and 5
		if r, err := strconv.ParseFloat(o.rating, 32); err != nil || r > 5 || r < 0 {
			return errors.New("Rating must be between 0 and 5.")
		}
	}
	if args["<review>"] != nil {
		o.reviewText = args["<review>"].(string)
	}

	o.list = args["list"].(bool) || args["ls"].(bool)
	if args["--incomplete"].(bool) {
		o.collection = o.collection.Incomplete()
	}
	if args["--retail"].(bool) {
		o.collection = o.collection.Retail()
	}
	if args["--nonretail"].(bool) {
		o.collection = o.collection.NonRetailOnly()
	}

	if args["<field_name>"] != nil {
		o.field = args["<field_name>"].(string)
		// TODO check it's valid
	}
	if args["<value>"] != nil {
		o.value = args["<value>"].(string)
	}
	o.edit = args["edit"].(bool)
	o.reset = args["reset"].(bool)
	o.set = args["set"].(bool)
	for _, p := range []string{"unread", "read", "reading", "shortlisted"} {
		if args[p].(bool) {
			o.progress = p
			break
		}
	}
	if o.set && !args["field"].(bool) && o.progress == "" {
		return errors.New("Invalid progress")
	}

	return nil
}
