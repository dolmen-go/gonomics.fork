package simpleGraph

import (
	"fmt"
	"github.com/vertgenlab/gonomics/cigar"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fastq"
	"github.com/vertgenlab/gonomics/giraf"
	"github.com/vertgenlab/gonomics/sam"
	"log"
	"math"
	"strings"
)

func GraphSmithWatermanToGiraf(gg *SimpleGraph, read *fastq.FastqBig, seedHash map[uint64][]uint64, seedLen int, stepSize int, scoreMatrix [][]int64, m [][]int64, trace [][]rune) *giraf.Giraf {
	//start := time.Now()
	var currBest giraf.Giraf = giraf.Giraf{
		QName:     read.Name,
		QStart:    0,
		QEnd:      0,
		PosStrand: true,
		Path:      &giraf.Path{},
		Aln:       []*cigar.Cigar{&cigar.Cigar{Op: '*'}},
		AlnScore:  0,
		MapQ:      255,
		Seq:       read.Seq,
		Qual:      read.Qual,
		Notes:     []giraf.Note{giraf.Note{Tag: "XO", Type: 'Z', Value: "~"}},
	}
	var leftAlignment, rightAlignment []*cigar.Cigar = []*cigar.Cigar{}, []*cigar.Cigar{}
	var minTarget, maxTarget int
	var minQuery, maxQuery int
	var leftScore, rightScore int64 = 0, 0
	var leftPath, rightPath []uint32
	var currScore int64 = 0
	perfectScore := perfectMatchBig(read, scoreMatrix)
	extension := int(perfectScore/600) + len(read.Seq)
	var seeds []*SeedDev
	seeds = findSeedsInSmallMapWithMemPool(seedHash, gg.Nodes, read, seedLen, perfectScore, scoreMatrix)
	SortSeedDevByLen(seeds)
	var tailSeed *SeedDev
	var seedScore int64
	var currSeq []dna.Base
	var currSeed *SeedDev
	//for currSeed = seeds; currSeed != nil && seedCouldBeBetterScores(int64(currSeed.TotalLength), int64(currBest.AlnScore), perfectScore, int64(len(read.Seq)), scoreMatrix); currSeed = currSeed.Next {
	for i := 0; i < len(seeds) && seedCouldBeBetter(int64(seeds[i].TotalLength), int64(currBest.AlnScore), perfectScore, int64(len(read.Seq)), 100, 90, -196, -296); i++ {
		currSeed = seeds[i]
		tailSeed = getLastPart(currSeed)
		if currSeed.PosStrand {
			currSeq = read.Seq
		} else {
			currSeq = read.SeqRc
		}
		seedScore = scoreSeedSeq(currSeq, currSeed.QueryStart, tailSeed.QueryStart+tailSeed.Length, scoreMatrix)
		if int(currSeed.TotalLength) == len(currSeq) {
			currScore = seedScore
			minTarget = int(currSeed.TargetStart)
			maxTarget = int(tailSeed.TargetStart + tailSeed.Length)
			minQuery = int(currSeed.QueryStart)
			maxQuery = int(currSeed.TotalLength - 1)
		} else {
			leftAlignment, leftScore, minTarget, minQuery, leftPath = AlignReverseGraphTraversal(gg.Nodes[currSeed.TargetId], []dna.Base{}, int(currSeed.TargetStart), []uint32{}, extension-int(currSeed.TotalLength), currSeq[:currSeed.QueryStart], m, trace)
			rightAlignment, rightScore, maxTarget, maxQuery, rightPath = AlignTraversalFwd(gg.Nodes[tailSeed.TargetId], []dna.Base{}, int(tailSeed.TargetStart+tailSeed.Length), []uint32{}, extension-int(currSeed.TotalLength), currSeq[tailSeed.QueryStart+tailSeed.Length:], m, trace)
		}
		currScore = leftScore + seedScore + rightScore
		if currScore > int64(currBest.AlnScore) {
			currBest.QStart = minQuery
			currBest.QEnd = maxQuery
			currBest.PosStrand = currSeed.PosStrand
			currBest.Path = setPath(currBest.Path, minTarget, CatPaths(CatPaths(leftPath, getSeedPath(currSeed)), rightPath), maxTarget)
			currBest.Aln = AddSClip(minQuery, len(currSeq), cigar.CatCigar(cigar.AddCigar(leftAlignment, &cigar.Cigar{RunLength: int64(sumLen(currSeed)), Op: 'M'}), rightAlignment))
			currBest.AlnScore = int(currScore)
			currBest.Seq = currSeq
			if gg.Nodes[currBest.Path.Nodes[0]].Info != nil {
				currBest.Notes[0].Value = fmt.Sprintf("%s=%d", gg.Nodes[currBest.Path.Nodes[0]].Name, gg.Nodes[currBest.Path.Nodes[0]].Info.Start)
				currBest.Notes = append(currBest.Notes, infoToNotes(gg.Nodes, currBest.Path.Nodes))
			} else {
				currBest.Notes[0].Value = fmt.Sprintf("%s=%d", gg.Nodes[currBest.Path.Nodes[0]].Name, 1)
			}
		}
	}
	if !currBest.PosStrand {
		fastq.ReverseQualUint8Record(currBest.Qual)
	}
	//end := time.Now()
	//fmt.Println("Read Time:", end.Nanosecond() - start.Nanosecond())
	//start = time.Now()
	currBest.Aln = GirafToExplicitCigar(&currBest, gg)
	//end = time.Now()
	//fmt.Println("Cigar Time:", end.Nanosecond() - start.Nanosecond())
	//fmt.Println()
	return &currBest
}

