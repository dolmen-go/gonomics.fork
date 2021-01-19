package main

import (
	"flag"
	"fmt"
	"github.com/vertgenlab/gonomics/bed"
	"github.com/vertgenlab/gonomics/chromInfo"
	"github.com/vertgenlab/gonomics/numbers"
	"github.com/vertgenlab/gonomics/wig"
	"log"
)

func bedMaxWig(infile string, database string, chromsizeFile string, outfile string, windowSize int) {
	var records []*bed.Bed = bed.Read(infile)
	var data []*wig.Wig = wig.Read(database)
	var sizes []*chromInfo.ChromInfo = chromInfo.ReadToSlice(chromsizeFile)
	var outlist []*bed.Bed
	var currentBed *bed.Bed = records[0]
	var chromSlice []float64
	var currentMax float64
	var currentStart, recordLength int
	var i, m int

	for i = 0; i < len(sizes); i++ {
		chromSlice = WigChromToSlice(data, sizes[i].Size, sizes[i].Name)
		for k := 0; k < len(records); k++ {
			if records[k].Chrom == sizes[i].Name {
				currentBed = records[k]
				recordLength = int(records[k].ChromEnd - records[k].ChromStart)
				if recordLength <= windowSize {
					currentBed.Annotation = append(currentBed.Annotation, fmt.Sprintf("%f", sliceRangeAverage(chromSlice, int(records[k].ChromStart), int(records[k].ChromEnd))))
				} else {
					currentMax = 0
					for m = 0; m < (recordLength - windowSize + 1); m++ {
						currentStart = int(records[k].ChromStart) + m
						currentMax = numbers.MaxFloat64(currentMax, sliceRangeAverage(chromSlice, currentStart, currentStart+windowSize))
					}
					currentBed.Annotation = append(currentBed.Annotation, fmt.Sprintf("%f", currentMax))

				}
				outlist = append(outlist, currentBed)
			}
		}
	}
	bed.Write(outfile, outlist, 7)
}

func WigChromToSlice(w []*wig.Wig, size int, chrom string) []float64 {
	output := make([]float64, size)
	for _, v := range w {
		if v.Chrom == chrom {
			for i := 0; i < len(output); i++ {
				output[i] = v.Values[i].Value
			}
		}
	}

	return output
}

func sliceRangeAverage(w []float64, start int, end int) float64 {
	length := end - start
	var sum float64
	var i int

	for i = 0; i < length; i++ {
		sum = sum + w[i+start]
	}
	return (sum / float64(length))
}

func usage() {
	fmt.Print(
		"bedMaxWig - Returns annotated bed with max wig score in bed entry range.\n" +
			"Usage:\n" +
			"bedMaxWig input.bed database.wig chrom.sizes output.bed\n" +
			"options:\n")
	flag.PrintDefaults()
}

func main() {
	var expectedNumArgs int = 4
	var windowSize *int = flag.Int("windowSize", 200, "Specify the window size")
	flag.Usage = usage
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	if len(flag.Args()) != expectedNumArgs {
		flag.Usage()
		log.Fatalf("Error: expecting %d arguments, but got %d\n",
			expectedNumArgs, len(flag.Args()))
	}

	infile := flag.Arg(0)
	database := flag.Arg(1)
	chromsize := flag.Arg(2)
	outfile := flag.Arg(3)

	bedMaxWig(infile, database, chromsize, outfile, *windowSize)
}
