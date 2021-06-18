//package bed provides functions for reading, writing, and manipulating Browser Extinsible Data (BED) format files.
//More information on the BED file format can be found at https://genome.ucsc.edu/FAQ/FAQformat.html#format1
package bed

import (
	"fmt"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/fileio"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

//Bed stores information about genomic regions, including their location, name, score, strand, and other annotations.
type Bed struct {
	Chrom             string
	ChromStart        int
	ChromEnd          int
	Name              string
	Score             int
	Strand            Strand
	FieldsInitialized int      //number of fields that are initialized, used for smart writing.
	Annotation        []string //long form for extra fields
}

//Strand stores strand state, which can be positive, negative, or none.
type Strand byte

const (
	Positive Strand = '+'
	Negative Strand = '-'
	None     Strand = '.'
)

// String converts a bed struct to a string so it will be automatically formatted when printing with the fmt package.
func (b *Bed) String() string {
	return ToString(b, b.FieldsInitialized)
}

//ToString converts a Bed struct into a BED file format string. Useful for writing to files or printing.
func ToString(bunk *Bed, fields int) string {
	switch fields {
	case 3:
		return fmt.Sprintf("%s\t%d\t%d", bunk.Chrom, bunk.ChromStart, bunk.ChromEnd)
	case 4:
		return fmt.Sprintf("%s\t%d\t%d\t%s", bunk.Chrom, bunk.ChromStart, bunk.ChromEnd, bunk.Name)
	case 5:
		return fmt.Sprintf("%s\t%d\t%d\t%s\t%d", bunk.Chrom, bunk.ChromStart, bunk.ChromEnd, bunk.Name, bunk.Score)
	case 6:
		return fmt.Sprintf("%s\t%d\t%d\t%s\t%d\t%c", bunk.Chrom, bunk.ChromStart, bunk.ChromEnd, bunk.Name, bunk.Score, bunk.Strand)
	case 7:
		var out string = fmt.Sprintf("%s\t%d\t%d\t%s\t%d\t%c", bunk.Chrom, bunk.ChromStart, bunk.ChromEnd, bunk.Name, bunk.Score, bunk.Strand)
		for i := 0; i < len(bunk.Annotation); i++ {
			out = fmt.Sprintf("%s\t%s", out, bunk.Annotation[i])
		}
		return out
	default:
		log.Fatalf("Error: expecting a request to print 3 to 7 bed fields, but got: %d\n", fields)
	}
	return ""
}

//WriteBed writes an input Bed struct to an os.File with a specified number of Bed fields.
func WriteBed(file *os.File, input *Bed, fields int) {
	var err error
	_, err = fmt.Fprintf(file, "%s\n", ToString(input, fields))
	common.ExitIfError(err)
}

//WriteToFileHandle writes an input Bed struct with a specified number of fields to an io.Writer
func WriteToFileHandle(file io.Writer, rec *Bed, fields int) {
	var err error
	_, err = fmt.Fprintf(file, "%s\n", ToString(rec, fields))
	common.ExitIfError(err)
}

//WriteSliceToFileHandle writes a slice of Bed structs with a specified number of fields to an io.Writer
func WriteSliceToFileHandle(file io.Writer, records []*Bed, fields int) {
	for _, rec := range records {
		WriteToFileHandle(file, rec, fields)
	}
}

//Write writes a slice of Bed structs with a specified number of fields to a specified filename.
func Write(filename string, records []*Bed, fields int) {
	file := fileio.EasyCreate(filename)
	defer file.Close()

	WriteSliceToFileHandle(file, records, fields)
}

//ReadLite reads a whole Bed file into memory, but only reads positional information (fields 1 to 3).
func ReadLite(filename string) []*Bed {
	var line string
	var answer []*Bed
	var doneReading bool = false

	file := fileio.EasyOpen(filename)
	defer file.Close()
	//reader := bufio.NewReader(file)

	for line, doneReading = fileio.EasyNextRealLine(file); !doneReading; line, doneReading = fileio.EasyNextRealLine(file) {
		current := processBedLineLite(line)
		answer = append(answer, current)
	}
	return answer
}

//Read returns a slice of Bed structs from an input filename.
func Read(filename string) []*Bed {
	var line string
	var answer []*Bed
	var doneReading bool = false

	file := fileio.EasyOpen(filename)
	defer file.Close()
	//reader := bufio.NewReader(file)

	for line, doneReading = fileio.EasyNextRealLine(file); !doneReading; line, doneReading = fileio.EasyNextRealLine(file) {
		current := processBedLine(line)
		answer = append(answer, current)
	}
	return answer
}

//processBedLineLite is like processBedLine, but only parses the first three fields of a Bed line.
func processBedLineLite(line string) *Bed {
	words := strings.Split(line, "\t")
	startNum := common.StringToInt(words[1])
	endNum := common.StringToInt(words[2])
	return &Bed{Chrom: words[0], ChromStart: startNum, ChromEnd: endNum, FieldsInitialized: 3}
}

//processBedLine is a helper function of Read that returns a Bed struct from an input line of a file.
func processBedLine(line string) *Bed {
	words := strings.Split(line, "\t")
	startNum := common.StringToInt(words[1])
	endNum := common.StringToInt(words[2])

	current := Bed{Chrom: words[0], ChromStart: startNum, ChromEnd: endNum, Strand: None, FieldsInitialized: len(words)}
	if len(words) >= 4 {
		current.Name = words[3]
	}
	if len(words) >= 5 {
		current.Score = common.StringToInt(words[4])
	}
	if len(words) >= 6 {
		current.Strand = StringToStrand(words[5])
	}
	if len(words) >= 7 {
		for i := 6; i < len(words); i++ {
			current.Annotation = append(current.Annotation, words[i])
		}
	}
	return &current
}

//StringToStrand parses a bed.Strand struct from an input string.
func StringToStrand(s string) Strand {
	switch s {
	case "+":
		return Positive
	case "-":
		return Negative
	case ".":
		return None
	default:
		log.Fatalf("Error: expected %s to be a strand that is either '+', '-', or '.'.\n", s)
		return None
	}
}

//NextBed returns a Bed struct from an input fileio.EasyReader. Returns a bool that is true when the reader is done.
func NextBed(reader *fileio.EasyReader) (*Bed, bool) {
	line, done := fileio.EasyNextLine(reader)
	if done {
		return nil, true
	}
	return processBedLine(line), false
}

//ReadToChan reads from a fileio.EasyReader to send Bed structs to a chan<- *Bed.
func ReadToChan(file *fileio.EasyReader, data chan<- *Bed, wg *sync.WaitGroup) {
	for curr, done := NextBed(file); !done; curr, done = NextBed(file) {
		data <- curr
	}
	file.Close()
	wg.Done()
}

//GoReadToChan reads Bed entries from an input filename to a <-chan *Bed.
func GoReadToChan(filename string) <-chan *Bed {
	file := fileio.EasyOpen(filename)
	var wg sync.WaitGroup
	data := make(chan *Bed)
	wg.Add(1)
	go ReadToChan(file, data, &wg)

	go func() {
		wg.Wait()
		close(data)
	}()

	return data
}
