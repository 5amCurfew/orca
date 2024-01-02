package lib

import (
	"errors"
	"fmt"
	"sync"
)

type depencyMap map[string]map[string]struct{}

func addEdge(dm depencyMap, from, to string) {
	nodes, ok := dm[from]
	if !ok {
		nodes = make(map[string]struct{})
		dm[from] = nodes
	}
	nodes[to] = struct{}{}
}

type Graph struct {
	Tasks    map[string]*Task `json:"tasks"`
	Parents  depencyMap       `json:"parents"`
	Children depencyMap       `json:"children"`
}

func NewGraph(filePath string) (*Graph, error) {

	tasks, err := parseTasks(filePath)
	if err != nil {
		return &Graph{}, errors.New(err.Error())
	}

	g := &Graph{
		Tasks:    tasks,
		Parents:  make(depencyMap),
		Children: make(depencyMap),
	}

	err = g.parseDependencies(filePath)
	if err != nil {
		return &Graph{}, errors.New(err.Error())
	}

	return g, nil
}

func (g *Graph) parseDependencies(filePath string) error {
	err := parseDependencies(filePath, g)
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func (g *Graph) DependOn(child, parent string) error {
	if child == parent {
		return errors.New("self-referential dependencies not allowed")
	}

	if g.dependsOn(parent, child) {
		return errors.New("circular dependencies not allowed")
	}

	// Add Edges
	addEdge(g.Parents, child, parent)
	addEdge(g.Children, parent, child)
	g.Tasks[parent].Children = append(g.Tasks[parent].Children, g.Tasks[child])

	return nil
}

func (g *Graph) dependsOn(child, parent string) bool {
	_, ok := g.dependencies(parent)[child]
	return ok
}

func (g *Graph) dependencies(root string) map[string]struct{} {
	out := make(map[string]struct{})
	g.findDependencies(root, out)
	return out
}

func (g *Graph) findDependencies(node string, out map[string]struct{}) {
	if _, ok := g.Tasks[node]; !ok {
		return
	}

	for key, nextNode := range g.Children[node] {
		if _, ok := out[key]; !ok {
			out[key] = nextNode
			g.findDependencies(key, out)
		}
	}
}

// ExecuteDAG orchestrates and executes the tasks given the DAG.
func (g *Graph) ExecuteDAG() {
	// Create Channel Map for task completion
	completionChannels := make(map[string]chan bool)

	// Initialise channel for each task
	for taskName := range g.Tasks {
		completionChannels[taskName] = make(chan bool, 1)
	}

	// Use a WaitGroup to wait for all tasks to complete
	var waitGroup sync.WaitGroup

	// Create goroutines for each task
	for taskName := range g.Tasks {
		// Increment the wait group for each task
		g.Tasks[taskName].Status = Pending
		waitGroup.Add(1)
		// Start goroutines for each task that are blocked until all parents have sent successful completion message
		go func(taskName string, completionChMap map[string]chan bool) {
			// Wait for all parents to complete
			for parent := range g.Parents[taskName] {
				<-completionChMap[parent]
			}

			// Execute the task
			fmt.Println("executing ", taskName)
			executeTask(g.Tasks[taskName], completionChMap)

			close(completionChMap[taskName])
			waitGroup.Done()
		}(taskName, completionChannels)
	}

	// Wait for all tasks to complete before exiting
	waitGroup.Wait()
}
