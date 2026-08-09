package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

// newSingleNodeGraph builds a graph whose exec state is fully initialised for
// running exactly one node with no parents or children.
func newSingleNodeGraph(node *Node, statusBuf int) *Graph {
	return &Graph{
		Nodes:    map[string]*Node{node.Spec.Name: node},
		Parents:  make(DependencyMap),
		Children: make(DependencyMap),
		exec: executionState{
			nodeRelay:     make(map[string]chan NodeStatus),
			statusChannel: make(chan NodeStatusMsg, statusBuf),
		},
	}
}

// collectStatuses closes the statusChannel and returns every NodeStatus sent.
func collectStatuses(g *Graph) []NodeStatus {
	close(g.exec.statusChannel)
	var out []NodeStatus
	for msg := range g.exec.statusChannel {
		out = append(out, msg.Status)
	}
	return out
}

// ──────────────────────────────────────────────────────────────────────────────
// execute — success path
// ──────────────────────────────────────────────────────────────────────────────

func TestNodeExecute_SuccessfulCommand(t *testing.T) {
	t.Chdir(t.TempDir())

	node := &Node{
		Spec:  NodeSpec{Name: "ok", Command: `echo "hello"`},
		State: NodeState{Status: Pending},
	}
	g := newSingleNodeGraph(node, 10)

	node.execute(g, time.Now())
	g.exec.notifyWG.Wait()

	if node.State.Status != Success {
		t.Errorf("expected Success, got %s", node.State.Status)
	}
	if g.exec.anyFailed.Load() {
		t.Error("anyFailed should be false after success")
	}

	statuses := collectStatuses(g)
	if len(statuses) < 2 {
		t.Fatalf("expected at least 2 status messages, got %d", len(statuses))
	}
	last := statuses[len(statuses)-1]
	if last != Success {
		t.Errorf("expected last status Success, got %s", last)
	}
}

func TestNodeExecute_RunningStatusEmitted(t *testing.T) {
	t.Chdir(t.TempDir())

	node := &Node{
		Spec:  NodeSpec{Name: "runner", Command: `true`},
		State: NodeState{Status: Pending},
	}
	g := newSingleNodeGraph(node, 10)

	node.execute(g, time.Now())
	g.exec.notifyWG.Wait()

	statuses := collectStatuses(g)
	var hasRunning bool
	for _, s := range statuses {
		if s == Running {
			hasRunning = true
		}
	}
	if !hasRunning {
		t.Error("expected at least one Running status message")
	}
}