func GirafToExplicitCigar(giraf *giraf.Giraf, graph *SimpleGraph) []*cigar.Cigar {
	var answer []*cigar.Cigar
	var seqIdx, refIdx, pathIdx int
	refIdx = giraf.Path.TStart
	var k, runLenCount int64

	if giraf.Aln[0].Op == '*' {
		return nil
	}

	for i := 0; i < len(giraf.Aln); i++ {
		switch giraf.Aln[i].Op {
		case 'M':
			runLenCount = 0

			for k = 0; k < giraf.Aln[i].RunLength; k++ {
				if refIdx > len(graph.Nodes[giraf.Path.Nodes[pathIdx]].Seq)-1 {
					pathIdx++
					refIdx = 0
				}
				if giraf.Seq[seqIdx] == graph.Nodes[giraf.Path.Nodes[pathIdx]].Seq[refIdx] {
					runLenCount++
				} else {
					if runLenCount > 0 {
						// Append the matching bases so far
						answer = append(answer, &cigar.Cigar{RunLength: runLenCount, Op: '='})
					}
					// Append the mismatch base
					if answer == nil {
						answer = append(answer, &cigar.Cigar{RunLength: 1, Op: 'X', Sequence: []dna.Base{giraf.Seq[k]}})
					} else if answer[len(answer)-1].Op == 'X' {
						answer[len(answer)-1].RunLength++
						answer[len(answer)-1].Sequence = append(answer[len(answer)-1].Sequence, giraf.Seq[k])
					} else {
						answer = append(answer, &cigar.Cigar{RunLength: 1, Op: 'X', Sequence: []dna.Base{giraf.Seq[k]}})
					}
					runLenCount = 0
				}
				seqIdx++
				refIdx++
			}

			if runLenCount > 0 {
				answer = append(answer, &cigar.Cigar{RunLength: runLenCount, Op: '='})
			}

		case 'I':
			var insSeq []dna.Base
			for k = 0; k < giraf.Aln[i].RunLength; k++ {
				insSeq = append(insSeq, giraf.Seq[seqIdx])
				seqIdx++
			}
			answer = append(answer, &cigar.Cigar{RunLength: giraf.Aln[i].RunLength, Op: 'I', Sequence: insSeq})

		case 'X':
			log.Println("WARNING: The input cigar already has explicit formatting")
			return giraf.Aln

		case '=':
			log.Println("WARNING: The input cigar already has explicit formatting")
			return giraf.Aln

		default:
			answer = append(answer, giraf.Aln[i])
			if cigar.ConsumesReference(giraf.Aln[i].Op) {
				refIdx += int(giraf.Aln[i].RunLength)
			}
			if cigar.ConsumesQuery(giraf.Aln[i].Op) {
				seqIdx += int(giraf.Aln[i].RunLength)
			}
		}
	}
	return answer
}

