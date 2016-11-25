package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fmt"

	b "github.com/barsanuphe/endive/book"
	"github.com/barsanuphe/endive/db"
	en "github.com/barsanuphe/endive/endive"
	l "github.com/barsanuphe/endive/library"
	"github.com/barsanuphe/endive/mock"
)

const (
	testIncorrectInput = "Incorrect input, expecting error."
	testAllSelected    = "All Books selected."
)

func TestCLI(t *testing.T) {
	fmt.Println("\n --- Testing CLI. ---")
	assert := assert.New(t)

	// config
	c := en.Config{}
	c.LibraryRoot = "test"
	c.DatabaseFile = "test/endive.json"
	c.RetailSource = []string{"test"}
	c.NonRetailSource = []string{"test"}
	c.EpubFilenameFormat = "$a - $t"
	// building endive struct
	db := &db.JSONDB{}
	db.SetPath(c.DatabaseFile)
	ui := &mock.UserInterface{}
	lib := l.Library{Collection: &b.Books{}, Config: c, Index: &mock.IndexService{}, UI: ui, DB: db}
	err := lib.Load()
	assert.Nil(err, "Error loading epubs from database")
	endive := &Endive{Config: c, UI: ui, Library: lib}

	// testing help
	fmt.Println(" + Testing config subcommand")
	cli := CLI{}
	err = cli.parseArgs(endive, []string{"-h"})
	assert.Nil(err)
	assert.True(cli.builtInCommand)

	// testing config
	fmt.Println(" + Testing config subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"config"})
	assert.Nil(err)
	assert.False(cli.builtInCommand)
	assert.True(cli.showConfig)

	// testing collection
	fmt.Println(" + Testing collection subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"collection", "check"})
	assert.Nil(err)
	assert.True(cli.checkCollection)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"collection", "refresh"})
	assert.Nil(err)
	assert.True(cli.refreshCollection)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"collection", "rebuild-index"})
	assert.Nil(err)
	assert.True(cli.rebuildIndex)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"collection", "check-index"})
	assert.Nil(err)
	assert.True(cli.checkIndex)

	// testing import
	fmt.Println(" + Testing import subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"import", "r"})
	assert.Nil(err)
	assert.True(cli.importEpubs)
	assert.True(cli.importRetail)
	assert.False(cli.listImport)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"import", "r", "hop.doc"})
	assert.NotNil(err, "Not an epub")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"i", "nonretail", "test/pg16328.epub", "test/pg17989.epub"})
	assert.Nil(err)
	assert.True(cli.importEpubs)
	assert.False(cli.importRetail)
	assert.False(cli.listImport)
	assert.Equal(2, len(cli.epubs))
	err = cli.parseArgs(endive, []string{"i", "retail", "--list"})
	assert.Nil(err)
	assert.True(cli.importEpubs)
	assert.True(cli.importRetail)
	assert.True(cli.listImport)

	// testing export
	fmt.Println(" + Testing export subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"export"})
	assert.NotNil(err)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"export", "title:thing"})
	assert.Nil(err)
	assert.True(cli.export)
	assert.Equal(1, len(cli.searchTerms))

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"export", "all"})
	assert.Nil(err)
	assert.True(cli.export)
	assert.Equal(2, len(cli.collection.Books()), testAllSelected)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"export", "id", "1"})
	assert.Nil(err)
	assert.True(cli.export)
	assert.Equal(1, len(cli.collection.Books()), "1 book selected.")

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"export", "id", "1", "--dir=/doesnotexist"})
	assert.NotNil(err)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"export", "id", "1", "--dir=/tmp"})
	assert.Nil(err)
	assert.True(cli.export)
	assert.Equal(1, len(cli.collection.Books()), "1 book selected.")
	assert.Equal("/tmp", cli.exportDirectory)

	// testing info
	fmt.Println(" + Testing info subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info"})
	assert.Nil(err)
	assert.Equal(cli.info, infoGeneral)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info", "1"})
	assert.Nil(err, "Book should be in collection")
	assert.Equal(cli.info, infoBook)
	assert.Equal(1, len(cli.books), "One book should be found.")
	assert.Equal(1, cli.books[0].ID(), "It should have ID 1.")

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info", "tags"})
	assert.Nil(err)
	assert.Equal(cli.info, infoTags)
	assert.Equal(10, len(cli.collectionMap), "10 different tags in test epubs")

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info", "series"})
	assert.Nil(err)
	assert.Equal(cli.info, infoSeries)
	assert.Equal(0, len(cli.collectionMap), "No series defined in test epubs")

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info", "authors"})
	assert.Nil(err)
	assert.Equal(cli.info, infoAuthors)
	assert.Equal(2, len(cli.collectionMap), "2 authors defined in test epubs")

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info", "publishers"})
	assert.Nil(err)
	assert.Equal(cli.info, infoPublishers)
	assert.Equal(1, len(cli.collectionMap), "Publisher 'unknown' defined in test epubs")

	// wrong input
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info", "33"})
	assert.NotNil(err, testIncorrectInput)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info", "zz"})
	assert.NotNil(err, testIncorrectInput)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info", "tags", "xx"})
	assert.NotNil(err, testIncorrectInput)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"info", "1", "2"})
	assert.NotNil(err, testIncorrectInput)

	// testing list
	fmt.Println(" + Testing list subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"list"})
	assert.Nil(err)
	assert.True(cli.list)
	assert.Equal(2, len(cli.collection.Books()), testAllSelected)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"ls", "-f", "2"})
	assert.Nil(err)
	assert.True(cli.list)
	assert.Equal(2, cli.firstN)
	assert.Equal(invalidLimit, cli.lastN)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"ls", "--last=5"})
	assert.Nil(err)
	assert.True(cli.list)
	assert.Equal(invalidLimit, cli.firstN)
	assert.Equal(5, cli.lastN)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"ls", "--last=5", "-s", "Year"})
	assert.Nil(err)
	assert.True(cli.list)
	assert.Equal(invalidLimit, cli.firstN)
	assert.Equal(5, cli.lastN)
	assert.Equal("year", cli.sortBy)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"ls", "--retail"})
	assert.Nil(err)
	assert.Equal(2, len(cli.collection.Books()), testAllSelected)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"ls", "--nonretail"})
	assert.Nil(err)
	assert.Equal(0, len(cli.collection.Books()), "No book selected.")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"ls", "--incomplete"})
	assert.Nil(err)
	assert.Equal(2, len(cli.collection.Books()), testAllSelected)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"ls", "--last=5", "-f", "1"})
	assert.NotNil(err, "Cannot have both last and first flags")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"ls", "-f"})
	assert.NotNil(err, testIncorrectInput)
	err = cli.parseArgs(endive, []string{"ls", "-f", "a"})
	assert.NotNil(err, testIncorrectInput)

	// testing search
	fmt.Println(" + Testing search subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"search"})
	assert.NotNil(err)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"search", "title:thing"})
	assert.Nil(err)
	assert.True(cli.search)
	assert.Equal(1, len(cli.searchTerms))

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"search", "-s", "Year", "title:thing", "+author:guy", "--first=2"})
	assert.Nil(err)
	assert.True(cli.search)
	assert.Equal(2, cli.firstN)
	assert.Equal(invalidLimit, cli.lastN)
	assert.Equal("year", cli.sortBy)
	assert.Equal(2, len(cli.searchTerms))

	// testing review
	fmt.Println(" + Testing review subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"review"})
	assert.NotNil(err)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"review", "1", "thing"})
	assert.NotNil(err)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"review", "100", "4.3"})
	assert.NotNil(err)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"review", "1", "2.5"})
	assert.Nil(err)
	assert.True(cli.review)
	assert.Equal("2.5", cli.rating)
	assert.Equal("", cli.reviewText)
	assert.Equal(1, len(cli.books))

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"review", "1", "2.5", "mediocre"})
	assert.Nil(err)
	assert.True(cli.review)
	assert.Equal("2.5", cli.rating)
	assert.Equal("mediocre", cli.reviewText)
	assert.Equal(1, len(cli.collection.Books()))

	// testing set
	fmt.Println(" + Testing set subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"set", "unread", "1", "20"})
	assert.NotNil(err)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"set", "unreadj", "1", "2"})
	assert.NotNil(err)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"set", "unread", "1", "2"})
	assert.Nil(err)
	assert.True(cli.set)
	assert.Equal("unread", cli.progress)
	assert.Equal(2, len(cli.books))

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"set", "read", "1"})
	assert.Nil(err)
	assert.True(cli.set)
	assert.Equal("read", cli.progress)
	assert.Equal(1, len(cli.books))

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"set", "field", "title", "thing", "1"})
	assert.Nil(err)
	assert.True(cli.set)
	assert.Equal("", cli.progress)
	assert.Equal(1, len(cli.books))
	assert.Equal("title", cli.field)
	assert.Equal("thing", cli.value)

	// testing edit
	fmt.Println(" + Testing edit subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"edit", "field", "title", "1", "20"})
	assert.NotNil(err)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"edit", "field", "1", "2"})
	assert.NotNil(err)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"edit", "title", "1", "20"})
	assert.NotNil(err)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"edit", "field", "title", "1", "2"})
	assert.Nil(err)
	assert.True(cli.edit)
	assert.Equal("title", cli.field)
	assert.Equal("", cli.value)
	assert.Equal(2, len(cli.books))

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"edit", "1", "2"})
	assert.Nil(err)
	assert.True(cli.edit)
	assert.Equal("", cli.field)
	assert.Equal("", cli.value)
	assert.Equal(2, len(cli.books))

	// testing reset
	fmt.Println(" + Testing reset subcommand")
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"reset", "field", "title", "1", "20"})
	assert.NotNil(err)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"reset", "field", "1", "2"})
	assert.NotNil(err)
	cli = CLI{}
	err = cli.parseArgs(endive, []string{"reset", "title", "1", "20"})
	assert.NotNil(err)

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"reset", "field", "title", "1", "2"})
	assert.Nil(err)
	assert.True(cli.reset)
	assert.Equal("title", cli.field)
	assert.Equal("", cli.value)
	assert.Equal(2, len(cli.books))

	cli = CLI{}
	err = cli.parseArgs(endive, []string{"reset", "1", "2"})
	assert.Nil(err)
	assert.True(cli.reset)
	assert.Equal("", cli.field)
	assert.Equal("", cli.value)
	assert.Equal(2, len(cli.books))
}
