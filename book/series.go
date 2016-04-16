package book

import (
	"fmt"
	"strconv"
)

// SingleSeries holds the name and index of a series a Book is part of.
type SingleSeries struct {
	Name  string `json:"seriesname"`
	Index string `json:"seriesindex"`
}

// String outputs a single series info.
func (s *SingleSeries) String() string {
	return fmt.Sprintf("%s (#%s)\n", s.Name, s.Index)
}

// Series can track a series and an epub's position.
type Series []SingleSeries

// String outputs a series info.
func (s Series) String() (description string) {
	for _, ss := range s {
		description += ss.String()
	}
	return
}

// Add a series
func (s *Series) Add(seriesName string, index float32) (seriesModified bool) {
	hasSeries, seriesIndex, currentIndex := s.Has(seriesName)
	indexStr := strconv.FormatFloat(float64(index), 'f', -1, 32)
	// if not HasSeries, create new Series and add
	if !hasSeries {
		ss := SingleSeries{Name: seriesName, Index: indexStr}
		*s = append(*s, ss)
		seriesModified = true
	} else {
		// if hasSeries, if index is different, update index
		if currentIndex != indexStr {
			(*s)[seriesIndex].Index = indexStr
			seriesModified = true
		}
	}
	return
}

// Remove series from the list
func (s *Series) Remove(seriesName ...string) (seriesRemoved bool) {
	for _, series := range seriesName {
		hasSeries, seriesIndex, _ := s.Has(series)
		if hasSeries {
			(*s)[seriesIndex] = (*s)[len(*s)-1]
			(*s) = (*s)[:len(*s)-1]
			seriesRemoved = true
		}
	}
	return
}

// Has checks if epub is part of a series
func (s *Series) Has(seriesName string) (hasSeries bool, index int, seriesIndex string) {
	for i, series := range *s {
		if series.Name == seriesName {
			return true, i, series.Index
		}
	}
	return
}

// HasAny checks if epub is part of any series
func (s *Series) HasAny() (hasSeries bool) {
	return len(*s) != 0
}
