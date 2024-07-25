package lib

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type ParentRule string
type TaskStatus string

const (
	Pending     TaskStatus = "pending"
	Running     TaskStatus = "running"
	Success     TaskStatus = "success"
	Skipped     TaskStatus = "skipped"
	Failed      TaskStatus = "failed"
	AllComplete ParentRule = "complete"
	AllSuccess  ParentRule = "success"
)

// Task represents a task in the DAG
type Task struct {
	Name       string     `json:"name,omitempty"`
	Desc       string     `json:"desc,omitempty"`
	Command    string     `json:"cmd,omitempty"`
	ParentRule ParentRule `json:"ParentRule,omitempty"`
	Status     TaskStatus `json:"status,omitempty"`
}

// ExecuteTask executes a Task's command
func (t *Task) execute(dagExecutionStartTime time.Time, g *Graph) {
	log.Infof("[START] %s task execution started", t.Name)

	cmdParts := []string{"bash", "-c", t.Command}
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	t.Status = Running

	// Create log directory if it doesn't exist
	logDir := fmt.Sprintf(".orca/%s/%s", g.Name, dagExecutionStartTime.Format("2006-01-02_15-04-05"))
	os.MkdirAll(logDir, os.ModePerm)

	// Create log file
	logFile, err := os.Create(fmt.Sprintf("%s/%s.log", logDir, t.Name))
	if err != nil {
		log.Errorf("error creating log output file: %s", err)
		t.Status = Failed
		for child := range g.Children[t.Name] {
			completionRelay[fmt.Sprint(t.Name, "->", child)] <- Failed
			close(completionRelay[fmt.Sprint(t.Name, "->", child)])
		}
		return
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		log.Errorf("[X FAILED] task %s execution failed", t.Name)
		t.Status = Failed
		for child := range g.Children[t.Name] {
			completionRelay[fmt.Sprint(t.Name, "->", child)] <- Failed
		}
	} else {
		log.Infof("[\u2714 SUCCESS] %s task execution successful", t.Name)
		t.Status = Success
		for child := range g.Children[t.Name] {
			completionRelay[fmt.Sprint(t.Name, "->", child)] <- Success
		}
	}
}
