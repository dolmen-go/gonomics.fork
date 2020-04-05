package simpleGraph

import (
	"fmt"
	"github.com/vertgenlab/gonomics/cigar"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fasta"
	"github.com/vertgenlab/gonomics/fastq"
	"github.com/vertgenlab/gonomics/sam"
	"github.com/vertgenlab/gonomics/vcf"
	"log"
	"strings"
	"time"
)

func FaToGenomeGraph(ref []*fasta.Fasta, vcfs []*vcf.Vcf) (*SimpleGraph, map[string]*Node) {
	gg := NewGraph()
	chrom := make(map[string]*Node)

	vcfSplit := vcf.VcfSplit(vcfs, ref)
	var i, nodeIdx int
	if len(vcfSplit) != len(ref) {
		log.Fatal("Slice of vcfs do not equal reference length")
	} else {
		for i = 0; i < len(ref); i++ {
			gg = vChrGraph(gg, ref[i], vcfSplit[i])
			_, ok := chrom[ref[i].Name]
			if !ok {
				chrom[ref[i].Name] = gg.Nodes[nodeIdx]
				nodeIdx = len(gg.Nodes)
			} else {
				log.Fatalf("Error: The indexing to keep track of the head nodes are off...")
			}
		}
	}
	return gg, chrom
}

func VariantGraph(reference []*fasta.Fasta, vcfs []*vcf.Vcf) *SimpleGraph {
	gg := NewGraph()
	vcfSplit := vcf.VcfSplit(vcfs, reference)
	var filterVcf []*vcf.Vcf
	if len(vcfSplit) != len(reference) {
		log.Fatal("Slice of vcfs do not equal reference length")
	} else {
		for i := 0; i < len(reference); i++ {
			if len(vcfSplit[i]) != 0 {
				vcfSplit[i] = append(vcfSplit[i], &vcf.Vcf{Chr: reference[i].Name, Pos: int64(len(reference[i].Seq))})
				filterVcf = vcf.FilterNs(vcfSplit[i])
				vcf.Sort(filterVcf)
				gg = vChrGraph(gg, reference[i], filterVcf)
			} else {
				//when only calling large SV, no variants called on chr is possible, entire chr becomes a node
				chrNode := &Node{Id: uint32(len(gg.Nodes)), Name: reference[i].Name, Seq: reference[i].Seq, Prev: nil, Next: nil, Info: &Annotation{Allele: 0, Start: 1, Variant: 0}}
				AddNode(gg, chrNode)
			}
		}
	}
	return gg
}

func SplitGraphChr(reference []*fasta.Fasta, vcfs []*vcf.Vcf) map[string]*SimpleGraph {
	gg := make(map[string]*SimpleGraph)
	vcfSplit := vcf.VcfSplit(vcfs, reference)
	if len(vcfSplit) != len(reference) {
		log.Fatal("Slice of vcfs do not equal reference length")
	} else {
		for i := 0; i < len(reference); i++ {
			gg[reference[i].Name] = vChrGraph(NewGraph(), reference[i], vcfSplit[i])
		}
	}
	return gg
}

