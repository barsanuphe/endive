package main

import (
	"fmt"
	"testing"
)

var testEpubs = []struct {
	filename              string
	expectedTitle         string
	expectedSha256        string
}{
	{
		"test/pg16328.epub",
		"Beowulf / An Anglo-Saxon Epic Poem",
		"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03",
	},
	{
		"test/pg17989.epub",
		"Le comte de Monte-Cristo, Tome I",
		"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a",
	},
}

func TestEpubGetHash(t *testing.T) {
	fmt.Println("+ Testing Epub.GetHash()...")
	for _, testEpub := range testEpubs {
		e := NewBook(testEpub.filename, standardTestConfig, true)
		err := e.RetailEpub.GetHash()
		if err != nil {
			t.Errorf("Error calculating hash for %s", e.getMainFilename())
		}
		if e.RetailEpub.Hash != testEpub.expectedSha256 {
			t.Errorf("GetHash(%s) returned %s, expected %s!", testEpub.filename, e.RetailEpub.Hash, testEpub.expectedSha256)
		}
	}
}
