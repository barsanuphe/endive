package book

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBookSet tests
func TestBookSet(t *testing.T) {
	fmt.Println("+ Testing Book.Set()...")
	assert := assert.New(t)
	e := NewBook(ui, 0, epubs[0].filename, standardTestConfig, isRetail)

	// set unknown field
	err := e.Set("ratingg", "a")
	assert.NotNil(err, invalidValue)

	// set Rating
	err = e.Set("rating", "a")
	assert.NotNil(err, invalidValue)
	err = e.Set("rating", "10")
	assert.NotNil(err, invalidValue)
	err = e.Set("rating", "-1")
	assert.NotNil(err, invalidValue)
	err = e.Set("rating", "3.5")
	assert.Nil(err, validValue)
	assert.Equal("3.5", e.Rating)

	// set Review
	err = e.Set("Review", "simple Review!")
	assert.Nil(err, validValue)
	assert.Equal("simple Review!", e.Review)

	// set Progress
	err = e.Set("progress", "a")
	assert.NotNil(err, invalidValue)
	for _, vp := range validProgress {
		err := e.Set("progress", vp)
		assert.Nil(err, validValue)
		assert.Equal(vp, e.Progress)
	}

	// set ReadDate
	err = e.Set("readdate", "a")
	assert.NotNil(err, invalidValue)
	err = e.Set("readdate", "2013-13-32")
	assert.NotNil(err, invalidValue)
	err = e.Set("readdate", "2013-12-15")
	assert.Nil(err, validValue)
	assert.Equal("2013-12-15", e.ReadDate)

	// set Metadata field
	err = e.Set("description", "something")
	assert.Nil(err, validValue)
	assert.Equal("something", e.Metadata.Description)
}
