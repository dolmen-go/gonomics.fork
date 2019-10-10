package bed

import (
	"strings"
	"sort"
	"github.com/vertgenlab/gonomics/common"
)

func Sort(bedFile []*Bed) {
	sort.Slice(bedFile, func(i, j int) bool { return Compare(bedFile[i], bedFile[j]) == -1 })
}

func MergeBeds(bedFile []*Bed) []*Bed {
	Sort(bedFile)
	var i, j int
	for i = 0; i < len(bedFile)-1; {
		if !(common.MaxInt64(bedFile[i].ChromStart, bedFile[i+1].ChromStart) < common.MinInt64(bedFile[i].ChromEnd, bedFile[i+1].ChromEnd)) || strings.Compare(bedFile[i].Chrom, bedFile[i+1].Chrom) != 0 {
			i++
		} else {
			bedFile[i].ChromStart, bedFile[i].ChromEnd, bedFile[i].Score = common.MinInt64(bedFile[i].ChromStart, bedFile[i+1].ChromStart), common.MaxInt64(bedFile[i].ChromEnd, bedFile[i+1].ChromEnd), bedFile[i].Score+bedFile[i+1].Score
			for j = i + 1; j < len(bedFile)-1; j++ {
				bedFile[j] = bedFile[j+1]
			}
			bedFile = bedFile[:len(bedFile)-1]
		}
	}
	return bedFile
}

func Compare(a *Bed, b *Bed) int {
	chromComp := strings.Compare(a.Chrom, b.Chrom)
	if chromComp != 0 {
		return chromComp
	}
	if a.ChromStart < b.ChromStart {
		return -1
	}
	if a.ChromStart > b.ChromStart {
		return 1
	}
	if a.ChromEnd < b.ChromEnd {
		return -1
	}
	if a.ChromEnd > b.ChromEnd {
		return 1
	}
	return 0
}

func AllAreEqual(a []*Bed, b []*Bed) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

func Equal(a *Bed, b *Bed) bool {
	if strings.Compare(a.Chrom, b.Chrom) != 0 {
		return false
	}
	if a.ChromStart != b.ChromStart {
		return false
	}
	if a.ChromEnd != b.ChromEnd {
		return false
	}
	return true
}
