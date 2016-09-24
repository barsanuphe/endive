package ui

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUIEdit(t *testing.T) {
	fmt.Println("+ Testing UI/Edit()...")

	assert := assert.New(t)
	ui := &UI{}

	// fake $EDITOR
	fakeCommand := "#!/bin/bash\necho fake > $1"
	editor := "test_edit"
	tmpfile, err := ioutil.TempFile("", editor)
	assert.Nil(err)
	_, err = tmpfile.Write([]byte(fakeCommand))
	assert.Nil(err)
	err = tmpfile.Close()
	assert.Nil(err)
	defer os.Remove(tmpfile.Name())
	// chmod +X
	err = os.Chmod(tmpfile.Name(), 0777)
	assert.Nil(err)
	// setting $EDITOR
	err = os.Setenv("EDITOR", tmpfile.Name())
	assert.Nil(err)

	// edit
	output, err := ui.Edit("input")
	assert.Nil(err, "Error editing file")
	assert.Equal("fake", output, "Expecting temp file contents to be: test")

	// unsetting editor, should fail
	err = os.Unsetenv("EDITOR")
	assert.Nil(err)
	_, err = ui.Edit("input")
	assert.NotNil(err, "Error editing file")
}
