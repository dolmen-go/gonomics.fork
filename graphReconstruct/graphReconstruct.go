package graphReconstruct

import (
	"github.com/vertgenlab/gonomics/dna"
	//"github.com/vertgenlab/gonomics/dna"
	"github.com/vertgenlab/gonomics/expandedTree"
	"github.com/vertgenlab/gonomics/fasta"
	"github.com/vertgenlab/gonomics/simpleGraph"
	"log"
)

type graphColumn struct {
	AlignId    int
	AlignNodes [][]*simpleGraph.Node
}

//returns the percentage accuracy by base returned by reconstruct of each node and of all nodes combined (usage in reconstruct_test.go)
func ReconAccuracy(simFilename string, reconFilename string) map[string]float64 {
	var allNodes string
	allNodes = "all Nodes"
	var found bool = false
	var total float64
	total = 0.0
	var mistakes float64
	sim := fasta.Read(simFilename)
	recon := fasta.Read(reconFilename)

	answer := make(map[string]float64)

	for i := 0; i < len(sim); i++ {
		mistakes = 0.0
		found = false
		for j := 0; j < len(recon); j++ {
			if sim[i].Name == recon[j].Name {
				found = true
				//DEBUG: log.Printf("\n%s \n%s \n", dna.BasesToString(sim[i].Seq), dna.BasesToString(recon[j].Seq))
				for k := 0; k < len(sim[0].Seq); k++ {
					if sim[i].Seq[k] != recon[j].Seq[k] {
						mistakes = mistakes + 1
					}
				}
			}
		}
		if found == false {
			log.Fatal("Did not find all simulated sequences in reconstructed fasta.")
		}
		accuracy := mistakes / float64(len(sim[i].Seq)) * 100.0
		//DEBUG: fmt.Printf("tot: %f, len(sim): %f, len(sim[0].Seq): %f \n", tot, float64(len(sim)), float64(len(sim[0].Seq)))
		acc := 100 - accuracy
		answer[sim[i].Name] = acc
		total = total + mistakes
	}
	accuracy := total / (float64(len(sim)) * float64(len(sim[0].Seq))) * 100.0
	//DEBUG: fmt.Printf("tot: %f, len(sim): %f, len(sim[0].Seq): %f \n", tot, float64(len(sim)), float64(len(sim[0].Seq)))
	acc := 100 - accuracy
	answer[allNodes] = acc
	return answer
}

//write assigned sequences at all nodes to a fasta file
func WriteTreeToFasta(tree *expandedTree.ETree, outFile string) {
	var fastas []*fasta.Fasta
	nodes := expandedTree.GetTree(tree)

	for i := 0; i < len(nodes); i++ {
		fastas = append(fastas, nodes[i].Fasta)
	}
	fasta.Write(outFile, fastas)
}

//write assigned sequences at leaf nodes to a fasta file
func WriteLeavesToFasta(tree *expandedTree.ETree, leafFile string) {
	var leafFastas []*fasta.Fasta
	nodes := expandedTree.GetLeaves(tree)

	for i := 0; i < len(nodes); i++ {
		leafFastas = append(leafFastas, nodes[i].Fasta)
	}
	fasta.Write(leafFile, leafFastas)
}

//calculate probability of switching from one base to another
func Prob(a int, b int, t float64) float64 {
	var p float64
	switch {
	case a > 3 || b > 3:
		p = 0
	case a == b:
		p = 1 - t
	default:
		p = t / 3
	}
	return p
}

//take in probability of all 4 bases return integer value of the most likely base
func Yhat(r []float64) int {
	var n float64
	n = 0
	var pos int
	for p, v := range r {
		if v > n {
			n = v
			pos = p
		}
	}
	return pos
}

func allZero(r []float64) bool {
	for _, v := range r {
		if v != 0 {
			return false
		}
	}
	return true
}

func GraphProb(a float32, b float64) float64 {
	//calc probability that the edge.Prob is unchanged in parent
	answer := float64(a) * (1 - b)
	return answer
}

