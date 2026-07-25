package lib

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// NodeStatusMsg is sent over the graph's statusChannel to report a change in
// a node's lifecycle state. It is consumed by DagModel.Update to update the
// live TUI table.
type NodeStatusMsg struct {
	NodeKey string
	Status  NodeStatus
	Pid     int
	Attempt string
}

// DagStartMsg is sent once at the beginning of execution to set the TUI header.
type DagStartMsg struct {
	Message string
}

// DagCompleteMsg is sent when all nodes have finished. Its message is rendered
// below the status table and it triggers tea.Quit to exit the TUI.
type DagCompleteMsg struct {
	Message string
}

// DagModel is the Bubble Tea model for the execution TUI. It tracks the
// current status, PID, attempt string, and start/end timestamps for every
// node, and advances a braille spinner on each tick for animated states.
type DagModel struct {
	Nodes          map[string]NodeStatus
	NodeOrder      []string
	NodeStartTimes map[string]time.Time
	NodeEndTimes   map[string]time.Time
	NodePids       map[string]int
	NodeAttempts   map[string]string
	NodeNameWidth  int // width of the node name column, derived from the longest name
	StartMsg       string
	CompleteMsg    string
	SpinnerFrame   int
}

// tickMsg is sent on a 50ms timer to drive the spinner animation.
type tickMsg struct{}

// spinnerFrames is the braille-dot animation sequence used for Running and
// UpForRetry states.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// rowStyles maps each NodeStatus to a lipgloss foreground style used to
// colour the entire row in View.
var rowStyles = map[NodeStatus]lipgloss.Style{
	Pending:    lipgloss.NewStyle().Foreground(lipgloss.Color("245")), // grey
	Running:    lipgloss.NewStyle().Foreground(lipgloss.Color("15")),  // white
	Success:    lipgloss.NewStyle().Foreground(lipgloss.Color("34")),  // green
	Failed:     lipgloss.NewStyle().Foreground(lipgloss.Color("160")), // red
	Skipped:    lipgloss.NewStyle().Foreground(lipgloss.Color("245")), // grey
	UpForRetry: lipgloss.NewStyle().Foreground(lipgloss.Color("220")), // yellow
}

// NewDagModel constructs a DagModel from the given Graph. Nodes are
// initialised to Pending and displayed in their YAML declaration order.
func NewDagModel(G *Graph) *DagModel {
	Nodes := make(map[string]NodeStatus)
	for k := range G.Nodes {
		Nodes[k] = Pending
	}

	// Compute the minimum column width needed to fit every node name.
	nodeNameWidth := len("Node") // never narrower than the header
	for _, k := range G.NodeOrder {
		if len(k) > nodeNameWidth {
			nodeNameWidth = len(k)
		}
	}

	return &DagModel{
		Nodes:          Nodes,
		NodeOrder:      G.NodeOrder,
		NodeNameWidth:  nodeNameWidth,
		NodeStartTimes: make(map[string]time.Time),
		NodeEndTimes:   make(map[string]time.Time),
		NodePids:       make(map[string]int),
		NodeAttempts:   make(map[string]string),
	}
}

// Init returns the first tick command, starting the 50ms spinner loop.
func (m *DagModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg { return tickMsg{} })
}

// Update handles incoming messages and mutates the model accordingly:
//   - DagStartMsg: records the header string
//   - tickMsg: advances the spinner frame and schedules the next tick
//   - NodeStatusMsg: updates node status, PID, attempt, and timestamps
//   - DagCompleteMsg: records the footer string and quits the program
//   - tea.KeyMsg ctrl+c: exits immediately
func (m *DagModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DagStartMsg:
		m.StartMsg = msg.Message
	case tickMsg:
		m.SpinnerFrame = (m.SpinnerFrame + 1) % len(spinnerFrames)
		return m, tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg { return tickMsg{} })
	case NodeStatusMsg:
		if msg.Status == Running {
			m.NodeStartTimes[msg.NodeKey] = time.Now()
		}
		if msg.Pid > 0 {
			m.NodePids[msg.NodeKey] = msg.Pid
		}
		if msg.Status == Success || msg.Status == Skipped || msg.Status == Failed {
			if _, exists := m.NodeEndTimes[msg.NodeKey]; !exists {
				m.NodeEndTimes[msg.NodeKey] = time.Now()
			}
		}
		m.Nodes[msg.NodeKey] = msg.Status
		m.NodeAttempts[msg.NodeKey] = msg.Attempt
	case DagCompleteMsg:
		m.CompleteMsg = msg.Message
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the current model state as a fixed-width table with one row per
// node, showing status (with spinner for animated states), PID, attempt
// counter, and start/end timestamps. Column widths are padded using
// go-runewidth to handle multi-byte rune characters correctly.
func (m *DagModel) View() string {
	var b strings.Builder
	if m.StartMsg != "" {
		fmt.Fprintf(&b, "\n%s\n", m.StartMsg)
	}

	// Total separator width: node col + 5 fixed cols + 5 single spaces between them.
	separatorWidth := m.NodeNameWidth + 1 + 12 + 1 + 10 + 1 + 10 + 1 + 15 + 1 + 15

	fmt.Fprintf(
		&b,
		"%s %-12s %-10s %-10s %-15s %-15s\n",
		runewidth.FillRight("Node", m.NodeNameWidth),
		"Status", "Pid", "Attempt", "Started", "Ended",
	)
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", separatorWidth))
	for _, k := range m.NodeOrder {
		v := m.Nodes[k]
		status := ""
		switch v {
		case Pending:
			status = "[ ] Pending "
		case Running:
			status = fmt.Sprintf(" %s  Running ", spinnerFrames[m.SpinnerFrame])
		case Success:
			status = "[✓] Success "
		case Failed:
			status = "[X] Failed "
		case Skipped:
			status = "[-] Skipped "
		case UpForRetry:
			status = fmt.Sprintf(" %s  Retry   ", spinnerFrames[m.SpinnerFrame])
		}
		status = runewidth.FillRight(status, 12)

		attempt := "-"
		if a, ok := m.NodeAttempts[k]; ok {
			a = runewidth.FillRight(a, 10)
			attempt = a
		}

		pid := "-"
		if p, ok := m.NodePids[k]; ok && p > 0 && v != Pending {
			pid = fmt.Sprintf("%d", p)
			pid = runewidth.FillRight(pid, 10)
		} else {
			pid = runewidth.FillRight("-", 10)
		}

		startTimestamp := ""
		if t, ok := m.NodeStartTimes[k]; ok {
			startTimestamp = t.Format("15:04:05.0000")
			startTimestamp = runewidth.FillRight(startTimestamp, 15)
		} else {
			startTimestamp = runewidth.FillRight("-", 15)
		}

		endTimestamp := ""
		if t, ok := m.NodeEndTimes[k]; ok {
			endTimestamp = t.Format("15:04:05.0000")
			endTimestamp = runewidth.FillRight(endTimestamp, 15)
		} else {
			endTimestamp = runewidth.FillRight("-", 15)
		}

		fmt.Fprintf(
			&b,
			"%s\n",
			rowStyles[v].Render(fmt.Sprintf(
				"%s %-12s %-10s %-10s %-15s %-15s",
				runewidth.FillRight(k, m.NodeNameWidth),
				status, pid, attempt, startTimestamp, endTimestamp,
			)),
		)
	}

	if m.CompleteMsg != "" {
		fmt.Fprintf(&b, "\n%s\n", m.CompleteMsg)
	}

	return b.String()
}
