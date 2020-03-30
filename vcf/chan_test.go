package vcf

import "testing"

func TestReadToChan(t *testing.T) {
	var actual []*Vcf
	for _, test := range readWriteTests {
		tempFile := test.filename //+ ".tmp"

		stream := ReadToChan(tempFile)

		for vcf := range(stream) {
			actual = append(actual, vcf)
		}

		PrintVcf(actual)

	}
}