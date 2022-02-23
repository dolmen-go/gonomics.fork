package convert

import (
	"strings"
	"testing"
)

var ChromNameTests = []struct {
	Ucsc    string
	Ensembl string
}{
	{"chr1", "1"},
	{"chr2", "2"},
	{"chr3", "3"},
	{"chr4", "4"},
	{"chr5", "5"},
	{"chr6", "6"},
	{"chr7", "7"},
	{"chr8", "8"},
	{"chr9", "9"},
	{"chr10", "10"},
	{"chr11", "11"},
	{"chr12", "12"},
	{"chr13", "13"},
	{"chr14", "14"},
	{"chr15", "15"},
	{"chr16", "16"},
	{"chr17", "17"},
	{"chr18", "18"},
	{"chr19", "19"},
	{"chr20", "20"},
	{"chr21", "21"},
	{"chr22", "22"},
	{"chrX", "X"},
	{"chrY", "Y"},
}

func TestEnsemblToUCSC(t *testing.T) {
	var curr string
	for _, v := range ChromNameTests {
		curr = EnsemblToUCSC(v.Ensembl)
		if strings.Compare(curr, v.Ucsc) != 0 {
			t.Errorf("Error in EnsemblToUcsc. Expected: %v. Found: %v.", v.Ucsc, curr)
		}
	}
}

func TestUCSCToEnsembl(t *testing.T) {
	var curr string
	for _, v := range ChromNameTests {
		curr = UCSCToEnsembl(v.Ucsc)
		if curr != v.Ensembl {
			t.Errorf("Error in UCSCToEnsembl. Expected: %v. Found: %v.", v.Ensembl, curr)
		}
	}
}
