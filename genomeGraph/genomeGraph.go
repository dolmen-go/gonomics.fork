package genomeGraph

import (
	"fmt"
	"github.com/vertgenlab/gonomics/common"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/dnaTwoBit"
	"github.com/vertgenlab/gonomics/exception"
	"github.com/vertgenlab/gonomics/fileio"
	"io"
	"strings"
)

// GenomeGraph struct contains a slice of Nodes
type GenomeGraph struct {
	Nodes []*Node
}

// Node is uniquely definded by Id and is encoded with information
// describing sequence order and orientation and annotated variance
type Node struct {
	Id uint32
	//Name      string
	Seq       []dna.Base        // only this field or the SeqThreeBit will be kept
	SeqTwoBit *dnaTwoBit.TwoBit // this will change to a ThreeBit or be removed
	Prev      []Edge
	Next      []Edge
	//Info      Annotation
}

// Edge describes the neighboring nodes and a weighted probabilty
// of the of the more likely path
type Edge struct {
	Dest *Node
	Prob float32
}

// Annotation struct is an uint64 encoding of allele id, starting position on linear reference and variant on node
// a single byte will represent what allele the node came from, uint32 will be used for starting postion of chromosome of the linear reference
// uint8 are variants are represented as follows: 0=match, 1=mismatch, 2=insert, 3=deletion, 4=hap
/*type Annotation struct {
	Start   uint32
	Allele  byte
	Variant uint8
}*/

// Read will process a simple graph formated text file and parse the data into graph fields
func Read(filename string) *GenomeGraph {
	simpleReader := fileio.NewByteReader(filename)
	genome := EmptyGraph()
	var currNode, homeNode, destNode *Node
	var weight float32
	var line string
	var words []string = make([]string, 0, 2)
	var seqIdx, homeNodeIdx, destNodeIdx uint32
	var i int

	for reader, done := fileio.ReadLine(simpleReader); !done; reader, done = fileio.ReadLine(simpleReader) {
		line = reader.String()
		switch true {
		case strings.HasPrefix(line, ">"):
			seqIdx = common.StringToUint32(line[1:])
			currNode = &Node{Id: seqIdx}
			AddNode(genome, currNode)
		case strings.Contains(line, "\t"):
			words = strings.Split(line, "\t")
			homeNodeIdx = common.StringToUint32(words[0])
			homeNode = genome.Nodes[homeNodeIdx]
			if len(words) > 2 {
				for i = 1; i < len(words); i += 2 {
					weight = common.StringToFloat32(words[i])
					destNodeIdx = common.StringToUint32(words[i+1])
					destNode = genome.Nodes[destNodeIdx]
					AddEdge(homeNode, destNode, weight)
				}
			}
		default:
			genome.Nodes[seqIdx].Seq = append(genome.Nodes[seqIdx].Seq, dna.ByteSliceToDnaBases(reader.Bytes())...)
		}
	}
	for i = 0; i < len(genome.Nodes); i++ {
		genome.Nodes[i].SeqTwoBit = dnaTwoBit.NewTwoBit(genome.Nodes[i].Seq)
	}
	return genome
}

// AddNode will append a new Node to a slice of nodes in GenomeGraph
func AddNode(g *GenomeGraph, n *Node) {
	g.Nodes = append(g.Nodes, n)
}

// AddEdge will append two edges one forward and one backwards for any two
// given node. Provide a probability float32 to specify a weight for an edge
// to describe the more likely path through the graph
func AddEdge(u, v *Node, p float32) {
	u.Next = append(u.Next, Edge{Dest: v, Prob: p})
	v.Prev = append(v.Prev, Edge{Dest: u, Prob: p})
}

// SetEvenWeights will loop through a slice of edges and set the probability weight
// divided by the length of the slice.
func SetEvenWeights(u *Node) {
	var edge int
	var weights float32 = 1 / float32(len(u.Next))
	for edge = 0; edge < len(u.Next); edge++ {
		u.Next[edge].Prob = weights
	}
}

// Write function will process GenomeGraph and write the data to a file
func Write(filename string, sg *GenomeGraph) {
	lineLength := 50
	file := fileio.EasyCreate(filename)
	defer file.Close()
	WriteToGraphHandle(file, sg, lineLength)
}

// EmptyGraph will allocate a new zero pointer to a simple graph and will allocate memory for the Nodes of the graph
func EmptyGraph() *GenomeGraph {
	var graph *GenomeGraph
	graph = &GenomeGraph{Nodes: make([]*Node, 0)}
	return graph
}

// PrintGraph will quickly print simpleGraph to standard out
func PrintGraph(gg *GenomeGraph) {
	Write("/dev/stdout", gg)
}

// WriteToGraphHandle will help with any error handling when writing GenomeGraph to file.
func WriteToGraphHandle(file io.Writer, gg *GenomeGraph, lineLength int) {
	var err error
	var i, j int
	for i = 0; i < len(gg.Nodes); i++ {
		_, err = fmt.Fprintf(file, ">%d\n", gg.Nodes[i].Id)
		exception.PanicOnErr(err)
		for j = 0; j < len(gg.Nodes[i].Seq); j += lineLength {
			if j+lineLength > len(gg.Nodes[i].Seq) {
				_, err = fmt.Fprintf(file, "%s\n", dna.BasesToString(gg.Nodes[i].Seq[j:]))
				exception.PanicOnErr(err)
			} else {
				_, err = fmt.Fprintf(file, "%s\n", dna.BasesToString(gg.Nodes[i].Seq[j:j+lineLength]))
				exception.PanicOnErr(err)
			}
		}
	}
	for i = 0; i < len(gg.Nodes); i++ {
		_, err = fmt.Fprintf(file, "%d", gg.Nodes[i].Id)
		exception.PanicOnErr(err)
		near := gg.Nodes[i].Next
		for j = 0; j < len(near); j++ {
			_, err = fmt.Fprintf(file, "\t%v\t%d", near[j].Prob, near[j].Dest.Id)
			exception.PanicOnErr(err)
		}
		_, err = fmt.Fprintf(file, "\n")
		exception.PanicOnErr(err)
	}
}

// BasesInGraph will calculate the number of bases contained in GenomeGraph using dnaTwoBit
func BasesInGraph(g *GenomeGraph) int {
	var i, baseCount int = 0, 0
	for i = 0; i < len(g.Nodes); i++ {
		baseCount += g.Nodes[i].SeqTwoBit.Len
	}
	return baseCount
}