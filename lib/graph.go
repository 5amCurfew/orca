package lib

import (
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type DepencyMap map[string]map[string]struct{}

var waitGroup sync.WaitGroup
var notifyWG sync.WaitGroup

// Create a map of channels for node task completion signals
var taskRelay = make(map[string]chan TaskStatus)

type Graph struct {
	File          string           `yml:"file"`
	Name          string           `yml:"name"`
	Tasks         map[string]*Task `yml:"tasks"`
	Parents       DepencyMap       `yml:"parents"`
	Children      DepencyMap       `yml:"children"`
	StatusChannel chan TaskStatusMsg
}

var withTaskFailures = false

// Execute directed acyclic graph
func (g *Graph) Execute() {
	dagExecutionStartTime := time.Now()

	model := NewDagModel(g)
	prog := tea.NewProgram(model)
	g.StatusChannel = make(chan TaskStatusMsg, len(g.Tasks)*2)

	// Forward status messages to Bubble Tea
	go func() {
		for msg := range g.StatusChannel {
			prog.Send(msg)
		}
	}()

	// Orchestrate tasks in a goroutine
	go func() {
		// Initialise channels for task dependencies
		for taskKey := range g.Tasks {
			for parent := range g.Parents[taskKey] {
				relayKey := edgeKey(parent, taskKey)
				taskRelay[relayKey] = make(chan TaskStatus, 1)
			}
		}

		for taskKey := range g.Tasks {
			waitGroup.Add(1)
			g.Tasks[taskKey].Status = Pending
			g.StatusChannel <- TaskStatusMsg{TaskKey: taskKey, Status: Pending}

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
		close(g.StatusChannel) // Close when done

		// Signal TUI to quit
		prog.Send(DagCompleteMsg{})
		var completeMsg string
		if withTaskFailures {
			completeMsg = "[⚠️  DAG COMPLETE] execution completed with failures"
		} else {
			completeMsg = "[✅ DAG COMPLETE] execution successful"
		}
		prog.Send(DagCompleteMsg{Message: completeMsg})
	}()

	// Run the TUI (this blocks until the program exits)
	if _, err := prog.Run(); err != nil {
		fmt.Println("Error running TUI:", err)
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
				return false
			}
		}
	}
	return true
}

func (g *Graph) skipTaskAndNotifyChildren(taskKey string) {
	g.StatusChannel <- TaskStatusMsg{TaskKey: taskKey, Status: Skipped}
	for child := range g.Children[taskKey] {
		taskRelay[edgeKey(taskKey, child)] <- Skipped
	}
}
