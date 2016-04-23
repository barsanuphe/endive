package book

import (
	"fmt"
	"testing"
)

// TestTags tests Add, Remove, Has and HasAny
func TestTags(t *testing.T) {
	fmt.Println("+ Testing Tags...")
	tag1 := Tag{Name: "test_é!/?*èç1"}
	tag2 := Tag{Name: "t54*ßèç©@1ϽƉ"}
	tag3 := Tag{Name: "forthcoming"}
	tag4 := Tag{Name: "sci-fi"}
	tag5 := Tag{Name: "science-fiction"}

	e := NewBook(0, epubs[0].filename, standardTestConfig, isRetail)

	// check empty tags
	isIn, _ := e.Metadata.Tags.Has(tag1)
	if isIn {
		t.Errorf("Error: did not expect to have tag1.")
	}
	// check adding 2 tags
	added := e.Metadata.Tags.Add(tag1, tag2)
	if !added {
		t.Errorf("Error: book should have tags now.")
	}
	isIn, _ = e.Metadata.Tags.Has(tag1)
	if !isIn {
		t.Errorf("Error: expected book to have tag1.")
	}
	isIn, _ = e.Metadata.Tags.Has(tag2)
	if !isIn {
		t.Errorf("Error: expected book to have tag2.")
	}
	// check string()
	if e.Metadata.Tags.String() != tag1.Name+", "+tag2.Name {
		t.Errorf("Error: Tags.String() returned %s, expected %s.", e.Metadata.Tags.String(), tag1.Name+" "+tag2.Name)
	}
	// adding more tags
	added = e.Metadata.Tags.Add(tag3, tag4)
	if !added {
		t.Errorf("Error: book should have tags now.")
	}
	// check Add only new tags
	added = e.Metadata.Tags.AddFromNames(tag3.Name)
	if added {
		t.Errorf("Error: book should not have added already known tag.")
	}
	// check clean
	isIn, _ = e.Metadata.Tags.Has(tag3)
	if !isIn {
		t.Errorf("Error: expected to have tag3.")
	}
	e.Metadata.Tags.Clean()
	isIn, _ = e.Metadata.Tags.Has(tag3)
	if isIn {
		t.Errorf("Error: expected tag3 to have been cleaned.")
	}
	isIn, _ = e.Metadata.Tags.Has(tag4)
	if isIn {
		t.Errorf("Error: expected tag5 to have been replaced by alias.")
	}
	isIn, _ = e.Metadata.Tags.Has(tag5)
	if !isIn {
		t.Errorf("Error: expected tag5 to have replaced its alias tag4.")
	}
	// test remove
	removed := e.Metadata.Tags.RemoveFromNames(tag5.Name, tag1.Name)
	if !removed {
		t.Errorf("Error: expecteds to be removed.")
	}
	isIn, _ = e.Metadata.Tags.Has(tag5)
	if isIn {
		t.Errorf("Error: expected tag5 to have been removed.")
	}

}
