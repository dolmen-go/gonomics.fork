package fastq

import (
	"fmt"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fileio"
)

type SingleCellPair struct {
	Reads PairedEnd
	Bx  []dna.Base
	Umi []dna.Base
}

func ReadToChanSingleCellPair(fileOne string, fileTwo string, barcodeLength int, umiLength int, tenXg chan<- SingleCellPair) {
	fwd, rev := fileio.EasyOpen(fileOne), fileio.EasyOpen(fileTwo)
	defer fwd.Close()
	defer rev.Close()

	for curr, done := NextFastqPair(fwd, rev); !done; curr, done = NextFastqPair(fwd, rev) {
		tenXg <- PairedEndToSingleCellPair(curr, barcodeLength, umiLength)
	}
	close(tenXg)
}

//10x linked-read library construct are modifications only on forward or read one, read two remains unchanged:
//1) first 16 bases: 10x barcode labeled either Bx or Rx tags
//2) Next 6 bases is a 6 base Umi
//3) finally adaptors and genomic sequence
//10x scRNA-seq is also a 16bp barcode (for the cell), and then UMIs of various lengths
//TODO: consider trimming off adapter sequences too
func PairedEndToSingleCellPair(fqPair PairedEnd, barcodeLength int, umiLength int) SingleCellPair {
	sc := SingleCellPair{Reads: fqPair, Bx: GetBarcode(fqPair.Fwd, 0, barcodeLength), Umi: GetBarcode(fqPair.Fwd, barcodeLength, barcodeLength+umiLength)}
	sc.Reads.Fwd = TrimFastq(fqPair.Fwd, barcodeLength+umiLength, len(fqPair.Fwd.Seq))
	bxTag := fmt.Sprintf("BX:%s", dna.BasesToString(sc.Bx))
	sc.Reads.Fwd.Name = fmt.Sprintf("%s_%s", fqPair.Fwd.Name, bxTag)
	sc.Reads.Rev.Name = fmt.Sprintf("%s_%s", fqPair.Rev.Name, bxTag)
	return sc
}

//TrimFastq trims fastq bases and quals given a start and end position zero base
func TrimFastq(fq *Fastq, start int, end int) *Fastq {
	fq.Seq = fq.Seq[start:end]
	fq.Qual = fq.Qual[start:end]
	return fq
}

//GetBarcode copies a subset of the given fastq record, at start and and end sites, zero base
//could be a UMI or 10x barcode
func GetBarcode(fq *Fastq, start int, end int) []dna.Base {
	answer := make([]dna.Base, end-start)
	copy(answer, fq.Seq[start:end])
	return answer
}

//UmiCollision takes a SingleCellPair and consults a map (with the string as the concatenated barcode and UMI strings), \\
//Returns true if the SingleCellPair UMI and barcode are found in the map, false otherwise. If a pair is not found in the map, the map is updated to include the current pair.
func UmiCollision(sc SingleCellPair, umiMap map[string]int) bool {
	currUmi := fmt.Sprintf("%s%s", sc.Bx, sc.Umi)
	count, found := umiMap[currUmi]
	if !found {
		umiMap[currUmi] = 1
	} else {
		umiMap[currUmi] = count + 1
	}
	return found
}

func fastqStats(fq *Fastq) string {
	return fmt.Sprintf("%s\t%d\n%s\n", fq.Name, len(fq.Seq), dna.BasesToString(fq.Seq))
}

func SingleCellPairPrettyString(lr *SingleCellPair) string {
	return fmt.Sprintf("FwdRead\t%s\nFwdSeq\t%s\nFwdQual\t%s\n10xG\t%s\nUmi\t%s\nRevRead\t%s\nRevSeq\t%s\nRevQual\t%s\n", lr.Reads.Fwd.Name, dna.BasesToString(lr.Reads.Fwd.Seq), QualString(lr.Reads.Fwd.Qual), dna.BasesToString(lr.Bx), dna.BasesToString(lr.Umi), lr.Reads.Rev.Name, dna.BasesToString(lr.Reads.Rev.Seq), QualString(lr.Reads.Rev.Qual))
}

//TODO: add a more stringent test to make sure we are trimming the reads correctly
func isLinkedRead(lr SingleCellPair) bool {
	if len(lr.Bx) != 16 {
		return false
	}
	if len(lr.Umi) != 6 {
		return false
	}
	return true
}
