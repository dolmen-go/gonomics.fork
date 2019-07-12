package main

import (
	"flag"
	"fmt"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fasta"
	"log"
	"strconv"
	"unicode/utf8"
)

func multFaVisualizer(infile string, start int64, end int64) {
	if !(start < end) {
		log.Fatalf("Invalid arguments, start must be lower than end")
	}

	var stop int
	records := fasta.Read(infile)

	for i := 1; i < len(records); i++ {
		for j := 0; j < len(records[0].Seq); j++ {
			if records[i].Seq[j] == records[0].Seq[j] {
				records[i].Seq[j] = dna.Dot
			}
		}
	}
	long := calculateLongestName(records)

	var refCounter int64 = 0
	var startCounter int64 = 0
	var endCounter int64 = 0

	for t := 0; refCounter < start; t++ {
		startCounter++
		if t == len(records[0].Seq) {
			log.Fatalf("Ran out of chromosome")
		} else if records[0].Seq[t] != dna.Gap {
			refCounter++
		}
	}

	fmt.Printf("Start: %d. refCounter: %d. alignCounter: %d\n", start, refCounter, startCounter)

	refCounter = 0
	for n := 0; refCounter < end; n++ {
		endCounter++
		if n == len(records[0].Seq) {
			log.Fatalf("Ran off the chromosome")
		} else if records[0].Seq[n] != dna.Gap {
			refCounter++
		}
	}

	for k := startCounter; k < endCounter; k = k + 100 {
		fmt.Printf("Position: %d\n", k)
		for m := 0; m < len(records); m++ {
			stop = common.Min(len(records[0].Seq), int(k+100))
			fmt.Printf("|%-*s| %s\n", long, records[m].Name, dna.BasesToString(records[m].Seq[k:stop]))
		}
		fmt.Printf("\n\n")
	}

}

func calculateLongestName(f []*fasta.Fasta) int {
	var ans int = 0
	var temp int
	for i := 0; i < len(f); i++ {
		temp = utf8.RuneCountInString(f[i].Name)
		if temp > ans {
			ans = temp
		}
	}
	return ans
}

func usage() {
	fmt.Print(
		"multFaVisualizer - Provides human-readable multiple alignment from a given .\n" +
			"Usage:\n" +
			"bedFilter mult.fa start end\n" +
			"options:\n")
	flag.PrintDefaults()
}

func main() {
	var expectedNumArgs int = 3
	flag.Usage = usage
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	if len(flag.Args()) != expectedNumArgs {
		flag.Usage()
		log.Fatalf("Error: expecting %d arguments, but got %d\n",
			expectedNumArgs, len(flag.Args()))
	}

	infile := flag.Arg(0)
	start, _ := strconv.ParseInt(flag.Arg(1), 10, 64)
	end, _ := strconv.ParseInt(flag.Arg(2), 10, 64)

	multFaVisualizer(infile, start, end)
}