package binaryGiraf

import (
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/fileio"
	"github.com/vertgenlab/gonomics/giraf"
	"github.com/vertgenlab/gonomics/simpleGraph"
	"io"
	"sync"
)

// Read will input a giraf.fe file and will decompress into a slice of giraf.Giraf.
// Requires the graph (.gg) file used to generate the giraf for decompression
func Read(giraffeFile string, graphFile string) []giraf.Giraf {
	var answer []giraf.Giraf
	file := fileio.EasyOpen(giraffeFile)
	graph := simpleGraph.Read(graphFile)
	defer file.Close()
	reader := NewBinReader(file.BuffReader)
	var curr giraf.Giraf
	var err error
	for curr, err = reader.Read(graph); err != io.EOF; curr, err = reader.Read(graph) {
		common.ExitIfError(err)
		answer = append(answer, curr)
	}

	// Close reader
	err = reader.bg.Close()
	common.ExitIfError(err)

	return answer
}

// ReadToChan will input a fileio.EasyReader and decompress girafs in the file into a stream (channel) of giraf.Giraf.
// ReadToChan is designed to be run as a goroutine. See GoReadToChan for internally handled goroutines.
func ReadToChan(file *fileio.EasyReader, graph *simpleGraph.SimpleGraph, data chan<- giraf.Giraf, wg *sync.WaitGroup) {
	var curr giraf.Giraf
	var err error
	reader := NewBinReader(file.BuffReader)
	for curr, err = reader.Read(graph); err != io.EOF; curr, err = reader.Read(graph) {
		common.ExitIfError(err)
		data <- curr
	}

	// Close reader
	err = reader.bg.Close()
	common.ExitIfError(err)

	file.Close()
	wg.Done()
}

// GoReadToChan is a wrapper for the ReadToChan function that handles goroutine creation internally. Decompresses input giraf.fe file
// into a buffered channel (len 1024) of giraf.Giraf records.
func GoReadToChan(giraffeFile string, graphFile string) <-chan giraf.Giraf {
	file := fileio.EasyOpen(giraffeFile)
	graph := simpleGraph.Read(graphFile)
	var wg sync.WaitGroup
	data := make(chan giraf.Giraf, 1024)
	wg.Add(1)
	go ReadToChan(file, graph, data, &wg)

	go func() {
		wg.Wait()
		close(data)
	}()

	return data
}

//TODO: move GoReadToChan functions in the giraf package from *Giraf to Giraf.
// GirafChanToBinary inputs a channel of *giraf.Giraf and compresses records from the stream into a output giraf.fe file.
func GirafChanToBinary(filename string, input <-chan *giraf.Giraf, wg *sync.WaitGroup) {
	file := fileio.EasyCreate(filename)
	writer := NewBinWriter(file)
	var err error
	defer file.Close()

	for record := range input {
		err = writer.Write(record)
		common.ExitIfError(err)
	}
	err = writer.bg.Close()
	common.ExitIfError(err)
	wg.Done()
}