//set up Stored list for each node in the tree with probability of each base
func SetState(node *expandedTree.ETree, position int, graph *simpleGraph.SimpleGraph) *simpleGraph.SimpleGraph {
	for gn := 0; gn < len(graph.Nodes); gn++ {
		currNode := graph.Nodes[gn]
		incoming := currNode.Prev
		sum := 0.0
		if node.Left != nil && node.Right != nil {
			SetState(node.Left, position, graph)
			SetState(node.Right, position, graph)
			//outgoing := currNode.Next
			for i := 0; i < len(incoming); i++ {
				//only want a single edge calculation in sum
				sum = GraphProb(incoming[i].Prob, node.Left.BranchLength) * GraphProb(incoming[i].Prob, node.Right.BranchLength)
				//the probability that this node was passed through in node.r/l and the prob that edge.prob has remained the same (BranchLength)
				incoming[i].Prob = float32(sum)
				//updates existing edge (either this node's graph has been copied from below, and needs to have edges changed
				//or the graph is a single graph for the whole tree that gets passed up to the parent and the edges need to be updated)
			}
			//for i := 0; i < 4; i++ {
			//	sum := 0.0
			//	for j := 0; j < 4; j++ {
			//		for k := 0; k < 4; k++ {
			//			sum = sum + Prob(i, j, node.Left.BranchLength)*node.Left.Stored[j]*Prob(i, k, node.Right.BranchLength)*node.Right.Stored[k]
			//		}
			//	}
			//	node.Stored[i] = sum
			//}
		} else if node.Left != nil {
			SetState(node.Left, position, graph)
			//if node.Left.Stored == nil {
			//	log.Fatal("no Stored values passed to internal node, left branch")
			//}
			for i := 0; i < len(incoming); i++ {
				//only want a single edge calculation in sum
				sum = GraphProb(incoming[i].Prob, node.Left.BranchLength) * GraphProb(incoming[i].Prob, node.Right.BranchLength)
				//missing probs needs to be derived from the same EDGE
				//the probability that this node was passed through in node.r/l and the modifier that changes that prob
				incoming[i].Prob = float32(sum)
			}
		} else if node.Right != nil {
			SetState(node.Right, position, graph)
			//if node.Right.Stored == nil {
			//	log.Fatal("no Stored values passed to internal node, right branch")
			//}
			for i := 0; i < 4; i++ {
				sum := 0.0
				for k := 0; k < 4; k++ {
					sum = sum + Prob(i, k, node.Right.BranchLength)*node.Right.Stored[k]
				}
				node.Stored[i] = sum
			}
		} else if node.Right == nil && node.Left == nil {
			//node.State = 4 // starts as N
			//node.Stored = []float64{0, 0, 0, 0}
			if len(graph.Nodes) <= gn {
				log.Fatal("node is out of range")
			} //not sure that i need an else (i don't need to use set or stored yet or maybe at all. Don't need to store the base,
			// should maybe return a graph and then that can get fed to GraphRecon to create a likely path

			//if len(node.Fasta.Seq) <= position {
			//	log.Fatal("position specified is out of range of sequence \n")
			//} else if len(node.Fasta.Seq) > position {
			//	node.State = int(node.Fasta.Seq[position])
			//	for i := 0; i < 4; i++ {
			//		if i == node.State {
			//			node.Stored[i] = 1
			//		} else {
			//			node.Stored[i] = 0
			//		}
			//	}
			//}
		}
	}
	return graph
}

//Bubble up the tree using the memory of the previous nodes
func BubbleUp(node *expandedTree.ETree, prevNode *expandedTree.ETree, scrap []float64) {
	tot := 0.0
	scrapNew := []float64{0, 0, 0, 0}
	for i := 0; i < 4; i++ {
		sum := 0.0
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				if prevNode.Up != nil {
					if prevNode == node.Left { //scrap is equal to one position of prevNode.Stored (Left or Right)
						sum = sum + Prob(i, j, node.Left.BranchLength)*Prob(i, k, node.Right.BranchLength)*scrap[j]*node.Right.Stored[k]
					} else if prevNode == node.Right {
						sum = sum + Prob(i, j, node.Left.BranchLength)*Prob(i, k, node.Right.BranchLength)*scrap[k]*node.Left.Stored[j]
					}
				} else if prevNode.Up == nil {
					sum = sum + Prob(i, j, node.Left.BranchLength)*Prob(i, k, node.Right.BranchLength)*node.Left.Stored[j]*node.Right.Stored[k]
				}
			}
		}
		scrapNew[i] = sum
	}
	if node.Up != nil {
		BubbleUp(node.Up, node, scrapNew)
	} else if node.Up == nil {
		tot = scrapNew[0] + scrapNew[1] + scrapNew[2] + scrapNew[3]
		node.Scrap = tot
	}
}

//fix each node and return the probabilities for each base at that site
func FixFc(root *expandedTree.ETree, node *expandedTree.ETree) []float64 {
	ans := []float64{0, 0, 0, 0}

	for i := 0; i < 4; i++ {
		scrap := []float64{0, 0, 0, 0} //checking one base at a time each time you call BubbleUp
		scrap[i] = node.Stored[i]
		if node.Up != nil {
			//Bubble up the tree using the memory of the previous node in relation to changing position taking in probabilities of bases
			//(node will be BubbleUp prevNode and node.Up will be the node being operated on)
			BubbleUp(node.Up, node, scrap) //node becomes PrevNode and scrap is set to one value of prevNode.Stored in BubbleUp
			ans[i] = root.Scrap            //root.Stored has previously assigned values (SetInternalState), you want to use whatever is returned by BubbleUp instead
		} else if node.Up == nil {
			ans[i] = root.Stored[i]
		}
	}
	return ans
}

