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
	time.Sleep(4 * time.Second)
	fmt.Printf("task[%v]: finished\n", t.Metadata.Name)
	return nil
}

var(
	rootTask = TestTask{ Id: 0, Metadata: TestTaskMetadata {Name: "TaskRoot"}}
	task1 = TestTask{ Id: 1, Metadata: TestTaskMetadata {"JobA"}}
	task2 = TestTask{ Id: 2, Metadata: TestTaskMetadata {"JobB"}}
	task3 = TestTask{ Id: 3, Metadata: TestTaskMetadata {"JobC"}}
	task4 = TestTask{ Id: 4, Metadata: TestTaskMetadata {"JobD"}}
	task5 = TestTask{ Id: 5, Metadata: TestTaskMetadata {"JobE"}}
)

// Test two node in this combination of graph
// O -> O
func TestTwoNode(t *testing.T) {
	nodes := map[TaskId]Task{
		task1.Id: &task1,
		task2.Id: &task2,
	}

	edges := make(map[TaskId][]TaskId, 0)

	edges[task1.Id] = []TaskId{ task2.Id }
	edges[rootTask.Id] = []TaskId{ task1.Id, task2.Id }

	dag := Dag{ &rootTask, nodes, edges, nil }

	dagRun := dag.DagRun()
	dagRun.Run()

	for err := range dagRun.errors {
		if err != nil {
			t.Error(err)
		}
	}
	t.Log("run successfully two node")
}

// Test four node in this combination of graph
//     O
//  /    \
// O      O
//  \    /
//    O
func TestFourNodes(t *testing.T) {

	nodes := map[TaskId]Task{
		task1.Id: &task1,
		task2.Id: &task2,
		task3.Id: &task3,
		task4.Id: &task4,
		rootTask.Id: &rootTask,
	}

	edges := make(map[TaskId][]TaskId, 0)

	edges[task1.Id] = []TaskId{}
	edges[task2.Id] = []TaskId{ task1.Id }
	edges[task3.Id] = []TaskId{ task1.Id }
	edges[task4.Id] = []TaskId{task2.Id, task3.Id }
	edges[rootTask.Id] = []TaskId{ task1.Id, task2.Id, task3.Id, task4.Id }

	dag := Dag{ &rootTask, nodes, edges, nil }

	dagRun := dag.DagRun()
	dagRun.Run()

	for err := range dagRun.errors {
		if err != nil {
			t.Error(err)
		}
	}
	t.Log("run successfully 4 nodes")
}

// Test five nodes in this combination of graph
//
//    O --- \
//  /   \    \
// O --- O - O
//  \   /    /
//    O -- /
//
func TestFiveNodes(t *testing.T) {

	nodes := map[TaskId]Task{
		task1.Id: &task1,
		task2.Id: &task2,
		task3.Id: &task3,
		task4.Id: &task4,
		task5.Id: &task5,
	}

	edges := make(map[TaskId][]TaskId, 0)

	edges[task1.Id] = []TaskId{}
	edges[task2.Id] = []TaskId{ task1.Id }
	edges[task3.Id] = []TaskId{ task1.Id }
	edges[task4.Id] = []TaskId{task1.Id, task2.Id, task3.Id }
	edges[task5.Id] = []TaskId{task2.Id, task3.Id, task4.Id }
	edges[rootTask.Id] = []TaskId{ task1.Id, task2.Id, task3.Id, task4.Id, task5.Id }

	dag := Dag{ &rootTask, nodes, edges, nil }

	dagRun := dag.DagRun()
	dagRun.Run()

	for err := range dagRun.errors {
		if err != nil {
			t.Error(err)
		}
	}
	t.Log("run successfully 5 nodes")
}
