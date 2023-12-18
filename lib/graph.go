package lib

import "errors"

type nodeSet map[string]struct{}
type depencyMap map[string]nodeSet

func addEdge(dm depencyMap, from, to string) {
	nodes, ok := dm[from]
	if !ok {
		nodes = make(nodeSet)
		dm[from] = nodes
	}
	nodes[to] = struct{}{}
}

type Graph struct {
	Tasks    map[string]*Task `json:"tasks"`
	Nodes    nodeSet          `json:"nodes"`
	Parents  depencyMap       `json:"parents"`
	Children depencyMap       `json:"children"`
	Layers   [][]string       `json:"layers"`
}

func NewGraph(tasks map[string]*Task) *Graph {
	return &Graph{
		Tasks:    tasks,
		Nodes:    make(nodeSet),
		Parents:  make(depencyMap),
		Children: make(depencyMap),
	}
}

func (g *Graph) AddNodes() {
	for task := range g.Tasks {
		g.Nodes[task] = struct{}{}
	}
}

func (g *Graph) DependOn(child, parent string) error {
	if child == parent {
		return errors.New("self-referential dependencies not allowed")
	}

	if g.dependsOn(parent, child) {
		return errors.New("circular dependencies not allowed")
	}

	// Add Nodes
	g.Nodes[parent] = struct{}{}
	g.Nodes[child] = struct{}{}

	// Add Edges
	addEdge(g.Parents, child, parent)
	addEdge(g.Children, parent, child)

	return nil
}

func (g *Graph) dependsOn(child, parent string) bool {
	deps := g.dependencies(child)
	_, ok := deps[parent]
	return ok
}

func (g *Graph) dependencies(root string) nodeSet {
	out := make(nodeSet)
	g.findDependencies(root, out)
	return out
}

func (g *Graph) findDependencies(node string, out nodeSet) {
	if _, ok := g.Nodes[node]; !ok {
		return
	}

	for key, nextNode := range g.Children[node] {
		if _, ok := out[key]; !ok {
			out[key] = nextNode
			g.findDependencies(key, out)
		}
	}
}

func (g *Graph) Leaves() []string {
	leaves := make([]string, 0)

	for node := range g.Nodes {
		if _, ok := g.Parents[node]; !ok {
			leaves = append(leaves, node)
		}
	}

	return leaves
}

func (g *Graph) CreateTopologicalLayers() {
	g.Layers = Sort(g)
}
