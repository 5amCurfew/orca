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
var NodeRelay = make(map[string]chan NodeStatus)

type Graph struct {
	File          string           `yml:"file"`
	Name          string           `yml:"name"`
	Nodes         map[string]*Node `yml:"nodes"`
	Parents       DepencyMap       `yml:"parents"`
	Children      DepencyMap       `yml:"children"`
	StatusChannel chan NodeStatusMsg
}

var withTaskFailures = false

// Execute directed acyclic graph
func (g *Graph) Execute() {
	dagExecutionStartTime := time.Now()

	model := NewDagModel(g)
	prog := tea.NewProgram(model)
	g.StatusChannel = make(chan NodeStatusMsg, len(g.Nodes)*2)
	done := make(chan struct{}) // Add a done channel for synchronization

	// ////////////////////////////////////////
	// Forward status messages to Bubble Tea
	// ////////////////////////////////////////
	go func() {
		prog.Send(DagStartMsg{Message: "[🚀 DAG START] executing tasks...\n"})

		for msg := range g.StatusChannel {
			prog.Send(msg)
		}
		done <- struct{}{} // Signal that all messages have been processed
	}()

	// ////////////////////////////////////////
	// Initialise channels for task dependencies
	// ////////////////////////////////////////
	for nodeKey := range g.Nodes {
		for parent := range g.Parents[nodeKey] {
			relayKey := edgeKey(parent, nodeKey)
			NodeRelay[relayKey] = make(chan NodeStatus, 1)
		}
	}

	// ////////////////////////////////////////
	// Orchestrate tasks in a goroutine
	// ////////////////////////////////////////
	go func() {
		for nodeKey := range g.Nodes {
			waitGroup.Add(1)
			g.Nodes[nodeKey].Status = Pending
			g.StatusChannel <- NodeStatusMsg{NodeKey: nodeKey, Status: Pending}

			go func(nodeKey string) {
				defer waitGroup.Done()

				if !g.waitForParents(nodeKey) {
					g.skipTaskAndNotifyChildren(nodeKey)
					return
				}

				g.Nodes[nodeKey].execute(dagExecutionStartTime)
			}(nodeKey)
		}

		waitGroup.Wait()
		notifyWG.Wait()

		close(g.StatusChannel) // Close when done
		<-done                 // Wait for all messages to be processed

		prog.Send(tickMsg{}) // Send a tick to ensure final updates are rendered
		time.Sleep(50 * time.Millisecond)

		// Signal TUI to quit
		var completeMsg string
		if withTaskFailures {
			completeMsg = "[⚠️  DAG COMPLETE] execution completed with failures\n"
		} else {
			completeMsg = "[✅ DAG COMPLETE] execution successful\n"
		}
		prog.Send(DagCompleteMsg{Message: completeMsg})
	}()

	// ////////////////////////////////////////
	// Run the TUI (this blocks until the program exits)
	// ////////////////////////////////////////
	if _, err := prog.Run(); err != nil {
		fmt.Println("Error running TUI:", err)
	}
}

// Wait for parent tasks to complete
func (g *Graph) waitForParents(nodeKey string) bool {
	for parent := range g.Parents[nodeKey] {
		signal := <-NodeRelay[edgeKey(parent, nodeKey)]

		if signal == Failed || signal == Skipped {
			withTaskFailures = withTaskFailures || signal == Failed

			if g.Nodes[nodeKey].ParentRule == AllSuccess {
				return false
			}
		}
	}
	return true
}

// Skip task and notify children
func (g *Graph) skipTaskAndNotifyChildren(nodeKey string) {
	g.StatusChannel <- NodeStatusMsg{NodeKey: nodeKey, Status: Skipped}
	for child := range g.Children[nodeKey] {
		NodeRelay[edgeKey(nodeKey, child)] <- Skipped
	}
}

// ////////////////////////////////////////
// Utility Functions
// ////////////////////////////////////////

// Generate edge key
func edgeKey(from, to string) string {
	return fmt.Sprintf("%s->%s", from, to)
}
