package book

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	sampleString = "simple text!"
)

// TestBookSet tests
func TestBookSetGet(t *testing.T) {
	fmt.Println("+ Testing Book.Set()/Get()...")
	assert := assert.New(t)
	e := NewBook(ui, 0, epubs[0].filename, standardTestConfig, isRetail)

	// set unknown field
	err := e.Set("ratingg", "a")
	assert.NotNil(err, invalidValue)
	// get unknown field
	_, err = e.Get("ratingg")
	assert.NotNil(err, invalidFieldT)

	// set Rating
	err = e.Set(ratingField, "a")
	assert.NotNil(err, invalidValue)
	err = e.Set(ratingField, "10")
	assert.NotNil(err, invalidValue)
	err = e.Set(ratingField, "-1")
	assert.NotNil(err, invalidValue)
	err = e.Set(ratingField, "3.5")
	assert.Nil(err, validValue)
	assert.Equal("3.5", e.Rating)
	// get Rating
	value, err := e.Get(ratingField)
	assert.Nil(err, validField)
	assert.Equal("3.5", value)

	// set Review
	err = e.Set("Review", sampleString)
	assert.Nil(err, validValue)
	assert.Equal(sampleString, e.Review)
	// get Review
	value, err = e.Get("Review")
	assert.Nil(err, validField)
	assert.Equal(sampleString, value)

	// set Progress
	err = e.Set(progressField, "a")
	assert.NotNil(err, invalidValue)
	for _, vp := range validProgress {
		err := e.Set(progressField, vp)
		assert.Nil(err, validValue)
		assert.Equal(vp, e.Progress)
	}
	// get Progress
	value, err = e.Get(progressField)
	assert.Nil(err, validField)
	assert.Equal(validProgress[len(validProgress)-1], value)

	// set ReadDate
	err = e.Set(readDateField, "a")
	assert.NotNil(err, invalidValue)
	err = e.Set(readDateField, "2013-13-32")
	assert.NotNil(err, invalidValue)
	err = e.Set(readDateField, "2013-12-15")
	assert.Nil(err, validValue)
	assert.Equal("2013-12-15", e.ReadDate)
	// get readdate
	value, err = e.Get(readDateField)
	assert.Nil(err, validField)
	assert.Equal("2013-12-15", value)

	// set Metadata field
	err = e.Set(descriptionField, sampleString)
	assert.Nil(err, validValue)
	assert.Equal(sampleString, e.Metadata.Description)
	// get metadata field
	value, err = e.Get(descriptionField)
	assert.Nil(err, validField)
	assert.Equal(sampleString, value)
}
