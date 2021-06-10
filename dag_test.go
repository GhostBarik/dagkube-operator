package main

import (
	"fmt"
	"testing"
	"time"
)

type TestTaskMetadata struct {
	Name            string
}

type TestTask struct {
	// generated during graph parsing, do not change!
	Id TaskId
	// task metadata (e.g. which image to run, arguments etc.)
	Metadata TestTaskMetadata
}

func (t *TestTask) GetId() TaskId {
	return t.Id
}

func (t *TestTask) Run() error {
	fmt.Printf("task[%v]: started\n", t.Metadata.Name)
	time.Sleep(3 * time.Second)
	fmt.Printf("task[%v]: finished\n", t.Metadata.Name)
	return nil
}

// Test two node in this combination of graph
func TestTwoNode(t *testing.T) {
	task1 := TestTask{ Id: 0, Metadata: TestTaskMetadata {"JobA"}}

	task2 := TestTask{ Id: 1, Metadata: TestTaskMetadata {"JobB"}}

	task3 := TestTask{ Id: 2, Metadata: TestTaskMetadata {"JobB"}}

	rootTask := TestTask{ Id: 4, Metadata: TestTaskMetadata {Name: "TaskRoot"}}

	nodes := map[TaskId]Task{
		task1.Id: &task1,
		task2.Id: &task2,
		task3.Id: &task3,
	}

	edges := make(map[TaskId][]TaskId, 0)

	edges[task1.Id] = []TaskId{ task2.Id }
	edges[rootTask.Id] = []TaskId{ task1.Id, task2.Id }

	dag := Dag{ &rootTask, nodes, edges, nil }

	dag.RunDag()

	if err := dag.ErrGroup.Wait(); err != nil {
		t.Fatal(err)
	} else {
		t.Log("dag successfully finished")
	}
}
