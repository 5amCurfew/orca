package lib

import (
	"os"
	"path/filepath"
	"testing"
)

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

// writeYAML writes content to a file named test.yml inside dir and returns its
// absolute path.
func writeYAML(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "test.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeYAML: %v", err)
	}
	return path
}

// newEmptyGraph returns a Graph with all maps initialised but no nodes or edges.
// Use this to test topology methods without hitting the filesystem.
func newEmptyGraph() *Graph {
	return &Graph{
		Nodes:    make(map[string]*Node),
		Parents:  make(DependencyMap),
		Children: make(DependencyMap),
		exec:     executionState{nodeRelay: make(map[string]chan NodeStatus)},
	}
}

// addTestNode appends a named node to g.
func addTestNode(g *Graph, name string) {
	g.Nodes[name] = &Node{Spec: NodeSpec{Name: name}}
	g.NodeOrder = append(g.NodeOrder, name)
}

// ──────────────────────────────────────────────────────────────────────────────
// NewGraph — file validation
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGraph_FileNotFound(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := NewGraph("nonexistent.yml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestNewGraph_NonYMLExtension(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	path := filepath.Join(tmpDir, "dag.yaml")
	if err := os.WriteFile(path, []byte("nodes: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := NewGraph(path)
	if err == nil {
		t.Fatal("expected error for non-.yml extension, got nil")
	}
}

func TestNewGraph_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	path := writeYAML(t, tmpDir, "key: [unclosed bracket\n")
	_, err := NewGraph(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// NewGraph — parsing and field population
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGraph_ValidLinearDAG(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := `
nodes:
  - name: node-a
    desc: first
    cmd: echo a
  - name: node-b
    desc: second
    cmd: echo b
dependencies:
  node-b: [node-a]
`
	path := writeYAML(t, tmpDir, content)
	g, err := NewGraph(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(g.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(g.Nodes))
	}
	if _, ok := g.Nodes["node-a"]; !ok {
		t.Error("node-a missing from graph")
	}
	if _, ok := g.Nodes["node-b"]; !ok {
		t.Error("node-b missing from graph")
	}

	// Declaration order preserved
	if len(g.NodeOrder) != 2 || g.NodeOrder[0] != "node-a" || g.NodeOrder[1] != "node-b" {
		t.Errorf("unexpected NodeOrder: %v", g.NodeOrder)
	}

	// Edges recorded in both directions
	if _, ok := g.Parents["node-b"]["node-a"]; !ok {
		t.Error("node-a should be a parent of node-b")
	}
	if _, ok := g.Children["node-a"]["node-b"]; !ok {
		t.Error("node-b should be a child of node-a")
	}
}

func TestNewGraph_NoDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := `
nodes:
  - name: standalone
    cmd: echo hi
`
	path := writeYAML(t, tmpDir, content)
	g, err := NewGraph(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(g.Nodes))
	}
	if len(g.Parents["standalone"]) != 0 {
		t.Error("standalone should have no parents")
	}
}

func TestNewGraph_GraphNameDerivedFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	path := filepath.Join(tmpDir, "mypipeline.yml")
	if err := os.WriteFile(path, []byte("nodes:\n  - name: a\n    cmd: echo a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	g, err := NewGraph(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Name != "mypipeline" {
		t.Errorf("expected Name %q, got %q", "mypipeline", g.Name)
	}
}

func TestNewGraph_OrcaDirectoryCreated(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	path := writeYAML(t, tmpDir, "nodes:\n  - name: a\n    cmd: echo a\n")
	if _, err := NewGraph(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(".orca"); os.IsNotExist(err) {
		t.Error(".orca directory was not created")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// NewGraph — node defaults
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGraph_DefaultParentRuleIsAllSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	path := writeYAML(t, tmpDir, "nodes:\n  - name: a\n    cmd: echo a\n")
	g, err := NewGraph(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Nodes["a"].Spec.ParentRule != AllSuccess {
		t.Errorf("expected AllSuccess, got %q", g.Nodes["a"].Spec.ParentRule)
	}
}

func TestNewGraph_ExplicitCompleteRulePreserved(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := "nodes:\n  - name: a\n    cmd: echo a\n    parentRule: complete\n"
	path := writeYAML(t, tmpDir, content)
	g, err := NewGraph(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Nodes["a"].Spec.ParentRule != AllComplete {
		t.Errorf("expected AllComplete, got %q", g.Nodes["a"].Spec.ParentRule)
	}
}

func TestNewGraph_RetryDelayDefaultsTen_WhenRetriesSet(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := "nodes:\n  - name: a\n    cmd: echo a\n    retries: 3\n"
	path := writeYAML(t, tmpDir, content)
	g, err := NewGraph(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Nodes["a"].Spec.RetryDelay != 10 {
		t.Errorf("expected retryDelay 10, got %d", g.Nodes["a"].Spec.RetryDelay)
	}
}

func TestNewGraph_ExplicitRetryDelayPreserved(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := "nodes:\n  - name: a\n    cmd: echo a\n    retries: 2\n    retryDelay: 5\n"
	path := writeYAML(t, tmpDir, content)
	g, err := NewGraph(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Nodes["a"].Spec.RetryDelay != 5 {
		t.Errorf("expected retryDelay 5, got %d", g.Nodes["a"].Spec.RetryDelay)
	}
}

func TestNewGraph_NoRetries_RetryDelayIsZero(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	path := writeYAML(t, tmpDir, "nodes:\n  - name: a\n    cmd: echo a\n")
	g, err := NewGraph(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Nodes["a"].Spec.RetryDelay != 0 {
		t.Errorf("expected retryDelay 0, got %d", g.Nodes["a"].Spec.RetryDelay)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// NewGraph — cycle / self-reference detection
// ──────────────────────────────────────────────────────────────────────────────

func TestNewGraph_SelfReferentialDependency(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := "nodes:\n  - name: a\n    cmd: echo a\ndependencies:\n  a: [a]\n"
	path := writeYAML(t, tmpDir, content)
	_, err := NewGraph(path)
	if err == nil {
		t.Fatal("expected error for self-referential dependency, got nil")
	}
}

func TestNewGraph_CircularDependency(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := `
nodes:
  - name: a
    cmd: echo a
  - name: b
    cmd: echo b
dependencies:
  a: [b]
  b: [a]
`
	path := writeYAML(t, tmpDir, content)
	_, err := NewGraph(path)
	if err == nil {
		t.Fatal("expected error for circular dependency, got nil")
	}
}

func TestNewGraph_ThreeNodeCycle(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := `
nodes:
  - name: a
    cmd: echo a
  - name: b
    cmd: echo b
  - name: c
    cmd: echo c
dependencies:
  b: [a]
  c: [b]
  a: [c]
`
	path := writeYAML(t, tmpDir, content)
	_, err := NewGraph(path)
	if err == nil {
		t.Fatal("expected error for 3-node cycle, got nil")
	}
}

func TestNewGraph_UndefinedChildInDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := `
nodes:
  - name: a
    cmd: echo a
dependencies:
  ghost: [a]
`
	path := writeYAML(t, tmpDir, content)
	_, err := NewGraph(path)
	if err == nil {
		t.Fatal("expected error for undefined child node in dependencies, got nil")
	}
}

func TestNewGraph_UndefinedParentInDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	content := `
nodes:
  - name: a
    cmd: echo a
dependencies:
  a: [ghost]
`
	path := writeYAML(t, tmpDir, content)
	_, err := NewGraph(path)
	if err == nil {
		t.Fatal("expected error for undefined parent node in dependencies, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// edgeKey
// ──────────────────────────────────────────────────────────────────────────────

func TestEdgeKey_Format(t *testing.T) {
	got := edgeKey("parent", "child")
	want := "parent->child"
	if got != want {
		t.Errorf("edgeKey: got %q, want %q", got, want)
	}
}

func TestEdgeKey_IsUnique(t *testing.T) {
	// Two edges that share a node but in opposite directions must have distinct keys.
	ab := edgeKey("a", "b")
	ba := edgeKey("b", "a")
	if ab == ba {
		t.Errorf("edgeKey should differ for opposite directions, both got %q", ab)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// addDependency
// ──────────────────────────────────────────────────────────────────────────────

func TestAddDependency_AddsEdgesInBothMaps(t *testing.T) {
	g := newEmptyGraph()
	addTestNode(g, "a")
	addTestNode(g, "b")

	if err := g.addDependency("b", "a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := g.Parents["b"]["a"]; !ok {
		t.Error("a should be in Parents[b]")
	}
	if _, ok := g.Children["a"]["b"]; !ok {
		t.Error("b should be in Children[a]")
	}
}

func TestAddDependency_SelfReference(t *testing.T) {
	g := newEmptyGraph()
	addTestNode(g, "a")
	if err := g.addDependency("a", "a"); err == nil {
		t.Error("expected error for self-referential dependency")
	}
}

func TestAddDependency_DirectCycle(t *testing.T) {
	g := newEmptyGraph()
	addTestNode(g, "a")
	addTestNode(g, "b")

	if err := g.addDependency("b", "a"); err != nil {
		t.Fatalf("first edge unexpected error: %v", err)
	}
	if err := g.addDependency("a", "b"); err == nil {
		t.Error("expected error when adding a cycle")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// dependsOn / findAllChildren
// ──────────────────────────────────────────────────────────────────────────────

func TestDependsOn_DirectEdge(t *testing.T) {
	g := newEmptyGraph()
	addTestNode(g, "a")
	addTestNode(g, "b")
	if err := g.addDependency("b", "a"); err != nil {
		t.Fatal(err)
	}
	// "b" is a child of "a" — b depends directly on a.
	if !g.dependsOn("b", "a") {
		t.Error("b should depend on a (direct edge)")
	}
}

func TestDependsOn_TransitiveEdge(t *testing.T) {
	g := newEmptyGraph()
	addTestNode(g, "a")
	addTestNode(g, "b")
	addTestNode(g, "c")
	if err := g.addDependency("b", "a"); err != nil {
		t.Fatal(err)
	}
	if err := g.addDependency("c", "b"); err != nil {
		t.Fatal(err)
	}
	if !g.dependsOn("c", "a") {
		t.Error("c should transitively depend on a")
	}
}

func TestDependsOn_NoRelation(t *testing.T) {
	g := newEmptyGraph()
	addTestNode(g, "a")
	addTestNode(g, "b")
	if g.dependsOn("a", "b") {
		t.Error("a should not depend on b when no edge exists")
	}
}

func TestFindAllChildren_Linear(t *testing.T) {
	g := newEmptyGraph()
	for _, n := range []string{"a", "b", "c"} {
		addTestNode(g, n)
	}
	if err := g.addDependency("b", "a"); err != nil {
		t.Fatal(err)
	}
	if err := g.addDependency("c", "b"); err != nil {
		t.Fatal(err)
	}

	all := make(map[string]struct{})
	g.findAllChildren("a", all)

	for _, want := range []string{"b", "c"} {
		if _, ok := all[want]; !ok {
			t.Errorf("expected %q in transitive children of a", want)
		}
	}
}

func TestFindAllChildren_BranchingDAG(t *testing.T) {
	g := newEmptyGraph()
	for _, n := range []string{"root", "left", "right", "leaf"} {
		addTestNode(g, n)
	}
	if err := g.addDependency("left", "root"); err != nil {
		t.Fatal(err)
	}
	if err := g.addDependency("right", "root"); err != nil {
		t.Fatal(err)
	}
	if err := g.addDependency("leaf", "left"); err != nil {
		t.Fatal(err)
	}

	all := make(map[string]struct{})
	g.findAllChildren("root", all)

	for _, want := range []string{"left", "right", "leaf"} {
		if _, ok := all[want]; !ok {
			t.Errorf("expected %q in transitive children of root", want)
		}
	}
}

func TestFindAllChildren_UnknownParent(t *testing.T) {
	g := newEmptyGraph()
	all := make(map[string]struct{})
	g.findAllChildren("ghost", all) // should not panic
	if len(all) != 0 {
		t.Errorf("expected empty set for unknown parent, got %v", all)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// waitForParents
// ──────────────────────────────────────────────────────────────────────────────

// makeParentChildGraph builds a 2-node graph, pre-populates the relay channel
// with parentStatus, and returns the graph ready for waitForParents.
func makeParentChildGraph(rule ParentRule, parentStatus NodeStatus) *Graph {
	g := &Graph{
		Nodes: map[string]*Node{
			"parent": {Spec: NodeSpec{Name: "parent"}},
			"child":  {Spec: NodeSpec{Name: "child", ParentRule: rule}},
		},
		Parents:  DependencyMap{"child": {"parent": struct{}{}}},
		Children: DependencyMap{"parent": {"child": struct{}{}}},
		exec: executionState{
			nodeRelay:     make(map[string]chan NodeStatus),
			statusChannel: make(chan NodeStatusMsg, 4),
		},
	}
	ch := make(chan NodeStatus, 1)
	ch <- parentStatus
	close(ch)
	g.exec.nodeRelay[edgeKey("parent", "child")] = ch
	return g
}

func TestWaitForParents_NoParents_AlwaysTrue(t *testing.T) {
	g := &Graph{
		Nodes:   map[string]*Node{"a": {Spec: NodeSpec{Name: "a"}}},
		Parents: make(DependencyMap),
		exec:    executionState{nodeRelay: make(map[string]chan NodeStatus)},
	}
	if !g.waitForParents("a") {
		t.Error("node with no parents should return true")
	}
}

func TestWaitForParents_ParentSucceeded_AllSuccessRule(t *testing.T) {
	g := makeParentChildGraph(AllSuccess, Success)
	if !g.waitForParents("child") {
		t.Error("expected true when parent succeeded under AllSuccess rule")
	}
}

func TestWaitForParents_ParentFailed_AllSuccessRule(t *testing.T) {
	g := makeParentChildGraph(AllSuccess, Failed)
	if g.waitForParents("child") {
		t.Error("expected false when parent failed under AllSuccess rule")
	}
}

func TestWaitForParents_ParentSkipped_AllSuccessRule(t *testing.T) {
	g := makeParentChildGraph(AllSuccess, Skipped)
	if g.waitForParents("child") {
		t.Error("expected false when parent was skipped under AllSuccess rule")
	}
}

func TestWaitForParents_ParentFailed_CompleteRule(t *testing.T) {
	g := makeParentChildGraph(AllComplete, Failed)
	if !g.waitForParents("child") {
		t.Error("expected true when parent failed but rule is AllComplete")
	}
}

func TestWaitForParents_ParentSkipped_CompleteRule(t *testing.T) {
	g := makeParentChildGraph(AllComplete, Skipped)
	if !g.waitForParents("child") {
		t.Error("expected true when parent skipped but rule is AllComplete")
	}
}

func TestWaitForParents_MultipleParents_AllSucceed(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"p1":    {Spec: NodeSpec{Name: "p1"}},
			"p2":    {Spec: NodeSpec{Name: "p2"}},
			"child": {Spec: NodeSpec{Name: "child", ParentRule: AllSuccess}},
		},
		Parents: DependencyMap{"child": {"p1": struct{}{}, "p2": struct{}{}}},
		exec: executionState{
			nodeRelay:     make(map[string]chan NodeStatus),
			statusChannel: make(chan NodeStatusMsg, 4),
		},
	}
	for _, p := range []string{"p1", "p2"} {
		ch := make(chan NodeStatus, 1)
		ch <- Success
		close(ch)
		g.exec.nodeRelay[edgeKey(p, "child")] = ch
	}
	if !g.waitForParents("child") {
		t.Error("expected true when all parents succeeded")
	}
}

func TestWaitForParents_MultipleParents_OneFails(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"p1":    {Spec: NodeSpec{Name: "p1"}},
			"p2":    {Spec: NodeSpec{Name: "p2"}},
			"child": {Spec: NodeSpec{Name: "child", ParentRule: AllSuccess}},
		},
		Parents: DependencyMap{"child": {"p1": struct{}{}, "p2": struct{}{}}},
		exec: executionState{
			nodeRelay:     make(map[string]chan NodeStatus),
			statusChannel: make(chan NodeStatusMsg, 4),
		},
	}
	statuses := map[string]NodeStatus{"p1": Success, "p2": Failed}
	for p, s := range statuses {
		ch := make(chan NodeStatus, 1)
		ch <- s
		close(ch)
		g.exec.nodeRelay[edgeKey(p, "child")] = ch
	}
	if g.waitForParents("child") {
		t.Error("expected false when one parent failed under AllSuccess rule")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// skipTaskAndNotifyChildren
// ──────────────────────────────────────────────────────────────────────────────

func TestSkipTaskAndNotifyChildren_NoChildren(t *testing.T) {
	g := &Graph{
		Nodes:    map[string]*Node{"a": {Spec: NodeSpec{Name: "a"}}},
		Parents:  make(DependencyMap),
		Children: make(DependencyMap),
		exec: executionState{
			nodeRelay:     make(map[string]chan NodeStatus),
			statusChannel: make(chan NodeStatusMsg, 4),
		},
	}

	g.skipTaskAndNotifyChildren("a")

	select {
	case msg := <-g.exec.statusChannel:
		if msg.NodeKey != "a" || msg.Status != Skipped {
			t.Errorf("unexpected message: %+v", msg)
		}
	default:
		t.Error("expected Skipped message on statusChannel")
	}
}

func TestSkipTaskAndNotifyChildren_PropagatesSkipToChildren(t *testing.T) {
	childCh := make(chan NodeStatus, 1)
	g := &Graph{
		Nodes: map[string]*Node{
			"a": {Spec: NodeSpec{Name: "a"}},
			"b": {Spec: NodeSpec{Name: "b"}},
		},
		Parents:  DependencyMap{"b": {"a": struct{}{}}},
		Children: DependencyMap{"a": {"b": struct{}{}}},
		exec: executionState{
			nodeRelay:     map[string]chan NodeStatus{edgeKey("a", "b"): childCh},
			statusChannel: make(chan NodeStatusMsg, 4),
		},
	}

	g.skipTaskAndNotifyChildren("a")

	// Status channel
	select {
	case msg := <-g.exec.statusChannel:
		if msg.Status != Skipped {
			t.Errorf("expected Skipped on statusChannel, got %s", msg.Status)
		}
	default:
		t.Error("expected message on statusChannel")
	}

	// Relay channel for child
	select {
	case status := <-childCh:
		if status != Skipped {
			t.Errorf("expected Skipped on relay channel, got %s", status)
		}
	default:
		t.Error("expected Skipped on child relay channel")
	}
}
