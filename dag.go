package main

import (
	"errors"
	"fmt"
)

type TaskResult struct {
	status bool
	err error
	message string
}

type TaskId int

type Task interface {
	GetId() TaskId
	Run() error
}

type Dag struct {
	// our graph has always single defined root node, from which we start
	RootTask     Task
	Tasks        map[TaskId]Task
	Dependencies map[TaskId][]TaskId
}

type WaitMap map[TaskId][]chan TaskResult
type NotifyMap map[TaskId][]chan TaskResult

func (g *Dag) DagRun() DagRun {

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
			ch := make(chan TaskResult, 1)

			// put channel to both maps (wait/notify)
			waitMap[startNodeId] = append(waitMap[startNodeId], ch)
			notifyMap[endNodeId] = append(notifyMap[endNodeId], ch)
		}
	}

	fmt.Printf("wait map: %v\n", waitMap)
	fmt.Printf("notify map: %v\n", notifyMap)

	dagRun := DagRun{
		dag:        *g,
		visitedNodes: &visitedSet,
		waitMap:      waitMap,
		notifyMap:    notifyMap,
		errors: make(chan error, len(g.Tasks)),
	}

	return dagRun
}

func (dagRun *DagRun) Run() {
	dagRun.processTask(dagRun.dag.RootTask.GetId())
}

type DagRun struct {
	dag        Dag
	visitedNodes *SynchronizedIntSet
	waitMap      WaitMap
	notifyMap    NotifyMap
	errors       chan error
}

func (dagRun *DagRun) processTask(taskId TaskId) {

	var(
		taskDependencyStatus bool = true
		taskStatus bool = true
	)
	if ok := dagRun.visitedNodes.addElement(int(taskId)); !ok {
		// node was already visited -> exit
		return
	}

	// run all children in parallel
	for _, childId := range dagRun.dag.Dependencies[taskId] {
		go dagRun.processTask(childId)
	}

	// wait for completion for all dependencies
	for _, resultCh := range dagRun.waitMap[taskId] {
		result := <-resultCh
		if result.err != nil {
			taskDependencyStatus = result.status
		}
	}

	fmt.Printf("task[%v]: state for all dependencies -  %v\n", taskId, taskDependencyStatus)

	// root task does not perform any processing
	if taskId != dagRun.dag.RootTask.GetId() && taskDependencyStatus {
		err := dagRun.dag.Tasks[taskId].Run()
		if err != nil {
			dagRun.errors <- err
			taskStatus = false
		}
	}

	// close errors channel
	if taskId == dagRun.dag.RootTask.GetId() {
		close(dagRun.errors)
	}

	// send completion signal to dependent nodes
	for _, channel := range dagRun.notifyMap[taskId] {
		if !taskStatus {
			err := errors.New(fmt.Sprintf("task[%v]: failed", taskId))
			channel <- TaskResult{
				status: false,
				err: err, // signal completion,
				message: err.Error(),
			}
		} else {
			channel <- TaskResult{status: true}
		}
	}
}
