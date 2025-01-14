// Command Group: "Data Conversion"

package main

import (
	"flag"
	"fmt"
	"github.com/vertgenlab/gonomics/axt"
	"github.com/vertgenlab/gonomics/chromInfo"
	"github.com/vertgenlab/gonomics/fasta"
	"github.com/vertgenlab/gonomics/fileio"
	"github.com/vertgenlab/gonomics/sam"
	"log"
	"sync"
)

func usage() {
	fmt.Print(
		"axtSam - convert axt alignments to sam format\n\n" +
			"Usage:\n" +
			"  axtSam [options] in.axt out.sam\n\n" +
			"Options:\n\n")
	flag.PrintDefaults()
	fmt.Print("Comming Soon: hard clip included sam cigar to represent entire query sequence\n\n")
}

func main() {
	var expectedNumArgs int = 2
	flag.Usage = usage
	log.SetFlags(log.Ldate | log.Ltime)
	var chrInfo *string = flag.String("chrom", "", "in order to write a proper sam header that is compatible with samtools,\na chrom`.sizes` file to determine target lengths must be provided")
	var faSeq *string = flag.String("fasta", "", "provide target sequences in `.fasta` format to obtain sequence lengths\nto process sam header\n")
	flag.Parse()

	if len(flag.Args()) != expectedNumArgs {
		flag.Usage()
		log.Fatalf("Error: expecting %d arguments, but got %d\n\n", expectedNumArgs, len(flag.Args()))
	} else {
		inFile, outFile := flag.Arg(0), flag.Arg(1)
		var headerInfo sam.Header

		if *chrInfo != "" {
			headerInfo = chromInfoSamHeader(*chrInfo)
		} else if *faSeq != "" {
			headerInfo = sam.GenerateHeader(fasta.ToChromInfo(fasta.Read(*faSeq)), nil, sam.Unsorted, sam.None)
		} else {
			log.Printf("Warning: no files were detected to support writing a proper sam header. Converted alignment formats will not be compatible with samtools\n")
		}
		log.Printf("Converting axt alignments into sam format...\n")

		axtToSam(inFile, headerInfo, outFile)
		log.Printf("Finished!\n")
	}
}

func axtToSam(axtfile string, header sam.Header, output string) {
	writer := fileio.EasyCreate(output)

	defer writer.Close()

	//setting up channels and wait groups
	results := make(chan sam.Sam, 824)
	var working, writingJob sync.WaitGroup

	data, _ := axt.GoReadToChan(axtfile)

	if header.Text != nil {
		sam.WriteHeaderToFileHandle(writer, header)
	}

	working.Add(1)
	go routineWorker(data, results, &working)

	writingJob.Add(1)
	go sam.WriteFromChan(results, output, header, &writingJob)

	working.Wait()
	close(results)
	writingJob.Wait()
}

func chromInfoSamHeader(filename string) sam.Header {
	return sam.GenerateHeader(chromInfo.ReadToSlice(filename), nil, sam.Unsorted, sam.None)
}

//Not sure if this is a potiential speed up, but i have fairly large axt files that come out of chain merge
//The idea is to provide at least 3 threads with some work, reading, axtToSam, writing
func routineWorker(readChannel <-chan axt.Axt, writingChannel chan<- sam.Sam, wg *sync.WaitGroup) {
	var numberComplete int = 0
	for eachAxt := range readChannel {
		writingChannel <- axt.ToSam(eachAxt)
		numberComplete++
	}
	wg.Done()
	log.Printf("Processed %d axt alignment into sam format\n", numberComplete)

}
