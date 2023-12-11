package lib

import "errors"

type nodeSet map[string]string
type depencyMap map[string]nodeSet

func addEdge(dm depencyMap, from, to string) {
	nodes, ok := dm[from]
	if !ok {
		nodes = make(nodeSet)
		dm[from] = nodes
	}
	nodes[to] = to
}

type Graph struct {
	Nodes    nodeSet    `json:"nodes"`
	Parents  depencyMap `json:"parents"`
	Children depencyMap `json:"children"`
}

func NewGraph() *Graph {
	return &Graph{
		Nodes:    make(nodeSet),
		Parents:  make(depencyMap),
		Children: make(depencyMap),
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
	g.Nodes[parent] = parent
	g.Nodes[child] = child

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

	for _, nextNode := range g.Children[node] {
		if _, ok := out[nextNode]; !ok {
			out[nextNode] = nextNode
			g.findDependencies(nextNode, out)
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