func TestNodeExecute_PidIsPositive(t *testing.T) {
	t.Chdir(t.TempDir())

	node := &Node{
		Spec:  NodeSpec{Name: "pid-test", Command: `sleep 0`},
		State: NodeState{Status: Pending},
	}
	g := newSingleNodeGraph(node, 10)

	node.execute(g, time.Now())
	g.exec.notifyWG.Wait()

	if node.State.Pid <= 0 {
		t.Errorf("expected positive PID after execution, got %d", node.State.Pid)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// execute — failure path
// ──────────────────────────────────────────────────────────────────────────────

func TestNodeExecute_FailedCommand(t *testing.T) {
	t.Chdir(t.TempDir())

	node := &Node{
		Spec:  NodeSpec{Name: "bad", Command: `exit 1`},
		State: NodeState{Status: Pending},
	}
	g := newSingleNodeGraph(node, 10)

	node.execute(g, time.Now())
	g.exec.notifyWG.Wait()

	if node.State.Status != Failed {
		t.Errorf("expected Failed, got %s", node.State.Status)
	}
	if !g.exec.anyFailed.Load() {
		t.Error("anyFailed should be true after failure")
	}

	statuses := collectStatuses(g)
	last := statuses[len(statuses)-1]
	if last != Failed {
		t.Errorf("expected last status Failed, got %s", last)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// execute — retry paths
// ──────────────────────────────────────────────────────────────────────────────

func TestNodeExecute_RetriesExhausted(t *testing.T) {
	t.Chdir(t.TempDir())

	node := &Node{
		Spec: NodeSpec{
			Name:       "always-fail",
			Command:    `exit 1`,
			Retries:    2,
			RetryDelay: 0,
		},
		State: NodeState{Status: Pending},
	}
	g := newSingleNodeGraph(node, 20)

	node.execute(g, time.Now())
	g.exec.notifyWG.Wait()

	if node.State.Status != Failed {
		t.Errorf("expected Failed after exhausted retries, got %s", node.State.Status)
	}

	statuses := collectStatuses(g)
	var upForRetryCount int
	for _, s := range statuses {
		if s == UpForRetry {
			upForRetryCount++
		}
	}
	if upForRetryCount != 2 {
		t.Errorf("expected 2 UpForRetry messages (Retries=2), got %d", upForRetryCount)
	}
}

func TestNodeExecute_RetryThenSucceed(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// First invocation creates the marker file and exits 1.
	// Second invocation finds the file and exits 0.
	counterFile := filepath.Join(tmpDir, "counter")
	cmd := fmt.Sprintf(`
		if [ ! -f %q ]; then
			echo 1 > %q
			exit 1
		fi
		echo "success on retry"
	`, counterFile, counterFile)

	node := &Node{
		Spec: NodeSpec{
			Name:       "retry-ok",
			Command:    cmd,
			Retries:    1,
			RetryDelay: 0,
		},
		State: NodeState{Status: Pending},
	}
	g := newSingleNodeGraph(node, 20)

	node.execute(g, time.Now())
	g.exec.notifyWG.Wait()

	if node.State.Status != Success {
		t.Errorf("expected Success after retry, got %s", node.State.Status)
	}
	if g.exec.anyFailed.Load() {
		t.Error("anyFailed should be false after eventual success")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// execute — log files
// ──────────────────────────────────────────────────────────────────────────────

func TestNodeExecute_LogFileCreatedForSuccessfulRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	node := &Node{
		Spec:  NodeSpec{Name: "log-node", Command: `echo "log me"`},
		State: NodeState{Status: Pending},
	}
	g := newSingleNodeGraph(node, 10)

	startTime := time.Now()
	node.execute(g, startTime)
	g.exec.notifyWG.Wait()

	logDir := filepath.Join(tmpDir, ".orca", startTime.Format("2006-01-02_15-04-05.000000000"))
	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("log directory not created at %s: %v", logDir, err)
	}
	var found bool
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "log-node_") && strings.HasSuffix(e.Name(), ".log") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("no log file for log-node found in %s", logDir)
	}
}

func TestNodeExecute_LogFileCreatedPerAttempt(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	node := &Node{
		Spec: NodeSpec{
			Name:       "multi-attempt",
			Command:    `exit 1`,
			Retries:    1,
			RetryDelay: 0,
		},
		State: NodeState{Status: Pending},
	}
	g := newSingleNodeGraph(node, 20)

	startTime := time.Now()
	node.execute(g, startTime)
	g.exec.notifyWG.Wait()

	logDir := filepath.Join(tmpDir, ".orca", startTime.Format("2006-01-02_15-04-05.000000000"))
	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("log directory not found: %v", err)
	}
	// Retries=1 means 2 attempts → 2 log files
	if len(entries) < 2 {
		t.Errorf("expected at least 2 log files for 2 attempts, got %d", len(entries))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// createLogFile
// ──────────────────────────────────────────────────────────────────────────────

func TestCreateLogFile_CreatesFile(t *testing.T) {
	t.Chdir(t.TempDir())

	node := &Node{Spec: NodeSpec{Name: "my-node"}}
	f, err := node.createLogFile(time.Now(), 1)
	if err != nil {
		t.Fatalf("createLogFile error: %v", err)
	}
	defer f.Close()

	if _, err := os.Stat(f.Name()); os.IsNotExist(err) {
		t.Errorf("log file %q was not created on disk", f.Name())
	}
}

func TestCreateLogFile_NameContainsNodeAndAttempt(t *testing.T) {
	t.Chdir(t.TempDir())

	node := &Node{Spec: NodeSpec{Name: "special-node"}}
	f, err := node.createLogFile(time.Now(), 3)
	if err != nil {
		t.Fatalf("createLogFile error: %v", err)
	}
	defer f.Close()

	base := filepath.Base(f.Name())
	if !strings.HasPrefix(base, "special-node_3") {
		t.Errorf("log filename %q should start with 'special-node_3'", base)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// notifyChildren
// ──────────────────────────────────────────────────────────────────────────────

func TestNotifyChildren_NoChildren_NoBlock(t *testing.T) {
	node := &Node{
		Spec:  NodeSpec{Name: "leaf"},
		State: NodeState{Status: Success},
	}
	g := &Graph{
		Nodes:    map[string]*Node{"leaf": node},
		Children: make(DependencyMap),
		exec:     executionState{nodeRelay: make(map[string]chan NodeStatus)},
	}
	node.notifyChildren(g)
	g.exec.notifyWG.Wait() // must not block
}

func TestNotifyChildren_SendsStatusToChild(t *testing.T) {
	childCh := make(chan NodeStatus, 1)
	node := &Node{
		Spec:  NodeSpec{Name: "parent"},
		State: NodeState{Status: Success},
	}
	g := &Graph{
		Nodes:    map[string]*Node{"parent": node, "child": {Spec: NodeSpec{Name: "child"}}},
		Children: DependencyMap{"parent": {"child": struct{}{}}},
		exec: executionState{
			nodeRelay: map[string]chan NodeStatus{edgeKey("parent", "child"): childCh},
		},
	}

	node.notifyChildren(g)
	g.exec.notifyWG.Wait()

	select {
	case s := <-childCh:
		if s != Success {
			t.Errorf("expected Success on relay, got %s", s)
		}
	default:
		t.Error("relay channel should have received a status")
	}
}

func TestNotifyChildren_ChannelClosedAfterSend(t *testing.T) {
	childCh := make(chan NodeStatus, 1)
	node := &Node{
		Spec:  NodeSpec{Name: "parent"},
		State: NodeState{Status: Failed},
	}
	g := &Graph{
		Nodes:    map[string]*Node{"parent": node, "child": {Spec: NodeSpec{Name: "child"}}},
		Children: DependencyMap{"parent": {"child": struct{}{}}},
		exec: executionState{
			nodeRelay: map[string]chan NodeStatus{edgeKey("parent", "child"): childCh},
		},
	}

	node.notifyChildren(g)
	g.exec.notifyWG.Wait()

	// Drain the value
	<-childCh
	// Channel should now be closed — a second receive should return zero value + ok=false
	_, ok := <-childCh
	if ok {
		t.Error("relay channel should be closed after notifyChildren")
	}
}
