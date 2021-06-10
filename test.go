package main

import (
	"fmt"
	"time"
)

type KubeTask struct {
	// generated during graph parsing / creation
	Id TaskId
	// task metadata (e.g. which image to run, arguments etc.)
	Metadata KubeTaskMetadata
}

type KubeTaskMetadata struct {
	Name            string
	Image           string   // url of the docker image (with tag)
	Args            []string // list of string arguments to pass
	NumberOfRetries int
	// TODO: add additional properties (e.g. container limits)
}

func (n *KubeTask) GetId() TaskId {
	return n.Id
}

func (n *KubeTask) Run() error {
	fmt.Printf("task[%v]: started\n", n.Metadata.Name)
	time.Sleep(1 * time.Second)
	fmt.Printf("task[%v]: finished\n", n.Metadata.Name)
	return nil
}

func createTestGraph() TaskDependencyGraph {

	nodeA := KubeTask{Id: 0, Metadata: KubeTaskMetadata{
		"JobA", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{"30", "0.8"}, 1,
	}}

	nodeB := KubeTask{Id: 1, Metadata: KubeTaskMetadata{
		"JobB", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{"30", "0.8"}, 1,
	}}

	nodeC := KubeTask{Id: 2, Metadata: KubeTaskMetadata{
		"JobC", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{"30", "0.8"}, 1,
	}}

	nodeD := KubeTask{Id: 3, Metadata: KubeTaskMetadata{
		"JobD", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{"30", "0.8"}, 1,
	}}

	rootNode := KubeTask{Id: 4, Metadata: KubeTaskMetadata{
		Name: "GraphRoot",
	}}

	nodes := map[TaskId]Task{
		nodeA.Id:    &nodeA,
		nodeB.Id:    &nodeB,
		nodeC.Id:    &nodeC,
		nodeD.Id:    &nodeD,
		rootNode.Id: &rootNode,
	}

	edges := make(map[TaskId][]TaskId, 0)

	edges[nodeA.Id] = []TaskId{nodeB.Id, nodeC.Id}
	edges[nodeB.Id] = []TaskId{nodeD.Id}
	edges[nodeC.Id] = []TaskId{nodeD.Id}
	edges[nodeD.Id] = []TaskId{}
	edges[rootNode.Id] = []TaskId{nodeA.Id, nodeB.Id, nodeC.Id, nodeD.Id}

	return TaskDependencyGraph{&rootNode, nodes, edges}
}
