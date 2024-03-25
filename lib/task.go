package lib

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type TaskStatus string

const (
	Pending TaskStatus = "pending"
	Running TaskStatus = "running"
	Success TaskStatus = "success"
	Failed  TaskStatus = "failed"
)

// Task represents a task in the DAG
type Task struct {
	Name    string     `json:"name,omitempty"`
	Desc    string     `json:"desc,omitempty"`
	Command string     `json:"cmd,omitempty"`
	Status  TaskStatus `json:"status,omitempty"`
}

// ExecuteTask executes a Task's command
func (t *Task) execute(dagExecutionStartTime time.Time, completionRelay map[string]chan bool, g *Graph) {
	log.Infof("%s task execution started", t.Name)

	cmdParts := []string{"bash", "-c", t.Command}
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	t.Status = Running

	// Create log directory if it doesn't exist
	logDir := fmt.Sprintf(".orca/%s/%s", g.Name, dagExecutionStartTime.Format("2006-01-02_15-04-05"))
	os.MkdirAll(logDir, os.ModePerm)

	// Create log file
	logFile, err := os.Create(fmt.Sprintf("%s/%s.log", logDir, t.Name))
	if err != nil {
		log.Warnf("Error creating log output file: %s", err)
		t.Status = Failed
		for child := range g.Children[t.Name] {
			completionRelay[fmt.Sprint(t.Name, "->", child)] <- false
		}
		return
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		log.Warnf("task %s execution failed", t.Name)
		t.Status = Failed
		for child := range g.Children[t.Name] {
			completionRelay[fmt.Sprint(t.Name, "->", child)] <- false
		}
	} else {
		log.Infof("%s task execution completed sucessfully", t.Name)
		t.Status = Success
		for child := range g.Children[t.Name] {
			completionRelay[fmt.Sprint(t.Name, "->", child)] <- true
		}
	}
}
