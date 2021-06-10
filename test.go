package main

func createTestGraph() Dag {

	nodeA := KubeTask{ Id: 0, Metadata: TaskMetadata {
		"JobA", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{ "30", "0.8" }, 1,
	}}

	nodeB := KubeTask{ Id: 1, Metadata: TaskMetadata {
		"JobB", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{ "30", "0.8" }, 1,
	}}

	nodeC := KubeTask{ Id: 2, Metadata: TaskMetadata {
		"JobC", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{ "30", "0.8" }, 1,
	}}

	nodeD := KubeTask{ Id: 3, Metadata: TaskMetadata {
		"JobD", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{ "30", "0.8" }, 1,
	}}

	rootNode := KubeTask{ Id: 4, Metadata: TaskMetadata {
		Name: "GraphRoot",
	}}

	nodes := map[TaskId]Task{
		nodeA.Id: &nodeA,
		nodeB.Id: &nodeB,
		nodeC.Id: &nodeC,
		nodeD.Id: &nodeD,
		rootNode.Id: &rootNode,
	}

	edges := make(map[TaskId][]TaskId, 0)

	edges[nodeA.Id] = []TaskId{ nodeB.Id, nodeC.Id }
	edges[nodeB.Id] = []TaskId{ nodeD.Id }
	edges[nodeC.Id] = []TaskId{ nodeD.Id }
	edges[nodeD.Id] = []TaskId{ }
	edges[rootNode.Id] = []TaskId{ nodeA.Id, nodeB.Id, nodeC.Id, nodeD.Id }

	return Dag{ &rootNode, nodes, edges }
}