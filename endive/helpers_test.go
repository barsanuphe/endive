package endive

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpersStringInSlice(t *testing.T) {
	fmt.Println("+ Testing Helpers/StringInSlice()...")
	candidates := []string{"one", "two"}
	idx, isIn := StringInSlice("one", candidates)
	if !isIn || idx != 0 {
		t.Error("Error finding string in slice")
	}
	idx, isIn = StringInSlice("One", candidates)
	if isIn || idx != -1 {
		t.Error("Error finding string in slice")
	}
}

func TestHelpersCSContains(t *testing.T) {
	fmt.Println("+ Testing Helpers/CaseInsensitiveContains()...")
	if !CaseInsensitiveContains("TestString", "test") {
		t.Error("Error, substring in string")
	}
	if !CaseInsensitiveContains("TestString", "stSt") {
		t.Error("Error, substring in string")
	}
	if CaseInsensitiveContains("TestString", "teest") {
		t.Error("Error, substring not in string")
	}
}

var isbns = []struct {
	candidate     string
	expectedISBN  string
	expectedError error
}{
	{
		"urn:ISBN: 9-780-5750-8-365-3  ",
		"9780575083653",
		nil,
	},
	{
		"9780575083653",
		"9780575083653",
		nil,
	},
	{
		"9780575083652",
		"",
		errors.New("ISBN-13 not found"),
	},
	{
		"0575083654",
		"9780575083653",
		nil,
	},
	{
		"0575083655",
		"",
		errors.New("ISBN-13 not found"),
	},
	{
		"A223234333432",
		"",
		errors.New("ISBN-13 not found"),
	},
	{
		"urn:isbn: 12-23-4-333-432  ",
		"",
		errors.New("ISBN-13 not found"),
	},
	{
		"urn:uuid:0adf2006-7812-4675-9c27-47699d21c4a2",
		"",
		errors.New("ISBN-13 not found"),
	},
}

func TestEpubCleanISBN(t *testing.T) {
	fmt.Println("+ Testing Info/CleanISBN()...")
	for _, c := range isbns {
		isbn, err := CleanISBN(c.candidate)
		if err == nil && c.expectedError != nil {
			t.Errorf("Unexpected error cleaning isbn %s", c.candidate)
		} else if err != nil && c.expectedError == nil {
			t.Errorf("Unexpected error cleaning isbn %s", c.candidate)
		} else if err != nil && c.expectedError != nil && c.expectedError.Error() != err.Error() {
			t.Errorf("Unexpected error cleaning isbn %s: got %s, expected %s", c.candidate, c.expectedError.Error(), err.Error())
		}
		assert.Equal(t, c.expectedISBN, isbn, "Error cleaning isbn")
	}
}

const (
	option1              = "option1"
	option2              = "option2"
	option3              = "option3"
	option4              = "option4"
	option5              = "option5"
	unexpectedOption     = "unexpected option label"
	expectedOptionsCount = "should have %d element(s)"
	expectedLocal        = "should be tagged local"
	expectedRemote       = "should be tagged remote"
)

func TestOptions(t *testing.T) {
	fmt.Println("+ Testing CleanSliceAndTagEntries/RemoveDuplicates()...")
	assert := assert.New(t)

	t1 := []string{option1, option1}
	RemoveDuplicates(&t1)
	assert.Equal(1, len(t1), fmt.Sprintf(expectedOptionsCount, 1))
	assert.Equal(option1, t1[0], unexpectedOption)

	t2 := []string{option1, option2, option3, option4, option5}
	RemoveDuplicates(&t2)
	assert.Equal(5, len(t2), fmt.Sprintf(expectedOptionsCount, 5))

	t3 := []string{option1, option2, option3, option4, option5, option5}
	RemoveDuplicates(&t3)
	assert.Equal(5, len(t3), fmt.Sprintf(expectedOptionsCount, 5))

	t4 := []string{option1, option1, option3, option4, option5, option5}
	RemoveDuplicates(&t4)
	assert.Equal(4, len(t4), fmt.Sprintf(expectedOptionsCount, 4))

	t5 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(option1, option3, &t5)
	assert.Equal(4, len(t5), fmt.Sprintf(expectedOptionsCount, 4))
	assert.Equal(LocalTag+option1, t5[0], expectedLocal)
	assert.Equal(RemoteTag+option3, t5[1], expectedRemote)
	assert.Equal(option4, t5[2], unexpectedOption)
	assert.Equal(option5, t5[3], unexpectedOption)

	t6 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(option1, option1, &t6)
	assert.Equal(4, len(t6), fmt.Sprintf(expectedOptionsCount, 4))
	assert.Equal(LocalTag+RemoteTag+option1, t6[0], expectedLocal+" & "+expectedRemote)
	assert.Equal(option3, t6[1], unexpectedOption)
	assert.Equal(option4, t6[2], unexpectedOption)
	assert.Equal(option5, t6[3], unexpectedOption)

	// + one tag is not found
	t7 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(option1, option1+"_", &t7)
	assert.Equal(4, len(t7), fmt.Sprintf(expectedOptionsCount, 4))
	assert.Equal(LocalTag+option1, t7[0], expectedLocal)
	assert.Equal(option3, t7[1], unexpectedOption)
	assert.Equal(option4, t7[2], unexpectedOption)
	assert.Equal(option5, t7[3], unexpectedOption)
}
