package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
)

type TaskId int

type Task interface {
	GetId() TaskId
	Run() error
}

type Dag struct {
	// our graph has always single defined root node, from which we start
	RootTask Task
	Tasks map[TaskId]Task
	Dependencies map[TaskId][]TaskId
	ErrGroup *errgroup.Group
}

type WaitMap map[TaskId][]chan bool
type NotifyMap map[TaskId][]chan bool


func (g *Dag) RunDag() {

	g.ErrGroup,_ = errgroup.WithContext(context.Background())

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
			ch := make(chan bool, 1)

			// put channel to both maps (wait/notify)
			waitMap[startNodeId] = append(waitMap[startNodeId], ch)
			notifyMap[endNodeId] = append(notifyMap[endNodeId], ch)
		}
	}

	fmt.Printf("wait map: %v\n", waitMap)
	fmt.Printf("notify map: %v\n", notifyMap)

	singleRun := DagRun{
		dag: *g,
		visitedNodes: &visitedSet,
		waitMap:      waitMap,
		notifyMap:    notifyMap,
	}

	singleRun.processTask(g.RootTask.GetId())
}

type DagRun struct {
	dag Dag
	visitedNodes *SynchronizedIntSet
	waitMap      WaitMap
	notifyMap    NotifyMap
}

func (dagRun *DagRun) processTask(taskId TaskId) error {

	if ok := dagRun.visitedNodes.addElement(int(taskId)); !ok {
		// node was already visited -> exit
		return nil
	}

	graph := dagRun.dag

	// run all children in parallel
	for _, childId := range graph.Dependencies[taskId] {
		dagRun.dag.ErrGroup.Go(func() error {
			if err := dagRun.processTask(childId); err!=nil {
				return err
			}
			return nil
		})
	}

	// wait for completion for all dependencies
	for _, resultCh := range dagRun.waitMap[taskId] {
		<-resultCh
	}

	// root node does not perform any processing
	if taskId != graph.RootTask.GetId() {
		if err:=graph.Tasks[taskId].Run();err!=nil {
			return err
		}
	}

	// send completion signal to dependent nodes
	for _, channel := range dagRun.notifyMap[taskId] {
		channel <- true // signal completion
	}

	return nil
}