func vChrGraph(genome *SimpleGraph, chr *fasta.Fasta, vcfsChr []*vcf.Vcf) *SimpleGraph {
	fasta.ToUpper(chr)
	var currMatch *Node = nil
	var lastMatch *Node = nil
	var refAllele, altAllele *Node
	var prev []*Node = nil
	var i, j, edge int
	var index int64 = 0
	var num int = 0
	for i = 0; i < len(vcfsChr)-1; i++ {
		if strings.Compare(chr.Name, vcfsChr[i].Chr) != 0 {
			log.Fatalf("Error: chromosome names do not match...\n")
		}
		if vcfsChr[i].Pos-index > 0 {
			currMatch = &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: chr.Seq[index : vcfsChr[i].Pos-1], Prev: nil, Next: make([]*Edge, 0, 2), Info: &Annotation{Allele: 0, Start: uint32(index + 1), Variant: 0}}
			if len(currMatch.Seq) == 0 {
				currMatch = lastMatch
				//assuming we already created the ref allele, we only need to create
				//the alt alleles this iteration
				if isSNP(vcfsChr[i]) {
					altAllele = &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Alt), Prev: nil, Next: nil, Info: &Annotation{Allele: 1, Start: uint32(vcfsChr[i].Pos), Variant: 1}}
					AddNode(genome, altAllele)
					AddEdge(currMatch, altAllele, 0.5)
				} else if isINS(vcfsChr[i]) {
					insertion := &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Alt)[1:], Prev: nil, Next: nil, Info: &Annotation{Allele: 1, Start: uint32(vcfsChr[i].Pos), Variant: 2}}
					AddNode(genome, insertion)
					AddEdge(currMatch, insertion, 1)

					index = vcfsChr[i].Pos - 1
				} else if isDEL(vcfsChr[i]) {
					deletion := &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Ref)[1:], Prev: nil, Next: nil, Info: &Annotation{Allele: 0, Start: uint32(vcfsChr[i].Pos), Variant: 3}}
					AddNode(genome, deletion)
					AddEdge(currMatch, deletion, 1)
					if strings.Contains(vcfsChr[i].Id, "pbsv") {
						index = common.MinInt64(vcfsChr[i].Pos+int64(len(deletion.Seq))-1, vcfsChr[i+1].Pos-1)
					} else {
						index = vcfsChr[i].Pos + int64(len(deletion.Seq))
					}
					//	index = common.MinInt64(getSvEnd(vcfsChr[i])-1, vcfsChr[i+1].Pos-1)
					//	} else {

					//	}
				} else if strings.Compare(vcfsChr[i].Format, "SVTYPE=SNP;INS") == 0 || strings.Compare(vcfsChr[i].Format, "SVTYPE=SNP;DEL") == 0 || strings.Compare(vcfsChr[i].Format, "SVTYPE=HAP") == 0 {
					altAllele := &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Alt), Prev: nil, Next: nil, Info: &Annotation{Allele: 1, Start: uint32(vcfsChr[i].Pos), Variant: 4}}
					AddNode(genome, altAllele)
					AddEdge(currMatch, altAllele, 1)
					index = vcfsChr[i].Pos + int64(len(refAllele.Seq)) - 1
				}
				lastMatch = currMatch
			} else if len(currMatch.Seq) > 0 {

				AddNode(genome, currMatch)
				if lastMatch != nil {
					if len(lastMatch.Next) > 0 {
						for edge = 0; edge < len(lastMatch.Next); edge++ {
							AddEdge(lastMatch.Next[edge].Dest, currMatch, 1)
						}
					}
					if i > 0 {
						if isSNP(vcfsChr[i-1]) || isHaplotypeBlock(vcfsChr[i-1]) {
							AddEdge(altAllele, currMatch, 1)
						}
					}
					AddEdge(lastMatch, currMatch, 1)
					SetEvenWeights(lastMatch)
				}
				if isSNP(vcfsChr[i]) {

					refAllele = &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Ref), Prev: nil, Next: nil, Info: &Annotation{Allele: 0, Start: uint32(vcfsChr[i].Pos), Variant: 1}}
					AddNode(genome, refAllele)
					AddEdge(currMatch, refAllele, 0.5)

					altAllele = &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Alt), Prev: nil, Next: nil, Info: &Annotation{Allele: 1, Start: uint32(vcfsChr[i].Pos), Variant: 1}}
					AddNode(genome, altAllele)
					AddEdge(currMatch, altAllele, 0.5)

					currMatch = refAllele
					index = vcfsChr[i].Pos
					for j = i + 1; j < len(vcfsChr)-1; j++ {
						if isSNP(vcfsChr[j-1]) && isSNP(vcfsChr[j]) && vcfsChr[j].Pos-1 == vcfsChr[j-1].Pos {
							refAllele.Seq = append(refAllele.Seq, dna.StringToBases(vcfsChr[j].Ref)...)
							altAllele.Seq = append(altAllele.Seq, dna.StringToBases(vcfsChr[j].Alt)...)
							index = vcfsChr[j].Pos
						} else {
							lastMatch = currMatch
							i = j - 1
							break
						}
					}
				} else if isINS(vcfsChr[i]) {
					//currMatch.Seq = append(currMatch.Seq, dna.StringToBases(vcfsChr[i].Ref)...)
					insertion := &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Alt), Prev: nil, Next: nil, Info: &Annotation{Allele: 1, Start: uint32(vcfsChr[i].Pos), Variant: 2}}
					AddNode(genome, insertion)
					AddEdge(currMatch, insertion, 1)
					index = vcfsChr[i].Pos - 1
					//if strings.Contains(vcfsChr[i].Id, "pbsv") {
					//	index = getSvEnd(vcfsChr[i]) - 1
					//}
				} else if isDEL(vcfsChr[i]) {
					//TODO: does not deal with ends of large deletions well, might contain extra sequence towards the end.
					//currMatch.Seq = append(currMatch.Seq, dna.StringToBases(vcfsChr[i].Alt)...)
					deletion := &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Ref), Prev: nil, Next: nil, Info: &Annotation{Allele: 0, Start: uint32(vcfsChr[i].Pos), Variant: 3}}
					AddNode(genome, deletion)
					AddEdge(currMatch, deletion, 1)
					if strings.Contains(vcfsChr[i].Id, "pbsv") {
						index = common.MinInt64(vcfsChr[i].Pos+int64(len(deletion.Seq))-1, vcfsChr[i+1].Pos-1)
					} else {
						index = vcfsChr[i].Pos + int64(len(deletion.Seq))
					}
				} else if isINV(vcfsChr[i]) {
					currMatch.Seq = append(currMatch.Seq, dna.StringToBases(vcfsChr[i].Ref)...)
					inversion := mkInversionNode(genome, vcfsChr[i], chr)
					AddNode(genome, inversion)
					AddEdge(currMatch, inversion, 1)
					index = getSvEnd(vcfsChr[i])
				} else if isCNV(vcfsChr[i]) || isDup(vcfsChr[i]) {
					currMatch.Seq = append(currMatch.Seq, dna.StringToBases(vcfsChr[i].Ref)...)
					copyVariance := &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: chr.Seq[vcfsChr[i].Pos:getSvEnd(vcfsChr[i])], Prev: nil, Next: nil, Info: &Annotation{Allele: 1, Start: uint32(vcfsChr[i].Pos), Variant: 6}}
					AddNode(genome, copyVariance)
					AddEdge(currMatch, copyVariance, 1)
					index = getSvEnd(vcfsChr[i])
				} else if isHaplotypeBlock(vcfsChr[i]) {
					refAllele = &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Ref), Prev: nil, Next: nil, Info: &Annotation{Allele: 0, Start: uint32(vcfsChr[i].Pos), Variant: 4}}
					AddNode(genome, refAllele)
					AddEdge(currMatch, refAllele, 1)
					altAllele = &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: dna.StringToBases(vcfsChr[i].Alt), Prev: nil, Next: nil, Info: &Annotation{Allele: 1, Start: uint32(vcfsChr[i].Pos), Variant: 4}}
					AddNode(genome, altAllele)
					AddEdge(currMatch, altAllele, 1)
					index = common.MinInt64(vcfsChr[i].Pos+int64(len(refAllele.Seq))-1, vcfsChr[i+1].Pos-1)
					//index = vcfsChr[i].Pos + int64(len(refAllele.Seq)) - 1
					currMatch = refAllele
				}
				lastMatch = currMatch
			} else {
				num++
				log.Printf("index=%d, vcfPos=%d\ndiff=%d\n%s", index, vcfsChr[i].Pos, vcfsChr[i].Pos-index, vcfsChr[i].Info)
			}
		}
	}
	//Case: last node
	lastNode := &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: chr.Seq[index:], Prev: make([]*Edge, 0, len(prev)), Next: nil, Info: &Annotation{Allele: 0, Start: uint32(index + 1), Variant: 0}}
	//weight = float32(1) / float32(len(lastMatch.Next))
	AddNode(genome, lastNode)
	for edge = 0; edge < len(lastMatch.Next); edge++ {
		AddEdge(lastMatch.Next[edge].Dest, lastNode, 1)
	}
	if isSNP(vcfsChr[len(vcfsChr)-1]) || isHaplotypeBlock(vcfsChr[len(vcfsChr)-1]) {
		AddEdge(altAllele, currMatch, 1)
	}
	//if !isSNP(vcfsChr[len(vcfsChr)-1]) {
	AddEdge(lastMatch, lastNode, 1)
	SetEvenWeights(lastMatch)
	log.Printf("%s Graph missed %d/%d\n", chr.Name, num, len(vcfsChr))
	return genome
}

