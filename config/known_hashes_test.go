package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var hashesFile = "../test/hashes.json"

// TestKnownHashes tests for all KnownHashes functions
func TestKnownHashes(t *testing.T) {
	fmt.Println("+ Testing KnownHashes...")
	assert := assert.New(t)
	c := KnownHashes{Filename: hashesFile}

	// check Load
	err := c.Load()
	assert.Nil(err, "Error loading hashes file")
	assert.Equal(len(c.Hashes), 2)
	assert.Equal(c.Count, 2)

	// check Save before modification
	wasModified, err := c.Save()
	assert.Nil(err, "Error saving")
	assert.False(wasModified, "Nothing has changed, should not have saved")

	// check IsIN
	isIn := c.IsIn("moijmoij")
	assert.False(isIn, "Bad hash cannot be in list")
	isIn = c.IsIn("74657165a56c9a54ed887cd895a0f67a70f29cbecaa96dfda840c76580da3dd8")
	assert.True(isIn, "Known hash should be detected")
	isIn = c.IsIn("74657165a56c9a54ed887cd895a0f67a70f29cbecaa96dfda840c76580da3dd9")
	assert.False(isIn, "Unknown hash should not be detected")

	// check Add
	added, err := c.Add("moijmoij")
	assert.NotNil(err, "Bad hash should have raised error")
	assert.False(added, "Bad hash should not have been added")
	added, err = c.Add("74657165a56c9a54ed887cd895a0f67a70f29cbecaa96dfda840c76580da3dd8")
	assert.Nil(err, "Hash should not have raised error")
	assert.False(added, "Known hash should not have been added")
	added, err = c.Add("74657165a56c9a54ed887cd895a0f67a70f29cbecaa96dfda840c76580da3dd9")
	assert.Nil(err, "Hash should not have raised error")
	assert.True(added, "Unknown hash should have been added")

	// check Save after modification
	c.Filename += "_temp"
	wasModified, err = c.Save()
	assert.Nil(err, "Error saving")
	assert.True(wasModified, "One hash was added, should have saved")

	// cleanup
	err = os.Remove(c.Filename)
	assert.Nil(err, "Error during cleanup of file "+c.Filename)
}
