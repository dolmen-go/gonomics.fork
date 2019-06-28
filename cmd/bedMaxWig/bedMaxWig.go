package main

import (
	"fmt"
	"log"
	"flag"
	"github.com/vertgenlab/gonomics/bed"
	"github.com/vertgenlab/gonomics/chromInfo"
	"github.com/vertgenlab/gonomics/wig"
	"github.com/vertgenlab/gonomics/common"
)

func bedMaxWig(infile string, database string, chromsize string, outfile string) {
	var records []*bed.Bed = bed.Read(infile)
	var data []*wig.Wig = wig.Read(database)
	var sizes []*chromInfo.ChromInfo = chromInfo.ReadToSlice(chromsize)
	var outlist []*bed.Bed
	var currentBed *bed.Bed = records[0]
	var chromSlice []float64
	var currentMax float64
	var currentStart int64
	var recordLength int64
	var i int64
	var m int64

	for i = 0; i < int64(len(sizes)); i++ {
		chromSlice = wig.WigToSlice(data, sizes[i].Size, sizes[i].Name)
		for k :=0; k < len(records); k++ {
			if records[k].Chrom == sizes[i].Name {
				currentBed = records[k]
				recordLength = records[k].ChromEnd - records[k].ChromStart
				if recordLength < 200 {
					currentBed.Annotation = append(currentBed.Annotation, fmt.Sprintf("%f", sliceRangeAverage(chromSlice, records[k].ChromStart, records[k].ChromEnd)))
					continue
				}

				currentMax = 0
				for m = 0; m < (recordLength-200); m++ {
					currentStart = records[k].ChromStart + int64(m)
					currentMax = common.MaxFloat64(currentMax, sliceRangeAverage(chromSlice, currentStart, currentStart + 200))
				}
				currentBed.Annotation = append(currentBed.Annotation, fmt.Sprintf("%f", currentMax))
				outlist = append(outlist, currentBed)
			}
		}
	}
	bed.Write(outfile, outlist, 7)
}

func sliceRangeAverage(w []float64, start int64, end int64) float64 {
	length := end - start
	var sum float64
	var i int64

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

	bedMaxWig(infile, database, chromsize, outfile)
}
