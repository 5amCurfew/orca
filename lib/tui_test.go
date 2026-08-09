package lib

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

// newTestModel returns a DagModel with two pending nodes ("a", "b") ready for
// Update / View testing.
func newTestModel() *DagModel {
	return &DagModel{
		Nodes:          map[string]NodeStatus{"a": Pending, "b": Pending},
		NodeOrder:      []string{"a", "b"},
		NodeStartTimes: make(map[string]time.Time),
		NodeEndTimes:   make(map[string]time.Time),
		NodePids:       make(map[string]int),
		NodeAttempts:   make(map[string]string),
		NodeNameWidth:  4,
	}
}

// updateModel is a convenience wrapper that type-asserts the returned tea.Model.
func updateModel(m *DagModel, msg tea.Msg) (*DagModel, tea.Cmd) {
	model, cmd := m.Update(msg)
	return model.(*DagModel), cmd
}

// ──────────────────────────────────────────────────────────────────────────────
// NewDagModel construction
// ──────────────────────────────────────────────────────────────────────────────

func TestNewDagModel_AllNodesPending(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"alpha": {Spec: NodeSpec{Name: "alpha"}},
			"beta":  {Spec: NodeSpec{Name: "beta"}},
		},
		NodeOrder: []string{"alpha", "beta"},
	}

	m := NewDagModel(g)

	if len(m.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(m.Nodes))
	}
	for _, name := range []string{"alpha", "beta"} {
		if m.Nodes[name] != Pending {
			t.Errorf("expected %s to start Pending, got %s", name, m.Nodes[name])
		}
	}
}

func TestNewDagModel_PreservesNodeOrder(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"z": {Spec: NodeSpec{Name: "z"}},
			"a": {Spec: NodeSpec{Name: "a"}},
		},
		NodeOrder: []string{"z", "a"},
	}

	m := NewDagModel(g)

	if len(m.NodeOrder) != 2 || m.NodeOrder[0] != "z" || m.NodeOrder[1] != "a" {
		t.Errorf("unexpected NodeOrder: %v", m.NodeOrder)
	}
}

func TestNewDagModel_NodeNameWidth_LongestName(t *testing.T) {
	g := &Graph{
		Nodes: map[string]*Node{
			"short":            {Spec: NodeSpec{Name: "short"}},
			"a-very-long-name": {Spec: NodeSpec{Name: "a-very-long-name"}},
		},
		NodeOrder: []string{"short", "a-very-long-name"},
	}

	m := NewDagModel(g)

	if m.NodeNameWidth != len("a-very-long-name") {
		t.Errorf("expected NodeNameWidth %d, got %d", len("a-very-long-name"), m.NodeNameWidth)
	}
}

