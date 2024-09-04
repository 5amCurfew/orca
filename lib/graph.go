package lib

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type DepencyMap map[string]map[string]struct{}

// Create a map of channels for task completion signals
var taskRelay = make(map[string]chan TaskStatus)

type Graph struct {
	File     string           `yml:"file"`
	Name     string           `yml:"name"`
	Tasks    map[string]*Task `yml:"tasks"`
	Parents  DepencyMap       `yml:"parents"`
	Children DepencyMap       `yml:"children"`
}

var withTaskFailures = false

// //////////////////////////////
// Execute DAG
// //////////////////////////////
func (g *Graph) Execute() {
	dagExecutionStartTime := time.Now()
	log.Print("[\u2714 DAG START] execution started")

	// Initialise channel for each task dependency signal
	for taskKey := range g.Tasks {
		for parent := range g.Parents[taskKey] {
			taskRelay[fmt.Sprint(parent, "->", taskKey)] = make(chan TaskStatus, 1)
		}
	}

	// Use a WaitGroup to wait for all tasks to complete
	var waitGroup sync.WaitGroup

	// Create & Start goroutines for each task
	for taskKey := range g.Tasks {
		// Increment the wait group
		waitGroup.Add(1)
		g.Tasks[taskKey].Status = Pending

		// Start goroutines that are blocked until all parents of the task have sent completion signal
		go func(taskKey string) {
			// Wait for all parents to complete
			defer waitGroup.Done()
			for parent := range g.Parents[taskKey] {
				signal := <-taskRelay[fmt.Sprint(parent, "->", taskKey)]

				if signal == Failed {
					withTaskFailures = true
					if g.Tasks[taskKey].ParentRule == AllSuccess {
						log.Warnf("[~ SKIPPED] parent task %s failed, skipping %s", parent, taskKey)
						for child := range g.Children[taskKey] {
							taskRelay[fmt.Sprint(taskKey, "->", child)] <- Skipped
						}
						close(taskRelay[fmt.Sprint(parent, "->", taskKey)])
						return
					}
				}

				if signal == Skipped {
					if g.Tasks[taskKey].ParentRule == AllSuccess {
						log.Warnf("[~ SKIPPED] parent task %s was skipped, skipping %s", parent, taskKey)
						for child := range g.Children[taskKey] {
							taskRelay[fmt.Sprint(taskKey, "->", child)] <- Skipped
						}
						close(taskRelay[fmt.Sprint(parent, "->", taskKey)])
						return
					}
				}

				close(taskRelay[fmt.Sprint(parent, "->", taskKey)])

			}

			g.Tasks[taskKey].execute(dagExecutionStartTime)
		}(taskKey)
	}

	// Wait for all tasks to complete before exiting
	waitGroup.Wait()
	if withTaskFailures {
		log.Warnf("[~ DAG COMPLETE] execution completed with failures")
	} else {
		log.Infof("[\u2714 DAG COMPLETE] execution successful")
	}
}
