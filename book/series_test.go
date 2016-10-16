package book

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	shouldBeModifiedError    = "series should be modified"
	shouldNotBeModifiedError = "series should not have been modified"
	wrongFormatError         = "adding series with wrong format should have failed"
	expectedSeriesError      = "expected epub %s to have series %s"
	addingSeriesError        = "error adding Series %s - %f for epub %s"
	seriesAtIndex0Error      = "expected epub %s to have series %s at index 0"
)

// TestSeries tests Add, Remove, Has and HasAny
func TestSeries(t *testing.T) {
	fmt.Println("+ Testing Series...")
	var err error
	var seriesModified bool
	assert := assert.New(t)

	for i, epub := range epubs {
		e := NewBook(ui, i, epub.filename, standardTestConfig, isRetail)
		seriesName := "test_é!/?*èç1"
		seriesName2 := "test2"

		hasAny := e.Metadata.Series.HasAny()
		assert.False(hasAny, "Error: did not expect to have any series.")

		// testing adding series
		seriesModified = e.Metadata.Series.add(seriesName, float64(i))
		assert.True(seriesModified, fmt.Sprintf(addingSeriesError, seriesName, float32(i), e.FullPath()))

		// testing adding second series
		seriesModified = e.Metadata.Series.add(seriesName2, float64(i))
		assert.True(seriesModified, fmt.Sprintf(addingSeriesError, seriesName2, float32(i), e.FullPath()))

		hasAny = e.Metadata.Series.HasAny()
		assert.True(hasAny, "Error: expected to have at least one series.")
		expectedString := fmt.Sprintf("%s #%d, %s #%d", seriesName, i, seriesName2, i)
		assert.Equal(e.Metadata.Series.String(), expectedString, "Error printing series info")

		// testing having series
		hasSeries, index, seriesIndex := e.Metadata.Series.Has(seriesName)
		assert.True(hasSeries, fmt.Sprintf(expectedSeriesError, e.FullPath(), seriesName))
		assert.Equal(index, 0, fmt.Sprintf(seriesAtIndex0Error, e.FullPath(), seriesName))
		assert.Equal(seriesIndex, strconv.Itoa(i), "Error:  expected epub %s to have series %s, book %f and not %s", e.FullPath(), seriesName, float32(i), seriesIndex)

		hasSeries, index, seriesIndex = e.Metadata.Series.Has(seriesName2)
		assert.True(hasSeries, fmt.Sprintf(expectedSeriesError, e.FullPath(), seriesName2))
		assert.Equal(index, 1, fmt.Sprintf(seriesAtIndex0Error, e.FullPath(), seriesName2))
		assert.Equal(seriesIndex, strconv.Itoa(i), "Error:  expected epub %s to have series %s, book %f and not %s", e.FullPath(), seriesName2, float32(i), seriesIndex)

		hasSeries, _, _ = e.Metadata.Series.Has(seriesName + "ç")
		assert.False(hasSeries, "Error:  did not expect epub %s to have series %s", e.FullPath(), seriesName+"ç")

		// testing updating series index
		seriesModified = e.Metadata.Series.add(seriesName, float64(i)+0.5)
		assert.True(seriesModified, fmt.Sprintf(addingSeriesError, seriesName, float32(i)+0.5, e.FullPath()))

		// testing having modified series
		hasSeries, index, seriesIndex = e.Metadata.Series.Has(seriesName)
		assert.True(hasSeries, fmt.Sprintf(expectedSeriesError, e.FullPath(), seriesName))
		assert.Equal(0, index, fmt.Sprintf(seriesAtIndex0Error, e.FullPath(), seriesName))
		expected := fmt.Sprintf("%s,%s", strconv.FormatFloat(float64(i), 'f', -1, 32), strconv.FormatFloat(float64(i)+0.5, 'f', -1, 32))
		assert.Equal(expected, seriesIndex, "Error:  expected epub %s to have series %s, book %f and not %s", e.FullPath(), seriesName, float32(i), seriesIndex)

		// testing adding from string
		e.Metadata.Series = Series{}
		entries := []string{"test:1.5", "test:2.5", "test2:", "test3", "test4:7-9", "test4:1", "series: with a semicolon :7-9"}
		for _, s := range entries {
			seriesModified, err = e.Metadata.Series.AddFromString(s)
			assert.True(seriesModified, shouldBeModifiedError)
			assert.Nil(err)
		}
		// testing adding already known index
		seriesModified, err = e.Metadata.Series.AddFromString("test4:8")
		assert.False(seriesModified, shouldNotBeModifiedError)
		assert.Nil(err)

		// testing output
		assert.Equal("test #1.5,2.5, test2 #0, test3 #0, test4 #1,7,8,9, series: with a semicolon #7,8,9", e.Metadata.Series.String())
		assert.Equal("test:1.5,2.5, test2:0, test3:0, test4:1,7,8,9, series: with a semicolon:7,8,9", e.Metadata.Series.rawString())

		// testing having modified series
		hasSeries, _, seriesIndex = e.Metadata.Series.Has("series: with a semicolon")
		assert.True(hasSeries, "Error:  expected epub %s to have series %s", e.FullPath(), "series: with a semicolon")
		assert.Equal("7,8,9", seriesIndex, "Error:  expected epub %s to have series %s, book %f and not %s", e.FullPath(), seriesName, float32(i), seriesIndex)

		// testing wrong inputs
		wrongEntries := []string{"test5:aoqoj", "test5:1-2-3", "test5:1-a", "test5:a-a"}
		for _, w := range wrongEntries {
			seriesModified, err = e.Metadata.Series.AddFromString(w)
			assert.NotNil(err, wrongFormatError)
			assert.False(seriesModified, shouldNotBeModifiedError)
		}
	}
}