func TestNewDagModel_NodeNameWidth_MinIsHeaderWidth(t *testing.T) {
	// "ab" is shorter than "Node" (4 chars); width must not go below 4.
	g := &Graph{
		Nodes:     map[string]*Node{"ab": {Spec: NodeSpec{Name: "ab"}}},
		NodeOrder: []string{"ab"},
	}

	m := NewDagModel(g)

	if m.NodeNameWidth < len("Node") {
		t.Errorf("NodeNameWidth should be at least %d, got %d", len("Node"), m.NodeNameWidth)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// DagModel.Init
// ──────────────────────────────────────────────────────────────────────────────

func TestDagModelInit_ReturnsNonNilCmd(t *testing.T) {
	m := newTestModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a non-nil tick command")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// DagModel.Update — message handling
// ──────────────────────────────────────────────────────────────────────────────

func TestUpdate_DagStartMsg_SetsStartMsg(t *testing.T) {
	m := newTestModel()
	updated, _ := updateModel(m, DagStartMsg{Message: "dag started"})
	if updated.StartMsg != "dag started" {
		t.Errorf("expected StartMsg %q, got %q", "dag started", updated.StartMsg)
	}
}

func TestUpdate_DagCompleteMsg_SetsCompleteMsgAndQuits(t *testing.T) {
	m := newTestModel()
	updated, cmd := updateModel(m, DagCompleteMsg{Message: "all done"})
	if updated.CompleteMsg != "all done" {
		t.Errorf("expected CompleteMsg %q, got %q", "all done", updated.CompleteMsg)
	}
	if cmd == nil {
		t.Error("DagCompleteMsg should return a non-nil quit command")
	}
}

func TestUpdate_NodeStatusMsg_Running_SetsStatusAndStartTime(t *testing.T) {
	m := newTestModel()
	before := time.Now()
	updated, _ := updateModel(m, NodeStatusMsg{NodeKey: "a", Status: Running, Pid: 42, Attempt: "1/1"})
	after := time.Now()

	if updated.Nodes["a"] != Running {
		t.Errorf("expected Running, got %s", updated.Nodes["a"])
	}
	if updated.NodePids["a"] != 42 {
		t.Errorf("expected pid 42, got %d", updated.NodePids["a"])
	}
	startTime, ok := updated.NodeStartTimes["a"]
	if !ok {
		t.Fatal("NodeStartTime should be set for Running status")
	}
	if startTime.Before(before) || startTime.After(after) {
		t.Error("NodeStartTime should be approximately time.Now()")
	}
}

func TestUpdate_NodeStatusMsg_Success_SetsEndTime(t *testing.T) {
	m := newTestModel()
	before := time.Now()
	updated, _ := updateModel(m, NodeStatusMsg{NodeKey: "a", Status: Success, Attempt: "1/1"})
	after := time.Now()

	if updated.Nodes["a"] != Success {
		t.Errorf("expected Success, got %s", updated.Nodes["a"])
	}
	endTime, ok := updated.NodeEndTimes["a"]
	if !ok {
		t.Fatal("NodeEndTime should be set for Success status")
	}
	if endTime.Before(before) || endTime.After(after) {
		t.Error("NodeEndTime should be approximately time.Now()")
	}
}

func TestUpdate_NodeStatusMsg_Failed_SetsEndTime(t *testing.T) {
	m := newTestModel()
	updated, _ := updateModel(m, NodeStatusMsg{NodeKey: "a", Status: Failed, Attempt: "1/1"})

	if updated.Nodes["a"] != Failed {
		t.Errorf("expected Failed, got %s", updated.Nodes["a"])
	}
	if _, ok := updated.NodeEndTimes["a"]; !ok {
		t.Error("NodeEndTime should be set for Failed status")
	}
}

func TestUpdate_NodeStatusMsg_Skipped_SetsEndTime(t *testing.T) {
	m := newTestModel()
	updated, _ := updateModel(m, NodeStatusMsg{NodeKey: "a", Status: Skipped})

	if updated.Nodes["a"] != Skipped {
		t.Errorf("expected Skipped, got %s", updated.Nodes["a"])
	}
	if _, ok := updated.NodeEndTimes["a"]; !ok {
		t.Error("NodeEndTime should be set for Skipped status")
	}
}

func TestUpdate_NodeStatusMsg_EndTimeSetOnlyOnce(t *testing.T) {
	// A terminal status message received a second time should not overwrite the
	// existing end timestamp.
	m := newTestModel()
	first := time.Now().Add(-time.Minute) // in the past
	m.NodeEndTimes["a"] = first

	updated, _ := updateModel(m, NodeStatusMsg{NodeKey: "a", Status: Success})

	if !updated.NodeEndTimes["a"].Equal(first) {
		t.Error("NodeEndTime should not be overwritten on second terminal status")
	}
}

func TestUpdate_NodeStatusMsg_AttemptRecorded(t *testing.T) {
	m := newTestModel()
	updated, _ := updateModel(m, NodeStatusMsg{NodeKey: "a", Status: Running, Attempt: "2/3"})

	if updated.NodeAttempts["a"] != "2/3" {
		t.Errorf("expected attempt %q, got %q", "2/3", updated.NodeAttempts["a"])
	}
}

func TestUpdate_TickMsg_AdvancesSpinnerFrame(t *testing.T) {
	m := newTestModel()
	m.SpinnerFrame = 0

	updated, cmd := updateModel(m, tickMsg{})

	if updated.SpinnerFrame != 1 {
		t.Errorf("expected SpinnerFrame 1, got %d", updated.SpinnerFrame)
	}
	if cmd == nil {
		t.Error("tickMsg should reschedule another tick")
	}
}

func TestUpdate_TickMsg_SpinnerWrapsAround(t *testing.T) {
	m := newTestModel()
	m.SpinnerFrame = len(spinnerFrames) - 1

	updated, _ := updateModel(m, tickMsg{})

	if updated.SpinnerFrame != 0 {
		t.Errorf("expected SpinnerFrame 0 after wrap, got %d", updated.SpinnerFrame)
	}
}

func TestUpdate_CtrlC_Quits(t *testing.T) {
	m := newTestModel()
	_, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should return a quit command")
	}
}

func TestUpdate_UnknownMsg_ReturnsNilCmd(t *testing.T) {
	type unknownMsg struct{}
	m := newTestModel()
	_, cmd := updateModel(m, unknownMsg{})
	if cmd != nil {
		t.Error("unknown message type should return nil cmd")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// DagModel.View
// ──────────────────────────────────────────────────────────────────────────────

func TestView_ContainsAllNodeNames(t *testing.T) {
	m := newTestModel()
	view := m.View()

	for _, name := range []string{"a", "b"} {
		if !strings.Contains(view, name) {
			t.Errorf("view should contain node name %q", name)
		}
	}
}

func TestView_ContainsColumnHeaders(t *testing.T) {
	m := newTestModel()
	view := m.View()

	for _, header := range []string{"Node", "Status", "Pid", "Attempt"} {
		if !strings.Contains(view, header) {
			t.Errorf("view should contain column header %q", header)
		}
	}
}

func TestView_ShowsStartMessage(t *testing.T) {
	m := newTestModel()
	m.StartMsg = "execution started"
	view := m.View()

	if !strings.Contains(view, "execution started") {
		t.Error("view should include the start message")
	}
}

func TestView_ShowsCompleteMessage(t *testing.T) {
	m := newTestModel()
	m.CompleteMsg = "execution complete"
	view := m.View()

	if !strings.Contains(view, "execution complete") {
		t.Error("view should include the complete message")
	}
}

func TestView_EmptyModel_DoesNotPanic(t *testing.T) {
	m := &DagModel{
		Nodes:          make(map[string]NodeStatus),
		NodeOrder:      []string{},
		NodeStartTimes: make(map[string]time.Time),
		NodeEndTimes:   make(map[string]time.Time),
		NodePids:       make(map[string]int),
		NodeAttempts:   make(map[string]string),
		NodeNameWidth:  4,
	}
	// Must not panic.
	_ = m.View()
}

func TestView_AllStatusStringsRendered(t *testing.T) {
	// Each status should produce a non-empty row in the view.
	allStatuses := []NodeStatus{Pending, Running, Success, Failed, Skipped, UpForRetry}
	for _, status := range allStatuses {
		t.Run(string(status), func(t *testing.T) {
			m := &DagModel{
				Nodes:          map[string]NodeStatus{"n": status},
				NodeOrder:      []string{"n"},
				NodeStartTimes: make(map[string]time.Time),
				NodeEndTimes:   make(map[string]time.Time),
				NodePids:       make(map[string]int),
				NodeAttempts:   make(map[string]string),
				NodeNameWidth:  4,
			}
			view := m.View()
			if !strings.Contains(view, "n") {
				t.Errorf("view for status %s does not contain node name", status)
			}
		})
	}
}
