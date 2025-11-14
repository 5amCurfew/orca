package lib

import (
	"fmt"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v2"
)

// Create Global Graph
var G Graph

type DepencyMap map[string]map[string]struct{}

var relayWG sync.WaitGroup
var notifyWG sync.WaitGroup
var statusWG sync.WaitGroup

// Create a map of channels for node task completion signals
var NodeRelay = make(map[string]chan NodeStatus)

type Graph struct {
	File     string           `yml:"file"`
	Name     string           `yml:"name"`
	Nodes    map[string]*Node `yml:"nodes"`
	Parents  DepencyMap       `yml:"parents"`
	Children DepencyMap       `yml:"children"`
}

var withTaskFailures = false

// ////////////////////////////////////////
// Graph Execution Function
// ////////////////////////////////////////
func (g *Graph) Execute() {
	dagExecutionStartTime := time.Now()

	model := NewDagModel(g)
	prog := tea.NewProgram(model)

	// runFinished receives the Run result when the TUI exits
	runFinished := make(chan error, 1)
	go func() {
		// Run blocks until the UI quits; run it in a goroutine so we can continue.
		_, err := prog.Run()
		runFinished <- err
	}()

	done := make(chan struct{})

	// ////////////////////////////////////////
	// Forward status messages to Bubble Tea
	// ////////////////////////////////////////
	// Send DAG start message
	prog.Send(DagStartMsg{Message: "[🚀 DAG START] executing tasks...\n"})

	for _, node := range g.Nodes {
		statusWG.Add(1)
		go func(n *Node) {
			defer statusWG.Done()
			for msg := range n.StatusChannel {
				prog.Send(msg) // Bubble Tea receives directly
			}
		}(node)
	}

	// Close done when all node consumers status messages have been processed
	go func() {
		statusWG.Wait()
		done <- struct{}{}
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
			relayWG.Add(1)
			g.Nodes[nodeKey].Status = Pending
			g.Nodes[nodeKey].StatusChannel <- NodeStatusMsg{NodeKey: nodeKey, Status: Pending}

			go func(nodeKey string) {
				defer close(g.Nodes[nodeKey].StatusChannel)
				defer relayWG.Done()

				if !g.waitForParents(nodeKey) {
					g.skipTaskAndNotifyChildren(nodeKey)
					return
				}

				g.Nodes[nodeKey].execute(dagExecutionStartTime)
			}(nodeKey)
		}

		relayWG.Wait()
		notifyWG.Wait()

		<-done // Wait for all messages to be processed

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
	// Wait for the TUI to finish before returning from Execute()
	if err := <-runFinished; err != nil {
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
	g.Nodes[nodeKey].StatusChannel <- NodeStatusMsg{NodeKey: nodeKey, Status: Skipped}
	for child := range g.Children[nodeKey] {
		NodeRelay[edgeKey(nodeKey, child)] <- Skipped
	}
}

// ////////////////////////////////////////
// Graph construction functions
// ////////////////////////////////////////

// Parse nodes from file
func (g *Graph) parseNodes() error {
	graphYML, err := os.ReadFile(g.File)
	if err != nil {
		return err
	}

	// Unmarshal YAML data into NodeYaml struct
	var nodeYaml NodeYaml
	err = yaml.Unmarshal(graphYML, &nodeYaml)
	if err != nil {
		return err
	}

	// Create a map of Nodes from the parsed YAML
	Nodes := make(map[string]*Node)
	for _, Node := range nodeYaml.Nodes {
		NodeCopy := Node // Create a copy of the Node
		NodeCopy.Status = Pending
		if NodeCopy.ParentRule == "" {
			NodeCopy.ParentRule = AllSuccess // Default parentRule if not set
		}

		if NodeCopy.Retries > 0 && NodeCopy.RetryDelay == 0 {
			NodeCopy.RetryDelay = 10 // Default retry delay if retries are set but delay is not
		} else if NodeCopy.Retries == 0 {
			NodeCopy.RetryDelay = 0 // No retry delay if no retries
		}

		Nodes[Node.Name] = &NodeCopy
	}

	g.Nodes = Nodes

	return nil
}

// Parse edges from File
func (g *Graph) parseEdges() error {
	// Open and read the YAML file
	graphYML, err := os.ReadFile(g.File)
	if err != nil {
		return err
	}

	// Define structure to match YAML format
	type DependencyYaml struct {
		Dependencies map[string][]string `yaml:"dependencies"`
	}

	// Unmarshal YAML data into DependencyYaml struct
	var dependencyYaml DependencyYaml
	err = yaml.Unmarshal(graphYML, &dependencyYaml)
	if err != nil {
		return err
	}

	// Loop through the dependencies and add them to the graph
	for Node, parents := range dependencyYaml.Dependencies {
		for _, parent := range parents {
			err := g.addDependency(Node, parent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Add edge to Graph
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

// True if child node depends on parent node (either directly or indirectly)
func (g *Graph) dependsOn(child, parent string) bool {
	allChildren := make(map[string]struct{})
	g.findAllChildren(parent, allChildren)
	_, isDependant := allChildren[child]
	return isDependant
}

// Find All Dependency Edges (direct and indriect)
func (g *Graph) findAllChildren(parent string, children map[string]struct{}) {
	if _, ok := g.Nodes[parent]; !ok {
		return
	}

	for child, nextChild := range g.Children[parent] {
		if _, ok := children[child]; !ok {
			children[child] = nextChild
			g.findAllChildren(child, children)
		}
	}
}

// ////////////////////////////////////////
// Utility Functions
// ////////////////////////////////////////

// Generate edge key
func edgeKey(from, to string) string {
	return fmt.Sprintf("%s->%s", from, to)
}

// Add edge
func addEdge(dm DepencyMap, from, to string) {
	nodes, ok := dm[from]
	if !ok {
		nodes = make(map[string]struct{})
		dm[from] = nodes
	}
	nodes[to] = struct{}{}
}
