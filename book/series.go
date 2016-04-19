package book

import (
	"fmt"
	"strconv"
	"strings"
)

// SingleSeries holds the name and index of a series a Book is part of.
type SingleSeries struct {
	Name     string `json:"seriesname" xml:"series>title"`
	Position string `json:"seriesindex" xml:"user_position"`
}

// String outputs a single series info.
func (s *SingleSeries) String() string {
	return fmt.Sprintf("%s (#%s)", s.Name, s.Position)
}

// Series can track a series and an epub's position.
type Series []SingleSeries

// String outputs a series info.
func (s Series) String() (description string) {
	series := []string{}
	for _, ss := range s {
		series = append(series, ss.String())
	}
	return strings.Join(series, ", ")
}

// Add a series
func (s *Series) Add(seriesName string, position float32) (seriesModified bool) {
	hasSeries, seriesIndex, currentIndex := s.Has(seriesName)
	indexStr := strconv.FormatFloat(float64(position), 'f', -1, 32)
	// if not HasSeries, create new Series and add
	if !hasSeries {
		ss := SingleSeries{Name: seriesName, Position: indexStr}
		*s = append(*s, ss)
		seriesModified = true
	} else {
		// if hasSeries, if index is different, update index
		if currentIndex != indexStr {
			(*s)[seriesIndex].Position = indexStr
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
func (s *Series) Has(seriesName string) (hasSeries bool, index int, position string) {
	for i, series := range *s {
		if series.Name == seriesName {
			return true, i, series.Position
		}
	}
	return
}

// HasAny checks if epub is part of any series
func (s *Series) HasAny() (hasSeries bool) {
	return len(*s) != 0
}
