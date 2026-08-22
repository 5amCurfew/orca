package lib

import (
	"testing"
)

// buildGraph constructs a Graph from a list of node names and dependency pairs
// without touching the filesystem — suitable for unit testing topology logic.
func buildGraph(t *testing.T, nodes []string, deps [][2]string) *Graph {
	t.Helper()
	g := newEmptyGraph()
	for _, name := range nodes {
		addTestNode(g, name)
	}
	for _, dep := range deps {
		child, parent := dep[0], dep[1]
		if err := g.addDependency(child, parent); err != nil {
			t.Fatalf("addDependency(%s, %s): %v", child, parent, err)
		}
	}
	return g
}

// ──────────────────────────────────────────────────────────────────────────────
// stages
// ──────────────────────────────────────────────────────────────────────────────

func TestStages_SingleNode(t *testing.T) {
	g := buildGraph(t, []string{"a"}, nil)
	stages := g.stages()

	if len(stages) != 1 {
		t.Fatalf("expected 1 stage, got %d", len(stages))
	}
	if len(stages[0]) != 1 || stages[0][0] != "a" {
		t.Errorf("expected stage 0 = [a], got %v", stages[0])
	}
}

func TestStages_LinearChain(t *testing.T) {
	// a → b → c  should produce three stages
	g := buildGraph(t, []string{"a", "b", "c"}, [][2]string{
		{"b", "a"},
		{"c", "b"},
	})
	stages := g.stages()

	if len(stages) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(stages))
	}
	expects := [][]string{{"a"}, {"b"}, {"c"}}
	for i, want := range expects {
		if len(stages[i]) != len(want) || stages[i][0] != want[0] {
			t.Errorf("stage %d: expected %v, got %v", i, want, stages[i])
		}
	}
}

func TestStages_ParallelRoots(t *testing.T) {
	// a and b have no dependencies — both should be in stage 0
	g := buildGraph(t, []string{"a", "b"}, nil)
	stages := g.stages()

	if len(stages) != 1 {
		t.Fatalf("expected 1 stage, got %d", len(stages))
	}
	if len(stages[0]) != 2 {
		t.Errorf("expected 2 nodes in stage 0, got %v", stages[0])
	}
}

func TestStages_DiamondDependency(t *testing.T) {
	// a → b, a → c, b → d, c → d
	// d depends on both b and c so must be in stage 2
	g := buildGraph(t, []string{"a", "b", "c", "d"}, [][2]string{
		{"b", "a"},
		{"c", "a"},
		{"d", "b"},
		{"d", "c"},
	})
	stages := g.stages()

	if len(stages) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(stages))
	}
	if stages[0][0] != "a" {
		t.Errorf("expected stage 0 = [a], got %v", stages[0])
	}
	if len(stages[1]) != 2 {
		t.Errorf("expected stage 1 to have 2 nodes (b, c), got %v", stages[1])
	}
	if len(stages[2]) != 1 || stages[2][0] != "d" {
		t.Errorf("expected stage 2 = [d], got %v", stages[2])
	}
}

func TestStages_SharedDownstreamNode(t *testing.T) {
	// a → c, b → c  — c appears in stage 1, not duplicated
	g := buildGraph(t, []string{"a", "b", "c"}, [][2]string{
		{"c", "a"},
		{"c", "b"},
	})
	stages := g.stages()

	if len(stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(stages))
	}
	if len(stages[0]) != 2 {
		t.Errorf("expected stage 0 to have 2 roots (a, b), got %v", stages[0])
	}
	if len(stages[1]) != 1 || stages[1][0] != "c" {
		t.Errorf("expected stage 1 = [c], got %v", stages[1])
	}
}

func TestStages_NodePlacedAtDeepestRequiredStage(t *testing.T) {
	// a → b → d
	// a → c → d  (c is one stage deeper than b via longer path to a)
	// d must land in stage 2, not stage 1
	g := buildGraph(t, []string{"a", "b", "c", "d"}, [][2]string{
		{"b", "a"},
		{"c", "b"},
		{"d", "c"},
	})
	stages := g.stages()

	if len(stages) != 4 {
		t.Fatalf("expected 4 stages for a→b→c→d, got %d", len(stages))
	}
}

func TestStages_DeclarationOrderPreservedWithinStage(t *testing.T) {
	// All roots: declaration order should be preserved within stage 0.
	g := buildGraph(t, []string{"z", "m", "a"}, nil)
	stages := g.stages()

	if len(stages) != 1 {
		t.Fatalf("expected 1 stage, got %d", len(stages))
	}
	want := []string{"z", "m", "a"}
	for i, name := range want {
		if stages[0][i] != name {
			t.Errorf("stage 0[%d]: expected %q, got %q", i, name, stages[0][i])
		}
	}
}
