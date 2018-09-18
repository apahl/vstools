package main

import (
	"github.com/apahl/vstools/internal/calculators"
)

// Precision used in LessF32
const Precision = 0.001

// ByScore implements the sort-by-score interface
type ByScore []Score

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].value < a[j].value }

// ByLE implements the sort-by-Ligand-Efficiency interface
type ByLE []Score

func (a ByLE) Len() int      { return len(a) }
func (a ByLE) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLE) Less(i, j int) bool {
	leA, errA := calculators.LigEffUI8(a[i].value, a[i].numHAtoms)
	leB, errB := calculators.LigEffUI8(a[j].value, a[j].numHAtoms)
	if errA != nil && errB != nil {
		return false
	}
	if errA != nil {
		return false
	}
	if errB != nil {
		return true
	}
	return leA < leB
}

// LessF32 compares two float32 values to a given precision
func LessF32(a, b float32) bool {
	if b-a > Precision {
		return true
	}
	return false
}
