package main

func createTestGraph() TaskDependencyGraph {

	namespace := "default"

	image := "acrdagkube.azurecr.io/dagkube-poc:v0.1.0"
	params := []string{"1", "0.8"}
	retries := 3

	jobClient := CreateJobClient(namespace)

	nodeA := KubeTask{Id: 0, Metadata: KubeTaskMetadata{
		"job-a", image, params, retries},
		jobClient: jobClient,
	}

	nodeB := KubeTask{Id: 1, Metadata: KubeTaskMetadata{
		"job-b", image, params, retries},
		jobClient: jobClient,
	}

	nodeC := KubeTask{Id: 2, Metadata: KubeTaskMetadata{
		"job-c", image, params, retries},
		jobClient: jobClient,
	}

	nodeD := KubeTask{Id: 3, Metadata: KubeTaskMetadata{
		"job-d", image, params, retries},
		jobClient: jobClient,
	}

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
