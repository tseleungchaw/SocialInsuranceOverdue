package main

import (
	"strconv"
	"testing"
)

func TestGenOverdueDays(t *testing.T) {
	data := [][]string{[]string{"20090503", "20090503"},
		[]string{"20090503", "20090504"},
		[]string{"20090503", "20090520"},
		[]string{"20090503", "20090804"},
		[]string{"20090503", "20100502"},
		[]string{"20090503", "20110804"}}
	result := []float64{0.00, 0.00, 6.00, 81.00, 352.00, 811.00}
	for i, j := range data {
		if genOverdueDays(j[0], j[1]) == result[i] {
			t.Log("test genOverdueDays(\"" + j[0] + "\", \"" + j[1] + "\") == " + strconv.FormatFloat(result[i], 'f', -1, 64) + " passed")
		} else {
			t.Error("test genOverdueDays(\"" + j[0] + "\", \"" + j[1] + "\") == " + strconv.FormatFloat(result[i], 'f', -1, 64) + " failed, got " + strconv.FormatFloat(genOverdueDays(j[0], j[1]), 'f', -1, 64))
		}
	}
}
