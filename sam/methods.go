package sam

import (
	"github.com/vertgenlab/gonomics/cigar"
	"github.com/vertgenlab/gonomics/fileio"
)

func (s *Aln) GetChrom() string {
	return s.RName
}

func (s *Aln) GetChromStart() int {
	return int(s.Pos - 1)
}

func (s *Aln) GetChromEnd() int {
	var runlength int = 0
	for i := 0; i < len(s.Cigar); i++ {
		if cigar.ConsumesReference(s.Cigar[i].Op) {
			runlength += int(s.Cigar[i].RunLength)
		}
	}
	return s.GetChromStart() + runlength
}

func (s *Aln) UpdateLift(c string, start int, end int) {
	s.RName = c
	s.Pos = uint32(start) + 1
}

type SamSlice []*Aln

func (v SamSlice) Len() int { return len(v) }

func (v SamSlice) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

func (v *SamSlice) Push(x interface{}) {
	answer := x.(*Aln)
	*v = append(*v, answer)
}

func (v *SamSlice) Pop() interface{} {
	oldQueue := *v
	n := len(oldQueue)
	answer := oldQueue[n-1]
	*v = oldQueue[:n-1]
	return answer
}

// TODO: modify sort/mergeSort.go to look for a header and save for output
func (s SamSlice) Write(filename string) { // Does not write header
	file := fileio.EasyCreate(filename)
	defer file.Close()
	for i := 0; i < s.Len(); i++ {
		s[i].WriteToFileHandle(file)
	}
}

func (s *Aln) WriteToFileHandle(file *fileio.EasyWriter) {
	WriteToFileHandle(file, *s)
}

func (s *Aln) NextRealRecord(file *fileio.EasyReader) bool {
	var done bool
	var next *Aln
	var curr Aln
	for nextBytes, err := file.Peek(1); next == nil && !done; nextBytes, err = file.Peek(1) {
		if err == nil && nextBytes[0] == '@' {
			fileio.EasyNextLine(file)
			continue
		}
		curr, done = ReadNext(file)
		next = &curr
	}
	if done {
		return true
	}
	*s = *next
	return done
}

func (s *Aln) Copy() interface{} {
	var answer *Aln = new(Aln)
	*answer = *s
	return answer
}

type ByGenomicCoordinates struct {
	SamSlice
}

func (g ByGenomicCoordinates) Less(i, j int) bool {
	// First sort criteria is chromosome
	if g.SamSlice[i].GetChrom() < g.SamSlice[j].GetChrom() {
		return true
	} else if g.SamSlice[i].GetChrom() == g.SamSlice[j].GetChrom() {
		// If chroms are equal then sort by start position
		if g.SamSlice[i].GetChromStart() < g.SamSlice[j].GetChromStart() {
			return true
		} else if g.SamSlice[i].GetChromStart() == g.SamSlice[j].GetChromStart() {
			// If start positions are equal then the shorter region wins
			if g.SamSlice[i].GetChromEnd() < g.SamSlice[j].GetChromEnd() {
				return true
			}
		}
	}
	return false
}
