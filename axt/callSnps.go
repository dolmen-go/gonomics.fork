package axt

import (
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/vcf"
	"strconv"
)

func AxtToVcf(axtFile *Axt) []*vcf.Vcf {
	var answer []*vcf.Vcf
	var curr *vcf.Vcf
	rCount := axtFile.RStart - 1
	qCount := axtFile.QStart - 1
	for i := 0; i < len(axtFile.RSeq); i++ {

		var infoTag string
		if axtFile.RSeq[i] != dna.Gap && axtFile.QSeq[i] != dna.Gap {
			rCount++
			qCount++
			//snp mismatch
			if axtFile.RSeq[i] != axtFile.QSeq[i] {
				infoTag = "POS=" + strconv.FormatInt(qCount, 10)
				curr = &vcf.Vcf{Chr: axtFile.RName, Pos: rCount, Id: axtFile.QName, Ref: dna.BaseToString(dna.ToUpper(axtFile.RSeq[i])), Alt: dna.BaseToString(dna.ToUpper(axtFile.QSeq[i])), Qual: 0, Filter: "PASS", Info: infoTag, Format: "SVTYPE=SNP", Unknown: "GT:DP:AD:RO:QR:AO:QA:GL"}
				//fmt.Println(snps[i].RefSub, snps[i].QuerySub)
				answer = append(answer, curr)
			}
		}
		//insertion in VCF record
		if axtFile.RSeq[i] != dna.Gap {
			var altTmp string
			qCount++
			//var refTmp string
			altTmp = dna.BaseToString(dna.ToUpper(axtFile.RSeq[i-1]))
			for j := i; j < len(axtFile.RSeq); j++ {
				if axtFile.RSeq[j] != dna.Gap {
					altTmp = altTmp + dna.BaseToString(dna.ToUpper(axtFile.QSeq[j]))
					qCount++
				} else {
					curr = &vcf.Vcf{Chr: axtFile.RName, Pos: rCount, Id: axtFile.QName, Ref: dna.BaseToString(dna.ToUpper(axtFile.RSeq[i-1])), Alt: altTmp, Qual: 0, Filter: "PASS", Info: infoTag, Format: "SVTYPE=INS", Unknown: "GT:DP:AD:RO:QR:AO:QA:GL"}
					answer = append(answer, curr)
					rCount++
					i = j
					break
				}
			}
		}
		//deleteion vcf record
		if axtFile.QSeq[i] != dna.Gap {
			//var refTmp string
			var altTmp string
			//rCount++
			altTmp = dna.BaseToString(dna.ToUpper(axtFile.RSeq[i-1]))
			for j := i; j < len(axtFile.RSeq); j++ {
				if axtFile.QSeq[j] != dna.Gap {
					altTmp = altTmp + dna.BaseToString(dna.ToUpper(axtFile.RSeq[j]))
				} else {
					curr = &vcf.Vcf{Chr: axtFile.RName, Pos: rCount, Id: axtFile.QName, Ref: altTmp, Alt: dna.BaseToString(dna.ToUpper(axtFile.RSeq[i-1])), Qual: 0, Filter: "PASS", Info: infoTag, Format: "SVTYPE=DEL", Unknown: "GT:DP:AD:RO:QR:AO:QA:GL"}
					rCount = rCount + int64(j)
					answer = append(answer, curr)
					i = j
					break
				}
			}
		}
	}
	return answer
}

func CallSnpsToVcf(axtList []*Axt) []*vcf.Vcf {
	var answer []*vcf.Vcf
	var curr []*vcf.Vcf
	for i := 0; i < len(axtList); i++ {
		curr = AxtToVcf(axtList[i])
		for _, j := range curr {
			answer = append(answer, j)
		}
	}
	return answer
}
