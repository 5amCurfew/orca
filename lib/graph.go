package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// DependencyMap is an adjacency set mapping a node name to the set of node names it is connected to.
type DependencyMap map[string]map[string]struct{}

// executionState holds the runtime-only machinery used during a single Execute
// call: channels, wait groups, and the failure flag. It is separate from the
// static graph topology so the two concerns do not mix.
type executionState struct {
	statusChannel chan NodeStatusMsg
	nodeRelay     map[string]chan NodeStatus // one buffered channel per directed edge, keyed by edgeKey
	anyFailed     atomic.Bool                // set to true in node.fail(); read after all goroutines finish
	waitGroup     sync.WaitGroup             // tracks all node execution goroutines
	notifyWG      sync.WaitGroup             // tracks all child-notification goroutines
}

// Graph is the in-memory representation of a DAG. It holds the parsed nodes
// and their dependency edges (topology), plus the runtime execution state.
type Graph struct {
	File      string
	Name      string
	Nodes     map[string]*Node
	NodeOrder []string // node names in their YAML declaration order
	Parents   DependencyMap
	Children  DependencyMap
	exec      executionState
}

// dagYAML is the top-level structure of a dag.yml file
type dagYAML struct {
	Nodes        []NodeSpec          `yaml:"nodes"`
	Dependencies map[string][]string `yaml:"dependencies"`
}

// NewGraph constructs a fully initialised Graph from the given YAML dag file.
// It validates the file path, parses nodes and dependency edges, and creates
// the .orca log directory. Returns an error if the file is missing, not a
// .yml file, or contains invalid node/dependency definitions.
func NewGraph(filePath string) (*Graph, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}
	if !strings.HasSuffix(filePath, ".yml") {
		return nil, fmt.Errorf("file must be a valid yaml file: %s", filePath)
	}

	g := &Graph{
		File:     filePath,
		Name:     strings.TrimSuffix(filepath.Base(filePath), ".yml"),
		Nodes:    make(map[string]*Node),
		Parents:  make(DependencyMap),
		Children: make(DependencyMap),
		exec:     executionState{nodeRelay: make(map[string]chan NodeStatus)},
	}

	if err := g.parse(); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(".orca", os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating .orca directory: %w", err)
	}

	return g, nil
}

// parse reads the dag file exactly once and populates g.Nodes, g.Parents,
// and g.Children. Node defaults (parentRule, retryDelay) are applied here.
func (g *Graph) parse() error {
	data, err := os.ReadFile(g.File)
	if err != nil {
		return fmt.Errorf("reading dag file: %w", err)
	}

	var dag dagYAML
	if err := yaml.Unmarshal(data, &dag); err != nil {
		return fmt.Errorf("parsing dag file: %w", err)
	}

	for _, n := range dag.Nodes {
		spec := n
		// Apply defaults for optional fields
		if spec.ParentRule == "" {
			spec.ParentRule = AllSuccess
		}
		if spec.Retries > 0 && spec.RetryDelay == 0 {
			spec.RetryDelay = 10
		} else if spec.Retries == 0 {
			spec.RetryDelay = 0
		}
		node := &Node{
			Spec:  spec,
			State: NodeState{Status: Pending},
		}
		g.Nodes[spec.Name] = node
		g.NodeOrder = append(g.NodeOrder, spec.Name)
	}

	for child, parents := range dag.Dependencies {
		if _, ok := g.Nodes[child]; !ok {
			return fmt.Errorf("dependency references undefined node: %q", child)
		}
		for _, parent := range parents {
			if _, ok := g.Nodes[parent]; !ok {
				return fmt.Errorf("dependency references undefined node: %q", parent)
			}
			if err := g.addDependency(child, parent); err != nil {
				return err
			}
		}
	}

	return nil
}

// ////////////////////////////////////////
// Graph Execution
// ////////////////////////////////////////

