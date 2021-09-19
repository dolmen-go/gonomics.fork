package align

import (
	//"math" //TODO: fix math.MinInt
)

var veryNegNum int = -9223372036854775808 / 2 //TODO: fix math.MinInt

// these are relative to the first seq.
// e.g. colI is an insertion in the second seq, relative to the first
type ColType uint8

const (
	ColM ColType = 0
	ColI ColType = 1
	ColD ColType = 2
)

type Cigar struct {
	RunLength int
	Op        ColType
}

// O=400 E=30
var DefaultScoreMatrix = [][]int{
	{91, -114, -31, -123, -44},
	{-114, 100, -125, -31, -43},
	{-31, -125, 100, -114, -43},
	{-123, -31, -114, 91, -44},
	{-44, -43, -43, -44, -43},
}

// O=400 E=30
var HoxD55ScoreMatrix = [][]int{
	{91, -114, -31, -123, 0},
	{-114, 100, -125, -31, 0},
	{-31, -125, 100, -114, 0},
	{-123, -31, -114, 91, 0},
	{0, 0, 0, 0, 0},
}

// O=600 E=55
var MouseRatScoreMatrix = [][]int{
	{91, -114, -31, -123, 0},
	{-114, 100, -125, -31, 0},
	{-31, -125, 100, -114, 0},
	{-123, -31, -114, 91, 0},
	{0, 0, 0, 0, 0},
}

// O=600 E=150
var HumanChimpTwoScoreMatrix = [][]int{
	{90, -330, -236, -356, -208},
	{-330, 100, -318, -236, -196},
	{-236, -318, 100, -330, -196},
	{-356, -236, -330, 90, -208},
	{-208, -196, -196, -208, -202},
}

/*
var StrictScoreMatrix = [][]int{
        {91, -114, -31, -123, -44},
        {-114, 100, -125, -31, -43},
        {-31, -125, 100, -114, -43},
        {-123, -31, -114, 91, -44},
        {-44, -43, -43, -44, -43},
}
*/

func tripleMaxTrace(a int, b int, c int) (int, ColType) {
	if a >= b && a >= c {
		return a, ColM
	} else if b >= c {
		return b, ColI
	} else {
		return c, ColD
	}
}

func reverseCigar(alpha []Cigar) {
	for i, j := 0, len(alpha)-1; i < j; i, j = i+1, j-1 {
		alpha[i], alpha[j] = alpha[j], alpha[i]
	}
}

func countAlignmentColumns(route []Cigar) int {
	var count int = 0
	for i := range route {
		count += route[i].RunLength
	}
	return count
}
