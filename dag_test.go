package main

import (
	"fmt"
	"testing"
	"time"
)

type TestTask struct {
	// generated during graph parsing, do not change!
	Id TaskId
	// task metadata (e.g. which image to run, arguments etc.)
	Metadata TaskMetadata
}

func (t *TestTask) GetId() TaskId {
	return t.Id
}

func (t *TestTask) Run() error {
	fmt.Printf("task[%v]: started\n", t.Metadata.Name)
	time.Sleep(1 * time.Second)
	fmt.Printf("task[%v]: finished\n", t.Metadata.Name)
	return nil
}

// Test two node in this combination of graph
func TestTwoNode(t *testing.T) {
	task1 := TestTask{ Id: 0, Metadata: TaskMetadata {
		"JobA", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{ "30", "0.8" }, 1,
	}}

	task2 := TestTask{ Id: 1, Metadata: TaskMetadata {
		"JobB", "acrdagkube.azurecr.io/dagkube-poc:v0.1.0", []string{ "30", "0.8" }, 1,
	}}

	rootTask := KubeTask{ Id: 4, Metadata: TaskMetadata {
		Name: "TaskRoot",
	}}

	nodes := map[TaskId]Task{
		task1.Id: &task1,
		task2.Id: &task2,
	}

	edges := make(map[TaskId][]TaskId, 0)

	edges[task1.Id] = []TaskId{ task2.Id }
	edges[rootTask.Id] = []TaskId{ task1.Id, task2.Id }

	dag := Dag{ &rootTask, nodes, edges }

	dag.RunDag()
}
