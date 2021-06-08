package main

import (
	"github.com/vertgenlab/gonomics/bed"
	"github.com/vertgenlab/gonomics/common"
	"os"
	"testing"
)

var bedMaxWigTests = []struct {
	inputBed        string
	inputWig        string
	inputChromSizes string
	outputFile      string
	expectedFile    string
}{
	{"testdata/testBed.bed", "testdata/startOneStepOne.wig", "testdata/fake.chrom.sizes", "testdata/testBMWOutput.bed", "testdata/testBMWExpected.bed"},
	//{"testdata/testBed.bed", "testdata/startNonOneStepNonOne.wig", "testdata/fake.chrom.sizes", "testdata/testBMWOutput.bed", "testdata/testBMWExpectedNonOneStartStep.bed"},
	//TODO: If bedMaxWig is written to handle a step other than 1, come back to ensure there is a test for this.
}

func TestBedMaxWig(t *testing.T) {
	var err error
	for _, v := range bedMaxWigTests {
		bedMaxWig(v.inputBed, v.inputWig, v.inputChromSizes, v.outputFile)
		records := bed.Read(v.outputFile)
		expected := bed.Read(v.expectedFile)
		if !bed.AllAreEqual(records, expected) {
			t.Errorf("Error in bedMaxWig, the output beds is not as expected")
		}
		err = os.Remove(v.outputFile)
		if err != nil {
			common.ExitIfError(err)
		}
	}
}
