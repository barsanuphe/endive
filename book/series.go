package book

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// SingleSeries holds the name and index of a series a Book is part of.
type SingleSeries struct {
	Name     string `json:"name" xml:"series>title"`
	Position string `json:"index" xml:"user_position"`
}

// String outputs a single series info.
func (s SingleSeries) String() string {
	return fmt.Sprintf("%s #%s", s.Name, s.Position)
}

// addIndex to SingleSeries Position
func (s *SingleSeries) addIndex(position float64) error {
	// convert Position to []float
	indexes := strings.Split(s.Position, ",")
	floatIndexes := []float64{}
	for _, v := range indexes {
		index, e := strconv.ParseFloat(v, 64)
		if e != nil {
			return errors.New("Invalid series index: " + v)
		}
		floatIndexes = append(floatIndexes, index)
	}

	modified := false
	// insert position
	sort.Float64s(floatIndexes)
	i := sort.SearchFloat64s(floatIndexes, position)
	if i == len(floatIndexes) {
		floatIndexes = append(floatIndexes, position)
		modified = true
	} else {
		// insert if not already in slice
		if floatIndexes[i] != position {
			// i is the index where position needs to be inserted
			floatIndexes = append(floatIndexes[:i], append([]float64{position}, floatIndexes[i:]...)...)
			modified = true
		}
	}
	// make into string again.
	s.Position = strconv.FormatFloat(floatIndexes[0], 'f', -1, 64)
	for _, v := range floatIndexes[1:] {
		s.Position += "," + strconv.FormatFloat(v, 'f', -1, 64)
	}
	if !modified {
		return errors.New("Series not modified")
	}
	return nil
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
				index1, e := strconv.ParseFloat(indexes[0], 64)
				if e != nil {
					return false, wrongFormatError
				}
				index2, e := strconv.ParseFloat(indexes[1], 64)
				if e != nil {
					return false, wrongFormatError
				}
				for i := index1; i <= index2; i++ {
					// if at least one index is added, series is modified
					if s.add(seriesName, i) {
						seriesModified = true
					}
				}
			} else {
				// case "series:float"
				index, e := strconv.ParseFloat(seriesIndex, 64)
				if e != nil {
					err = wrongFormatError
				} else {
					seriesModified = s.add(seriesName, index)
				}
			}
		}
	}
	return
}

// add a series with a float index
func (s *Series) add(seriesName string, position float64) (seriesModified bool) {
	hasSeries, seriesIndex, _ := s.Has(seriesName)
	indexStr := strconv.FormatFloat(position, 'f', -1, 64)
	// if not HasSeries, create new Series and add
	if !hasSeries {
		ss := SingleSeries{Name: seriesName, Position: indexStr}
		*s = append(*s, ss)
		seriesModified = true
	} else {
		err := (*s)[seriesIndex].addIndex(position)
		if err == nil {
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