func GraphToFa(gg *SimpleGraph) []*fasta.Fasta {
	var answer []*fasta.Fasta
	for i := 0; i < len(gg.Nodes); i++ {
		if len(gg.Nodes[i].Seq) == 0 {
			log.Printf("Empty node: %s_%d_%d_%d_%d", gg.Nodes[i].Name, gg.Nodes[i].Id, gg.Nodes[i].Info.Allele, gg.Nodes[i].Info.Variant, gg.Nodes[i].Info.Start)
		}
		//if (gg.Nodes[i].Info.Variant == 2 || gg.Nodes[i].Info.Variant == 3) && (len(gg.Nodes[i].Seq) > 0) {
		fragment := &fasta.Fasta{Name: fmt.Sprintf("%s_%d_%d_%d_%d", gg.Nodes[i].Name, gg.Nodes[i].Id, gg.Nodes[i].Info.Allele, gg.Nodes[i].Info.Variant, gg.Nodes[i].Info.Start), Seq: gg.Nodes[i].Seq}
		answer = append(answer, fragment)
		//}
	}
	return answer
}

//new nodes are treated as insertion
func isINV(v *vcf.Vcf) bool {
	var truth bool = false
	data := strings.Split(v.Info, ";")
	if strings.Compare(v.Alt, "<INV>") == 0 || strings.Compare(data[0], "SVTYPE=INV") == 0 {
		truth = true
	}
	return truth
}

func isSNP(v *vcf.Vcf) bool {
	var truth bool = false
	if strings.Compare(v.Format, "SVTYPE=SNP") == 0 {
		truth = true
	}
	return truth
}

func isINS(v *vcf.Vcf) bool {
	var truth bool = false
	//data := strings.Split(v.Info, ";")
	if strings.Compare(v.Format, "SVTYPE=INS") == 0 {
		truth = true
	}
	if strings.Contains(v.Info, "SVTYPE=INS") {
		truth = true
	}
	return truth
}

func isDEL(v *vcf.Vcf) bool {
	var truth bool = false
	//data := strings.Split(v.Info, ";")
	if strings.Compare(v.Format, "SVTYPE=DEL") == 0 {
		truth = true
	}
	if strings.Contains(v.Info, "SVTYPE=DEL") {
		truth = true
	}
	return truth
}

func isDup(v *vcf.Vcf) bool {
	var truth bool = false
	if strings.Contains(v.Info, "SVTYPE=DUP") {
		truth = true
	}
	return truth
}

func isCNV(v *vcf.Vcf) bool {
	var truth bool = false
	if strings.Contains(v.Info, "SVTYPE=CNV") {
		truth = true
	}
	return truth
}

func getSvEnd(v *vcf.Vcf) int64 {
	if !strings.Contains(v.Info, "END=") {
		log.Fatalf("Error: Vcf might not be from PBSV...")
	} else {
		words := strings.Split(v.Info, ";")
		for i := 0; i < len(words); i++ {
			if strings.Contains(words[i], "END=") {
				text := strings.Split(words[i], "END=")
				return common.StringToInt64(text[1])
			}
		}
	}
	return 0
}

func mkInversionNode(genome *SimpleGraph, v *vcf.Vcf, chr *fasta.Fasta) *Node {
	invSeq := make([]dna.Base, 0)
	invSeq = append(invSeq, chr.Seq[v.Pos:getSvEnd(v)]...)
	dna.ReverseComplement(invSeq)
	inversion := &Node{Id: uint32(len(genome.Nodes)), Name: chr.Name, Seq: invSeq, Prev: nil, Next: nil, Info: &Annotation{Allele: 1, Start: uint32(v.Pos), Variant: 5}}
	return inversion
}

func createSNP(sg *SimpleGraph, snp *vcf.Vcf, chr string) (*Node, *Node) {
	refAllele := &Node{Id: uint32(len(sg.Nodes)), Name: chr, Seq: dna.StringToBases(snp.Ref), Next: nil, Prev: nil}
	altAllele := &Node{Id: uint32(len(sg.Nodes) + 1), Name: chr, Seq: dna.StringToBases(snp.Alt), Next: nil, Prev: nil}
	AddNode(sg, refAllele)
	AddNode(sg, altAllele)
	return refAllele, altAllele
}

func NodeSplitByNs(sg *SimpleGraph, currMatch *Node, chr *fasta.Fasta, index int64, end int64) *Node {
	var inRegion bool = false
	var start int64 = 0
	for ; index < end; index++ {
		if dna.DefineBase(chr.Seq[start]) && inRegion == false {
			inRegion = false
			currMatch.Seq = chr.Seq[start:index]
			start = index
		} else if dna.DefineBase(chr.Seq[start]) && inRegion == true {
			newMatch := &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name}
			AddNode(sg, newMatch)
			AddEdge(currMatch, newMatch, 1)
			inRegion = true
			currMatch = newMatch
		}
	}
	return currMatch
}

type noNsBed struct {
	Start int32
	End   int32
}

