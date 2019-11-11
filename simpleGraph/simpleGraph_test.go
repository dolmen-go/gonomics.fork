package simpleGraph

import (
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/sam"
	"github.com/vertgenlab/gonomics/fastq"
	//"github.com/vertgenlab/gonomics/align"
	"log"
	"os"
	"testing"
	"runtime"
	"runtime/pprof"
	"flag"
	"fmt"
)

var seqOneA = dna.StringToBases("ACGTACGTCATCATCATTACTACTAC")
var seqOneB = dna.StringToBases("ACGTACGT")
var seqOneC = dna.StringToBases("ACGTACGTACGTT")
var readWriteTests = []struct {
	filename string // input
	data     []*Node
}{
	{"testdata/testOne.sg", []*Node{{0, "seqOneA", seqOneA, nil, nil}, {1, "seqOneB", seqOneB, nil, nil}, {2, "seqOneC", seqOneC, nil, nil}}},
}

func TestRead(t *testing.T) {
	for _, test := range readWriteTests {
		actual := Read(test.filename)
		if !AllAreEqual(test.data, actual.Nodes) {
			t.Errorf("The %s file was not read correctly.", test.filename)
		}
	}
}

func TestWriteAndRead(t *testing.T) {
	//var actual []*Node
	for _, test := range readWriteTests {
		tempFile := test.filename + ".tmp"
		Write(tempFile, test.data)
		actual := Read(tempFile)
		if !AllAreEqual(test.data, actual.Nodes) {
			t.Errorf("The %s file was not read correctly.", test.filename)
		}
		err := os.Remove(tempFile)
		if err != nil {
			t.Errorf("Deleting temp file %s gave an error.", tempFile)
		}
	}
}

func TestGraphTraversal(t *testing.T) {
	gg := NewGraph()

	var seqOne = dna.StringToBases("ATG")
	var seqTwo = dna.StringToBases("GGA")
	var seqThree = dna.StringToBases("CCC")
	var seqFour = dna.StringToBases("TTG")
	var seqFive = dna.StringToBases("TGT")

	testFastq := fastq.Fastq{Name: "TestSeq", Seq: dna.StringToBases("CCCTTG"), Qual: []rune("JJJJJJ")}
	nA := Node{0, "A", seqOne, nil, nil}
    nB := Node{1, "B", seqTwo, nil, nil}
    nC := Node{3, "C", seqThree, nil, nil}
    nD := Node{4, "D", seqFour, nil, nil}
    nE := Node{5, "E", seqFive, nil, nil}

    AddNode(gg, &nA)
    AddNode(gg, &nB)
    AddNode(gg, &nC)
    AddNode(gg, &nD)
    AddNode(gg, &nE)

    //AddEdge(gg, &nA, &nB, 1)
    //AddEdge(gg, &nB, &nC, 1)

    AddEdge(&nA, &nB, 1)
    AddEdge(&nA, &nC, 1)
    AddEdge(&nA, &nE, 1)
    AddEdge(&nB, &nD, 1)
    AddEdge(&nC, &nD, 1)
    AddEdge(&nE, &nD, 1)
    var query []dna.Base
    var path string = ""
    //fmt.Printf("Start node: %s\n", dna.BasesToString(gg.Nodes[3].Seq))
    path, query = ReverseGraphTraversal(gg, gg.Nodes[0], query, path, 1, 9)
    GraphTraversalFwd(gg,gg.Nodes[0], query, path, 1, 9)
    m, trace := swMatrixSetup(10000)
    var mappedRead *sam.SamAln =  &sam.SamAln{QName: testFastq.Name, Flag: 0, RName: "", Pos: 0, MapQ: 255, RNext: "*", PNext: 0, TLen: 0, Seq: []dna.Base{}, Qual: "", Extra: ""}
    var score int64
    mappedRead, score = AlignTraversalFwd(gg, gg.Nodes[0], query, 0, "", 11, testFastq, m, trace, mappedRead, score)
    fmt.Printf("%s\n", sam.SamAlnToString(mappedRead))

    //log.Printf("Indexing the graphGenome...\n")
    //ham5 := make(map[uint64][]uint64)
    //ham5 = GraphTraversalHashMap(gg, gg.Nodes[0], 0, 2, ham5)
    //for i := 0; i < len(ham5[dnaToNumber(dna.StringToBases("AC"), 0, 2)]); i++ {
    //	nodeIdx, pos := numberToChromAndPos(ham5[dnaToNumber(dna.StringToBases("AC"), 0, 2)][i])
    //	fmt.Printf("NodeIdx: %d Position: %d", nodeIdx, pos)
    //}
    

   // mappedRead, score = AlignReverseGraphTraversal(gg, gg.Nodes[3], query, 0, "", 9, testFastq, m, trace, mappedRead, score)
    //fmt.Printf("%s\n", sam.SamAlnToString(mappedRead))
}

func TestAligning(t *testing.T) {
	var tileSize int = 30
	var readLength int = 150
	var numberOfReads int = 1
	var mappedRead *sam.SamAln

	log.Printf("Reading in the genome (simple graph)...\n")
	genome := Read("testdata/bigGenome.sg")

	log.Printf("Indexing the genome...\n")
	tiles := indexGenome(genome.Nodes, tileSize)

	log.Printf("Simulating reads...\n")
	simReads := RandomReads(genome.Nodes, readLength, numberOfReads)
	m, trace := swMatrixSetup(10000)


	//code block for program profiling
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

	flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal("could not create CPU profile: ", err)
        }
        defer f.Close()
        if err := pprof.StartCPUProfile(f); err != nil {
            log.Fatal("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
    }


	log.Printf("Aligning reads...\n")

	for i := 0; i < len(simReads); i++ {
		mappedRead = MapSingleFastq(genome.Nodes, tiles, simReads[i], tileSize, m, trace)
		fmt.Printf("%s\n", sam.SamAlnToString(mappedRead))
	}
	log.Printf("Done mapping %d reads\n", numberOfReads)
	//PrintGraph(genome)

	//code block for program profiling
	if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal("could not create memory profile: ", err)
        }
        defer f.Close()
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            log.Fatal("could not write memory profile: ", err)
        }
    }
}

func BenchmarkAligning(b *testing.B) {
	var tileSize int = 30
	var readLength int = 150
	var numberOfReads int = 100
	var mappedReads []*sam.SamAln = make([]*sam.SamAln, numberOfReads)

	genome := Read("testdata/bigGenome.sg")
	tiles := indexGenome(genome.Nodes, tileSize)
	simReads := RandomReads(genome.Nodes, readLength, numberOfReads)
	//var seeds []Seed = make([]Seed, 256)
	m, trace := swMatrixSetup(10000)
	//var seeds []Seed = make([]Seed, 256)
	b.ResetTimer()
	
	var i int

	for n := 0; n < b.N; n++ {
		for i = 0; i < len(simReads); i++ {
			mappedReads[i] = MapSingleFastq(genome.Nodes, tiles, simReads[i], tileSize, m, trace)
			//log.Printf("%s\n", sam.SamAlnToString(mappedReads[i]))
		}
	}
}
