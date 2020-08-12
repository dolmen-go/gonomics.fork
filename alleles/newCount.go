package alleles

import (
	"github.com/vertgenlab/gonomics/cigar"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fasta"
	"github.com/vertgenlab/gonomics/sam"
	"sync"
)

// sendPassedPositions sends positions that have been passed in the file
func sendPassedPositions(answer chan<- *Allele, aln *sam.SamAln, samFilename string, runningCount []*Location, currAlleles map[Location]*AlleleCount) []*Location {
	for i := 0; i < len(runningCount); i++ {

		if runningCount[i].Chr != aln.RName {
			answer <- &Allele{samFilename, currAlleles[*runningCount[i]], runningCount[i]}
			delete(currAlleles, *runningCount[i])

			// Catch instance where every entry in running count is sent
			// Delete all of runningCount
			if i == len(runningCount)-1 {
				runningCount = nil
			}
			continue
		}

		if runningCount[i].Pos < (aln.Pos - 1) {
			answer <- &Allele{samFilename, currAlleles[*runningCount[i]], runningCount[i]}
			delete(currAlleles, *runningCount[i])

			// Catch instance where every entry in running count is sent
			// Delete all of runningCount
			if i == len(runningCount)-1 {
				runningCount = nil
			}

		} else {
			// Remove sent values from count
			runningCount = runningCount[i:]
			break
		}
	}
	return runningCount
}

