package main

import (
	"fmt"
	"golang.org/x/sync/errgroup"
)

type TaskId int

type TaskMetadata struct {
	Name string
	Image string  // url of the docker image (with tag)
	Args []string // list of string arguments to pass
	NumberOfRetries int
	// TODO: add additional properties (e.g. container limits)
}

type Task interface {
	GetId() TaskId
	Run() error
}

type Dag struct {
	// our graph has always single defined root node, from which we start
	RootTask Task
	Tasks map[TaskId]Task
	Edges map[TaskId][]TaskId
	Errors e
}

// FIXME: temporary structure, remove
type Task2 struct {
	NodeId TaskId
	result chan bool
}

// FIXME: remove task2
type WaitMap map[TaskId][]Task2
type NotifyMap map[TaskId][]chan bool

func (g *Dag) RunDag() {

	// start from root node
	fmt.Println("start graph processing")

	// create set for storing visited nodes
	visitedSet := SynchronizedIntSet{ set: make(map[int]bool) }

	// create wait map (node -> list of nodes to wait for)
	waitMap := make(WaitMap)
	for startNodeId, children := range g.Edges {
		waitMap[startNodeId] = make([]Task2, 0)
		for _, endNodeId := range children {
			// create buffered channel with buffer size == 1
			// this will prevent notifier nodes from blocking
			// (i.e. while iterating through list of node to notify - in notifyMap[nodeId])
			task := Task2{ endNodeId, make(chan bool, 1) }
			waitMap[startNodeId] = append(waitMap[startNodeId], task)
		}
	}

	fmt.Printf("wait map: %v\n", waitMap)

	// create notify map (node -> list of nodes to notify)
	notifyMap := make(NotifyMap)
	for startNodeId, _ := range waitMap {
		notifyMap[startNodeId] = make([]chan bool, 0)
	}
	for _, tasks := range waitMap {
		for _, task := range tasks {
			endNode := task.NodeId
			notifyMap[endNode] = append(notifyMap[endNode], task.result)
		}
	}

	fmt.Printf("notify map: %v\n", notifyMap)

	singleRun := DagRun{
		dag: *g,
		visitedNodes: &visitedSet,
		waitMap: waitMap,
		notifyMap: notifyMap,
	}

	singleRun.processTask(g.RootTask.GetId())
}

type DagRun struct {
	dag Dag
	visitedNodes *SynchronizedIntSet
	waitMap WaitMap
	notifyMap NotifyMap
}

func (dagRun *DagRun) processTask (taskId TaskId) {

	if ok := dagRun.visitedNodes.addElement(int(taskId)); !ok {
		// node was already visited -> exit
		return
	}

	graph := dagRun.dag

	// run all children in parallel
	for _, childId := range graph.Edges[taskId] {
		go dagRun.processTask(childId)
	}

	// wait for completion for all dependencies
	for _, resultCh := range dagRun.waitMap[taskId] {
		<- resultCh.result
	}

	// root node does not perform any processing
	if taskId != graph.RootTask.GetId() {
		// TODO: handle error (not necessary now)
		_ = graph.Tasks[taskId].Run()
	}

	// send completion signal to dependent nodes
	for _, channel := range dagRun.notifyMap[taskId] {
		channel <- true // signal completion
	}
}


