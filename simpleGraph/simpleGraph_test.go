package simpleGraph

import (
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/sam"
	"log"
	"os"
	"testing"
)

var seqOneA = dna.StringToBases("ACGTACGTCATCATCATTACTACTAC")
var seqOneB = dna.StringToBases("ACGTACGT")
var seqOneC = dna.StringToBases("ACGTACGTACGTT")
var readWriteTests = []struct {
	filename string // input
	data     []*Node
}{
	{"testdata/testOne.sg", []*Node{{0, seqOneA}, {1, seqOneB}, {2, seqOneC}}},
}

func TestRead(t *testing.T) {
	for _, test := range readWriteTests {
		actual := Read(test.filename)
		if !AllAreEqual(test.data, actual) {
			t.Errorf("The %s file was not read correctly.", test.filename)
		}
	}
}

func TestWriteAndRead(t *testing.T) {
	var actual []*Node
	for _, test := range readWriteTests {
		tempFile := test.filename + ".tmp"
		Write(tempFile, test.data)
		actual = Read(tempFile)
		if !AllAreEqual(test.data, actual) {
			t.Errorf("The %s file was not read correctly.", test.filename)
		}
		err := os.Remove(tempFile)
		if err != nil {
			t.Errorf("Deleting temp file %s gave an error.", tempFile)
		}
	}
}

func TestAligning(t *testing.T) {
	var tileSize int = 30
	var readLength int = 150
	var numberOfReads int = 10
	var mappedRead *sam.SamAln

	log.Printf("Reading in the genome (simple graph)...\n")
	genome := Read("testdata/bigGenome.sg")

	log.Printf("Indexing the genome...\n")
	tiles := indexGenome(genome, tileSize)

	log.Printf("Simulating reads...\n")
	simReads := RandomReads(genome, readLength, numberOfReads)

	log.Printf("Aligning reads...\n")
	for i := 0; i < len(simReads); i++ {
		mappedRead = MapSingleFastq(genome, tiles, simReads[i], tileSize)
		log.Printf("%s\n", sam.SamAlnToString(mappedRead))
	}
	log.Printf("Done mapping\n")
}

func BenchmarkAligning(b *testing.B) {
	var tileSize int = 30
	var readLength int = 150
	var numberOfReads int = 100
	var mappedReads []*sam.SamAln = make([]*sam.SamAln, numberOfReads)

	genome := Read("testdata/bigGenome.sg")
	tiles := indexGenome(genome, tileSize)
	simReads := RandomReads(genome, readLength, numberOfReads)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(simReads); i++ {
			mappedReads[i] = MapSingleFastq(genome, tiles, simReads[i], tileSize)
		}
	}
}
