package graphReconstruct

import (
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/dnaTwoBit"

	//"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/expandedTree"
	"github.com/vertgenlab/gonomics/simpleGraph"
)

type graphColumn struct {
	AlignId    int
	AlignNodes map[string][]*simpleGraph.Node
}

//BuildNodes uses a graphColumn to create nodes for an ancestor's graph seq that represents all the unique sequences in an aligned graph
func BuildNodes(root *expandedTree.ETree, column graphColumn, id uint32) uint32 {
	var nodeInfo = make(map[string]bool)
	for _, nodes := range column.AlignNodes { //nodes is all nodes for an individual species
		for n := range nodes { //n is an individual node of an individual species
			stringSeq := dna.BasesToString(nodes[n].Seq)
			nodeInfo[stringSeq] = true
		}
	}
	for seq, _ := range nodeInfo {
		id += 1
		var newNode *simpleGraph.Node
		newNode = &simpleGraph.Node{Id: id, Name: root.Name, Seq: dna.StringToBases(seq), SeqTwoBit: dnaTwoBit.NewTwoBit(dna.StringToBases(seq)), Next: nil, Prev: nil, Info: simpleGraph.Annotation{}}
		for name, info := range column.AlignNodes { //add a new node to the existing nodes for this species
			if name == root.Name {
				info = append(info, newNode)
				column.AlignNodes[root.Name] = info
			} else { //create a record for this species nodes in GraphColumn
				column.AlignNodes[root.Name] = []*simpleGraph.Node{newNode}
			}
		}
	}
	return id
}

//BuildEdges connects the nodes of a species' graph that are stored in GraphColumns
//func BuildEdges
//FindAncSeq creates a graph from the node records stored in GraphColumns and then calls PathFinder and seqOfPath to determine the most likley seq of the ancestor before assigning that
//seq to the Fasta field of the ancestors tree node
//func FindAncSeq will loop through aligncolumns and build a single graph of all of the nodes that belong to the ancestor species after edges are created
//run PathFinder on the graph for the anc, run seqOfPath, then turn that to a fasta for that node of the tree

//seqOfPath takes in a graph and a path specified by the Node IDs and returns the seq of the path through the graph
func seqOfPath(g *simpleGraph.SimpleGraph, path []uint32) []dna.Base {
	var seq []dna.Base
	for n := 0; n < len(g.Nodes); n++ {
		for p := 0; p < len(path); p++ {
			if g.Nodes[n].Id == path[p] {
				seq = append(seq, g.Nodes[n].Seq...)
			}
		}
	}
	return seq
}

//PathFinder takes a graph and returns the most likely path through that graph after checking all possible paths from the first node to the last
func PathFinder(g *simpleGraph.SimpleGraph) ([]uint32, float32) {
	var finalPath []uint32
	var finalProb float32
	var tempPath = make([]uint32, 0)

	for n := 0; n < len(g.Nodes); n++ {
		if g.Nodes[n].Id == 0 {
			finalProb, finalPath = bestPath(g.Nodes[n], 1, tempPath)
		}
	}
	return finalPath, finalProb
}

//bestPath is the helper function for PathFinder, and recursively traverses the graph depth first to determine the most likely path from start to finish
func bestPath(node *simpleGraph.Node, prevProb float32, prevPath []uint32) (prob float32, path []uint32) {
	var tempProb float32 = 0
	var tempPath []uint32
	var finalProb float32
	var finalPath []uint32

	tempPath = append(tempPath, prevPath...)
	tempPath = append(tempPath, node.Id)
	if len(node.Next) == 0 {
		return prevProb, tempPath
	}
	for i, _ := range node.Next {
		tempProb = node.Next[i].Prob * prevProb
		currentProb, currentPath := bestPath(node.Next[i].Dest, tempProb, tempPath)
		if currentProb > finalProb {
			finalProb = currentProb
			finalPath = currentPath
		}
	}
	return finalProb, finalPath
}
