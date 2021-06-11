package main

import (
	"errors"
	"fmt"
)

type TaskId int

type Task interface {
	GetId() TaskId
	Run() error
}

type TaskDependencyGraph struct {
	// our graph has always single defined root node, from which we start
	RootTask     Task
	Tasks        map[TaskId]Task
	Dependencies map[TaskId][]TaskId
}

type WaitMap map[TaskId][]chan error
type NotifyMap map[TaskId][]chan error

func (g *TaskDependencyGraph) runGraph() DagRun {

	// start from root node
	fmt.Println("start graph processing")

	// create set for storing visited nodes
	visitedSet := SynchronizedIntSet{set: make(map[int]bool)}

	// create wait map (node -> list of nodes to wait for)
	waitMap := make(WaitMap)

	// create notify map (node -> list of nodes to notify)
	notifyMap := make(NotifyMap)

	// initialize both maps with channels
	for startNodeId, children := range g.Dependencies {
		for _, endNodeId := range children {

			// create buffered channel with buffer size == 1
			// this will prevent notifier nodes from blocking
			// (i.e. while iterating through list of node to notify - in notifyMap[nodeId])
			ch := make(chan error, 1)

			// put channel to both maps (wait/notify)
			waitMap[startNodeId] = append(waitMap[startNodeId], ch)
			notifyMap[endNodeId] = append(notifyMap[endNodeId], ch)
		}
	}

	fmt.Printf("wait map: %v\n", waitMap)
	fmt.Printf("notify map: %v\n", notifyMap)

	singleRun := DagRun{
		graph:        *g,
		visitedNodes: &visitedSet,
		waitMap:      waitMap,
		notifyMap:    notifyMap,
		errors: make(chan error, len(g.Tasks)),
	}

	return singleRun
}

func (dagRun *DagRun) run () {
	dagRun.processTask(dagRun.graph.RootTask.GetId())
}

type DagRun struct {
	graph        TaskDependencyGraph
	visitedNodes *SynchronizedIntSet
	waitMap      WaitMap
	notifyMap    NotifyMap
	errors       chan error
}

func (dagRun *DagRun) processTask(taskId TaskId) {

	if ok := dagRun.visitedNodes.addElement(int(taskId)); !ok {
		// node was already visited -> exit
		return
	}

	graph := dagRun.graph

	// run all children in parallel
	for _, childId := range graph.Dependencies[taskId] {
		go dagRun.processTask(childId)
	}

	var dependenciesWereSuccessful = true

	// wait for completion for all dependencies
	for _, resultCh := range dagRun.waitMap[taskId] {
		result := <-resultCh
		if result != nil {
			dependenciesWereSuccessful = false
		}
	}

	fmt.Printf("task[%v]: state for all dependencies -  %v\n", taskId, dependenciesWereSuccessful)

	var currentJobSuccess = dependenciesWereSuccessful

	// root task does not perform any processing
	if taskId != graph.RootTask.GetId() && dependenciesWereSuccessful {
		err := graph.Tasks[taskId].Run()
		dagRun.errors <- err
		if err != nil {
			currentJobSuccess = false
		}
	}

	// close errors channel
	if taskId == graph.RootTask.GetId() {
		close(dagRun.errors)
	}

	// send completion signal to dependent nodes
	for _, channel := range dagRun.notifyMap[taskId] {
		if !currentJobSuccess {
			channel <- errors.New(fmt.Sprintf("task[%v]: failed", taskId)) // signal completion
		} else {
			channel <- nil
		}
	}
}
