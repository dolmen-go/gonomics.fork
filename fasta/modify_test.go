package fasta

import (
	"fmt"
	"github.com/vertgenlab/gonomics/dna"
	"testing"
)

var seqThreeA = dna.StringToBases("ACGTacgTCATCATCATTACTACTAC")
var seqThreeB = dna.StringToBases("acgtACGTACGT")
var seqThreeC = dna.StringToBases("ACGTACGTACGTT")
var rcSeqThreeA = dna.StringToBases("GTAGTAGTAATGATGATGAcgtACGT")
var rcSeqThreeB = dna.StringToBases("ACGTACGTacgt")
var rcSeqThreeC = dna.StringToBases("AACGTACGTACGT")
var allRevCompTests = []struct {
	input    []*Fasta
	expected []*Fasta
}{
	{[]*Fasta{{"apple", seqThreeA}, {"banana", seqThreeB}, {"carrot", seqThreeC}}, []*Fasta{{"apple", rcSeqThreeA}, {"banana", rcSeqThreeB}, {"carrot", rcSeqThreeC}}},
}

func TestReverseComplement(t *testing.T) {
	for _, test := range allRevCompTests {
		ReverseComplementAll(test.input)
		if !AllAreEqual(test.input, test.expected) {
			t.Errorf("Expected reverse complement to give %v, but got %v.", test.input, test.expected)
		}
	}
}

func TestChangeName(t *testing.T) {
	fa := Read("testdata/testOne.fa")
	newFa := Read("testdata/testOne.fa")
	ChangePrefix(newFa, "Library")
	for i := 0; i < len(newFa); i++ {
		fmt.Printf("Name of fasta record: %s\n", newFa[i].Name)
		if newFa[i].Name == fa[i].Name {
			t.Errorf("Fasta record change was not successful")
		}
	}
}