// Execute runs all nodes in the DAG concurrently, respecting their dependency
// order, and renders progress via the Bubble Tea TUI. It blocks until all
// nodes have finished (or been skipped) and the TUI exits.
func (g *Graph) Execute() {
	dagExecutionStartTime := time.Now()

	model := NewDagModel(g)
	prog := tea.NewProgram(model)

	// Size the buffer to hold every possible status message without blocking:
	// each node sends 1 initial Pending + up to (2*Retries + 2) execution messages.
	bufSize := 0
	for _, node := range g.Nodes {
		bufSize += 2*node.Spec.Retries + 3
	}
	g.exec.statusChannel = make(chan NodeStatusMsg, bufSize)
	done := make(chan struct{})

	// Forward status messages to Bubble Tea
	go func() {
		prog.Send(DagStartMsg{Message: "[🚀 DAG START] executing tasks...\n"})
		for msg := range g.exec.statusChannel {
			prog.Send(msg)
		}
		done <- struct{}{}
	}()

	// Orchestrate tasks
	go func() {
		// Initialise relay channels before any node goroutine is launched,
		// ensuring every channel exists before a node can attempt to read from it.
		for nodeKey := range g.Nodes {
			for parent := range g.Parents[nodeKey] {
				g.exec.nodeRelay[edgeKey(parent, nodeKey)] = make(chan NodeStatus, 1)
			}
		}

		for nodeKey := range g.Nodes {
			g.exec.waitGroup.Add(1)
			g.Nodes[nodeKey].State.Status = Pending
			g.exec.statusChannel <- NodeStatusMsg{NodeKey: nodeKey, Status: g.Nodes[nodeKey].State.Status}

			go func(nodeKey string) {
				defer g.exec.waitGroup.Done()

				if !g.waitForParents(nodeKey) {
					g.skipTaskAndNotifyChildren(nodeKey)
					return
				}

				g.Nodes[nodeKey].execute(g, dagExecutionStartTime)
			}(nodeKey)
		}

		g.exec.waitGroup.Wait()
		g.exec.notifyWG.Wait()

		close(g.exec.statusChannel)
		<-done

		prog.Send(tickMsg{})
		time.Sleep(50 * time.Millisecond)

		var completeMsg string
		if g.exec.anyFailed.Load() {
			completeMsg = "[⚠️  DAG COMPLETE] execution completed with failures\n"
		} else {
			completeMsg = "[✅ DAG COMPLETE] execution successful\n"
		}
		prog.Send(DagCompleteMsg{Message: completeMsg})
	}()

	if _, err := prog.Run(); err != nil {
		log.Errorf("Error running TUI: %v", err)
	}
}

// waitForParents blocks until every parent of nodeKey has sent a completion
// signal on its relay channel. Returns false if the node should be skipped
// (i.e. a parent failed or was skipped and the node's parentRule is AllSuccess).
// It has no side effects beyond reading from the relay channels.
func (g *Graph) waitForParents(nodeKey string) bool {
	for parent := range g.Parents[nodeKey] {
		signal := <-g.exec.nodeRelay[edgeKey(parent, nodeKey)]
		if signal == Failed || signal == Skipped {
			if g.Nodes[nodeKey].Spec.ParentRule == AllSuccess {
				return false
			}
		}
	}
	return true
}

// skipTaskAndNotifyChildren marks nodeKey as Skipped, forwards the Skipped
// signal to all direct children, and closes each relay channel — matching the
// lifecycle of the success/fail path in notifyChildren.
func (g *Graph) skipTaskAndNotifyChildren(nodeKey string) {
	g.exec.statusChannel <- NodeStatusMsg{NodeKey: nodeKey, Status: Skipped}
	for child := range g.Children[nodeKey] {
		key := edgeKey(nodeKey, child)
		g.exec.nodeRelay[key] <- Skipped
		close(g.exec.nodeRelay[key])
	}
}

// ////////////////////////////////////////
// Graph construction helpers
// ////////////////////////////////////////

// addDependency registers a directed edge from parent to child in both the
// Parents and Children adjacency maps. Returns an error on self-reference or
// if the edge would introduce a cycle.
func (g *Graph) addDependency(child, parent string) error {
	if child == parent {
		return fmt.Errorf("self-referential dependency: %s", child)
	}
	if g.dependsOn(parent, child) {
		return fmt.Errorf("circular dependency: %s, %s", child, parent)
	}
	addEdge(g.Parents, child, parent)
	addEdge(g.Children, parent, child)
	return nil
}

// dependsOn reports whether child is a transitive descendant of parent.
func (g *Graph) dependsOn(child, parent string) bool {
	allChildren := make(map[string]struct{})
	g.findAllChildren(parent, allChildren)
	_, isDependant := allChildren[child]
	return isDependant
}

// findAllChildren recursively collects all transitive children of parent into
// the provided set.
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

// edgeKey returns the map key used to identify the relay channel for the
// directed edge from -> to.
func edgeKey(from, to string) string {
	return fmt.Sprintf("%s->%s", from, to)
}

// addEdge inserts to into the adjacency set of from within dm.
func addEdge(dm DependencyMap, from, to string) {
	nodes, ok := dm[from]
	if !ok {
		nodes = make(map[string]struct{})
		dm[from] = nodes
	}
	nodes[to] = struct{}{}
}
