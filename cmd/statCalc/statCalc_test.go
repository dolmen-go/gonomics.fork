package main

import (
	"github.com/vertgenlab/gonomics/exception"
	"github.com/vertgenlab/gonomics/fileio"
	"os"
	"testing"
)

var StatCalcTests = []struct {
	Args         []string
	Normal       string
	Binomial     string
	Poisson      string
	Beta         string
	Gamma        string
	SampleAfs    string
	SampleBeta   string
	SampleGamma  string
	SampleNormal string
	SetSeed      int64
	OutFile      string
	ExpectedFile string
}{
	{[]string{"1"},
		"0,1",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.normalDensity.txt",
		"testdata/expected.normalDensity.txt"},
	{[]string{"1", "inf"},
		"0,1",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.normalIntegral.txt",
		"testdata/expected.normalIntegral.txt"},
	{[]string{"3"},
		"",
		"10,0.5",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.binomialDist.txt",
		"testdata/expected.binomialDist.txt"},
	{[]string{"3", "n"},
		"",
		"10,0.5",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.binomialSum.txt",
		"testdata/expected.binomialSum.txt"},
	{[]string{"4"},
		"",
		"",
		"4",
		"",
		"",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.poissonDist.txt",
		"testdata/expected.poissonDist.txt"},
	{[]string{"4", "inf"},
		"",
		"",
		"4",
		"",
		"",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.poissonIntegral.txt",
		"testdata/expected.poissonIntegral.txt"},
	{[]string{"0.3"},
		"",
		"",
		"",
		"2,3",
		"",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.betaDist.txt",
		"testdata/expected.betaDist.txt"},
	{[]string{"0.3", "1"},
		"",
		"",
		"",
		"2,3",
		"",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.betaIntegral.txt",
		"testdata/expected.betaIntegral.txt"},
	{[]string{"3.5"},
		"",
		"",
		"",
		"",
		"4,2",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.gammaDist.txt",
		"testdata/expected.gammaDist.txt"},
	{[]string{"8", "inf"},
		"",
		"",
		"",
		"",
		"4,2",
		"",
		"",
		"",
		"",
		1,
		"testdata/tmp.gammaIntegral.txt",
		"testdata/expected.gammaIntegral.txt"},
	{[]string{},
		"",
		"",
		"",
		"",
		"",
		"0.02,10,1000,1000,0.001,0.999",
		"",
		"",
		"",
		1,
		"testdata/tmp.sampleAfs.txt",
		"testdata/expected.sampleAfs.txt"},
	{[]string{},
		"",
		"",
		"",
		"",
		"",
		"",
		"4,4,10",
		"",
		"",
		1,
		"testdata/tmp.sampleBeta.txt",
		"testdata/expected.sampleBeta.txt"},
	{[]string{},
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"4,4,10",
		"",
		1,
		"testdata/tmp.sampleGamma.txt",
		"testdata/expected.sampleGamma.txt"},
	{[]string{},
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"0,1,20",
		1,
		"testdata/tmp.sampleNormal.txt",
		"testdata/expected.sampleNormal.txt"},
}

func TestStatCalc(t *testing.T) {
	var err error
	var s Settings
	for _, v := range StatCalcTests {
		s = Settings{
			Args:         v.Args,
			Normal:       v.Normal,
			Binomial:     v.Binomial,
			Poisson:      v.Poisson,
			Beta:         v.Beta,
			Gamma:        v.Gamma,
			SampleAfs:    v.SampleAfs,
			SampleBeta:   v.SampleBeta,
			SampleGamma:  v.SampleGamma,
			SampleNormal: v.SampleNormal,
			SetSeed:      v.SetSeed,
			OutFile:      v.OutFile,
		}
		statCalc(s)
		if !fileio.AreEqual(v.OutFile, v.ExpectedFile) {
			t.Errorf("Error in statCalc. Output did not match expected.")
		} else {
			err = os.Remove(v.OutFile)
			exception.PanicOnErr(err)
		}
	}
}
