package main

import (
	"fmt"
	"sync"
	"time"
)

type NodeId int

type Node struct {
	Id NodeId
	Name string
}

type Graph struct {
	RootNode Node
	Nodes map[NodeId]Node
	Edges map[NodeId][]NodeId
}

type SynchronizedIntSet struct {
	smu sync.Mutex
	set map[int]bool
}

func (s *SynchronizedIntSet) AddElement(elem int) bool {

	// lock `visited` set to check its value
	s.smu.Lock()

	// release lock at the end of function call
	defer s.smu.Unlock()

	// check if not is already visited
	if _, ok := s.set[elem]; ok {
		// element is already in set -> return false
		return false
	}

	// mark node as 'visited' to avoid running processing multiple times
	s.set[elem] = true

	// element was successfully added to set -> return true
	return true
}

type Task struct {
	NodeId NodeId
	result chan bool
}

type WaitMap map[NodeId][]Task
type NotifyMap map[NodeId][]chan bool


func (g *Graph) RunGraph() {

	// start from root node
	fmt.Println("start graph processing")

	// create set for storing visited nodes
	visitedSet := SynchronizedIntSet{ set: make(map[int]bool) }

	// create wait map (node -> list of nodes to wait for)
	waitMap := make(WaitMap)
	for startNodeId, children := range g.Edges {
		waitMap[startNodeId] = make([]Task, 0)
		for _, endNodeId := range children {
			// create buffered channel with buffer size == 1
			// this will prevent notifier nodes from blocking
			// (i.e. while iterating through list of node to notify - in notifyMap[nodeId])
			task := Task{ endNodeId, make(chan bool, 1) }
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
			notifyMap[task.NodeId] = append(notifyMap[task.NodeId], task.result)
		}
	}

	fmt.Printf("notify map: %v\n", notifyMap)

	g.ProcessNode(g.RootNode.Id, &visitedSet, waitMap, notifyMap)

	fmt.Println("graph processing has finished")
}

func (g *Graph) ProcessNode(nodeId NodeId, visitedNodes *SynchronizedIntSet, waitMap WaitMap, notifyMap NotifyMap) {

	if ok := visitedNodes.AddElement(int(nodeId)); !ok {
		// node was already visited -> exit
		return
	}

	// run all children in parallel
	for _, childId := range g.Edges[nodeId] {
		go g.ProcessNode(childId, visitedNodes, waitMap, notifyMap)
	}

	// wait for completion for all dependencies
	for _, resultCh := range waitMap[nodeId] {
		<- resultCh.result
	}

	// TODO: add Kubernetes API client logic ///
	// root node does not perform any processing
	if nodeId != g.RootNode.Id {
		// simulate some work for non-root note
		node := g.Nodes[nodeId]
		fmt.Printf("node[%v]: processing ...\n", node)
		time.Sleep(2 * time.Second)
		fmt.Printf("node[%v]: finished\n", node)
	}

	// send completion signal to dependent nodes
	for _, channel := range notifyMap[nodeId] {
		channel <- true // signal completion
	}
}

