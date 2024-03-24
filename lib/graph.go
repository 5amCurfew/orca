package lib

import (
	"errors"
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

// //////////////////////////////
// Execute DAG
// //////////////////////////////
func (g *Graph) Execute(dagExecutionStartTime time.Time) {
	log.Printf("%s execution started", g.Name)

	// Create a Map of Channels for task completion
	completionChannels := make(map[string]chan bool)

	// Initialise channel for each task
	for taskName := range g.Tasks {
		completionChannels[taskName] = make(chan bool, 1)
	}

	// Use a WaitGroup to wait for all tasks to complete
	var waitGroup sync.WaitGroup

	// Create & Start goroutines for each task
	for taskName := range g.Tasks {
		// Increment the wait group
		waitGroup.Add(1)
		g.Tasks[taskName].Status = Pending

		// Start goroutines that are blocked until all parents have sent successful completion message
		go func(taskName string, completionChMap map[string]chan bool) {
			// Wait for all parents to complete
			for parent := range g.Parents[taskName] {
				<-completionChMap[parent]
			}

			g.Tasks[taskName].execute(dagExecutionStartTime, completionChMap, g)

			close(completionChMap[taskName])
			waitGroup.Done()
		}(taskName, completionChannels)
	}

	// Wait for all tasks to complete before exiting
	waitGroup.Wait()
	log.Printf("%s execution complete at %s\n", g.Name, time.Now().Format("2006-01-02 15:04:05"))
}

// //////////////////////////////
// Add Dependency Edges to Graph
// //////////////////////////////
func (g *Graph) addDependency(child, parent string) error {
	if child == parent {
		return errors.New("self-referential dependencies not allowed")
	}

	if g.dependsOn(parent, child) {
		return errors.New("circular dependencies not allowed")
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