func findUngap(fa *fasta.Fasta, index int64, end int64) []*noNsBed {
	var answer []*noNsBed
	var inRegion bool = false
	var startIndex int32 = 0
	for ; index < end; index++ {
		if dna.DefineBase(fa.Seq[index]) && inRegion == false {
			inRegion = true
			startIndex = int32(index)
		} else if !(dna.DefineBase(fa.Seq[index])) && inRegion == true {
			answer = append(answer, &noNsBed{Start: startIndex, End: int32(index)})
			inRegion = false
		}

	}
	if inRegion == true {
		answer = append(answer, &noNsBed{Start: int32(startIndex), End: int32(index)})
	}
	return answer
}

func createINS(sg *SimpleGraph, v *vcf.Vcf, chr string) *Node {
	curr := Node{Id: uint32(len(sg.Nodes)), Name: chr, Seq: dna.StringToBases(v.Alt)}
	AddNode(sg, &curr)
	return &curr
}

func isHaplotypeBlock(v *vcf.Vcf) bool {
	if strings.Compare(v.Format, "SVTYPE=SNP;INS") == 0 || strings.Compare(v.Format, "SVTYPE=SNP;DEL") == 0 || strings.Compare(v.Format, "SVTYPE=HAP") == 0 {
		return true
	} else {
		return false
	}
}

func createDEL(sg *SimpleGraph, v *vcf.Vcf, chr string) *Node {
	curr := Node{Id: uint32(len(sg.Nodes)), Name: chr, Seq: dna.StringToBases(v.Ref[1:len(v.Ref)])}
	AddNode(sg, &curr)
	return &curr
}

func goGraphSmithWatermanMap(gg *SimpleGraph, read *fastq.Fastq, seedHash map[uint64][]*SeedBed, seedLen int, stepSize int, m [][]int64, trace [][]rune) *sam.SamAln {
	var currBest sam.SamAln = sam.SamAln{QName: read.Name, Flag: 0, RName: "", Pos: 0, MapQ: 255, Cigar: []*cigar.Cigar{}, RNext: "*", PNext: 0, TLen: 0, Seq: read.Seq, Qual: string(read.Qual), Extra: ""}
	var leftAlignment, rightAlignment []*cigar.Cigar = []*cigar.Cigar{}, []*cigar.Cigar{}
	var i, minTarget int
	var minQuery int
	var leftScore, rightScore, seedScore, bestScore int64
	var leftPath, rightPath, bestPath []uint32

	var currScore, maxScore int64 = 0, 0
	for bases := 0; bases < len(read.Seq); bases++ {
		maxScore += HumanChimpTwoScoreMatrix[read.Seq[bases]][read.Seq[bases]]
	}
	ext := int(maxScore/600) + len(read.Seq)

	//seedStart := time.Now()
	var currRead *fastq.Fastq = nil
	var seeds []*SeedDev = findSeedsInMapDev(seedHash, read, seedLen, stepSize, true)
	extendSeedsDev(seeds, gg, read)
	revCompRead := fastq.Copy(read)
	fastq.ReverseComplement(revCompRead)
	var revCompSeeds []*SeedDev = findSeedsInMapDev(seedHash, revCompRead, seedLen, stepSize, false)
	extendSeedsDev(revCompSeeds, gg, revCompRead)
	seeds = append(seeds, revCompSeeds...)
	SortSeedDevByLen(seeds)
	//seedEnd := time.Now()
	//seedDuration := seedEnd.Sub(seedStart)
	//log.Printf("Have %d seeds, where best is of length %d, and it took %s\n", len(seeds), seeds[0].Length, seedDuration)

	//swStart := time.Now()
	for i = 0; i < len(seeds) && seedCouldBeBetter(int64(seeds[i].TotalLength), bestScore, maxScore, int64(len(read.Seq)), 100, 90, -196, -296); i++ {
		if seeds[i].PosStrand {
			currRead = read
		} else {
			currRead = revCompRead
		}
		seedScore = scoreSeed(seeds[i], currRead, HumanChimpTwoScoreMatrix)
		leftAlignment, leftScore, minTarget, minQuery, leftPath = AlignReverseGraphTraversal(gg.Nodes[seeds[i].TargetId], []dna.Base{}, int(seeds[i].TargetStart), []uint32{}, ext, currRead.Seq[:seeds[i].QueryStart], m, trace)
		rightAlignment, rightScore, _, rightPath = AlignTraversalFwd(gg.Nodes[seeds[i].TargetId], []dna.Base{}, int(seeds[i].TargetStart+seeds[i].Length), []uint32{}, ext, currRead.Seq[seeds[i].QueryStart+seeds[i].Length:], m, trace)
		currScore = leftScore + seedScore + rightScore
		if currScore > bestScore {
			bestScore = currScore
			if seeds[i].PosStrand {
				currBest.Flag = 0
			} else {
				currBest.Flag = 1
			}
			bestPath = CatPaths(AddPath(seeds[i].TargetId, leftPath), rightPath)

			currBest.RName = gg.Nodes[bestPath[0]].Name
			currBest.Pos = int64(minTarget)
			currBest.Cigar = cigar.CatCigar(cigar.AddCigar(leftAlignment, &cigar.Cigar{RunLength: int64(seeds[i].Length), Op: 'M'}), rightAlignment)
			currBest.Cigar = AddSClip(int(minQuery), len(currRead.Seq), currBest.Cigar)
			currBest.Extra = PathToString(bestPath, gg)
		}
	}
	//swEnd := time.Now()
	//duration := swEnd.Sub(swStart)
	//log.Printf("sw took %s\n", duration)
	return &currBest
}

