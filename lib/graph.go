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

// Create a Map of Channels for task completion
var completionRelay = make(map[string]chan TaskStatus)

// //////////////////////////////
// Execute DAG
// //////////////////////////////
func (g *Graph) Execute() {
	dagExecutionStartTime := time.Now()
	log.Printf("[\u2714 DAG START] %s execution started", g.Name)

	// Initialise channel for each task dependency signal
	for taskKey := range g.Tasks {
		for parent := range g.Parents[taskKey] {
			completionRelay[fmt.Sprint(parent, "->", taskKey)] = make(chan TaskStatus, 1)
		}
	}

	// Use a WaitGroup to wait for all tasks to complete
	var waitGroup sync.WaitGroup

	// Create & Start goroutines for each task
	for taskKey := range g.Tasks {
		// Increment the wait group
		waitGroup.Add(1)
		g.Tasks[taskKey].Status = Pending

		// Start goroutines that are blocked until all parents have sent completion signal
		go func(taskKey string) {
			// Wait for all parents to complete
			defer waitGroup.Done()
			for parent := range g.Parents[taskKey] {
				signal := <-completionRelay[fmt.Sprint(parent, "->", taskKey)]

				if signal == Failed {
					withTaskFailures = true
					log.Warnf("[~ SKIPPED] parent task %s failed, skipping %s", parent, taskKey)
					for child := range g.Children[taskKey] {
						completionRelay[fmt.Sprint(taskKey, "->", child)] <- Skipped
					}
					close(completionRelay[fmt.Sprint(parent, "->", taskKey)])
					return
				}

				if signal == Skipped {
					log.Warnf("[~ SKIPPED] parent task %s skipped, skipping %s", parent, taskKey)
					for child := range g.Children[taskKey] {
						completionRelay[fmt.Sprint(taskKey, "->", child)] <- Skipped
					}
					close(completionRelay[fmt.Sprint(parent, "->", taskKey)])
					return
				}

				close(completionRelay[fmt.Sprint(parent, "->", taskKey)])
			}

			g.Tasks[taskKey].execute(dagExecutionStartTime, g)
		}(taskKey)
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
