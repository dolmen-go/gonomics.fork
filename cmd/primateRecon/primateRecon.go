// Command Group: "Sequence Evolution & Reconstruction"

package main

import (
	"flag"
	"fmt"
	"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/exception"
	"github.com/vertgenlab/gonomics/expandedTree"
	"github.com/vertgenlab/gonomics/fasta"
	"github.com/vertgenlab/gonomics/reconstruct"
	"log"
)

// likelihoodsToBase takes the un-normalized likelihoods for A, C, G, T as well
// as the index of the human base (0 for A, 1 for C, etc), and the probability
// threshold for when we will call it the mle base instead of the human base
// and gives back the reconstructed base for the hca
func likelihoodsToBase(likes []float64, humanBase dna.Base, probThreshold float64, messyToN bool) dna.Base {
	var total, bestProb float64
	var i int
	var answer dna.Base

	if messyToN {
		answer = dna.N
	} else {
		answer = humanBase
	}

	for i = range likes {
		total += likes[i]
	}
	for i = range likes {
		if likes[i]/total >= probThreshold && likes[i] > bestProb {
			bestProb = likes[i]
			answer = dna.Base(i)
		}
	}
	return answer
}

func hcaIsPresent(human, bonobo, chimp, gorilla, organutan dna.Base) bool {
	if dna.DefineBase(human) && (dna.DefineBase(bonobo) || dna.DefineBase(chimp)) {
		return true
	}
	if (dna.DefineBase(human) || dna.DefineBase(bonobo) || dna.DefineBase(chimp)) && (dna.DefineBase(gorilla) || dna.DefineBase(organutan)) {
		return true
	}
	return false
}

func reconHcaBase(root, humanNode, nodeToRecon *expandedTree.ETree, position int, probThreshold float64, messyToN bool) {
	reconstruct.SetState(root, position)
	likelihoods := reconstruct.FixFc(root, nodeToRecon)
	nodeToRecon.Fasta.Seq = append(nodeToRecon.Fasta.Seq, likelihoodsToBase(likelihoods, humanNode.Fasta.Seq[position], probThreshold, messyToN))
}

func primateReconMle(inFastaFilename string, inTreeFilename string, probThreshold float64, messyToN bool, outputFastaFilename string) {
	var tree, humanNode, bonoboNode, chimpNode, gorillaNode, orangutanNode, hcaNode *expandedTree.ETree
	var err error
	var i int

	tree, err = expandedTree.ReadTree(inTreeFilename, inFastaFilename)
	exception.FatalOnErr(err)

	// roll call to make sure everyone is here and will need them later
	humanNode = expandedTree.FindNodeName(tree, "hg38")
	if humanNode == nil {
		log.Fatalf("Didn't find hg38 in the tree\n")
	}
	bonoboNode = expandedTree.FindNodeName(tree, "panPan2")
	if bonoboNode == nil {
		log.Fatalf("Didn't find panPan2 in the tree\n")
	}
	chimpNode = expandedTree.FindNodeName(tree, "panTro6")
	if chimpNode == nil {
		log.Fatalf("Didn't find panTro6 in the tree\n")
	}
	gorillaNode = expandedTree.FindNodeName(tree, "gorGor5")
	if gorillaNode == nil {
		log.Fatalf("Didn't find gorGor5 in the tree\n")
	}
	orangutanNode = expandedTree.FindNodeName(tree, "ponAbe3")
	if orangutanNode == nil {
		log.Fatalf("Didn't find ponAbe3 in the tree\n")
	}
	hcaNode = expandedTree.FindNodeName(tree, "hca")
	if hcaNode == nil {
		log.Fatalf("Didn't find hca in the tree\n")
	}

	for i = range humanNode.Fasta.Seq {
		if hcaIsPresent(humanNode.Fasta.Seq[i], bonoboNode.Fasta.Seq[i], chimpNode.Fasta.Seq[i], gorillaNode.Fasta.Seq[i], orangutanNode.Fasta.Seq[i]) {
			reconHcaBase(tree, humanNode, hcaNode, i, probThreshold, messyToN)
		} else {
			hcaNode.Fasta.Seq = append(hcaNode.Fasta.Seq, dna.Gap)
		}
	}
	fasta.Write(outputFastaFilename, []fasta.Fasta{*humanNode.Fasta, *hcaNode.Fasta})
}

func primateRecon(infile string, outfile string, messyToN bool) {
	records := fasta.Read(infile)
	output := append(records, reconstruct.PrimateRecon(records, messyToN))
	fasta.Write(outfile, output)
}

func usage() {
	fmt.Print(
		"primateRecon - Returns maximum likelihood sequence from an HBCGO primate alignment\n" +
			"Usage:\n" +
			"primateRecon input.fa output.fa\n" +
			"options:\n")
	flag.PrintDefaults()
}

func main() {
	var expectedNumArgs int = 2
	var messyToN *bool = flag.Bool("messyToN", false, "Sets messy bases to Ns in the output file.")
	var mle *string = flag.String("mle", "", "Does a maximum likelihood estimate if newick tree filename provided.")
	var probThreshold *float64 = flag.Float64("probThreshold", 0.0, "The probability that a base other than human must pass to be considered a true change in the hca.")
	flag.Usage = usage
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	if len(flag.Args()) != expectedNumArgs {
		flag.Usage()
		log.Fatalf("Error: expecting %d arguments, but got %d\n",
			expectedNumArgs, len(flag.Args()))
	}

	inFile := flag.Arg(0)
	outFile := flag.Arg(1)

	if *mle != "" {
		primateReconMle(inFile, *mle, *probThreshold, *messyToN, outFile)
	} else {
		primateRecon(inFile, outFile, *messyToN)
	}
}
