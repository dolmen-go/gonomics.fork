package main

import (
	"github.com/vertgenlab/gonomics/exception"
	"github.com/vertgenlab/gonomics/fileio"
	"math"
	"os"
	"testing"
)

var WigMathTests = []struct {
	InFile                 string
	OutFile                string
	BedMask                string
	MinValue               float64
	MaxValue               float64
	ExpectedFile           string
	ScalarMultiply         float64
	ScalarDivide           float64
	AddFile                string
	SubtractFile           string
	MovingAverageSmoothing int
	AbsoluteError          string
	AbsolutePercentError   string
	Missing                float64
	MissingBed             bool
	Pearson                string
	SamplingFrequency      float64
}{
	{
		InFile:                 "testdata/in.wig",
		OutFile:                "testdata/tmp.add.wig",
		ExpectedFile:           "testdata/expected.add.wig",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "testdata/second.wig",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.wig",
		OutFile:                "testdata/tmp.wig",
		ExpectedFile:           "testdata/expected.subtract.wig",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "testdata/second.wig",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/unsmooth.wig",
		OutFile:                "testdata/tmp.smooth.wig",
		ExpectedFile:           "testdata/expected.smooth.wig",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 5,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.wig",
		OutFile:                "testdata/tmp.absError.wig",
		ExpectedFile:           "testdata/expected.absError.wig",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "testdata/second.wig",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.wig",
		OutFile:                "testdata/tmp.absPercentError.wig",
		ExpectedFile:           "testdata/expected.absPercentError.wig",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "testdata/second.wig",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.wig",
		OutFile:                "testdata/tmp.Pearson.txt",
		ExpectedFile:           "testdata/expected.Pearson.txt",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "testdata/second.wig",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.wig",
		OutFile:                "testdata/tmp.Mult55.wig",
		ExpectedFile:           "testdata/expected.mult55.wig",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         55,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.wig",
		OutFile:                "testdata/tmp.divide4.wig",
		ExpectedFile:           "testdata/expected.divide4.wig",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           4,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.wig",
		OutFile:                "testdata/tmp.min25.wig",
		ExpectedFile:           "testdata/expected.min25.wig",
		BedMask:                "",
		MinValue:               25,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.wig",
		OutFile:                "testdata/tmp.max300.wig",
		ExpectedFile:           "testdata/expected.max300.wig",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               300,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.missing.wig",
		OutFile:                "testdata/tmp.missing.bed",
		ExpectedFile:           "testdata/expected.missing.bed",
		BedMask:                "",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             true,
		Pearson:                "",
		SamplingFrequency:      1,
	},
	{
		InFile:                 "testdata/in.missing.wig",
		OutFile:                "testdata/tmp.masked.wig",
		ExpectedFile:           "testdata/expected.masked.wig",
		BedMask:                "testdata/mask.bed",
		MinValue:               -1 * math.MaxFloat64,
		MaxValue:               math.MaxFloat64,
		ScalarMultiply:         1,
		ScalarDivide:           1,
		AddFile:                "",
		SubtractFile:           "",
		MovingAverageSmoothing: 1,
		AbsoluteError:          "",
		AbsolutePercentError:   "",
		Missing:                -10,
		MissingBed:             false,
		Pearson:                "",
		SamplingFrequency:      1,
	},
}

func TestWigMath(t *testing.T) {
	var err error
	var s Settings
	for _, v := range WigMathTests {
		s = Settings{
			InFile:                 v.InFile,
			OutFile:                v.OutFile,
			BedMask:                v.BedMask,
			MinValue:               v.MinValue,
			MaxValue:               v.MaxValue,
			ScalarMultiply:         v.ScalarMultiply,
			ScalarDivide:           v.ScalarDivide,
			ElementWiseAdd:         v.AddFile,
			ElementWiseSubtract:    v.SubtractFile,
			MovingAverageSmoothing: v.MovingAverageSmoothing,
			AbsoluteError:          v.AbsoluteError,
			AbsolutePercentError:   v.AbsolutePercentError,
			Missing:                v.Missing,
			MissingBed:             v.MissingBed,
			Pearson:                v.Pearson,
			SamplingFrequency:      v.SamplingFrequency,
		}
		wigMath(s)
		if !fileio.AreEqual(v.OutFile, v.ExpectedFile) {
			t.Errorf("Error in wigMath. Output was not as expected.")
		} else {
			err = os.Remove(v.OutFile)
			exception.PanicOnErr(err)
		}
	}
}
