package axt

import (
	"fmt"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fileio"
	"log"
	"strconv"
	"strings"
)

// Naming convention is hard here because UCSC website does not
// match the UCSC Kent source tree.
type Axt struct {
	RName      string
	RStart     int64
	REnd       int64
	QName      string
	QStart     int64
	QEnd       int64
	QStrandPos bool // true is positive strand, false is negative strand
	Score      int64
	RSeq       []dna.Base
	QSeq       []dna.Base
}

func Read(filename string) []*Axt {
	var answer []*Axt
	var header, rSeq, qSeq, blank string
	var err, startErr, endErr error
	var hDone, rDone, qDone, bDone bool
	var words []string

	file := fileio.EasyOpen(filename)
	defer file.Close()
	for header, hDone = fileio.EasyNextRealLine(file); !hDone; header, hDone = fileio.EasyNextRealLine(file) {
		rSeq, rDone = fileio.EasyNextRealLine(file)
		qSeq, qDone = fileio.EasyNextRealLine(file)
		blank, bDone = fileio.EasyNextRealLine(file)
		if rDone || qDone || bDone {
			log.Fatalf("Error: lines in %s, must be a multiple of four\n", filename)
		}
		if blank != "" {
			log.Fatalf("Error: every fourth line in %s should be blank\n", filename)
		}

		words = strings.Split(header, " ")
		if len(words) != 9 {
			log.Fatalf("Error: sequences in %s should be the same length\n", header)
		}

		curr := Axt{}
		curr.RName = words[1]
		curr.RStart, startErr = strconv.ParseInt(words[2], 10, 64)
		curr.REnd, endErr = strconv.ParseInt(words[3], 10, 64)
		if startErr != nil || endErr != nil {
			log.Fatalf("Error: trouble parsing reference start and end in %s\n", header)
		}
		curr.QName = words[4]
		curr.QStart, startErr = strconv.ParseInt(words[5], 10, 64)
		curr.QEnd, endErr = strconv.ParseInt(words[6], 10, 64)
		if startErr != nil || endErr != nil {
			log.Fatalf("Error: trouble parsing query start and end in %s\n", header)
		}
		switch words[7] {
		case "+":
			curr.QStrandPos = true
		case "-":
			curr.QStrandPos = false
		default:
			log.Fatalf("Error: did not recognize strand in %s\n", header)
		}
		curr.Score, err = strconv.ParseInt(words[8], 10, 64)
		if err != nil {
			log.Fatalf("Error: trouble parsing the score in %s\n", header)
		}
		curr.RSeq = dna.StringToBases(rSeq)
		curr.QSeq = dna.StringToBases(qSeq)

		answer = append(answer, &curr)
	}
	return answer
}

func ReadToChan(reader *fileio.EasyReader, answer chan<- *Axt) {
	for data, err := NextAxt(reader); !err; data, err = NextAxt(reader) {
		answer <- data
	}
	close(answer)
}

func NextAxt(reader *fileio.EasyReader) (*Axt, bool) {
	header, hDone := fileio.EasyNextRealLine(reader)
	rSeq, rDone := fileio.EasyNextRealLine(reader)
	qSeq, qDone := fileio.EasyNextRealLine(reader)
	blank, bDone := fileio.EasyNextRealLine(reader)
	if blank != "" {
		log.Fatalf("Error: every fourth line should be blank\n")
	}
	if hDone || rDone || qDone || bDone {
		return nil, true
	}
	return axtHelper(header, rSeq, qSeq, blank), false
}

func axtHelper(header string, rSeq string, qSeq string, blank string) *Axt {
	var words []string = strings.Split(header, " ")
	if len(words) != 9 || rSeq == "" || qSeq == "" {
		log.Fatalf("Error: missing fields in header or sequences\n")
	}
	var answer *Axt = &Axt{
		RName:      words[1],
		RStart:     common.StringToInt64(words[2]),
		REnd:       common.StringToInt64(words[3]),
		QName:      words[4],
		QStart:     common.StringToInt64(words[5]),
		QEnd:       common.StringToInt64(words[6]),
		QStrandPos: common.StringToStrand(words[7]),
		Score:      common.StringToInt64(words[8]),
		RSeq:       dna.StringToBases(rSeq),
		QSeq:       dna.StringToBases(qSeq),
	}
	return answer
}

func WriteToFileHandle(file *fileio.EasyWriter, input *Axt, alnNumber int) {
	_, err := fmt.Fprintf(file, "%s", ToString(input, alnNumber))
	common.ExitIfError(err)
}

func ToString(input *Axt, id int) string {
	return fmt.Sprintf("%d %s %d %d %s %d %d %c %d\n%s\n%s\n\n", id, input.RName, input.RStart, input.REnd, input.QName, input.QStart, input.QEnd, common.StrandToRune(input.QStrandPos), input.Score, dna.BasesToString(input.RSeq), dna.BasesToString(input.QSeq))
}

func Write(filename string, data []*Axt) {
	file := fileio.EasyCreate(filename)
	defer file.Close()

	for i, _ := range data {
		WriteToFileHandle(file, data[i], i)
	}
}

func AxtInfo(input *Axt) string {
	var text string = ""
	text = fmt.Sprintf("%s;%d;%d;%s;%d;%d;%t;%d", input.RName, input.RStart, input.REnd, input.QName, input.QStart, input.QEnd, input.QStrandPos, input.Score)
	return text
}

func QuerySwap(in *Axt, tLen int64, qLen int64) *Axt {
	var ans *Axt = &Axt{}
	//perform swap
	ans.RName, ans.QName = in.QName, in.RName

	//set to positive because target is always positive
	ans.QStrandPos = in.QStrandPos
	ans.Score = in.Score

	ans.RSeq = make([]dna.Base, len(in.QSeq))
	copy(ans.RSeq, in.QSeq)

	ans.QSeq = make([]dna.Base, len(in.RSeq))
	copy(ans.QSeq, in.RSeq)

	
	if !in.QStrandPos {
		ans.RStart, ans.REnd = qLen-in.QEnd+1, qLen-in.QStart+1
		ans.QStart, ans.QEnd = tLen-in.REnd+1, tLen-in.RStart+1
		dna.ReverseComplement(ans.RSeq)
		dna.ReverseComplement(ans.QSeq)
	} else {
		ans.RStart, ans.REnd = in.QStart, in.QEnd
		ans.QStart, ans.QEnd = in.RStart, in.REnd
	}
	return ans
}