type ScoreMatrixHelper struct {
	Matrix                         [][]int64
	MaxMatch                       int64
	MinMatch                       int64
	LeastSevereMismatch            int64
	LeastSevereMatchMismatchChange int64
}

func getScoreMatrixHelp(scoreMatrix [][]int64) *ScoreMatrixHelper {
	help := ScoreMatrixHelper{Matrix: scoreMatrix}
	help.MaxMatch, help.MinMatch, help.LeastSevereMismatch, help.LeastSevereMatchMismatchChange = MismatchStats(scoreMatrix)
	return &help
}

func MismatchStats(scoreMatrix [][]int64) (int64, int64, int64, int64) {
	var maxMatch int64 = 0
	var minMatch int64
	var leastSevereMismatch int64 = scoreMatrix[0][1]
	var i, j int
	for i = 0; i < len(scoreMatrix); i++ {
		for j = 0; j < len(scoreMatrix[i]); j++ {
			if scoreMatrix[i][j] > maxMatch {
				minMatch = maxMatch
				maxMatch = scoreMatrix[i][j]
			} else {
				if scoreMatrix[i][j] < 0 && leastSevereMismatch < scoreMatrix[i][j] {
					leastSevereMismatch = scoreMatrix[i][j]
				}
			}

		}
	}
	var leastSevereMatchMismatchChange int64 = leastSevereMismatch - maxMatch
	return maxMatch, minMatch, leastSevereMismatch, leastSevereMatchMismatchChange
}

func WrapPairGiraf(gg *SimpleGraph, readPair *fastq.PairedEndBig, seedHash map[uint64][]uint64, seedLen int, stepSize int, scoreMatrix [][]int64, m [][]int64, trace [][]rune) *giraf.GirafPair {
	var mappedPair giraf.GirafPair = giraf.GirafPair{Fwd: nil, Rev: nil}
	mappedPair.Fwd = GraphSmithWatermanToGiraf(gg, readPair.Fwd, seedHash, seedLen, stepSize, scoreMatrix, m, trace)
	mappedPair.Rev = GraphSmithWatermanToGiraf(gg, readPair.Rev, seedHash, seedLen, stepSize, scoreMatrix, m, trace)
	setGirafFlags(&mappedPair)
	return &mappedPair
}

// setGirafFlags generates the appropriate flags for each giraf in a pair
func setGirafFlags(pair *giraf.GirafPair) {
	pair.Fwd.Flag = getGirafFlags(pair.Fwd)
	pair.Rev.Flag = getGirafFlags(pair.Rev)
	pair.Fwd.Flag += 8 // Forward
	pair.Fwd.Flag += 16 // Paired Reads
	pair.Fwd.Flag += 16 // Paired Reads
	if isProperPairAlign(pair) {
		pair.Fwd.Flag += 1 // Properly Aligned
		pair.Rev.Flag += 1 // Properly Aligned
	}
}