//called by reconstructSeq.go on each base of the modern (leaf) seq. Loop over the nodes of the tree to return most probable base to the Fasta
//func LoopNodes(root *expandedTree.ETree, position int) {
//	internalNodes := expandedTree.GetBranch(root)
//	SetState(root, position)
//	for k := 0; k < len(internalNodes); k++ {
//		fix := FixFc(root, internalNodes[k])
//		yHat := Yhat(fix)
//		internalNodes[k].Fasta.Seq = append(internalNodes[k].Fasta.Seq, []dna.Base{dna.Base(yHat)}...)
//	}
//}

//call for GraphRecon will be a lot like loopNodes call but instead of calling it on each position it will be called for each graph column
//might not need a graph arg
func GraphRecon(root *expandedTree.ETree, column graphColumn, graph simpleGraph.SimpleGraph) {
	internalNodes := expandedTree.GetBranch(root)
	var rightPresent = false
	var leftPresent = false
	if root.Right != nil && root.Left != nil {
		for c := 0; c < len(column.AlignNodes); c++ {
			for s := 0; s < len(column.AlignNodes[c]); s++ {
				if column.AlignNodes[c][s].Name == root.Right.Name { //node names must contain a species name
					rightPresent = true
				}
				if column.AlignNodes[c][s].Name == root.Left.Name {
					leftPresent = true
				}
			}
		}
	}
	if rightPresent && leftPresent {
		var newNode *simpleGraph.Node
		var newAlign []*simpleGraph.Node
		//newNode := *simpleGraph.Node{Id, root.Name, seq, seqTwoBit, Prev, Next, info} //Id should probably be whatever number node this is for this species somehow
		//seq can just be determined the same way it always has been i guess? using stored and the old logic to determine most likely based at each position, node by node so we don't have to realign
		simpleGraph.AddNode(&graph, newNode)
		newAlign = append(newAlign, newNode)
		column.AlignNodes = append(column.AlignNodes, newAlign)
	}
}

//seqOfPath takes in a graph and, using PathFinder, returns the seq of the best path through the graph
func seqOfPath(g *simpleGraph.SimpleGraph) []dna.Base {
	var seq []dna.Base
	path, _ := PathFinder(g)
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
	//tempPath, tempProb and existingPaths variables are necessary for recursive calls of bestPath
	var tempPath = make([]uint32, 0)
	var tempProb float32
	var existingPaths = make(map[float32][]uint32)

	for n := 0; n < len(g.Nodes); n++ {
		if g.Nodes[n].Id == 0 {
			finalProb, finalPath = bestPath(g.Nodes[n], tempProb, tempPath, existingPaths)
		}
	}

	return finalPath, finalProb
}

//bestPath is the helper function for PathFinder, and recursively traverses the graph depth first to determine the most likely path from start to finish
func bestPath(node *simpleGraph.Node, prevProb float32, prevPath []uint32, existingPaths map[float32][]uint32) (prob float32, path []uint32) {
	//TODO: same prob values will overwrite existing value if they are equal, i need a way around that
	currentNode := node
	var currentProb float32 = 0
	currentPath := prevPath
	var tempProb float32 = 1
	var tempPath []uint32
	potentialPaths := existingPaths
	var finalProb float32
	var finalPath []uint32

	if prevProb != 0 && len(prevPath) > 0 {
		tempProb = prevProb
	}
	if len(currentNode.Next) == 0 {
		currentPath = append(currentPath, currentNode.Id)
		potentialPaths[tempProb] = currentPath
	}
	for i, _ := range currentNode.Next {
		tempPath = append(tempPath, currentPath...)
		tempPath = append(tempPath, currentNode.Id)
		if currentNode.Next[i].Prob*tempProb > currentProb {
			currentProb = currentNode.Next[i].Prob * tempProb
			checkPath := len(currentPath)
			if checkPath == 0 {
				currentPath = append(currentPath, currentNode.Id)
			} else if currentPath[checkPath-1] != currentNode.Id {
				currentPath = append(currentPath, currentNode.Id)
			}
			bestPath(currentNode.Next[i].Dest, currentProb, currentPath, potentialPaths)
		}
	}
	for prob, path := range potentialPaths {
		if prob > finalProb {
			finalPath = make([]uint32, 0)
			finalProb = prob
			finalPath = append(finalPath, path...)
		}
	}
	return finalProb, finalPath
}
