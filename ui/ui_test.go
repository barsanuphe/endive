package ui

import (
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
)

func TestUIEdit(t *testing.T) {
	fmt.Println("+ Testing UI/Edit()...")

	assert := assert.New(t)
	ui := &UI{}

	// fake $EDITOR
	fakeCommand := "#!/bin/bash\necho test > $1"
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

	// os.Setenv("PATH", "/usr/bin:/sbin")

	// edit
	output, err := ui.Edit(tmpfile.Name(), "input")
	assert.Nil(err, "Error editing file")
	assert.Equal("test\n", output, "Expecting temp file contents to be: test")

}
