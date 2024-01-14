package lib

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/5amCurfew/orca/util"
)

type Graph struct {
	Name     string           `json:"name"`
	Tasks    map[string]*Task `json:"tasks"`
	Parents  util.DepencyMap  `json:"parents"`
	Children util.DepencyMap  `json:"children"`
}

func NewGraph(filePath string) (*Graph, error) {
	tasks, err := parseTasks(filePath)
	if err != nil {
		return &Graph{}, errors.New(err.Error())
	}

	g := &Graph{
		Name:     filePath[5:strings.Index(filePath, ".orca")],
		Tasks:    tasks,
		Parents:  make(util.DepencyMap),
		Children: make(util.DepencyMap),
	}

	err = g.parseDependencies(filePath)
	if err != nil {
		return &Graph{}, errors.New(err.Error())
	}

	dirPath := fmt.Sprintf("logs/%s", g.Name)
	err2 := os.MkdirAll(dirPath, os.ModePerm)
	if err2 != nil {
		fmt.Printf("Error creating logs directory: %s\n", err2)
	}

	return g, nil
}

// ExecuteDAG orchestrates and executes the tasks in the DAG
func (g *Graph) Execute(dagExecutionStartTime time.Time) {

	log.Printf("Execution start at %s (%s)\n", time.Now().Format("2006-01-02 15:04:05"), g.Name)

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

		// Start goroutines that is blocked until all parents have sent successful completion message
		go func(taskName string, completionChMap map[string]chan bool) {
			// Wait for all parents to complete
			for parent := range g.Parents[taskName] {
				<-completionChMap[parent]
			}

			// Execute
			task := g.Tasks[taskName]
			task.execute(g.Name, dagExecutionStartTime, completionChMap)

			close(completionChMap[taskName])
			waitGroup.Done()
		}(taskName, completionChannels)
	}

	// Wait for all tasks to complete before exiting
	waitGroup.Wait()
	log.Printf("Execution complete at %s (%s)\n", time.Now().Format("2006-01-02 15:04:05"), g.Name)
}

func (g *Graph) parseDependencies(filePath string) error {
	err := parseDependencies(filePath, g)
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func (g *Graph) dependOn(child, parent string) error {
	if child == parent {
		return errors.New("self-referential dependencies not allowed")
	}

	if g.dependsOn(parent, child) {
		return errors.New("circular dependencies not allowed")
	}

	// Add Edges
	util.AddEdge(g.Parents, child, parent)
	util.AddEdge(g.Children, parent, child)
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
