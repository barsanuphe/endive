package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	docopt "github.com/docopt/docopt-go"

	b "github.com/barsanuphe/endive/book"
	en "github.com/barsanuphe/endive/endive"
)

const (
	incorrectInput        = "Incorrect input, check endive -h for complete help."
	incorrectIDValue      = "Incorrect ID."
	noBookFound           = "Book with ID %d cannot be found"
	numberOfBooksHeader   = "# of Books"
	incorrectFlag         = "--first and --last only support integer values"
	directoryDoesNotExist = "Directory %s does not exist"
	invalidLimit          = -1
	infoTags              = "Tags"
	infoSeries            = "Series"
	infoPublishers        = "Publishers"
	infoAuthors           = "Authors"
	infoBook              = "Book"
	infoGeneral           = "General"

	endiveVersion = "Endive -- CLI Epub collection manager -- v1.0."
	endiveUsage   = `
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
	endive (export|x) (all|(id <ID>...)|<search-criteria>...) [--dir=DIRECTORY]
	endive info [tags|series|authors|publishers] [<ID>]
	endive (list|ls) [--incomplete|--nonretail|--retail] [--first=N|--last=N] [--sort=SORT]
	endive (search|s) <search-criteria>... [--first=N|--last=N] [--sort=SORT]
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
    --dir=DIRECTORY      Override the export directory in the configuration file.
	-f N --first=N       Filter only the n first books.
	-l N --last=N        Filter only the n last books.
	-s SORT --sort=SORT  Sort results [default: id].
	--incomplete         Filter books with incomplete metadata.
	--retail             Only show retail books.
	--nonretail          Only show non-retail books.`
)

// CLI sorts and checks user input
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
	export          bool
	exportDirectory string
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
	var ok bool
	// checking if IDs were given, getting relevant *Book-s
	if args["<ID>"] != nil {
		// test if string or []string
		idsString, ok := args["<ID>"].([]string)
		if !ok {
			return errors.New(incorrectInput)
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
				bk, err := o.collection.FindByID(id)
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
	if args["--dir"] != nil {
		exportDir := args["--dir"].(string)
		// check it exists, do not create in case of error
		if !en.DirectoryExists(exportDir) {
			return fmt.Errorf(directoryDoesNotExist, o.exportDirectory)
		}
		o.exportDirectory = exportDir
	}

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
	o.searchTerms, ok = args["<search-criteria>"].([]string)
	if ok && o.search && len(o.searchTerms) == 0 {
		return errors.New("No search terms found.")
	}

	// if export all: o.collection is set to complete collection by default
	// if export ids: o.collection is already set
	// if export search: same for o.searchTerms
	o.export = args["export"].(bool) || args["x"].(bool)

	if args["info"].(bool) {
		if args["tags"].(bool) {
			o.info = infoTags
			o.collectionMap = o.collection.Tags()
		} else if args["series"].(bool) {
			o.info = infoSeries
			o.collectionMap = o.collection.Series()
		} else if args["authors"].(bool) {
			o.info = infoAuthors
			o.collectionMap = o.collection.Authors()
		} else if args["publishers"].(bool) {
			o.info = infoPublishers
			o.collectionMap = o.collection.Publishers()
		} else {
			if len(o.books) != 0 {
				o.info = infoBook
			} else {
				o.info = infoGeneral
			}
		}
	}

	o.review = args["review"].(bool)
	o.rating, ok = args["<rating>"].(string)
	if ok {
		// checking rating is between 0 and 5
		if r, err := strconv.ParseFloat(o.rating, 32); err != nil || r > 5 || r < 0 {
			return errors.New("Rating must be between 0 and 5.")
		}
	}
	o.reviewText, _ = args["<review>"].(string)

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

	o.field, ok = args["<field_name>"].(string)
	if ok {
		o.field = strings.ToLower(o.field)
	}
	// check it's a valid field name
	if ok && !b.CheckValidField(o.field) {
		return errors.New("Invalid field!")
	}
	o.value, _ = args["<value>"].(string)

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
