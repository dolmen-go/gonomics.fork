package gtf

import (
	"errors"
	"fmt"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/interval"
	"github.com/vertgenlab/gonomics/vcf"
	"math"
	"reflect"
)

type Variant struct {
	vcf.Vcf
	RefId       string // e.g. NC_000023.10, LRG_199, NG_012232.1, NM_004006.2, LRG-199t1, NR_002196.1, NP_003997.1, etc.
	Gene        string
	PosStrand   bool
	NearestCDS  *CDS
	CDNAPos     int // 1-base
	AAPos       int // 1-base
	AARef       []dna.AminoAcid
	AAAlt       []dna.AminoAcid
	VariantType string // e.g. Silent, Missense, Nonsense, Frameshift, Intergenic, Intronic, Splice (1-2 away), FarSplice (3-10 away)
}

// GenesToIntervalTree builds a fractionally cascaded 2d interval tree for efficiently identifying genes that overlap a variant
func GenesToIntervalTree(genes map[string]*Gene) map[string]*interval.IntervalNode {
	MoveAllCanonicalToZero(genes)
	intervals := make([]interval.Interval, len(genes))
	var i int = 0
	for k, g := range genes {
		intervals[i] = g
		delete(genes, k)
		i++
	}
	return interval.BuildTree(intervals)
}

// VcfToVariant determines the effects of a variant on the cDNA and amino acid sequence by querying genes in the tree made by GenesToIntervalTree
// Note that if multiple genes are found to overlap a variant this function will return the variant based on the first queried gene and throw an error
func VcfToVariant(v *vcf.Vcf, tree map[string]*interval.IntervalNode, seq map[string][]dna.Base) (*Variant, error) {
	var answer *Variant
	var err error

	overlappingGenes := interval.Query(tree, v, "any")

	if len(overlappingGenes) > 1 {
		err = errors.New(fmt.Sprintf("Variant overlaps with multiple genes. Mutation will be based on first gene."))
	}

	var annotatingGene *Gene
	if len(overlappingGenes) > 0 {
		annotatingGene = overlappingGenes[0].(*Gene)
		answer = vcfToVariant(v, annotatingGene, seq)
	} else {
		answer = &Variant{Vcf: *v}
	}
	addVariantType(answer)
	return answer, err
}

// vcfToVariant is a helper function that annotates the Variant struct with information from the Vcf and Gtf input
func vcfToVariant(v *vcf.Vcf, gene *Gene, seq map[string][]dna.Base) *Variant {
	answer := new(Variant)
	answer.Vcf = *v
	answer.RefId = gene.Transcripts[0].TranscriptID
	answer.Gene = gene.GeneID
	answer.PosStrand = gene.Transcripts[0].Strand
	vcfCdsIntersect(v, gene, answer)
	if int(v.Pos) >= answer.NearestCDS.Start && int(v.Pos) <= answer.NearestCDS.End {
		findAAChange(answer, seq)
	}
	return answer
}

// vcfCdsIntersect annotates the Variant struct with the cDNA position of the vcf as well as the CDS nearest to the vcf
func vcfCdsIntersect(v *vcf.Vcf, gene *Gene, answer *Variant) {
	var cdsPos int
	var exon *Exon
	//TODO: this code may be able to be compressed
	if answer.PosStrand {
		for i := 0; i < len(gene.Transcripts[0].Exons); i++ {
			exon = gene.Transcripts[0].Exons[i]
			if exon.Cds != nil && int(v.Pos) > exon.Cds.End { // variant is further in gene
				cdsPos += exon.Cds.End - exon.Cds.Start + 1
				answer.NearestCDS = exon.Cds // Store most recent exon and move on // Catches variants past the last exon
			} else if exon.Cds != nil && int(v.Pos) <= exon.Cds.End { // variant is before end of this exon
				if int(v.Pos) < exon.Cds.Start { // Variant is NOT in CDS
					if exon.Cds.Prev == nil || exon.Cds.Start-int(v.Pos) < int(v.Pos)-gene.Transcripts[0].Exons[i-1].Cds.Start {
						answer.NearestCDS = exon.Cds
					} else {
						answer.NearestCDS = gene.Transcripts[0].Exons[i-1].Cds
					}
					break
				}
				cdsPos += int(v.Pos) - exon.Cds.Start + 1
				answer.CDNAPos = cdsPos
				answer.NearestCDS = exon.Cds
			}
		}
	} else {
		for i := 0; i < len(gene.Transcripts[0].Exons); i++ {
			exon = gene.Transcripts[0].Exons[len(gene.Transcripts[0].Exons)-1-i]
			if exon.Cds != nil && int(v.Pos) < exon.Cds.Start { // variant is further in gene
				cdsPos += exon.Cds.End - exon.Cds.Start + 1
				answer.NearestCDS = exon.Cds // Store most recent exon and move on // Catches variants past the last exon
			} else if exon.Cds != nil && int(v.Pos) >= exon.Cds.Start { // variant is before end of this exon
				if int(v.Pos) > exon.Cds.End { // Variant is NOT in CDS
					if exon.Cds.Next == nil || int(v.Pos)-exon.Cds.End < gene.Transcripts[0].Exons[len(gene.Transcripts[0].Exons)-1-i+1].Cds.Start-int(v.Pos) {
						answer.NearestCDS = exon.Cds
					} else {
						answer.NearestCDS = gene.Transcripts[0].Exons[len(gene.Transcripts[0].Exons)-1-i+1].Cds
					}
					break
				}
				// Variant IS in CDS
				cdsPos += exon.Cds.End - int(v.Pos) + 1
				answer.CDNAPos = cdsPos
				answer.NearestCDS = exon.Cds
			}
		}
	}

}

