package book

import (
	"errors"
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
func (s SingleSeries) String() string {
	return fmt.Sprintf("%s #%s", s.Name, s.Position)
}

// Series can track a series and an epub's position.
type Series []SingleSeries

// String outputs a series info.
func (s Series) String() string {
	series := []string{}
	for _, ss := range s {
		series = append(series, ss.String())
	}
	return strings.Join(series, ", ")
}

// rawString outputs a series info in raw form: series:index.
func (s Series) rawString() string {
	series := []string{}
	for _, ss := range s {
		series = append(series, fmt.Sprintf("%s:%s", ss.Name, ss.Position))
	}
	return strings.Join(series, ", ")
}

// AddFromString a series, checking for correct form.
func (s *Series) AddFromString(candidate string) (seriesModified bool, err error) {
	wrongFormatError := errors.New("Series index must be empty, a float, or a range, got: " + candidate)

	candidate = strings.TrimSpace(candidate)
	lastSemiColonIndex := strings.LastIndex(candidate, ":")
	if lastSemiColonIndex == -1 {
		// case "series"
		s.add(strings.TrimSpace(candidate), 0)
		seriesModified = true
	} else {
		seriesName := strings.TrimSpace(candidate[:lastSemiColonIndex])
		if lastSemiColonIndex == len(candidate)-1 {
			// case "series:"
			s.add(seriesName, 0)
			seriesModified = true
		} else {
			seriesIndex := strings.TrimSpace(candidate[lastSemiColonIndex+1:])
			// case "series:index1-index2
			if strings.Contains(seriesIndex, "-") {
				indexes := strings.Split(seriesIndex, "-")
				if len(indexes) != 2 {
					return false, wrongFormatError
				}
				// parse as float both indexes
				index1, e := strconv.ParseFloat(indexes[0], 32)
				if e != nil {
					return false, wrongFormatError
				}
				index2, e := strconv.ParseFloat(indexes[1], 32)
				if e != nil {
					return false, wrongFormatError
				}
				for i := index1; i <= index2; i++ {
					s.add(seriesName, float32(i))
				}
				seriesModified = true
			} else {
				// case "series:float"
				index, e := strconv.ParseFloat(seriesIndex, 32)
				if e != nil {
					err = wrongFormatError
				} else {
					s.add(seriesName, float32(index))
					seriesModified = true
				}
			}
		}
	}
	return
}

// add a series with a float index
func (s *Series) add(seriesName string, position float32) (seriesModified bool) {
	hasSeries, seriesIndex, currentIndex := s.Has(seriesName)
	indexStr := strconv.FormatFloat(float64(position), 'f', -1, 32)
	// if not HasSeries, create new Series and add
	if !hasSeries {
		ss := SingleSeries{Name: seriesName, Position: indexStr}
		*s = append(*s, ss)
		seriesModified = true
	} else {
		// if hasSeries, if index is different, update index
		// TODO will not work is already contains several indexes
		if currentIndex != indexStr {
			// TODO order indexes
			(*s)[seriesIndex].Position += "," + indexStr
			seriesModified = true
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
func (s *Series) HasAny() bool {
	return len(*s) != 0
}
