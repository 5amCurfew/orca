package lib

import (
	"fmt"
	"strings"
)

// stages computes which execution stage each node belongs to.
// A node's stage is max(parent stages) + 1; roots are stage 0.
// Returns a slice of stages, each containing node names in declaration order.
func (g *Graph) stages() [][]string {
	stage := make(map[string]int, len(g.Nodes))
	for range g.NodeOrder {
		for _, name := range g.NodeOrder {
			for parent := range g.Parents[name] {
				if stage[parent]+1 > stage[name] {
					stage[name] = stage[parent] + 1
				}
			}
		}
	}

	maxStage := 0
	for _, s := range stage {
		if s > maxStage {
			maxStage = s
		}
	}
	result := make([][]string, maxStage+1)
	for _, name := range g.NodeOrder {
		s := stage[name]
		result[s] = append(result[s], name)
	}
	return result
}

// Visualise prints the DAG as annotated stages. Each stage groups nodes that
// can run in parallel; each node lists its direct parents to make edges explicit.
func (g *Graph) Visualise() {
	stages := g.stages()

	maxNameLen := 0
	for _, name := range g.NodeOrder {
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}

	fmt.Printf("dag: %s\n", g.Name)
	for i, nodes := range stages {
		fmt.Printf("\n  stage %d\n", i)
		for j, name := range nodes {
			connector := "├── "
			if j == len(nodes)-1 {
				connector = "└── "
			}

			var parents []string
			for _, n := range g.NodeOrder {
				if _, ok := g.Parents[name][n]; ok {
					parents = append(parents, n)
				}
			}

			if len(parents) == 0 {
				fmt.Printf("  %s%s\n", connector, name)
			} else {
				fmt.Printf("  %s%-*s  ← %s\n", connector, maxNameLen, name, strings.Join(parents, ", "))
			}
		}
	}
	fmt.Println()
}