func GirafToSam(ag *giraf.Giraf) *sam.SamAln {
	curr := &sam.SamAln{QName: ag.QName, Flag: 4, RName: "*", Pos: 0, MapQ: 255, Cigar: []*cigar.Cigar{&cigar.Cigar{Op: '*'}}, RNext: "*", PNext: 0, TLen: 0, Seq: ag.Seq, Qual: fastq.Uint8QualToString(ag.Qual), Extra: "BZ:i:0\tGP:Z:-1\tXO:Z:~"}
	//read is unMapped
	if strings.Compare(ag.Notes[0].Value, "~") == 0 {
		return curr
	} else {
		target := strings.Split(ag.Notes[0].Value, "=")
		curr.RName = target[0]
		curr.Pos = int64(ag.Path.TStart) + common.StringToInt64(target[1])
		curr.Flag = getSamFlags(ag)
		curr.Cigar = ag.Aln

		if len(ag.Notes) == 2 {
			curr.Extra = fmt.Sprintf("BZ:i:%d\tGP:Z:%s\tXO:Z:%d\t%s", ag.AlnScore, PathToString(ag.Path.Nodes), ag.Path.TStart, giraf.NoteToString(ag.Notes[1]))
		} else {
			curr.Extra = fmt.Sprintf("BZ:i:%d\tGP:Z:%s\tXO:Z:%d", ag.AlnScore, PathToString(ag.Path.Nodes), ag.Path.TStart)
		}
	}
	return curr
}

func GirafPairToSam(ag *giraf.GirafPair) *sam.PairedSamAln {
	var mappedPair sam.PairedSamAln = sam.PairedSamAln{FwdSam: &sam.SamAln{}, RevSam: &sam.SamAln{}}
	mappedPair.FwdSam = GirafToSam(ag.Fwd)
	mappedPair.RevSam = GirafToSam(ag.Rev)
	mappedPair.FwdSam.Flag += 64
	mappedPair.RevSam.Flag += 128
	if isProperPairAlign(ag) {
		mappedPair.FwdSam.Flag += 2
		mappedPair.RevSam.Flag += 2
	}
	return &mappedPair
}

func isProperPairAlign(mappedPair *giraf.GirafPair) bool {
	if math.Abs(float64(mappedPair.Fwd.Path.TStart-mappedPair.Rev.Path.TStart)) < 10000 {
		if mappedPair.Fwd.Path.TStart < mappedPair.Rev.Path.TStart && mappedPair.Fwd.PosStrand && !mappedPair.Rev.PosStrand {
			return true
		}
		if mappedPair.Fwd.Path.TStart > mappedPair.Rev.Path.TStart && !mappedPair.Fwd.PosStrand && mappedPair.Rev.PosStrand {
			return true
		}
	}
	return false
}

func getGirafFlags(ag *giraf.Giraf) uint8 {
	var answer uint8
	if ag.PosStrand {
		answer += 4 // Positive Strand
	}
	if ag.AlnScore < 1200 {
		answer += 2 // Unmapped
	}
	return answer
}

func getSamFlags(ag *giraf.Giraf) int64 {
	var answer int64
	if !ag.PosStrand {
		answer += 16
	}
	if ag.AlnScore < 1200 {
		answer += 4
	}
	return answer
}

func setPath(p *giraf.Path, targetStart int, nodes []uint32, targetEnd int) *giraf.Path {
	p.TStart = targetStart
	p.Nodes = nodes
	p.TEnd = targetEnd
	return p
}

func vInfoToValue(n *Node) string {
	var answer string
	switch {
	case n.Info.Variant == 1:
		answer = fmt.Sprintf("%d=%s", n.Id, "snp")
	case n.Info.Variant == 2:
		answer = fmt.Sprintf("%d=%s", n.Id, "ins")
	case n.Info.Variant == 3:
		answer = fmt.Sprintf("%d=%s", n.Id, "del")
	}
	return answer
}

func infoToNotes(nodes []*Node, path []uint32) giraf.Note {
	var vInfo giraf.Note = giraf.Note{Tag: "XV", Type: 'Z'}
	vInfo.Value = fmt.Sprintf("%d_%d", nodes[0].Info.Allele, nodes[0].Info.Variant)
	if len(path) > 0 {
		for i := 1; i < len(path); i++ {
			if nodes[i].Info.Variant > 0 {
				vInfo.Value += fmt.Sprintf(",%s", vInfoToValue(nodes[path[i]]))
			} else {
				vInfo.Value += fmt.Sprintf(",%d_%d", nodes[i].Info.Allele, nodes[path[i]].Info.Variant)
			}

		}
	}
	return vInfo
}
