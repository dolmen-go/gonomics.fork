package giraf

import (
	"fmt"
	"github.com/vertgenlab/gonomics/cigar"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fileio"
	"os"
	"sync"
)

type Giraf struct {
	QName     string
	QStart    int
	QEnd      int
	PosStrand bool
	Path      *Path
	Aln       []*cigar.Cigar // current cigar will need to be expanded
	AlnScore  int
	MapQ      uint8
	Seq       []dna.Base // dnaTwoBit?
	Qual      []uint8
	Notes     []Note // Similar to sam, this is should be a list of notes.
	// Each note should be of the form TAG:TYPE:VALUE
	// TAG is two characters
	// TYPE is a single character
	// VALUE will be stored as a string and can then be de-coded based on type
	// An example would be "BZ:i:4000
}

type Path struct {
	TStart int      // The path starts on the TStart base (0-based, closed) of Nodes[0]
	Nodes  []uint32 // The node Id/Index of all the nodes in the path
	TEnd   int      // The path ends on the TEnd base (0-based, open) of Nodes[len(Nodes)-1]
}

type Note struct {
	Tag   string
	Type  rune
	Value string
}

type GirafPair struct {
	Fwd *Giraf
	Rev *Giraf
}

func Read(filename string) []*Giraf {
	var answer []*Giraf
	file := fileio.EasyOpen(filename)
	defer file.Close()
	var curr *Giraf
	var done bool
	for curr, done = NextGiraf(file); !done; curr, done = NextGiraf(file) {
		answer = append(answer, curr)
	}
	return answer
}

func NextGiraf(reader *fileio.EasyReader) (*Giraf, bool) {
	line, done := fileio.EasyNextLine(reader)
	if done {
		return nil, true
	}
	return stringToGiraf(line), false
}

func Write(filename string, gfs []*Giraf) {
	file, _ := os.Create(filename)
	defer file.Close()
	for i := 0; i < len(gfs); i++ {
		WriteGriafHelper(file, gfs[i])
	}
}

func GirafChanToFile(filename string, input <-chan *Giraf, wg *sync.WaitGroup) {
	file, _ := os.Create(filename)
	defer file.Close()
	for line := range input {
		WriteGriafHelper(file, line)
	}
	wg.Done()
}

func GirafPairChanToFile(filename string, input <-chan *GirafPair, wg *sync.WaitGroup) {
	file, _ := os.Create(filename)
	defer file.Close()
	for pair := range input {
		WriteGriafHelper(file, pair.Fwd)
		WriteGriafHelper(file, pair.Rev)
	}
	wg.Done()
}

func WriteGriafHelper(file *os.File, gf *Giraf) {
	_, err := fmt.Fprintf(file, "%s\n", GriafToString(gf))
	common.ExitIfError(err)
}
