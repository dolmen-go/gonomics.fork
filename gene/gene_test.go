package gene

import (
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fasta"
	"github.com/vertgenlab/gonomics/gtf"
	"testing"
)

func TestGtfToGoGene(t *testing.T) {
	g := gtf.Read("testdata/test.gtf")
	f := fasta.Read("testdata/test.fasta")
	var ok bool

	// Positive posStrand test
	answerPos := GtfToGene(g["test_gene_id"], f)

	correctBackup := goGeneBackup{
		startPos:     0,
		cdsStarts:    []int{2, 7, 11},
		cdsEnds:      []int{4, 9, 13},
		genomeSeq:    dna.StringToBases("ACATGCACCGTTAACG"),
		cdnaSeq:      dna.StringToBases("ATGCCGTAA"),
		featureArray: []Feature{-5, -5, 0, 1, 2, -1, -1, 3, 4, 5, -1, 6, 7, 8, -3, -3},
	}

	var correctPos Gene = Gene{
		id:           "test_gene_id",
		startPos:     0,
		posStrand:    true,
		cdsStarts:    []int{2, 7, 11},
		cdsEnds:      []int{4, 9, 13},
		genomeSeq:    dna.StringToBases("ACATGCACCGTTAACG"),
		cdnaSeq:      dna.StringToBases("ATGCCGTAA"),
		featureArray: []Feature{-5, -5, 0, 1, 2, -1, -1, 3, 4, 5, -1, 6, 7, 8, -3, -3},
		orig:         correctBackup,
	}

	if ok, _ = equal(answerPos, &correctPos); !ok {
		t.Error("ERROR: Trouble converting gtf to Gene on positive posStrand")
	}

	// Negative posStrand test
	answerNeg := GtfToGene(g["test_gene_id_negative"], f)

	var correctNeg Gene = Gene{
		id:           "test_gene_id_negative",
		startPos:     15,
		posStrand:    false,
		cdsStarts:    correctPos.cdsStarts,
		cdsEnds:      correctPos.cdsEnds,
		genomeSeq:    correctPos.genomeSeq,
		cdnaSeq:      correctPos.cdnaSeq,
		featureArray: correctPos.featureArray,
		orig:         correctBackup,
	}

	correctNeg.orig.startPos = 15

	if ok, _ = equal(answerNeg, &correctNeg); !ok {
		t.Error("ERROR: Trouble converting gtf to Gene on negative posStrand")
	}
}

func TestPositionConversion(t *testing.T) {
	g := gtf.Read("testdata/test.gtf")
	f := fasta.Read("testdata/test.fasta")

	// Positive Tests
	genePos := GtfToGene(g["test_gene_id"], f)
	var testsPass bool = true

	var answerPos, answerOffset int
	var err error

	answerPos, answerOffset, err = GenomicPosToCdna(genePos, 0)
	if answerPos != 0 || answerOffset != -2 || err != nil {
		testsPass = false
	}

	answerPos, answerOffset, err = GenomicPosToCdna(genePos, 3)
	if answerPos != 1 || answerOffset != 0 || err != nil {
		testsPass = false
	}

	answerPos, answerOffset, err = GenomicPosToCdna(genePos, 10)
	if answerPos != 5 || answerOffset != 1 || err != nil {
		testsPass = false
	}

	answerPos, answerOffset, err = GenomicPosToCdna(genePos, 14)
	if answerPos != 8 || answerOffset != 1 || err != nil {
		testsPass = false
	}

	answerPos, err = CdnaPosToGenomic(genePos, 0)
	if answerPos != 2 || err != nil {
		testsPass = false
	}

	answerPos, err = CdnaPosToGenomic(genePos, 3)
	if answerPos != 7 || err != nil {
		testsPass = false
	}

	answerPos, err = CdnaPosToGenomic(genePos, 6)
	if answerPos != 11 || err != nil {
		testsPass = false
	}

	answerPos, err = CdnaPosToGenomic(genePos, 8)
	if answerPos != 13 || err != nil {
		testsPass = false
	}

	if !testsPass {
		t.Error("ERROR: Trouble converting positions on positive posStrand")
	}

	// Negative Tests
	testsPass = true
	geneNeg := GtfToGene(g["test_gene_id_negative"], f)

	answerPos, answerOffset, err = GenomicPosToCdna(geneNeg, 0)
	if answerPos != 8 || answerOffset != 2 || err != nil {
		testsPass = false
	}

	answerPos, answerOffset, err = GenomicPosToCdna(geneNeg, 3)
	if answerPos != 7 || answerOffset != 0 || err != nil {
		testsPass = false
	}

	answerPos, answerOffset, err = GenomicPosToCdna(geneNeg, 10)
	if answerPos != 2 || answerOffset != 1 || err != nil {
		testsPass = false
	}

	answerPos, answerOffset, err = GenomicPosToCdna(geneNeg, 14)
	if answerPos != 0 || answerOffset != -1 || err != nil {
		testsPass = false
	}

	answerPos, err = CdnaPosToGenomic(geneNeg, 0)
	if answerPos != 13 || err != nil {
		testsPass = false
	}

	answerPos, err = CdnaPosToGenomic(geneNeg, 3)
	if answerPos != 8 || err != nil {
		testsPass = false
	}

	answerPos, err = CdnaPosToGenomic(geneNeg, 6)
	if answerPos != 4 || err != nil {
		testsPass = false
	}

	answerPos, err = CdnaPosToGenomic(geneNeg, 8)
	if answerPos != 2 || err != nil {
		testsPass = false
	}

	if !testsPass {
		t.Error("ERROR: Trouble converting positions on negative posStrand")
	}
}