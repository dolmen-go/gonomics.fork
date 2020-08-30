package giraf

import (
	"fmt"
	"github.com/vertgenlab/gonomics/cigar"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/fastq"
	"log"
	"strconv"
	"strings"
)

func GirafToString(g *Giraf) string {
	var answer string
	answer += fmt.Sprintf("%s\t%d\t%d\t%d\t%c\t%s\t%s\t%d\t%d\t%s\t%s%s", g.QName, g.QStart, g.QEnd, g.Flag, strandToRune(g.PosStrand), PathToString(g.Path), cigar.ByteCigarToString(g.Cigar), g.AlnScore, g.MapQ, dna.BasesToString(g.Seq), fastq.QualString(g.Qual), NotesToString(g.Notes))
	return answer
}

func stringToGiraf(line string) *Giraf {
	var curr *Giraf
	data := strings.SplitN(line, "\t", 12)
	if len(data) > 10 {

		curr = &Giraf{
			QName:     data[0],
			QStart:    common.StringToInt(data[1]),
			QEnd:      common.StringToInt(data[2]),
			Flag:      common.StringToUint8(data[3]),
			PosStrand: StringToPos(data[4]),
			Path:      FromStringToPath(data[5]),
			Cigar:     cigar.ReadToBytesCigar([]byte(data[6])),
			AlnScore:  common.StringToInt(data[7]),
			MapQ:      uint8(common.StringToInt(data[8])),
			Seq:       dna.StringToBases(data[9]),
			Qual:      fastq.ToQualUint8([]rune(data[10]))}

		if len(data) == 12 {
			curr.Notes = FromStringToNotes(data[11])
		}
	} else {
		log.Fatalf("Error: Expecting at least 11 columns, but only found %d on %s", len(data), line)
	}
	return curr
}

func PathToString(p Path) string {
	nodeString := make([]string, len(p.Nodes))
	for i, v := range p.Nodes {
		nodeString[i] = strconv.Itoa(int(v))
	}
	return fmt.Sprintf("%d:%s:%d", p.TStart, strings.Join(nodeString, ">"), p.TEnd)
}

func FromStringToPath(column string) Path {
	words := strings.Split(column, ":")
	if len(words) != 3 {
		log.Fatalf("Error: Needs exact 3 values, only found %d", len(words))
	}
	nodes := strings.Split(words[1], ">")
	answer := Path{TStart: common.StringToInt(words[0]), Nodes: make([]uint32, len(nodes)), TEnd: common.StringToInt(words[2])}

	for i, v := range nodes {
		answer.Nodes[i] = common.StringToUint32(v)
	}
	return answer
}

func NotesToString(notes []Note) string {
	var answer string = ""
	if len(notes) == 0 {
		return answer
	} else {
		for i := 0; i < len(notes); i++ {
			answer += fmt.Sprintf("\t%s", NoteToString(notes[i]))
		}
	}
	return answer
}

func strandToRune(posStrand bool) rune {
	if posStrand {
		return '+'
	} else {
		return '-'
	}
}

func StringToPos(column string) bool {
	var answer bool
	if strings.Compare(column, "+") == 0 {
		answer = true
	} else if strings.Compare(column, "-") == 0 {
		answer = false
	} else {
		log.Fatalf("Error did not find stand information in this column...\n")
	}
	return answer
}

func NoteToString(txt Note) string {
	var answer string = fmt.Sprintf("%s:%c:%s", txt.Tag, txt.Type, txt.Value)
	return answer
}

func FromStringToNotes(s string) []Note {
	words := strings.Split(s, "\t")
	answer := make([]Note, len(words))
	var text []string
	for i, v := range words {
		text = strings.SplitN(v, ":", 3)
		answer[i] = Note{Tag: []byte(text[0])[:1], Type: []byte(text[1])[0], Value: text[2]}
	}
	return answer
}

func qualToString(qual []uint8) string {
	answer := make([]string, len(qual))
	for i, v := range qual {
		answer[i] = strconv.Itoa(int(v))
	}
	return strings.Join(answer, ",")
}

func fromStringToQual(s string) []uint8 {
	words := []rune(s)
	answer := make([]uint8, 0)
	for _, v := range words {
		answer = append(answer, uint8(v))
	}
	return answer
}

func ToString(g *Giraf) string {
	buf := &strings.Builder{}
	buf.Grow(400)
	var err error
	_, err = buf.WriteString(g.QName)
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteString(strconv.Itoa(g.QStart))
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteString(strconv.Itoa(g.QEnd))
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteString(strconv.Itoa(int(g.Flag)))
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteRune(common.StrandToRune(g.PosStrand))
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteString(PathToString(g.Path))
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteString(cigar.ByteCigarToString(g.Cigar))
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteString(strconv.Itoa(g.AlnScore))
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteString(strconv.Itoa(int(g.MapQ)))
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteString(dna.ByteDnaBasesToString(g.Seq))
	common.ExitIfError(err)
	err = buf.WriteByte('\t')
	common.ExitIfError(err)
	_, err = buf.WriteString(fastq.QualString(g.Qual))
	common.ExitIfError(err)
	_, err = buf.WriteString(NotesToString(g.Notes))
	common.ExitIfError(err)
	return buf.String()
}
