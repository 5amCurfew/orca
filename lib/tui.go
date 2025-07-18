package lib

import (
	"fmt"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

type TaskStatusMsg struct {
	TaskKey string
	Status  TaskStatus
}

type DagStartMsg struct {
	Message string
}

type DagCompleteMsg struct {
	Message string
}

type DagModel struct {
	Tasks          map[string]TaskStatus
	TaskOrder      []string
	TaskStartTimes map[string]time.Time
	TaskEndTimes   map[string]time.Time
	StartMsg       string
	CompleteMsg    string
	SpinnerFrame   int
}

type tickMsg struct{}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func NewDagModel(G *Graph) *DagModel {
	tasks := make(map[string]TaskStatus)
	order := make([]string, 0, len(G.Tasks))
	startTimes := make(map[string]time.Time)
	endTimes := make(map[string]time.Time)
	for k := range G.Tasks {
		tasks[k] = Pending
		order = append(order, k)
	}

	slices.Sort(order)
	return &DagModel{
		Tasks:          tasks,
		TaskOrder:      order,
		TaskStartTimes: startTimes,
		TaskEndTimes:   endTimes,
	}
}

func (m *DagModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*120, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *DagModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DagStartMsg:
		m.StartMsg = msg.Message
	case tickMsg:
		m.SpinnerFrame = (m.SpinnerFrame + 1) % len(spinnerFrames)
		return m, tea.Tick(time.Millisecond*120, func(time.Time) tea.Msg { return tickMsg{} })
	case TaskStatusMsg:
		if msg.Status == Running {
			if _, exists := m.TaskStartTimes[msg.TaskKey]; !exists {
				m.TaskStartTimes[msg.TaskKey] = time.Now()
			}
		}
		if msg.Status == Success || msg.Status == Skipped || msg.Status == Failed {
			if _, exists := m.TaskEndTimes[msg.TaskKey]; !exists {
				m.TaskEndTimes[msg.TaskKey] = time.Now()
			}
		}
		m.Tasks[msg.TaskKey] = msg.Status
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

	fmt.Fprintf(&b, "%-20s %-12s %-15s %-15s\n", "Task", "Status", "Started", "Ended")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 65))
	for _, k := range m.TaskOrder {
		v := m.Tasks[k]
		status := ""
		switch v {
		case Pending:
			status = "⏳ Pending"
		case Running:
			status = fmt.Sprintf("%s Running", spinnerFrames[m.SpinnerFrame])
		case Success:
			status = "✅ Success"
		case Failed:
			status = "❌ Failed"
		case Skipped:
			status = "⚠️  Skipped"
		}
		status = runewidth.FillRight(status, 14)

		startTimestamp := ""
		if t, ok := m.TaskStartTimes[k]; ok {
			startTimestamp = t.Format("15:04:05.0000")
		} else {
			startTimestamp = runewidth.FillRight("-", 15)
		}
		endTimestamp := ""
		if t, ok := m.TaskEndTimes[k]; ok {
			endTimestamp = t.Format("15:04:05.0000")
		} else {
			endTimestamp = runewidth.FillRight("-", 15)
		}

		fmt.Fprintf(&b, "%-20s %-12s %-15s %-15s\n", k, status, startTimestamp, endTimestamp)
	}
	if m.CompleteMsg != "" {
		fmt.Fprintf(&b, "\n%s\n", m.CompleteMsg)
	}
	return b.String()
}

// Helper to center a string in a field of given width
func centerString(s string, width int) string {
	displayWidth := runewidth.StringWidth(s)
	padding := width - displayWidth
	if padding <= 0 {
		return s
	}
	left := padding / 2
	right := padding - left
	return fmt.Sprintf("%s%s%s", strings.Repeat(" ", left), s, strings.Repeat(" ", right))
}