/*func goGraphSmithWaterman(gg *SimpleGraph, read *fastq.Fastq, seedHash [][]*SeedBed, seedLen int, m [][]int64, trace [][]rune) *sam.SamAln {
	var currBest sam.SamAln = sam.SamAln{QName: read.Name, Flag: 0, RName: "", Pos: 0, MapQ: 255, Cigar: []*cigar.Cigar{}, RNext: "*", PNext: 0, TLen: 0, Seq: read.Seq, Qual: string(read.Qual), Extra: ""}
	var leftAlignment, rightAlignment []*cigar.Cigar = []*cigar.Cigar{}, []*cigar.Cigar{}
	var i, minTarget int
	var minQuery int
	var leftScore, rightScore, seedScore, bestScore int64
	var leftPath, rightPath, bestPath []uint32

	var currScore, maxScore int64 = 0, 0
	for bases := 0; bases < len(read.Seq); bases++ {
		maxScore += HumanChimpTwoScoreMatrix[read.Seq[bases]][read.Seq[bases]]
	}
	ext := int(maxScore/600) + len(read.Seq)

	var currRead *fastq.Fastq = nil
	var seeds []*SeedDev = findSeedsInSlice(seedHash, read, seedLen, true)
	revCompRead := fastq.Copy(read)
	fastq.ReverseComplement(revCompRead)
	var revCompSeeds []*SeedDev = findSeedsInSlice(seedHash, revCompRead, seedLen, false)
	seeds = append(seeds, revCompSeeds...)
	SortSeedDevByLen(seeds)

	for i = 0; i < len(seeds) && seedCouldBeBetter(seeds[i], bestScore, maxScore, int64(len(read.Seq)), 100, 90, -196, -296); i++ {
		if seeds[i].PosStrand {
			currRead = read
		} else {
			currRead = revCompRead
		}
		seedScore = scoreSeed(seeds[i], currRead, HumanChimpTwoScoreMatrix)
		leftAlignment, leftScore, minTarget, minQuery, leftPath = AlignReverseGraphTraversal(gg.Nodes[seeds[i].TargetId], []dna.Base{}, int(seeds[i].TargetStart), []uint32{}, ext, currRead.Seq[:seeds[i].QueryStart], m, trace)
		rightAlignment, rightScore, _, rightPath = AlignTraversalFwd(gg.Nodes[seeds[i].TargetId], []dna.Base{}, int(seeds[i].TargetStart+seeds[i].Length), []uint32{}, ext, currRead.Seq[seeds[i].QueryStart+seeds[i].Length:], m, trace)
		currScore = leftScore + seedScore + rightScore
		if currScore > bestScore {
			bestScore = currScore
			if seeds[i].PosStrand {
				currBest.Flag = 0
			} else {
				currBest.Flag = 1
			}
			bestPath = CatPaths(AddPath(seeds[i].TargetId, leftPath), rightPath)

			currBest.RName = gg.Nodes[bestPath[0]].Name
			currBest.Pos = int64(minTarget)
			currBest.Cigar = cigar.CatCigar(cigar.AddCigar(leftAlignment, &cigar.Cigar{RunLength: int64(seeds[i].Length), Op: 'M'}), rightAlignment)
			currBest.Cigar = AddSClip(minQuery, len(currRead.Seq), currBest.Cigar)
			currBest.Extra = PathToString(bestPath, gg)
		}
	}
	return &currBest
}*/

func wrapMap(ref *SimpleGraph, r *fastq.Fastq, seedHash map[uint64][]*SeedBed, seedLen int, stepSize int, c chan *sam.SamAln) {
	var mappedRead *sam.SamAln
	m, trace := swMatrixSetup(10000)
	mappedRead = goGraphSmithWatermanMap(ref, r, seedHash, seedLen, stepSize, m, trace)
	c <- mappedRead
	//log.Printf("%s\n", sam.SamAlnToString(mappedRead))
}

/*func wrap(ref *SimpleGraph, r *fastq.Fastq, seedHash [][]*SeedBed, seedLen int, c chan *sam.SamAln) {
	var mappedRead *sam.SamAln
	m, trace := swMatrixSetup(10000)
	mappedRead = goGraphSmithWaterman(ref, r, seedHash, seedLen, m, trace)
	c <- mappedRead
	//log.Printf("%s\n", sam.SamAlnToString(mappedRead))
}*/

func wrapNoChanMap(ref *SimpleGraph, r *fastq.Fastq, seedHash map[uint64][]*SeedBed, seedLen int, stepSize int) {
	var mappedRead *sam.SamAln
	startOne := time.Now()
	m, trace := swMatrixSetup(10000)
	stopOne := time.Now()
	durationOne := stopOne.Sub(startOne)
	startTwo := time.Now()
	mappedRead = goGraphSmithWatermanMap(ref, r, seedHash, seedLen, stepSize, m, trace)
	stopTwo := time.Now()
	durationTwo := stopTwo.Sub(startTwo)
	log.Printf("%s %s %s\n", sam.SamAlnToString(mappedRead), durationOne, durationTwo)
}

/*func wrapNoChan(ref *SimpleGraph, r *fastq.Fastq, seedHash [][]*SeedBed, seedLen int) {
	var mappedRead *sam.SamAln
	m, trace := swMatrixSetup(10000)
	mappedRead = goGraphSmithWaterman(ref, r, seedHash, seedLen, m, trace)
	log.Printf("%s\n", sam.SamAlnToString(mappedRead))
}*/

