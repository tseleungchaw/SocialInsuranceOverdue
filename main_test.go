package main

import (
	"strconv"
	"testing"
	"time"
)

func TestGenOverdueDays(t *testing.T) {
	data := [][]time.Time{[]time.Time{time.Date(2009, time.Month(5), 3, 0, 0, 0, 0, time.UTC), time.Date(2009, time.Month(5), 3, 24, 0, 0, 0, time.UTC)},
		[]time.Time{time.Date(2009, time.Month(5), 3, 0, 0, 0, 0, time.UTC), time.Date(2009, time.Month(5), 4, 24, 0, 0, 0, time.UTC)},
		[]time.Time{time.Date(2009, time.Month(5), 3, 0, 0, 0, 0, time.UTC), time.Date(2009, time.Month(5), 20, 24, 0, 0, 0, time.UTC)},
		[]time.Time{time.Date(2009, time.Month(5), 3, 0, 0, 0, 0, time.UTC), time.Date(2009, time.Month(8), 4, 24, 0, 0, 0, time.UTC)},
		[]time.Time{time.Date(2009, time.Month(5), 3, 0, 0, 0, 0, time.UTC), time.Date(2010, time.Month(5), 2, 24, 0, 0, 0, time.UTC)},
		[]time.Time{time.Date(2009, time.Month(5), 3, 0, 0, 0, 0, time.UTC), time.Date(2011, time.Month(8), 4, 24, 0, 0, 0, time.UTC)}}
	result := []float64{0.00, 0.00, 6.00, 81.00, 352.00, 811.00}
	for i, j := range data {
		if genOverdueDays(j[0], j[1]) == result[i] {
			t.Log("test genOverdueDays(\"" + j[0].Format("20060102") + "\", \"" + j[1].Format("20060102") + "\") == " + strconv.FormatFloat(result[i], 'f', -1, 64) + " passed")
		} else {
			t.Error("test genOverdueDays(\"" + j[0].Format("20060102") + "\", \"" + j[1].Format("20060102") + "\") == " + strconv.FormatFloat(result[i], 'f', -1, 64) + " failed, got " + strconv.FormatFloat(genOverdueDays(j[0], j[1]), 'f', -1, 64))
		}
	}
}
