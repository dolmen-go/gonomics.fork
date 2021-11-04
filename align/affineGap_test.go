package align

import (
	"fmt"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fasta"
	"testing"
)

var affineAlignTests = []struct {
	seqOne string
	seqTwo string
	aln    string
}{
	{"ACGT", "ACGT", "ACGT\nACGT\n"},
	{"ACGT", "CGT", "ACGT\n-CGT\n"},
	{"ACGT", "ACG", "ACGT\nACG-\n"},
	{"CGT", "ACGT", "-CGT\nACGT\n"},
	{"ACG", "ACGT", "ACG-\nACGT\n"},
	{"AGT", "ACGT", "A-GT\nACGT\n"},
	{"ACT", "ACGT", "AC-T\nACGT\n"},
	{"CGCGCGCGCG", "CGCGCGTTTTCGCG", "CGCGCG----CGCG\nCGCGCGTTTTCGCG\n"},
	{"CGCGCGCGCG", "CGAAAACGCGTTTTCGCG", "CG----CGCG----CGCG\nCGAAAACGCGTTTTCGCG\n"},
}

var affineAlignChunkTests = []struct {
	seqOne string
	seqTwo string
	aln    string
}{
	{"ACG", "ACG", "ACG\nACG\n"},
	{"ACG", "CCG", "ACG\nCCG\n"},
	{"TTGTTCTTCTTCTTC", "TTGTTCTTCTTATTATTATTCTTC", "TTGTTCTTC---------TTCTTC\nTTGTTCTTCTTATTATTATTCTTC\n"},
	{"ACAACAATAAGAAAAACAAAA", "ACAACAAAAACAAAA", "ACAACAATAAGAAAAACAAAA\nACAACA------AAAACAAAA\n"},
}

var affineScoreTests = []struct {
	filename string
}{
	{"testdata/tmp.aln.fa"},
	{"testdata/hand.aln.fa"},
}

func TestAffineGap(t *testing.T) {
	for _, test := range affineAlignTests {
		basesOne := dna.StringToBases(test.seqOne)
		basesTwo := dna.StringToBases(test.seqTwo)
		_, cigar := AffineGap_highMem(basesOne, basesTwo, DefaultScoreMatrix, -400, -30)
		prettyAlignment := View(basesOne, basesTwo, cigar)
		if prettyAlignment != test.aln {
			t.Errorf("The affine gap alignment of %s and %s gave %s, but %s was expected", test.seqOne, test.seqTwo, prettyAlignment, test.aln)
		}
	}
}

func TestAffineGap_lowMem(t *testing.T) {
	for _, test := range affineAlignTests {
		basesOne := dna.StringToBases(test.seqOne)
		basesTwo := dna.StringToBases(test.seqTwo)
		score_highest_highMem, route_highMem := AffineGap_highMem(basesOne, basesTwo, DefaultScoreMatrix, -400, -30)
		score_highest_lowMem, route_lowMem := AffineGap(basesOne, basesTwo, DefaultScoreMatrix, -400, -30)
		score_highest_lowMem_customizeCheckersize, route_lowMem_customizeCheckersize := AffineGap_customizeCheckersize(basesOne, basesTwo, DefaultScoreMatrix, -400, -30, 3, 3)
		if score_highest_lowMem != score_highest_highMem {
			t.Errorf("score_highest_lowMem gave %d, but expected it to be the same as score_highest_highMem which gave %d\n", score_highest_lowMem, score_highest_highMem)
		}
		if score_highest_lowMem_customizeCheckersize != score_highest_highMem {
			t.Errorf("score_highest_lowMem_customizeCheckersize gave %d, but expected it to be the same as score_highest_highMem which gave %d\n", score_highest_lowMem_customizeCheckersize, score_highest_highMem)
		}
		for i := range route_lowMem {
			if route_lowMem[i] != route_highMem[i] {
				t.Errorf("route_lowMem gave %v, but expected it to be the same as route_highMem which gave %v\n", route_lowMem, route_highMem)
			}
		}
		for i := range route_lowMem_customizeCheckersize {
			if route_lowMem_customizeCheckersize[i] != route_highMem[i] {
				t.Errorf("route_lowMem_customizeCheckersize gave %v, but expected it to be the same as route_highMem which gave %v\n", route_lowMem_customizeCheckersize, route_highMem)
			}
		}
	}
}

func TestAffineGapChunk(t *testing.T) {
	for _, test := range affineAlignChunkTests {
		basesOne := dna.StringToBases(test.seqOne)
		basesTwo := dna.StringToBases(test.seqTwo)
		_, cigar := AffineGapChunk(basesOne, basesTwo, DefaultScoreMatrix, -400, -30, 3)
		prettyAlignment := View(basesOne, basesTwo, cigar)
		if prettyAlignment != test.aln {
			t.Errorf("The affine gap chunk alignment of %s and %s gave %s, but %s was expected", test.seqOne, test.seqTwo, prettyAlignment, test.aln)
		}
	}
}

func TestAffineGapMulti(t *testing.T) {
	for _, test := range affineAlignTests {
		basesOne := dna.StringToBases(test.seqOne)
		basesTwo := dna.StringToBases(test.seqTwo)
		one := []fasta.Fasta{{Name: "one", Seq: basesOne}}
		two := []fasta.Fasta{{Name: "two", Seq: basesTwo}}
		_, cigar := multipleAffineGap(one, two, DefaultScoreMatrix, -400, -30)
		answer := mergeMultipleAlignments(one, two, cigar)
		pretty := fmt.Sprintf("%s\n%s\n", dna.BasesToString(answer[0].Seq), dna.BasesToString(answer[1].Seq))
		if pretty != test.aln {
			t.Errorf("The affine gap alignment of %s and %s gave %s, but %s was expected.  Cigar was: %v", test.seqOne, test.seqTwo, pretty, test.aln, cigar)
		}
	}
}

func TestAffineScore(t *testing.T) {
	for _, test := range affineScoreTests {
		aln := fasta.Read(test.filename)
		//common.ExitIfError(err)
		score, err := scoreAffineAln(aln[0], aln[1], DefaultScoreMatrix, -10, -1)
		common.ExitIfError(err)
		fmt.Printf("Score of alignment in %s is: %d\n", test.filename, score)
	}
}