// findAAChange annotates the Variant struct with the amino acids changed by a given variant
func findAAChange(variant *Variant, seq map[string][]dna.Base) {
	var refBases = make([]dna.Base, 0)
	var altBases = make([]dna.Base, 0)
	var seqPos int = int(variant.Pos) - 1
	var currCDS *CDS = variant.NearestCDS
	if variant.PosStrand {
		seqPos -= determineFrame(variant)
		for ; seqPos < int(variant.Pos-1); seqPos++ {
			if seqPos < currCDS.Start-1 {
				seqPos = currCDS.Prev.End - 1
				currCDS = currCDS.Prev
			} else if seqPos > currCDS.End-1 {
				seqPos = currCDS.Next.Start - 1
				currCDS = currCDS.Next
			}
			refBases = append(refBases, seq[variant.Chr][seqPos])
			altBases = append(altBases, seq[variant.Chr][seqPos])
		}
		refBases = append(refBases, dna.StringToBases(variant.Ref)...)
		altBases = append(altBases, dna.StringToBases(variant.Alt)...)
		seqPos += len(dna.StringToBases(variant.Ref))
		var offset int
		for offset = 0; len(refBases)%3 != 0; offset++ {
			if seqPos+offset > currCDS.End-1 {
				seqPos = currCDS.Next.Start - 1
				currCDS = currCDS.Next
			}
			refBases = append(refBases, seq[variant.Chr][seqPos+offset])
		}
		for offset = 0; len(altBases)%3 != 0; offset++ {
			if seqPos+offset > currCDS.End-1 {
				seqPos = currCDS.Next.Start - 1
				currCDS = currCDS.Next
			}
			altBases = append(altBases, seq[variant.Chr][seqPos+offset])
		}
		variant.AARef = dna.TranslateSeq(refBases)
		variant.AAAlt = dna.TranslateSeq(altBases)
		variant.AAPos = int(math.Round((float64(variant.CDNAPos) / 3) + 0.4)) // Add 0.4 so pos will always round up
	} else {
		seqPos += determineFrame(variant)
		for ; seqPos > int(variant.Pos-1); seqPos-- {
			if seqPos < currCDS.Start-1 {
				seqPos = currCDS.Prev.End - 1
				currCDS = currCDS.Prev
			} else if seqPos > currCDS.End-1 {
				seqPos = currCDS.Next.Start - 1
				currCDS = currCDS.Next
			}
			refBases = append(refBases, seq[variant.Chr][seqPos])
			altBases = append(altBases, seq[variant.Chr][seqPos])
		}
		refBases = append(refBases, reverse(dna.StringToBases(variant.Ref))...)
		altBases = append(altBases, reverse(dna.StringToBases(variant.Alt))...)
		seqPos -= len(dna.StringToBases(variant.Ref))
		var offset int
		for offset = 0; len(refBases)%3 != 0; offset++ {
			if seqPos-offset < currCDS.Start-1 {
				seqPos = currCDS.Prev.End - 1
				currCDS = currCDS.Prev
			}
			refBases = append(refBases, seq[variant.Chr][seqPos-offset])
		}
		for offset = 0; len(altBases)%3 != 0; offset++ {
			if seqPos-offset < currCDS.Start-1 {
				seqPos = currCDS.Prev.End - 1
				currCDS = currCDS.Prev
			}
			altBases = append(altBases, seq[variant.Chr][seqPos-offset])
		}
		dna.Complement(refBases)
		dna.Complement(altBases)
		variant.AARef = dna.TranslateSeq(refBases)
		variant.AAAlt = dna.TranslateSeq(altBases)
		variant.AAPos = int(math.Round((float64(variant.CDNAPos) / 3) + 0.4)) // Add 0.4 so pos will always round up
	}
}