/*
func VcfNodesToGraph(sg *SimpleGraph, chr *fasta.Fasta, vcfs []*vcf.Vcf) *SimpleGraph {
	fasta.ToUpper(chr)
	var vcfFlag = -1
	var curr *Node
	var prev *Node
	var currMatch *Node
	var lastMatch *Node
	var refAllele *Node
	var altAllele *Node

	var idx int64 = 0
	var i int
	//var j int
	for i = 0; i < len(vcfs); i++ {
		if i > 0 {
			if vcfs[i-1].Pos == vcfs[i].Pos {
				log.Printf("Warning: there are two vcf records in %s, postion %d...\n", chr.Name, vcfs[i].Pos)
				fmt.Printf("%s\t%s\t%d\t%s\t%s\n%s\t%s\t%d\t%s\t%s\n", vcfs[i-1].Format, vcfs[i-1].Chr, vcfs[i-1].Pos, vcfs[i-1].Ref, vcfs[i-1].Alt, vcfs[i].Format, vcfs[i].Chr, vcfs[i].Pos, vcfs[i].Ref, vcfs[i].Alt)
				//
				//
				continue
			}
		}
		idx, vcfFlag, curr, prev, currMatch, lastMatch, refAllele, altAllele = devNodesToGraph(sg, chr, vcfFlag, vcfs[i], idx, curr, prev, currMatch, lastMatch, refAllele, altAllele)

		if idx > int64(len(chr.Seq)) {
			return sg
		}
	}
	lastNode := &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: chr.Seq[idx:]}
	AddNode(sg, lastNode)
	if vcfFlag == 1 {
		lastNode.Seq = lastNode.Seq[1:]
		AddEdge(refAllele, lastNode, 1)
		AddEdge(altAllele, lastNode, 1)
	} else {
		AddEdge(lastMatch, lastNode, 0.5)
		AddEdge(prev, lastNode, 1)
	}
	return sg
}

func NodesToGraph(sg *SimpleGraph, chr *fasta.Fasta, vcfFlag int, v *vcf.Vcf, idx int64, curr *Node, prev *Node, currMatch *Node, lastMatch *Node, refAllele *Node, altAllele *Node) (int64, int, *Node, *Node, *Node, *Node, *Node, *Node) {
	if chr.Name != v.Chr {
		log.Fatalf("Fasta %s does not match vcf name %s", chr.Name, v.Chr)
	}
	if vcfFlag == -1 && v.Pos == 1 {
		if strings.Compare(v.Format, "SVTYPE=SNP") == 0 {
			refAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
			AddNode(sg, refAllele)
			altAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt), Next: nil, Prev: nil}

			AddNode(sg, altAllele)
			vcfFlag = 1
			idx++
		}
		if strings.Compare(v.Format, "SVTYPE=INS") == 0 {
			lastMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
			AddNode(sg, lastMatch)
			prev = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt[1:len(v.Alt)]), Next: nil, Prev: nil}
			AddNode(sg, prev)
			AddEdge(lastMatch, prev, 0.5)
			idx = v.Pos
			vcfFlag = 2
		}
		if strings.Compare(v.Format, "SVTYPE=DEL") == 0 {
			lastMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt), Next: nil, Prev: nil}
			AddNode(sg, lastMatch)
			prev = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt[1:len(v.Ref)]), Next: nil, Prev: nil}
			AddNode(sg, prev)
			AddEdge(lastMatch, prev, 0.5)

			idx = v.Pos + int64(len(prev.Seq)) - 1
			vcfFlag = 3
		}
	} else {
		if v.Pos-1-idx > 0 {
			currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: chr.Seq[idx:v.Pos], Next: nil, Prev: nil}
			AddNode(sg, currMatch)

			if lastMatch != nil && vcfFlag != 1 {
				AddEdge(lastMatch, currMatch, 0.5)
			}
			if vcfFlag == 1 {
				AddEdge(refAllele, currMatch, 1)
				AddEdge(altAllele, currMatch, 1)
			}
			if vcfFlag == 2 || vcfFlag == 3 {
				AddEdge(prev, currMatch, 1)
			} //AddEdge(lastMatch, currMatch, 0.5)
			vcfFlag = 0
		}
		if strings.Compare(v.Format, "SVTYPE=SNP") == 0 {

			if vcfFlag == 0 {
				currMatch.Seq = currMatch.Seq[:len(currMatch.Seq)-1]
				refAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
				AddNode(sg, refAllele)
				altAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt), Next: nil, Prev: nil}
				AddNode(sg, altAllele)

				AddEdge(currMatch, refAllele, 0.5)
				AddEdge(currMatch, altAllele, 0.5)
			} else if vcfFlag == 1 {
				curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
				AddNode(sg, curr)
				AddEdge(refAllele, curr, 1)
				refAllele = curr

				curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt), Next: nil, Prev: nil}
				AddNode(sg, curr)
				AddEdge(altAllele, curr, 1)
				altAllele = curr
			} else if vcfFlag == 2 {

				refAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
				AddNode(sg, refAllele)

				altAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt), Next: nil, Prev: nil}
				AddNode(sg, altAllele)

				AddEdge(lastMatch, refAllele, 0.5)
				AddEdge(prev, altAllele, 1)
				curr = refAllele

			} else if vcfFlag == 3 {
				refAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
				AddNode(sg, refAllele)
				altAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt), Next: nil, Prev: nil}
				AddNode(sg, altAllele)

				AddEdge(prev, refAllele, 1)
				AddEdge(lastMatch, altAllele, 0.5)
				curr = refAllele
			} else {
				log.Fatal("Flag was not set up correctly")
			}
			vcfFlag = 1
			idx = v.Pos
		}
		if strings.Compare(v.Format, "SVTYPE=INS") == 0 {
			curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt[1:len(v.Alt)]), Next: nil, Prev: nil}

			if vcfFlag == 0 {
				AddNode(sg, curr)
				AddEdge(currMatch, curr, 0.5)

				vcfFlag = 2
			} else if vcfFlag == 1 {
				AddNode(sg, curr)
				AddEdge(altAllele, curr, 0.5)
				altAllele = curr
				//flag as SNP because we still need to connect two nodes to the next match
				vcfFlag = 1
			} else if vcfFlag == 2 {

				currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
				AddNode(sg, currMatch)
				curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt[1:len(v.Alt)]), Next: nil, Prev: nil}
				AddNode(sg, curr)
				AddEdge(lastMatch, currMatch, 0.5)
				AddEdge(prev, currMatch, 1)

				if int64(len(chr.Seq))-v.Pos == 0 {
					AddEdge(currMatch, curr, 1)
				} else {
					AddEdge(currMatch, curr, 0.5)
				}
				vcfFlag = 2
			} else if vcfFlag == 3 {
				currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt), Next: nil, Prev: nil}
				AddNode(sg, currMatch)
				curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref[1:len(v.Ref)]), Next: nil, Prev: nil}
				AddNode(sg, curr)
				AddEdge(lastMatch, currMatch, 0.5)
				AddEdge(prev, currMatch, 1)

				if int64(len(chr.Seq))-v.Pos == 0 {
					AddEdge(currMatch, curr, 1)
				} else {
					AddEdge(currMatch, curr, 0.5)
				}

				vcfFlag = 2
			} else {
				log.Fatal("Flag was not set up correctly")
			}
			idx = v.Pos
		}
		if strings.Compare(v.Format, "SVTYPE=DEL") == 0 {

			curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref[1:len(v.Ref)]), Next: nil, Prev: nil}

			if vcfFlag == 0 {
				AddNode(sg, curr)
				AddEdge(currMatch, curr, 0.5)
				vcfFlag = 3
			} else if vcfFlag == 1 {
				AddNode(sg, curr)
				AddEdge(refAllele, curr, 0.5)
				refAllele = curr
				vcfFlag = 1
			} else if vcfFlag == 2 {

				currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt), Next: nil, Prev: nil}
				AddNode(sg, currMatch)
				curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref[1:len(v.Ref)]), Next: nil, Prev: nil}
				AddNode(sg, curr)
				AddEdge(lastMatch, currMatch, 0.5)
				AddEdge(prev, currMatch, 1)

				if int64(len(chr.Seq))-v.Pos-1 == 0 {
					AddEdge(currMatch, curr, 1)
				} else {
					AddEdge(currMatch, curr, 0.5)
				}
				vcfFlag = 3
			} else if vcfFlag == 3 {
				currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
				AddNode(sg, currMatch)
				curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt[1:len(v.Alt)]), Next: nil, Prev: nil}
				AddNode(sg, curr)
				AddEdge(lastMatch, currMatch, 0.5)
				AddEdge(prev, currMatch, 1)

				if int64(len(chr.Seq))-v.Pos == 0 {
					AddEdge(currMatch, curr, 1)
				} else {
					AddEdge(currMatch, curr, 0.5)
				}
				vcfFlag = 3
			} else {
				log.Fatal("Flag was not set up correctly")
			}

			idx = v.Pos + int64(len(curr.Seq))
		}
		prev = curr
		lastMatch = currMatch

	}

	return idx, vcfFlag, curr, prev, currMatch, lastMatch, refAllele, altAllele

}

func devNodesToGraph(sg *SimpleGraph, chr *fasta.Fasta, vcfFlag int, v *vcf.Vcf, idx int64, curr *Node, prev *Node, currMatch *Node, lastMatch *Node, refAllele *Node, altAllele *Node) (int64, int, *Node, *Node, *Node, *Node, *Node, *Node) {
	//fasta.ToUpper(chr)
	if chr.Name != v.Chr {
		log.Fatalf("Fasta %s does not match vcf name %s", chr.Name, v.Chr)
	}
	if vcfFlag == -1 {
		if isSNP(v) {
			altAllele, refAllele = createSNP(sg, v, chr.Name)
			vcfFlag = 1
			idx++
		}
		if isINS(v) {
			lastMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
			AddNode(sg, lastMatch)
			prev = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt[1:len(v.Alt)]), Next: nil, Prev: nil}
			AddNode(sg, prev)
			AddEdge(lastMatch, prev, 0.5)
			idx = v.Pos
			vcfFlag = 2
		}
		if isDEL(v) {
			lastMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
			AddNode(sg, lastMatch)
			prev = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt)[1:], Next: nil, Prev: nil}
			AddNode(sg, prev)
			AddEdge(lastMatch, prev, 0.5)

			idx = v.Pos + int64(len(prev.Seq)) - 1
			vcfFlag = 3
		}
	} else {
		if v.Pos-idx > 1 {
			currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: chr.Seq[idx:v.Pos], Next: nil, Prev: nil}
			dna.AllToUpper(currMatch.Seq)
			AddNode(sg, currMatch)

			if lastMatch != nil && vcfFlag != 1 {
				AddEdge(lastMatch, currMatch, 0.5)
			}
			if vcfFlag == 1 {
				AddEdge(refAllele, currMatch, 1)
				AddEdge(altAllele, currMatch, 1)
			}
			if vcfFlag == 2 || vcfFlag == 3 {
				AddEdge(prev, currMatch, 1)
			} //AddEdge(lastMatch, currMatch, 0.5)
			vcfFlag = 0
		}
		if isSNP(v) {
			if vcfFlag == 0 {
				currMatch.Seq = currMatch.Seq[:len(currMatch.Seq)-1]
				altAllele, refAllele = createSNP(sg, v, chr.Name)
				AddEdge(currMatch, refAllele, 0.5)
				AddEdge(currMatch, altAllele, 0.5)
			} else if vcfFlag == 1 && (idx+1 == v.Pos) {
				refAllele.Seq = append(refAllele.Seq, dna.StringToBases(v.Ref)...)
				altAllele.Seq = append(altAllele.Seq, dna.StringToBases(v.Alt)...)
			} else if vcfFlag == 2 {
				refAllele, altAllele = createSNP(sg, v, chr.Name)
				AddEdge(lastMatch, refAllele, 0.5)
				AddEdge(prev, altAllele, 1)
				curr = refAllele
			} else if vcfFlag == 3 {
				refAllele, altAllele = createSNP(sg, v, chr.Name)
				AddEdge(prev, refAllele, 1)
				AddEdge(lastMatch, altAllele, 0.5)
				curr = refAllele
			} else {
				log.Fatal("Flag was not set up correctly")
			}
			vcfFlag = 1
			idx = v.Pos
		}
		if isINS(v) {
			curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt[1:len(v.Alt)]), Next: nil, Prev: nil}
			if vcfFlag == 0 {
				AddNode(sg, curr)
				AddEdge(currMatch, curr, 0.5)
			} else {
				if idx+1 == v.Pos {
					currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref), Next: nil, Prev: nil}
					AddNode(sg, currMatch)
					if vcfFlag == 1 {
						AddEdge(refAllele, currMatch, 1)
						AddEdge(altAllele, currMatch, 1)
					} else {
						AddEdge(prev, currMatch, 1)
					}
				}
				curr.Id++
				AddNode(sg, curr)
				AddEdge(currMatch, curr, 0.5)

			}
			vcfFlag = 2
			idx = v.Pos
		}
		if isDEL(v) {
			curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Ref[1:len(v.Ref)]), Next: nil, Prev: nil}
			if vcfFlag == 0 {
				AddNode(sg, curr)
				AddEdge(currMatch, curr, 0.5)
			} else {
				if idx+1 == v.Pos {
					//make a new curr match for the one base shared between ttwo genomes
					//using v.Alt because it is easy to grap that then risk going index out of bound
					currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(v.Alt), Next: nil, Prev: nil}
					AddNode(sg, currMatch)
					if vcfFlag == 1 {
						AddEdge(refAllele, currMatch, 1)
						AddEdge(altAllele, currMatch, 1)
					} else {
						AddEdge(prev, currMatch, 0.5)
					}
				}
				curr.Id++
				AddNode(sg, curr)
				AddEdge(currMatch, curr, 0.5)
			}
			vcfFlag = 3
			idx = v.Pos + int64(len(curr.Seq))
		}
		prev = curr
		lastMatch = currMatch
	}
	return idx, vcfFlag, curr, prev, currMatch, lastMatch, refAllele, altAllele
}

func SnpToNodesEdges(sg *SimpleGraph, snp *vcf.Vcf, lastNodeFlag int, chr *fasta.Fasta, curr *Node, prev *Node, currMatch *Node, lastMatch *Node, refAllele *Node, altAllele *Node) (*Node, *Node) {
	//Special case if last variant called was a SNP
	if lastNodeFlag == 1 {
		refAllele.Seq = append(refAllele.Seq, dna.StringToBases(snp.Ref)...)
		altAllele.Seq = append(altAllele.Seq, dna.StringToBases(snp.Alt)...)
		//curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(snp.Ref)}
		//AddNode(sg, curr)
		//AddEdge(refAllele, curr, 1)
		//refAllele = curr

		//curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(snp.Alt)}
		//AddNode(sg, curr)
		//AddEdge(altAllele, curr, 1)
		//altAllele = curr
	} else {
		refAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(snp.Ref)}
		AddNode(sg, refAllele)
		altAllele = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(snp.Alt)}
		AddNode(sg, altAllele)
		//last node was a match
		if lastNodeFlag == 0 {
			AddEdge(currMatch, refAllele, 0.5)
			AddEdge(currMatch, altAllele, 0.5)
		} //last node was an insertion
		if lastNodeFlag == 2 {
			AddEdge(lastMatch, refAllele, 0.5)
			AddEdge(prev, altAllele, 1)
		} //last node was a deletion
		if lastNodeFlag == 3 {
			AddEdge(prev, refAllele, 1)
			AddEdge(lastMatch, altAllele, 0.5)
		}
	}
	return refAllele, altAllele
}

func InsToNodesEdges(sg *SimpleGraph, ins *vcf.Vcf, vcfFlag int, chr *fasta.Fasta, curr *Node, prev *Node, currMatch *Node, lastMatch *Node, refAllele *Node, altAllele *Node) (*Node, int) {
	if vcfFlag == 0 || vcfFlag == 1 {
		curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(ins.Alt[1:])}
		AddNode(sg, curr)
		if vcfFlag == 0 {
			AddEdge(currMatch, curr, 0.5)
			vcfFlag = 2
		} else {
			//flag as SNP because we still need to connect two nodes to the next match
			AddEdge(altAllele, curr, 0.5)
			altAllele = curr
			vcfFlag = 1
		}
	} else if vcfFlag == 2 || vcfFlag == 3 {
		if vcfFlag == 2 {
			currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(ins.Ref)}
			curr = &Node{Id: uint32(len(sg.Nodes) + 1), Name: chr.Name, Seq: dna.StringToBases(ins.Alt[1:])}
		} else {
			currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(ins.Alt)}
			curr = &Node{Id: uint32(len(sg.Nodes) + 1), Name: chr.Name, Seq: dna.StringToBases(ins.Ref[1:])}
		}
		AddNode(sg, currMatch)
		AddNode(sg, curr)
		AddEdge(lastMatch, currMatch, 0.5)
		AddEdge(prev, currMatch, 1)
		if int64(len(chr.Seq))-ins.Pos == 0 {
			AddEdge(currMatch, curr, 1)
		} else {
			AddEdge(currMatch, curr, 0.5)
		}
		vcfFlag = 2
	}
	return curr, vcfFlag
}

func DelToNodesEdges(sg *SimpleGraph, del *vcf.Vcf, vcfFlag int, chr *fasta.Fasta, curr *Node, prev *Node, currMatch *Node, lastMatch *Node, refAllele *Node, altAllele *Node) (*Node, int) {
	if vcfFlag == 0 || vcfFlag == 1 {
		curr = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(del.Ref[1:])}
		AddNode(sg, curr)
		if vcfFlag == 0 {
			AddEdge(currMatch, curr, 0.5)
			vcfFlag = 3
		} else {
			//flag as SNP because we still need to connect two nodes to the next match
			AddEdge(refAllele, curr, 0.5)
			altAllele = curr
			vcfFlag = 1
		}
	} else if vcfFlag == 2 || vcfFlag == 3 {
		if vcfFlag == 2 {
			currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(del.Alt)}
			curr = &Node{Id: uint32(len(sg.Nodes) + 1), Name: chr.Name, Seq: dna.StringToBases(del.Ref[1:])}
		} else {
			currMatch = &Node{Id: uint32(len(sg.Nodes)), Name: chr.Name, Seq: dna.StringToBases(del.Ref)}
			curr = &Node{Id: uint32(len(sg.Nodes) + 1), Name: chr.Name, Seq: dna.StringToBases(del.Alt[1:])}
		}
		AddNode(sg, currMatch)
		AddNode(sg, curr)
		AddEdge(lastMatch, currMatch, 0.5)
		AddEdge(prev, currMatch, 1)
		if int64(len(chr.Seq))-del.Pos == 0 {
			AddEdge(currMatch, curr, 1)
		} else {
			AddEdge(currMatch, curr, 0.5)
		}
		vcfFlag = 3
	}
	return curr, vcfFlag
}*/
