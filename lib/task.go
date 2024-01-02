package lib

import (
	"os"
	"os/exec"
)

type TaskStatus string

const (
	Pending TaskStatus = "pending"
	Running TaskStatus = "running"
	Success TaskStatus = "success"
	Failed  TaskStatus = "failed"
)

// Task represents a task in the DAG.
type Task struct {
	Name     string        `json:"name,omitempty"`
	Desc     string        `json:"desc,omitempty"`
	Command  string        `json:"cmd,omitempty"`
	Children []interface{} `json:"children,omitempty"`
	Status   TaskStatus    `json:"status,omitempty"`
}

// ExecuteTask simulates the execution of a task
func executeTask(task *Task, completionChMap map[string]chan bool) {
	cmdParts := []string{"bash", "-c", task.Command}
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	task.Status = Running
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		completionChMap[task.Name] <- false
		task.Status = Failed
	} else {
		completionChMap[task.Name] <- true
		task.Status = Success
	}
}
