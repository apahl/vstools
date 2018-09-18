package calculators

import (
	"errors"
)

// LEformula describes the current implementation of LE
const LEformula = "LE = 10 * score / (number of heavy atoms)"

// LigEffF32 calculates the ligand scoring efficiency
// LE = 10 * score / ((number of heavy atoms) * 2)
// Because the score is negative, LE is also negative,
// usually between -6 .. 0
// Takes numHA as float32
func LigEffF32(score, numHA float32) (float32, error) {
	var result float32
	if numHA < 1.0 {
		return 0.0, errors.New("number of heavy atoms must be at least 1")
	}
	result = 10 * score / numHA
	return result, nil
}

// LigEffUI8 calculates the ligand scoring efficiency
// Takes numHA as uint8
func LigEffUI8(score float32, numHA uint8) (float32, error) {
	return LigEffF32(score, float32(numHA))
}

// MaxUI8 returns the larger of two uint8 numbers
func MaxUI8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

// MinUI8 returns the smaller of two uint8 numbers
func MinUI8(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}
