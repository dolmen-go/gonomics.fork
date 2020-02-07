package qDna

import (
	//"fmt"
	//"github.com/vertgenlab/gonomics/align"
	"github.com/vertgenlab/gonomics/dna"
	//"github.com/vertgenlab/gonomics/fasta"
	"github.com/vertgenlab/gonomics/fastq"
	"testing"
)

var convertTests = []struct {
	filename string // input
}{
	{"testdata/seq.fa"},
}

var bases = dna.StringToBases("ATGCGCG")
var testFasta = "testdata/seq.fa"
var printTest = QFrag{Seq: FromDna(bases), From: nil, Fwd: nil, Rev: nil}

/*
func TestPrint(t *testing.T) {
	printFasta := fasta.Read(testFasta)
	//fmt.Println(printFasta)
	seq1 := FromFasta(printFasta[0])
	seq2 := FromFasta(printFasta[1])
	fmt.Println("Sequence 1")
	for i := 0; i < len(seq1.Seq); i++ {
		fmt.Println(seq1.Seq[i])
	}
	fmt.Println("Sequence 2")
	for j := 0; j < len(seq2.Seq); j++ {
		fmt.Println(seq2.Seq[j])
	}
	score, alignment, _, maxI, _, _ := SmithWaterman(seq1.Seq, seq2.Seq, HumanChimpTwoScoreMatrix, -600)
	fmt.Println(score, alignment, maxI)
	fmt.Println(align.LocalView(mostLikelySeq(seq1.Seq), mostLikelySeq(seq2.Seq), alignment, maxI))
	//fmt.Println(LocalMaxScore(-1,-1100,-91))
	//fmt.Println("Testing Affine Gap: \n\n ")
	//blastz, alignment2 := AffineGap(seq1.Seq, seq2.Seq, HumanChimpTwoScoreMatrix, 600, 150)
	//fmt.Println(blastz, alignment2)
	//fmt.Println(align.View(mostLikelySeq(seq1.Seq), mostLikelySeq(seq2.Seq), alignment2))

}*/
/*
func TestAlign(t *testing.T) {
	fastaFile := FromFastaSlice(fasta.Read("testdata/multiCHr_test.fa"))
	fastq := fastq.Read("testdata/test.fastq")
	samTest := GSW(fastaFile, fastq)
	for i := 0; i < len(samTest); i++ {
		fmt.Println(samTest[i])
	}
}*/

func TestQDnaScoreLoop(t *testing.T) {
	_ = fastq.Read("testdata/CL12-3_16w_L001_R1.fastq.gz.sorted.nodups.mapped.R1.fastq")
	//var testScore float64
	//for i:= 0; i < len(fq[0].Seq); i++ {
	//	testScore += QDnaFasterScore(FromFastq(fq[0])[i], FromFastq(fq[0])[i], HumanChimpTwoScoreMatrix)
	//}
	//fmt.Println(testScore)

}

/*
func TestConvert(t *testing.T) {
	for _, test := range convertTests {
		actual := fasta.Read(test.filename)
		qfrag := FromFastaSlice(actual)
		qfragOutput := toFastaList(qfrag)
		if !fasta.AllAreEqual(qfragOutput, actual) {
			t.Errorf("The %s file did not convert properly.", test.filename)
		}

		//a := &QBase{A: 1, C: 0, G: 0, T: 0}
		//b := &QBase{A: 0, C: 1, G: 0, T: 0}
		//fmt.Println(QDnaScore(a, b, HumanChimpTwoScoreMatrix))
		//fmt.Println("DONE")
	}

}*/