package fileio

import (
	"bufio"
	"compress/gzip"
	"errors"
	"github.com/vertgenlab/gonomics/exception"
	"log"
	"os"
	"strings"
)

// EasyReader provides a simplified wrapper of the builtin golang io functions.
// Will silently handle reading gziped files. EasyReader is a valid io.Reader.
type EasyReader struct {
	File         *os.File
	internalGzip *gzip.Reader
	BuffReader   *bufio.Reader
}

// EasyWriter provides a simplified wrapper of the builtin golang io functions.
// Will silently gzip output files when the input filename ends with '.gz'.
// EasyWriter is a valid io.Writer.
type EasyWriter struct {
	File         *os.File
	internalBuff *bufio.Writer
	internalGzip *gzip.Writer
}

// EasyOpen opens the input file. Panics if errors are encountered.
func EasyOpen(filename string) *EasyReader {
	if strings.Contains(filename, "http") {
		return EasyHttp(filename)
	}

	answer := EasyReader{}
	if strings.HasPrefix(filename, "stdin") {
		answer.File = os.Stdin
	} else {
		answer.File = MustOpen(filename)
	}
	var err error

	if strings.HasSuffix(filename, ".gz") {
		answer.internalGzip, err = gzip.NewReader(answer.File)
		exception.PanicOnErr(err)
		answer.BuffReader = bufio.NewReader(answer.internalGzip)
	} else {
		answer.BuffReader = bufio.NewReader(answer.File)
		answer.internalGzip = nil
	}
	return &answer
}

// EasyCreate creates a file with the input name. Panics if errors are encountered.
func EasyCreate(filename string) *EasyWriter {
	answer := EasyWriter{}

	switch {
	case strings.HasPrefix(filename, "stdout"):
		answer.File = os.Stdout
	case strings.HasPrefix(filename, "stderr"):
		answer.File = os.Stderr
	default:
		answer.File = MustCreate(filename)
	}

	answer.internalBuff = bufio.NewWriter(answer.File)

	if strings.HasSuffix(filename, ".gz") {
		answer.internalGzip = gzip.NewWriter(answer.internalBuff)
	} else {
		answer.internalGzip = nil
	}
	return &answer
}

// EasyNextLine returns the next line of the input EasyReader.
// Returns true at EOF.
func EasyNextLine(file *EasyReader) (string, bool) {
	return NextLine(file.BuffReader)
}

// EasyNextRealLine returns the next line of the input EasyReader that does not begin with '#'.
// Returns true at EOF.
func EasyNextRealLine(file *EasyReader) (string, bool) {
	return NextRealLine(file.BuffReader)
}

// EasyPeekReal will advance a reader past any lines beginning with '#' and read the first n bytes without advancing the reader.
func EasyPeekReal(file *EasyReader, n int) ([]byte, error) {
	return PeekReal(file.BuffReader, n)
}

func EasyReadHeader(file *EasyReader) ([]string, error) {
	return ReadHeader(file.BuffReader)
}

// EasyRemove deletes the input file.
func EasyRemove(filename string) {
	MustRemove(filename)
}

// Close the receiving EasyReader.
func (er *EasyReader) Close() error {
	var gzErr, fileErr error
	if er.internalGzip != nil {
		gzErr = er.internalGzip.Close()
	}
	if er.File != nil {
		fileErr = er.File.Close()
	} else {
		return errors.New("no file found")
	}

	switch { // Handle error returns. Priority is gzErr > fileErr
	case gzErr != nil:
		return gzErr

	case fileErr != nil:
		log.Println("WARNING: attempted to close file, but file already closed")
		return nil

	default:
		return nil
	}
}

// Read retrieves n bytes from the receiving EasyReader.
func (er *EasyReader) Read(p []byte) (n int, err error) {
	if er.internalGzip != nil {
		return er.internalGzip.Read(p)
	} else {
		return er.BuffReader.Read(p)
	}
}

// Close the receiving EasyWriter.
func (ew *EasyWriter) Close() error {
	var gzErr, bufErr, fileErr error
	if ew.internalGzip != nil {
		gzErr = ew.internalGzip.Close() // Serious write errors possible.
	}
	if ew.internalBuff != nil {
		bufErr = ew.internalBuff.Flush() // Serious write errors possible.
	}
	if ew.File != nil {
		fileErr = ew.File.Close() // The only possible err is that the file has already been closed.
	} else {
		return errors.New("no open file")
	}

	switch { // Handle error returns. Priority is gzErr > bufErr > fileErr
	case gzErr != nil:
		return gzErr

	case bufErr != nil:
		return bufErr

	case fileErr != nil:
		log.Println("WARNING: attempted to close file, but file already closed")
		return nil

	default:
		return nil
	}
}

// Write bytes to the receiving EasyWriter.
func (ew *EasyWriter) Write(p []byte) (n int, err error) {
	if ew.internalGzip != nil {
		return ew.internalGzip.Write(p)
	} else {
		return ew.internalBuff.Write(p)
	}
}

// Peek retrieves the next n bytes of the receiving EasyReader without advancing the reader.
func (er *EasyReader) Peek(n int) ([]byte, error) {
	return er.BuffReader.Peek(n)
}
