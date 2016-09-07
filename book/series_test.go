package book

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSeries tests Add, Remove, Has and HasAny
func TestSeries(t *testing.T) {
	fmt.Println("+ Testing Series...")
	assert := assert.New(t)
	for i, epub := range epubs {
		e := NewBook(i, epub.filename, standardTestConfig, isRetail)
		seriesName := "test_é!/?*èç1"
		seriesName2 := "test2"

		hasAny := e.Metadata.Series.HasAny()
		assert.False(hasAny, "Error: did not expect to have any series.")

		// testing adding series
		seriesModified := e.Metadata.Series.Add(seriesName, float32(i))
		assert.True(seriesModified, "Error adding Series %s - %f for epub %s", seriesName, float32(i), e.FullPath())

		// testing adding second series
		seriesModified = e.Metadata.Series.Add(seriesName2, float32(i))
		assert.True(seriesModified, "Error adding Series %s - %f for epub %s", seriesName2, float32(i), e.FullPath())

		hasAny = e.Metadata.Series.HasAny()
		assert.True(hasAny, "Error: expected to have at least one series.")
		expectedString := fmt.Sprintf("%s #%d, %s #%d", seriesName, i, seriesName2, i)
		assert.Equal(e.Metadata.Series.String(), expectedString, "Error printing series info")

		// testing having series
		hasSeries, index, seriesIndex := e.Metadata.Series.Has(seriesName)
		assert.True(hasSeries, "Error:  expected epub %s to have series %s", e.FullPath(), seriesName)
		assert.Equal(index, 0, "Error:  expected epub %s to have series %s at index 0", e.FullPath(), seriesName)
		assert.Equal(seriesIndex, strconv.Itoa(i), "Error:  expected epub %s to have series %s, book %f and not %s", e.FullPath(), seriesName, float32(i), seriesIndex)

		hasSeries, index, seriesIndex = e.Metadata.Series.Has(seriesName2)
		assert.True(hasSeries, "Error:  expected epub %s to have series %s", e.FullPath(), seriesName2)
		assert.Equal(index, 1, "Error:  expected epub %s to have series %s at index 0", e.FullPath(), seriesName2)
		assert.Equal(seriesIndex, strconv.Itoa(i), "Error:  expected epub %s to have series %s, book %f and not %s", e.FullPath(), seriesName2, float32(i), seriesIndex)

		hasSeries, _, _ = e.Metadata.Series.Has(seriesName + "ç")
		assert.False(hasSeries, "Error:  did not expect epub %s to have series %s", e.FullPath(), seriesName+"ç")

		// testing updating series index
		seriesModified = e.Metadata.Series.Add(seriesName, float32(i)+0.5)
		assert.True(seriesModified, "Error adding Series %s - %f for epub %s", seriesName, float32(i)+0.5, e.FullPath())

		// testing having modified series
		hasSeries, index, seriesIndex = e.Metadata.Series.Has(seriesName)
		assert.True(hasSeries, "Error:  expected epub %s to have series %s", e.FullPath(), seriesName)
		assert.Equal(0, index, "Error:  expected epub %s to have series %s at index 0", e.FullPath(), seriesName)
		expected := fmt.Sprintf("%s,%s", strconv.FormatFloat(float64(i), 'f', -1, 32), strconv.FormatFloat(float64(i)+0.5, 'f', -1, 32))
		assert.Equal(expected, seriesIndex, "Error:  expected epub %s to have series %s, book %f and not %s", e.FullPath(), seriesName, float32(i), seriesIndex)

		// testing removing series
		seriesRemoved := e.Metadata.Series.Remove(seriesName)
		assert.True(seriesRemoved, "Error removing Series %s for epub %s", seriesName, e.FullPath())
		hasSeries, _, _ = e.Metadata.Series.Has(seriesName)
		assert.False(hasSeries, "Error:  did not expect epub %s to have series %s", e.FullPath(), seriesName)

		// testing adding from string
		e.Metadata.Series = Series{}
		seriesModified, err := e.Metadata.Series.AddFromString("test:1.5")
		assert.True(seriesModified, "Error: series should be modified")
		assert.Nil(err)

		seriesModified, err = e.Metadata.Series.AddFromString("test:2.5")
		assert.True(seriesModified, "Error: series should be modified")
		assert.Nil(err)

		seriesModified, err = e.Metadata.Series.AddFromString("test2:")
		assert.True(seriesModified, "Error: series should be modified")
		assert.Nil(err)

		seriesModified, err = e.Metadata.Series.AddFromString("test3")
		assert.True(seriesModified, "Error: series should be modified")
		assert.Nil(err)

		seriesModified, err = e.Metadata.Series.AddFromString("test4:7-9")
		assert.True(seriesModified, "Error: series should be modified")
		assert.Nil(err)

		assert.Equal("test #1.5,2.5, test2 #0, test3 #0, test4 #7,8,9", e.Metadata.Series.String())
	}
}
