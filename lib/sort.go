package lib

func Sort(g Graph) [][]string {
	gTemp := copyGraph(g)
	layers := [][]string{}

	for {
		leaves := gTemp.Leaves()
		if len(leaves) == 0 {
			break
		}

		layers = append(layers, leaves)
		for _, leafNode := range leaves {
			remove(gTemp, leafNode)
		}
	}

	return layers
}

func remove(g Graph, node string) {
	// Remove edges from things that depend on `node`.
	for dependent := range g.Children[node] {
		removeFromDepmap(g.Parents, dependent, node)
	}
	delete(g.Children, node)

	// Remove all edges from node to the things it depends on.
	for dependency := range g.Parents[node] {
		removeFromDepmap(g.Children, dependency, node)
	}
	delete(g.Parents, node)

	// Finally, remove the node itself.
	delete(g.Nodes, node)
}

func removeFromDepmap(dm depencyMap, key, node string) {
	nodes := dm[key]
	if len(nodes) == 1 {
		// The only element in the nodeset must be `node`, so we
		// can delete the entry entirely.
		delete(dm, key)
	} else {
		// Otherwise, remove the single node from the nodeset.
		delete(nodes, node)
	}
}

func copyNodeset(s nodeSet) nodeSet {
	out := make(nodeSet, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}

func copyDepencyMap(m depencyMap) depencyMap {
	out := make(depencyMap, len(m))
	for k, v := range m {
		out[k] = copyNodeset(v)
	}
	return out
}

func copyGraph(g Graph) Graph {
	return Graph{
		Parents:  copyDepencyMap(g.Parents),
		Children: copyDepencyMap(g.Children),
		Nodes:    copyNodeset(g.Nodes),
	}
}
