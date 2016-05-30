package book

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTags tests Add, Remove, Has and HasAny
func TestTags(t *testing.T) {
	fmt.Println("+ Testing Tags...")
	assert := assert.New(t)
	tag1 := Tag{Name: "test_é!/?*èç1"}
	tag2 := Tag{Name: "t54*ßèç©@1ϽƉ"}
	tag3 := Tag{Name: "forthcoming"}
	tag4 := Tag{Name: "sci-fi"}
	tag5 := Tag{Name: "science-fiction"}
	e := NewBook(0, epubs[0].filename, standardTestConfig, isRetail)

	// check empty tags
	isIn, _ := e.Metadata.Tags.Has(tag1)
	assert.False(isIn, "Error: did not expect to have tag1.")
	// check adding 2 tags
	added := e.Metadata.Tags.Add(tag1, tag2)
	assert.True(added, "Error: book should have tags now.")
	isIn, _ = e.Metadata.Tags.Has(tag1)
	assert.True(isIn, "Error: expected to have tag1.")
	isIn, _ = e.Metadata.Tags.Has(tag2)
	assert.True(isIn, "Error: expected to have tag2.")
	// check string()
	assert.Equal(tag1.Name+", "+tag2.Name, e.Metadata.Tags.String(), "Error generating String()")
	// adding more tags
	added = e.Metadata.Tags.Add(tag3, tag4)
	assert.True(added, "Error: book should have more tags now.")
	// check Add only new tags
	added = e.Metadata.Tags.AddFromNames(tag3.Name)
	assert.False(added, "Error: book should not have added already known tag.")
	// check clean
	isIn, _ = e.Metadata.Tags.Has(tag3)
	assert.True(isIn, "Error: expected to have tag3.")
	e.Metadata.Clean(standardTestConfig)
	isIn, _ = e.Metadata.Tags.Has(tag3)
	assert.False(isIn, "Error: expected tag3 to have been cleaned.")
	isIn, _ = e.Metadata.Tags.Has(tag4)
	assert.False(isIn, "Error: expected tag4 to have been replaced by alias.")
	isIn, _ = e.Metadata.Tags.Has(tag5)
	assert.True(isIn, "Error: expected tag5 to have replaced its alias tag4.")
	// test remove
	removed := e.Metadata.Tags.RemoveFromNames(tag5.Name, tag1.Name)
	assert.True(removed, "Error: expecteds to be removed.")
	isIn, _ = e.Metadata.Tags.Has(tag5)
	assert.False(isIn, "Error: expected tag5 to have been removed.")
}
