package lib

import "fmt"

// Add edges to Graph
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
	if _, ok := g.Tasks[parent]; !ok {
		return
	}

	for child, nextChild := range g.Children[parent] {
		if _, ok := children[child]; !ok {
			children[child] = nextChild
			g.findAllChildren(child, children)
		}
	}
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
