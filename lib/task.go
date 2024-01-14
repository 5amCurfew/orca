package lib

import (
	"fmt"
	"os"
	"os/exec"
	"time"
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
	Name     string        `json:"name,omitempty"`
	Desc     string        `json:"desc,omitempty"`
	Command  string        `json:"cmd,omitempty"`
	Children []interface{} `json:"children,omitempty"`
	Status   TaskStatus    `json:"status,omitempty"`
}

// ExecuteTask executes a Task's command
func (t *Task) execute(graphName string, dagExecutionStartTime time.Time, completionChMap map[string]chan bool) {
	cmdParts := []string{"bash", "-c", t.Command}
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	t.Status = Running

	// Create log directory if it doesn't exist
	logDir := fmt.Sprintf("logs/%s/%s", graphName, dagExecutionStartTime.Format("2006-01-02_15-04-05"))
	os.MkdirAll(logDir, os.ModePerm)

	// Create log file
	logFile, err := os.Create(fmt.Sprintf("%s/%s.log", logDir, t.Name))
	if err != nil {
		fmt.Println("Error creating log output file:", err)
		completionChMap[t.Name] <- false
		t.Status = Failed
		return
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		completionChMap[t.Name] <- false
		t.Status = Failed
	} else {
		completionChMap[t.Name] <- true
		t.Status = Success
	}
}
