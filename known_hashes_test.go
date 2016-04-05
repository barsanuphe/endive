package main

import (
	"fmt"
	"testing"
	"os"
)

var hashesFile string = "test/hashes.json"

// TestKnownHashes tests for all KnownHashes functions
func TestKnownHashes(t *testing.T) {
	fmt.Println("+ Testing KnownHashes...")
	c := KnownHashes{Filename: hashesFile}

	// check Load
	err := c.Load()
	if err != nil {
		t.Errorf("Error loading hashes file: %s", err.Error())
	}
	if len(c.Hashes) != 2 && c.Count != 2 {
		t.Errorf("Error loading hashes, should have found 2 hashes instead of %d", len(c.Hashes))
	}

	// check Save before modification
	wasModified, err := c.Save()
	if err != nil {
		t.Errorf("Error saving")
	}
	if wasModified {
		t.Errorf("Nothing has changed, should not have saved")
	}

	// check IsIN
	isIn := c.IsIn("moijmoij")
	if isIn {
		t.Errorf("Bad hash cannot be in list")
	}
	isIn = c.IsIn("74657165a56c9a54ed887cd895a0f67a70f29cbecaa96dfda840c76580da3dd8")
	if !isIn {
		t.Errorf("Known hash should be detected")
	}
	isIn = c.IsIn("74657165a56c9a54ed887cd895a0f67a70f29cbecaa96dfda840c76580da3dd9")
	if isIn {
		t.Errorf("Unknown hash should not be detected")
	}

	// check Add
	added, err := c.Add("moijmoij")
	if err == nil {
		t.Errorf("Bad hash should have raised error")
	}
	if added {
		t.Errorf("Bad hash should not have been added")
	}
	added, err = c.Add("74657165a56c9a54ed887cd895a0f67a70f29cbecaa96dfda840c76580da3dd8")
	if err != nil {
		t.Errorf("Hash should not have raised error")
	}
	if added {
		t.Errorf("Known hash should not have been added")
	}
	added, err = c.Add("74657165a56c9a54ed887cd895a0f67a70f29cbecaa96dfda840c76580da3dd9")
	if err != nil {
		t.Errorf("Hash should not have raised error")
	}
	if !added {
		t.Errorf("Unknown hash should have been added")
	}

	// check Save after modification
	c.Filename += "_temp"
	wasModified, err = c.Save()
	if err != nil {
		t.Errorf("Error saving")
	}
	if !wasModified {
		t.Errorf("One hash was added, should have saved")
	}
	// cleanup
	err = os.Remove(c.Filename)
	if err != nil {
		t.Errorf("Error during cleanup of file " + c.Filename)
	}
}