func countRead(aln *sam.SamAln, currAlleles map[Location]*AlleleCount, runningCount []*Location, ref map[string][]dna.Base, minMapQ int64, progress int) (map[Location]*AlleleCount, []*Location) {
	var RefIndex, SeqIndex int64
	var currentSeq []dna.Base
	var i, j, k int
	var currentIndel Indel
	var indelSeq []dna.Base
	var OrigRefIndex int64
	var Match bool

	// Count the bases
	progress++
	SeqIndex = 0
	RefIndex = aln.Pos - 1

	if aln.Cigar[0].Op == '*' {
		return currAlleles, runningCount
	}

	if aln.MapQ < minMapQ {
		return currAlleles, runningCount
	}

	for i = 0; i < len(aln.Cigar); i++ {
		currentSeq = aln.Seq

		if aln.Cigar[i].Op == 'D' {
			OrigRefIndex = RefIndex
			indelSeq = make([]dna.Base, 1)

			// First base in indel is the base prior to the indel sequence per VCF standard format
			indelSeq[0] = ref[aln.RName][OrigRefIndex-1]

			for k = 0; k < int(aln.Cigar[i].RunLength); k++ {

				// If the position has already been added to the map, move along
				_, ok := currAlleles[Location{aln.RName, RefIndex}]

				// If the position is NOT in the map, initialize
				if !ok {
					currAlleles[Location{aln.RName, RefIndex}] = &AlleleCount{
						Ref: ref[aln.RName][RefIndex], Counts: 0, BaseAF: 0, BaseCF: 0, BaseGF: 0, BaseTF: 0, BaseAR: 0, BaseCR: 0, BaseGR: 0, BaseTR: 0, Indel: make([]Indel, 0)}
					runningCount = append(runningCount, &Location{aln.RName, RefIndex})
				}

				// Keep track of deleted sequence
				indelSeq = append(indelSeq, ref[aln.RName][RefIndex])

				currAlleles[Location{aln.RName, RefIndex}].Counts++
				RefIndex++
			}

			Match = false
			for j = 0; j < len(currAlleles[Location{aln.RName, OrigRefIndex}].Indel); j++ {
				// If the deletion has already been seen before, increment the existing entry
				// For a deletion the indelSeq should match the Ref
				if dna.CompareSeqsIgnoreCase(indelSeq, currAlleles[Location{aln.RName, OrigRefIndex}].Indel[j].Ref) == 0 &&
					dna.CompareSeqsIgnoreCase(indelSeq[:1], currAlleles[Location{aln.RName, OrigRefIndex}].Indel[j].Alt) == 0 {
					if sam.IsForwardRead(aln) == true {
						currAlleles[Location{aln.RName, OrigRefIndex}].Indel[j].CountF++
					} else if sam.IsReverseRead(aln) == true {
						currAlleles[Location{aln.RName, OrigRefIndex}].Indel[j].CountR++
					}

					Match = true
					break
				}
			}

			// If the deletion has not been seen before, then append it to the Del slice
			// For Alt indelSeq[:1] is used to give me a slice of just the first base in the slice which we defined earlier
			if Match == false {

				currentIndel = Indel{indelSeq, indelSeq[:1], 0, 0}
				if sam.IsForwardRead(aln) == true {
					currentIndel.CountF++
				} else if sam.IsReverseRead(aln) == false {
					currentIndel.CountR++
				}
				currAlleles[Location{aln.RName, OrigRefIndex}].Indel = append(currAlleles[Location{aln.RName, OrigRefIndex}].Indel, currentIndel)
			}

			//Handle insertion relative to ref
			//The base after the inserted sequence is annotated with an Ins read
		} else if aln.Cigar[i].Op == 'I' {

			// If the position has already been added to the map, move along
			_, ok := currAlleles[Location{aln.RName, RefIndex}]

			// If the position is NOT in the map, initialize
			if !ok {
				currAlleles[Location{aln.RName, RefIndex}] = &AlleleCount{
					Ref: ref[aln.RName][RefIndex], Counts: 0, BaseAF: 0, BaseCF: 0, BaseGF: 0, BaseTF: 0, BaseAR: 0, BaseCR: 0, BaseGR: 0, BaseTR: 0, Indel: make([]Indel, 0)}
				runningCount = append(runningCount, &Location{aln.RName, RefIndex})
			}

			// Loop through read sequence and keep track of the inserted bases
			indelSeq = make([]dna.Base, 1)

			// First base in indel is the base prior to the indel sequence per VCF standard format
			indelSeq[0] = ref[aln.RName][RefIndex-1]

			// Keep track of inserted sequence by moving along the read
			for k = 0; k < int(aln.Cigar[i].RunLength); k++ {
				indelSeq = append(indelSeq, currentSeq[SeqIndex])
				SeqIndex++
			}

			Match = false
			for j = 0; j < len(currAlleles[Location{aln.RName, RefIndex}].Indel); j++ {
				// If the inserted sequence matches a previously inserted sequence, then increment the count
				// For an insertion, the indelSeq should match the Alt
				if dna.CompareSeqsIgnoreCase(indelSeq, currAlleles[Location{aln.RName, RefIndex}].Indel[j].Alt) == 0 &&
					dna.CompareSeqsIgnoreCase(indelSeq[:1], currAlleles[Location{aln.RName, RefIndex}].Indel[j].Ref) == 0 {
					if sam.IsForwardRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].Indel[j].CountF++
					} else if sam.IsReverseRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].Indel[j].CountR++
					}
					Match = true
					break
				}
			}

			if Match == false {
				currentIndel = Indel{indelSeq[:1], indelSeq, 0, 0}
				if sam.IsForwardRead(aln) == true {
					currentIndel.CountF++
				} else if sam.IsReverseRead(aln) == true {
					currentIndel.CountR++
				}
				currAlleles[Location{aln.RName, RefIndex}].Indel = append(currAlleles[Location{aln.RName, RefIndex}].Indel, currentIndel)
			}

			// Note: Insertions do not contribute to the total counts as the insertion is associated with the previous reference base

			//Handle matching pos relative to ref
		} else if cigar.CigarConsumesReference(*aln.Cigar[i]) {

			for k = 0; k < int(aln.Cigar[i].RunLength); k++ {

				//if the position has already been added to the matrix, move along
				_, ok := currAlleles[Location{aln.RName, RefIndex}]

				//if the position is NOT in the matrix, add it
				if !ok {
					currAlleles[Location{aln.RName, RefIndex}] = &AlleleCount{
						Ref: ref[aln.RName][RefIndex], Counts: 0, BaseAF: 0, BaseCF: 0, BaseGF: 0, BaseTF: 0, BaseAR: 0, BaseCR: 0, BaseGR: 0, BaseTR: 0, Indel: make([]Indel, 0)}
					runningCount = append(runningCount, &Location{aln.RName, RefIndex})
				}

				switch currentSeq[SeqIndex] {
				case dna.A:
					if sam.IsForwardRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].BaseAF++
					} else if sam.IsReverseRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].BaseAR++
					}
					currAlleles[Location{aln.RName, RefIndex}].Counts++
				case dna.T:
					if sam.IsForwardRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].BaseTF++
					} else if sam.IsReverseRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].BaseTR++
					}
					currAlleles[Location{aln.RName, RefIndex}].Counts++
				case dna.G:
					if sam.IsForwardRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].BaseGF++
					} else if sam.IsReverseRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].BaseGR++
					}
					currAlleles[Location{aln.RName, RefIndex}].Counts++
				case dna.C:
					if sam.IsForwardRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].BaseCF++
					} else if sam.IsReverseRead(aln) == true {
						currAlleles[Location{aln.RName, RefIndex}].BaseCR++
					}
					currAlleles[Location{aln.RName, RefIndex}].Counts++
				}
				SeqIndex++
				RefIndex++
			}
		} else if aln.Cigar[i].Op != 'H' {
			SeqIndex = SeqIndex + aln.Cigar[i].RunLength
		}
	}
	return currAlleles, runningCount
}

func GoCountAlleles(samFilename string, reference []*fasta.Fasta, minMapQ int64) <-chan *Allele {
	answer := make(chan *Allele)
	var wg sync.WaitGroup
	wg.Add(1)
	go NewCountAlleles(answer, samFilename, reference, minMapQ, &wg)

	go func() {
		wg.Wait()
		close(answer)
	}()

	return answer
}

func NewCountAlleles(answer chan<- *Allele, samFilename string, reference []*fasta.Fasta, minMapQ int64, wg *sync.WaitGroup) {
	samChan, _ := sam.GoReadToChan(samFilename)
	var currAlleles = make(map[Location]*AlleleCount)
	var runningCount = make([]*Location, 0)
	var progress int // TODO: Make option to print progress

	fasta.AllToUpper(reference)
	ref := fasta.FastaMap(reference)

	for read := range samChan {
		runningCount = sendPassedPositions(answer, read, samFilename, runningCount, currAlleles)
		currAlleles, runningCount = countRead(read, currAlleles, runningCount, ref, minMapQ, progress)
	}
	wg.Done()
}