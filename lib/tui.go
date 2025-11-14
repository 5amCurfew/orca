package lib

import (
	"fmt"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

type NodeStatusMsg struct {
	NodeKey string
	Status  NodeStatus
	Pid     int
	Attempt string
}

type DagStartMsg struct {
	Message string
}

type DagCompleteMsg struct {
	Message string
}

type DagModel struct {
	Nodes          map[string]NodeStatus
	NodeOrder      []string
	NodeStartTimes map[string]time.Time
	NodeEndTimes   map[string]time.Time
	NodePids       map[string]int
	NodeAttempts   map[string]string
	StartMsg       string
	CompleteMsg    string
	SpinnerFrame   int
}

type tickMsg struct{}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func NewDagModel(G *Graph) *DagModel {
	Nodes := make(map[string]NodeStatus)
	order := make([]string, 0, len(G.Nodes))
	for k := range G.Nodes {
		Nodes[k] = Pending
		order = append(order, k)
	}

	slices.Sort(order)
	return &DagModel{
		Nodes:          Nodes,
		NodeOrder:      order,
		NodeStartTimes: make(map[string]time.Time),
		NodeEndTimes:   make(map[string]time.Time),
		NodePids:       make(map[string]int),
		NodeAttempts:   make(map[string]string),
	}
}

func (m *DagModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(time.Time) tea.Msg { return tickMsg{} })
}

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

func (m *DagModel) View() string {
	var b strings.Builder
	if m.StartMsg != "" {
		fmt.Fprintf(&b, "\n%s\n", m.StartMsg)
	}

	fmt.Fprintf(
		&b,
		"%-20s %-12s %-10s %-10s %-15s %-15s\n",
		"Node", "Status", "Pid", "Attempt", "Started", "Ended",
	)

	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 85))
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
			"%-20s %-12s %-10s %-10s %-15s %-15s\n",
			k, status, pid, attempt, startTimestamp, endTimestamp,
		)
	}

	if m.CompleteMsg != "" {
		fmt.Fprintf(&b, "\n%s\n", m.CompleteMsg)
	}

	return b.String()
}
