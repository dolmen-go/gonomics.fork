package main

import (
	"flag"
	"fmt"
	"github.com/vertgenlab/gonomics/expandedTree"
	"github.com/vertgenlab/gonomics/fasta"
	"github.com/vertgenlab/gonomics/fileio"
	"github.com/vertgenlab/gonomics/reconstruct"
	"github.com/vertgenlab/gonomics/simulate"
	"log"
)

func SimRecon(rootFastaFile string, treeFile string, gp string, simOutFile string, leafOutFile string, reconOutFile string, accuracyOutFile string) {
	//SimEvolution
	tree := expandedTree.ReadTree(treeFile, rootFastaFile)
	var fastas []*fasta.Fasta
	var leafFastas []*fasta.Fasta
	simulate.Simulate(rootFastaFile, tree, gp)
	nodes := expandedTree.GetTree(tree)

	for i := 0; i < len(nodes); i++ {
		fastas = append(fastas, nodes[i].Fasta)
		if nodes[i].Left == nil && nodes[i].Right == nil {
			leafFastas = append(leafFastas, nodes[i].Fasta)
		}
	}

	fasta.Write(simOutFile, fastas)
	fasta.Write(leafOutFile, leafFastas)

	//ReconSeq
	leaves := expandedTree.GetLeaves(tree)
	branches := expandedTree.GetBranch(tree)
	var treeFastas []*fasta.Fasta

	for i := 0; i < len(leaves[0].Fasta.Seq); i++ {
		reconstruct.LoopNodes(tree, i)
	}
	for j := 0; j < len(leaves); j++ {
		treeFastas = append(treeFastas, leaves[j].Fasta)
	}
	for k := 0; k < len(branches); k++ {
		treeFastas = append(treeFastas, branches[k].Fasta)
	}
	fasta.Write(reconOutFile, treeFastas)

	//ReconAccuracy
	//TODO: this code will need to change drastically for sequences of varying lengths.
	//The loop through the sequence is restricted by a single fasta and the tot calculation will need to calculate the total number of bps
	//ReconAccuracy calculates the total number of incorrectly reconstructed base pairs in a tree and returns a percentage of correct base calls
	answer := reconstruct.ReconAccuracy(simOutFile, reconOutFile)
	out := fileio.EasyCreate(accuracyOutFile)
	defer out.Close()

	for name, accuracy := range answer {
		fmt.Fprintf(out, "%s\t%f\n", name, accuracy)
	}
}

func usage() {
	fmt.Print(
		"ReconAccuracy takes in a fasta file of simulated evolution along a tree, and the reconstructed fastas of the same tree and returns the percentage accuracy of the sequences of each node and all nodes in the tree.\n" +
			"reconAccuracy <simulationOut.fasta> <reconstructionOut.fasta> <outFilename.txt> \n" +
			"options:\n")
	flag.PrintDefaults()
}

func main() {
	var expectedNumArgs = 7

	flag.Usage = usage
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	if len(flag.Args()) != expectedNumArgs {
		flag.Usage()
		log.Fatalf("Error: expecting %d arguments, but got %d\n",
			expectedNumArgs, len(flag.Args()))
	}

	rootFastaFile := flag.Arg(0)
	treeFile := flag.Arg(1)
	gp := flag.Arg(2)
	simOutFile := flag.Arg(3)
	leafOutFile := flag.Arg(4)
	reconOutFile := flag.Arg(5)
	accuracyOutFile := flag.Arg(6)

	SimRecon(rootFastaFile, treeFile, gp, simOutFile, leafOutFile, reconOutFile, accuracyOutFile)
}
