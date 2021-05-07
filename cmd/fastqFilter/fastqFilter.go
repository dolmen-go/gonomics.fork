package main

import (
	"flag"
	"fmt"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/fastq"
	"github.com/vertgenlab/gonomics/fileio"
	"github.com/vertgenlab/gonomics/numbers"
	"log"
	"math/rand"
	"sync"
)

type Settings struct {
	InFile    string
	OutFile   string
	R1InFile  string
	R2InFile  string
	R1OutFile string
	R2OutFile string
	PairedEnd bool
	SubSet    float64
	RandSeed  bool
	SetSeed   int64
	MinSize   int
	MaxSize   int
}

func fastqFilter(s Settings) {
	common.RngSeed(s.RandSeed, s.SetSeed)
	var r float64

	if s.PairedEnd {
		ReadCh := make(chan *fastq.PairedEnd, 100000)
		WriteCh := make(chan *fastq.PairedEnd, 100000)
		var wg sync.WaitGroup
		wg.Add(1)
		go fastq.PairEndToChan(s.R1InFile, s.R2InFile, ReadCh)
		go fastq.WritingChan(s.R1OutFile, s.R2OutFile, WriteCh, &wg)
		for i := range ReadCh {
			r = rand.Float64()
			if r > s.SubSet {
				continue
			}
			if len(i.Fwd.Seq) < s.MinSize {
				continue
			}
			if len(i.Rev.Seq) < s.MinSize {
				continue
			}
			if len(i.Fwd.Seq) > s.MaxSize {
				continue
			}
			if len(i.Rev.Seq) > s.MaxSize {
				continue
			}
			WriteCh <- i
		}
		close(WriteCh)
		wg.Wait()
	} else {
		f := fastq.GoReadToChan(s.InFile)
		out := fileio.EasyCreate(s.OutFile)
		defer out.Close()

		for i := range f {
			r = rand.Float64()
			if r > s.SubSet {
				continue
			}
			if len(i.Seq) < s.MinSize {
				continue
			}
			if len(i.Seq) > s.MaxSize {
				continue
			}
			fastq.WriteToFileHandle(out, i)
		}
	}
}

func usage() {
	fmt.Print(
		"fastqFilter - Returns a filtered fastq based on option parameters.\n" +
			"Usage:\n" +
			"fastqFilter input.fastq output.fastq\n" +
			"OR\n" +
			"fastqFilter -pairedEnd R1.fastq R2.fastq out1.fastq out2.fastq\n" +
			"options:\n")
	flag.PrintDefaults()
}

func main() {
	var expectedNumArgs int = 2
	var pairedEnd *bool = flag.Bool("pairedEnd", false, "Paired end reads, use two input and output fastq files.")
	var subSet *float64 = flag.Float64("subSet", 1.0, "Proportion of reads to retain in output.")
	var randSeed *bool = flag.Bool("randSeed", false, "Uses a random seed for the RNG.")
	var setSeed *int64 = flag.Int64("setSeed", -1, "Use a specific seed for the RNG.")
	var minSize *int = flag.Int("minSize", 0, "Retain fastq reads above this size.")
	var maxSize *int = flag.Int("maxSize", numbers.MaxInt, "Retain fastq reads below this size.")
	flag.Usage = usage
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	if *subSet < 0 || *subSet > 1 {
		log.Fatalf(fmt.Sprintf("The subSet option must be between 0 and 1, received %v.", *subSet))
	}

	if *pairedEnd {
		expectedNumArgs = 4
	}

	if len(flag.Args()) != expectedNumArgs {
		flag.Usage()
		log.Fatalf("Error: expecting %d arguments, but got %d\n",
			expectedNumArgs, len(flag.Args()))
	}

	s := Settings{
		InFile:    "",
		OutFile:   "",
		R1InFile:  "",
		R2InFile:  "",
		R1OutFile: "",
		R2OutFile: "",
		PairedEnd: *pairedEnd,
		SubSet:    *subSet,
		RandSeed:  *randSeed,
		SetSeed:   *setSeed,
		MinSize:   *minSize,
		MaxSize:   *maxSize,
	}

	if *pairedEnd {
		s.R1InFile = flag.Arg(0)
		s.R2InFile = flag.Arg(1)
		s.R1OutFile = flag.Arg(2)
		s.R2OutFile = flag.Arg(3)
	} else {
		s.InFile = flag.Arg(0)
		s.OutFile = flag.Arg(1)
	}

	fastqFilter(s)
}
