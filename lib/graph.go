package lib

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type DepencyMap map[string]map[string]struct{}

type Graph struct {
	File     string           `json:"file"`
	Name     string           `json:"name"`
	Tasks    map[string]*Task `json:"tasks"`
	Parents  DepencyMap       `json:"parents"`
	Children DepencyMap       `json:"children"`
}

var withTaskFailures = false

// //////////////////////////////
// Execute DAG
// //////////////////////////////
func (g *Graph) Execute() {
	dagExecutionStartTime := time.Now()
	log.Printf("[\u2714 DAG START] %s execution started", g.Name)

	// Create a Map of Channels for task completion
	completionRelay := make(map[string]chan bool)

	// Initialise channel for each task dependency signal
	for taskKey := range g.Tasks {
		for parent := range g.Parents[taskKey] {
			completionRelay[fmt.Sprint(parent, "->", taskKey)] = make(chan bool, 1)
		}
	}

	// Use a WaitGroup to wait for all tasks to complete
	var waitGroup sync.WaitGroup

	// Create & Start goroutines for each task
	for taskKey := range g.Tasks {
		// Increment the wait group
		waitGroup.Add(1)
		g.Tasks[taskKey].Status = Pending

		// Start goroutines that are blocked until all parents have sent successful completion message
		go func(taskKey string, completionChMap map[string]chan bool) {
			// Wait for all parents to complete
			for parent := range g.Parents[taskKey] {
				successSignal := <-completionRelay[fmt.Sprint(parent, "->", taskKey)]
				if !successSignal {
					withTaskFailures = true
					log.Warnf("[~ SKIPPED] parent task %s failed, aborting %s", parent, taskKey)
					for child := range g.Children[taskKey] {
						completionRelay[fmt.Sprint(taskKey, "->", child)] <- false
					}
					close(completionChMap[fmt.Sprint(parent, "->", taskKey)])
					waitGroup.Done()
					return
				}
				close(completionChMap[fmt.Sprint(parent, "->", taskKey)])
			}

			g.Tasks[taskKey].execute(dagExecutionStartTime, completionChMap, g)
			waitGroup.Done()
		}(taskKey, completionRelay)
	}

	// Wait for all tasks to complete before exiting
	waitGroup.Wait()
	if withTaskFailures {
		log.Warnf("[~ DAG COMPLETE] %s.orca execution completed with failures", g.Name)
	} else {
		log.Infof("[\u2714 DAG COMPLETE] %s.orca execution successful", g.Name)
	}
}

// //////////////////////////////
// Add Dependency Edges to Graph
// //////////////////////////////
func (g *Graph) addDependency(child, parent string) error {
	if child == parent {
		return fmt.Errorf("self-referential dependency: %s", child)
	}

	if g.dependsOn(parent, child) {
		return fmt.Errorf("circular dependency: %s, %s", child, parent)
	}

	// Add Edges
	addEdge(g.Parents, child, parent)
	addEdge(g.Children, parent, child)

	return nil
}

// //////////////////////////////
// "Does <CHILD> depend on <PARENT>?"
// //////////////////////////////
func (g *Graph) dependsOn(child, parent string) bool {
	allChildren := make(map[string]struct{})
	g.findAllChildren(parent, allChildren)
	_, isDependant := allChildren[child]
	return isDependant
}

// //////////////////////////////
// Find All Dependency Edges (direct and indriect)
// //////////////////////////////
func (g *Graph) findAllChildren(parent string, children map[string]struct{}) {
	if _, ok := g.Tasks[parent]; !ok {
		return
	}

	for child, nextChild := range g.Children[parent] {
		if _, ok := children[child]; !ok {
			children[child] = nextChild
			g.findAllChildren(child, children)
		}
	}
}

// //////////////////////////////
// Add Dependency Edge
// //////////////////////////////
func addEdge(dm DepencyMap, from, to string) {
	nodes, ok := dm[from]
	if !ok {
		nodes = make(map[string]struct{})
		dm[from] = nodes
	}
	nodes[to] = struct{}{}
}
