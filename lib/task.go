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
func executeTask(task *Task, graphName string, completionChMap map[string]chan bool) {
	cmdParts := []string{"bash", "-c", task.Command}
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	task.Status = Running

	// Create log
	logFile, err := os.Create(fmt.Sprintf("logs/%s/%s_%s.log", graphName, task.Name, time.Now().Format("2006_01_02_15_04_05")))
	if err != nil {
		fmt.Println("Error creating output file:", err)
		completionChMap[task.Name] <- false
		task.Status = Failed
		return
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		completionChMap[task.Name] <- false
		task.Status = Failed
	} else {
		completionChMap[task.Name] <- true
		task.Status = Success
	}
}
