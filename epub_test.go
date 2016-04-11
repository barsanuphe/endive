package main

import (
	"fmt"
	"testing"
)

func TestEpubGetHash(t *testing.T) {
	fmt.Println("+ Testing Epub.GetHash()...")
	for _, testEpub := range epubs {
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
