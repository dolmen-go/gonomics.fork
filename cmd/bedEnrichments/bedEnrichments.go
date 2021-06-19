package main

import (
	"flag"
	"fmt"
	"github.com/vertgenlab/gonomics/bed"
	"github.com/vertgenlab/gonomics/exception"
	"github.com/vertgenlab/gonomics/fileio"
	"log"
)

func bedEnrichments(inFile string, secondFile string, noGapFile string, outFile string) {
	elementsOne := bed.ReadLite(inFile)
	elementsTwo := bed.ReadLite(secondFile)
	noGapRegions := bed.ReadLite(noGapFile)

	bed.SortByCoord(elementsOne)
	bed.SortByCoord(elementsTwo)
	bed.SortByCoord(noGapRegions)

	//preflight checks: check for error in user input. Beds should not be self-overlapping.
	if bed.IsSelfOverlapping(elementsOne) {
		log.Fatalf("Elements in bedEnrichments must not be self-overlapping. Self-overlap found in inFile1.bed.")
	}
	if bed.IsSelfOverlapping(elementsTwo) {
		log.Fatalf("Elements in bedEnrichments must not be self-overlapping. Self-overlap found in inFile2.bed.")
	}
	if bed.IsSelfOverlapping(noGapRegions) {
		log.Fatalf("Elements in bedEnrichments must not be self-overlapping. Self-overlap found in noGap.bed.")
	}

	overlapCount := bed.OverlapCount(elementsOne, elementsTwo)
	probs := bed.ElementOverlapProbabilities(elementsOne, elementsTwo, noGapRegions)
	summarySlice := bed.EnrichmentPValue(probs, len(elementsTwo), overlapCount)

	var err error
	out := fileio.EasyCreate(outFile)
	_, err = fmt.Fprintf(out, "#Filename1\tFilename2\tLenElements1\tLenElements2\tOverlapCount\tDebugCheck\tExpectedOverlap\tEnrichment\tpValue\n")
	exception.PanicOnErr(err)
	_, err = fmt.Fprintf(out, "%s\t%s\t%d\t%d\t%d\t%f\t%f\t%f\t%e\n", inFile, secondFile, len(elementsOne), len(elementsTwo), overlapCount, summarySlice[0], summarySlice[1], float64(overlapCount)/summarySlice[1], summarySlice[2])
	exception.PanicOnErr(err)
	err = out.Close()
	exception.PanicOnErr(err)
}

func usage() {
	fmt.Print(
		"bedEnrichments - Returns the p-value of enrichment for overlaps between the elements in two input bed files.\n" +
			"noGap.bed represents a bed of all regions in the search space of the genome.\n" +
			"out.txt is in the form of a tab-separated value file with a header line starting with '#'.\n" +
			"Usage:\n" +
			"bedEnrichments elements1.bed elements2.bed noGap.bed out.txt\n")
	flag.PrintDefaults()
}

func main() {
	var expectedNumArgs int = 4

	flag.Usage = usage
	//log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetFlags(0)
	flag.Parse()

	if len(flag.Args()) != expectedNumArgs {
		flag.Usage()
		log.Fatalf("Error: expecting %d arguments, but got %d\n",
			expectedNumArgs, len(flag.Args()))
	}
	inFile := flag.Arg(0)
	secondFile := flag.Arg(1)
	noGapFile := flag.Arg(2)
	outFile := flag.Arg(3)

	bedEnrichments(inFile, secondFile, noGapFile, outFile)
}
