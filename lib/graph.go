package lib

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type DepencyMap map[string]map[string]struct{}

// Create a map of channels for node task completion signals
var taskRelay = make(map[string]chan TaskStatus)

type Graph struct {
	File     string           `yml:"file"`
	Name     string           `yml:"name"`
	Tasks    map[string]*Task `yml:"tasks"`
	Parents  DepencyMap       `yml:"parents"`
	Children DepencyMap       `yml:"children"`
}

var withTaskFailures = false

// Execute directed acyclic graph
func (g *Graph) Execute() {
	dagExecutionStartTime := time.Now()
	log.Print("[\u2714 DAG START] execution started")

	// Initialise channels for task dependencies
	for taskKey := range g.Tasks {
		for parent := range g.Parents[taskKey] {
			relayKey := edgeKey(parent, taskKey)
			taskRelay[relayKey] = make(chan TaskStatus, 1)
		}
	}

	var waitGroup sync.WaitGroup
	for taskKey := range g.Tasks {
		waitGroup.Add(1)
		g.Tasks[taskKey].Status = Pending

		go func(taskKey string) {
			defer waitGroup.Done()

			if !g.waitForParents(taskKey) {
				g.skipTaskAndNotifyChildren(taskKey)
				return
			}

			g.Tasks[taskKey].execute(dagExecutionStartTime)
		}(taskKey)
	}

	waitGroup.Wait()
	if withTaskFailures {
		log.Warnf("[~ DAG COMPLETE] execution completed with failures")
	} else {
		log.Infof("[\u2714 DAG COMPLETE] execution successful")
	}
}

func edgeKey(from, to string) string {
	return fmt.Sprintf("%s->%s", from, to)
}

// Wait for parent tasks to complete
func (g *Graph) waitForParents(taskKey string) bool {
	for parent := range g.Parents[taskKey] {
		signal := <-taskRelay[edgeKey(parent, taskKey)]

		if signal == Failed || signal == Skipped {
			withTaskFailures = withTaskFailures || signal == Failed

			if g.Tasks[taskKey].ParentRule == AllSuccess {
				log.Warnf("[~ SKIPPED] parent task %s status %s, skipping %s", parent, signal, taskKey)
				return false
			}
		}
	}
	return true
}

func (g *Graph) skipTaskAndNotifyChildren(taskKey string) {
	for child := range g.Children[taskKey] {
		taskRelay[edgeKey(taskKey, child)] <- Skipped
	}
}
