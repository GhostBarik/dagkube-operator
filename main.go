package main

import (
	"fmt"
	"time"
)

func createUserGraph() Graph {

	nodeA := Node{Id: 0, Name: "JobA"}
	nodeB := Node{Id: 1, Name: "JobB"}
	nodeC := Node{Id: 2, Name: "JobC"}
	nodeD := Node{Id: 3, Name: "JobD"}
	nodeE := Node{Id: 3, Name: "JobE"}

	rootNode := Node{Id: 4, Name: "GraphRoot"}

	nodes := map[NodeId]Node{
		nodeA.Id: nodeA,
		nodeB.Id: nodeB,
		nodeC.Id: nodeC,
		nodeD.Id: nodeD,
		nodeE.Id: nodeE,
		rootNode.Id: rootNode,
	}

	edges := make(map[NodeId][]NodeId, 0)

	edges[nodeA.Id] = []NodeId{ nodeB.Id, nodeC.Id }
	edges[nodeB.Id] = []NodeId{ nodeE.Id }
	edges[nodeC.Id] = []NodeId{ nodeE.Id, nodeB.Id }
	edges[nodeD.Id] = []NodeId{ nodeE.Id, nodeB.Id }
	edges[nodeE.Id] = []NodeId{ }
	edges[rootNode.Id] = []NodeId{ nodeA.Id, nodeB.Id, nodeC.Id, nodeD.Id, nodeE.Id }

	return Graph{ rootNode, nodes, edges }
}


func main() {

	t1 := time.Now()

	g := createUserGraph()
	fmt.Printf("graph: %v\n", g)

	g.RunGraph()
	fmt.Printf("the end, processing took: %v\n", time.Now().Sub(t1))
}

