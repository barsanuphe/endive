package main

import (
	"fmt"
	"strconv"
	"testing"
)

var seriesTestEpubs = []string{"test/pg16328.epub", "test/pg17989.epub"}

// TestSeries tests Add, Remove, Has and HasAny
func TestSeries(t *testing.T) {
	fmt.Println("+ Testing Series...")
	for i, testEpub := range seriesTestEpubs {
		e := NewBook(testEpub, standardTestConfig, isRetail)
		seriesName := "test_é!/?*èç1"
		seriesName2 := "test2"

		hasAny := e.Series.HasAny()
		if hasAny {
			t.Errorf("Error: did not expect to have any series.")
		}

		// testing adding series
		seriesModified := e.Series.Add(seriesName, float32(i))
		if !seriesModified {
			t.Errorf("Error adding Series %s - %f for epub %s", seriesName, float32(i), e.getMainFilename())
		}
		// testing adding second series
		seriesModified = e.Series.Add(seriesName2, float32(i))
		if !seriesModified {
			t.Errorf("Error adding Series %s - %f for epub %s", seriesName2, float32(i), e.getMainFilename())
		}
		hasAny = e.Series.HasAny()
		if !hasAny {
			t.Errorf("Error: expected to have at least one series.")
		}

		// testing having series
		hasSeries, index, seriesIndex := e.Series.Has(seriesName)
		if !hasSeries {
			t.Errorf("Error:  expected epub %s to have series %s", e.getMainFilename(), seriesName)
		}
		if index != 0 {
			t.Errorf("Error:  expected epub %s to have series %s at index 0", e.getMainFilename(), seriesName)
		}
		if seriesIndex != strconv.Itoa(i) {
			t.Errorf("Error:  expected epub %s to have series %s, book %f and not %f", e.getMainFilename(), seriesName, float32(i), seriesIndex)
		}
		hasSeries, index, seriesIndex = e.Series.Has(seriesName2)
		if !hasSeries {
			t.Errorf("Error:  expected epub %s to have series %s", e.getMainFilename(), seriesName2)
		}
		if index != 1 {
			t.Errorf("Error:  expected epub %s to have series %s at index 1", e.getMainFilename(), seriesName2)
		}
		if seriesIndex != strconv.Itoa(i) {
			t.Errorf("Error:  expected epub %s to have series %s, book %f and not %f", e.getMainFilename(), seriesName2, float32(i), seriesIndex)
		}

		hasSeries, _, _ = e.Series.Has(seriesName + "ç")
		if hasSeries {
			t.Errorf("Error:  did not expect epub %s to have series %s", e.getMainFilename(), seriesName+"ç")
		}

		// testing updating series index
		seriesModified = e.Series.Add(seriesName, float32(i)+0.5)
		if !seriesModified {
			t.Errorf("Error adding Series %s - %f for epub %s", seriesName, float32(i)+0.5, e.getMainFilename())
		}
		// testing having modified series
		hasSeries, index, seriesIndex = e.Series.Has(seriesName)
		if !hasSeries {
			t.Errorf("Error:  expected epub %s to have series %s", e.getMainFilename(), seriesName)
		}
		if index != 0 {
			t.Errorf("Error:  expected epub %s to have series %s at index 0", e.getMainFilename(), seriesName)
		}
		if seriesIndex != strconv.FormatFloat(float64(i)+0.5, 'f', -1, 32) {
			t.Errorf("Error:  expected epub %s to have series %s, book %f and not %f", e.getMainFilename(), seriesName, float32(i)+0.5, seriesIndex)
		}

		// testing removing series
		seriesRemoved := e.Series.Remove(seriesName)
		if !seriesRemoved {
			t.Errorf("Error removing Series %s for epub %s", seriesName, e.getMainFilename())
		}
		hasSeries, _, _ = e.Series.Has(seriesName)
		if hasSeries {
			t.Errorf("Error: did not expect epub %s to have series %s", e.getMainFilename(), seriesName)
		}
	}
}