// addVariantType annotates the Variant struct with the VariantType
// Valid types include: Silent, Missense, Nonsense, Frameshift, Intergenic, Intronic, Splice, FarSplice
// Splice is defined as 1-2 bases away from intron-exon border, Farsplice is 3-10 bases away from intron-exon border
func addVariantType(v *Variant) {
	cdsDist := getCdsDist(v)
	switch {
	case v.Gene == "":
		v.VariantType = "Intergenic"
	case cdsDist > 0 && cdsDist <= 2:
		v.VariantType = "Splice"
	case cdsDist > 0 && cdsDist <= 10:
		v.VariantType = "FarSplice"
	case v.AARef == nil:
		v.VariantType = "Intronic"
	case isFrameshift(v):
		v.VariantType = "Frameshift"
	case isNonsense(v):
		v.VariantType = "Nonsense"
	case !reflect.DeepEqual(v.AARef, v.AAAlt):
		v.VariantType = "Missense"
	case reflect.DeepEqual(v.AARef, v.AAAlt):
		v.VariantType = "Silent"
	default:
		v.VariantType = "Unrecognized"
	}
}

// reverse reverses the order of a slice of dna.Base
// e.g. [0 1 2] -> [2 1 0]
func reverse(s []dna.Base) []dna.Base {
	for i := 0; i < len(s)/2; i++ {
		s[i], s[len(s)-1-i] = s[len(s)-1-i], s[i]
	}
	return s
}

// determineFrame will determine the position of the variant in a codon
// This is used to determine how many bases before the variant must be retrieved to get the full codon
func determineFrame(v *Variant) int {
	if v.PosStrand {
		return (int(v.Pos)-v.NearestCDS.Start)%3 + ((3 - v.NearestCDS.Frame) % 3)
	} else {
		return (v.NearestCDS.End-int(v.Pos))%3 + ((3 - v.NearestCDS.Frame) % 3)
	}
}

// getCdsDist determines the distance of the variant from the nearest CDS
// Returns 0 if the variant is inside the CDS
func getCdsDist(v *Variant) int {
	switch {
	case int(v.Pos) >= v.NearestCDS.Start && int(v.Pos) <= v.NearestCDS.End: // Variant is in CDS
		return 0

	case int(v.Pos) < v.NearestCDS.Start: // Variant is before nearest CDS
		return v.NearestCDS.Start - int(v.Pos)

	default:
		return int(v.Pos) - v.NearestCDS.End // Variant is after nearest CDS
	}
}

// isFrameshift returns true if the variant shifts the reading frame
func isFrameshift(v *Variant) bool {
	refBases := dna.StringToBases(v.Ref)
	altBases := dna.StringToBases(v.Alt)

	start := int(v.Pos)
	refEnd := start + len(refBases) - 1
	altEnd := start + len(altBases) - 1

	var refBasesInCds int
	var altBasesInCds int

	var startOffset int
	if start < v.NearestCDS.Start {
		startOffset = v.NearestCDS.Start - start
	}

	if refEnd <= v.NearestCDS.End {
		refBasesInCds = len(refBases) - startOffset
	} else if refEnd > v.NearestCDS.End {
		refBasesInCds = len(refBases) - (refEnd - v.NearestCDS.End) - startOffset
	}
	if altEnd <= v.NearestCDS.End {
		altBasesInCds = len(altBases) - startOffset
	} else if altEnd > v.NearestCDS.End {
		altBasesInCds = len(altBases) - (altEnd - v.NearestCDS.End) - startOffset
	}

	shift := altBasesInCds - refBasesInCds
	return shift%3 != 0
}

// isNonsense returns true if the variant creates a premature stop codon
func isNonsense(v *Variant) bool {
	for _, val := range v.AAAlt {
		if val == dna.Stop {
			return true
		}
	}
	return false
}
