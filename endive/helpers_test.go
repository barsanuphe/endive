package endive

import (
	"errors"
	"fmt"
	"testing"

	h "github.com/barsanuphe/helpers"
	"github.com/stretchr/testify/assert"
)

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
)

func TestOptions(t *testing.T) {
	fmt.Println("+ Testing CleanSliceAndTagEntries/RemoveDuplicates()...")
	assert := assert.New(t)

	t1 := []string{option1, option1}
	h.RemoveDuplicates(&t1)
	assert.Equal(1, len(t1), fmt.Sprintf(expectedOptionsCount, 1))
	assert.Equal(option1, t1[0], unexpectedOption)

	t2 := []string{option1, option2, option3, option4, option5}
	h.RemoveDuplicates(&t2)
	assert.Equal(5, len(t2), fmt.Sprintf(expectedOptionsCount, 5))

	t3 := []string{option1, option2, option3, option4, option5, option5}
	h.RemoveDuplicates(&t3)
	assert.Equal(5, len(t3), fmt.Sprintf(expectedOptionsCount, 5))

	t4 := []string{option1, option1, option3, option4, option5, option5}
	h.RemoveDuplicates(&t4)
	assert.Equal(4, len(t4), fmt.Sprintf(expectedOptionsCount, 4))
}
