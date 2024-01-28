package lib

import (
	"fmt"
	"log"
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
	Name    string     `json:"name,omitempty"`
	Desc    string     `json:"desc,omitempty"`
	Command string     `json:"cmd,omitempty"`
	Status  TaskStatus `json:"status,omitempty"`
}

// ExecuteTask executes a Task's command
func (t *Task) execute(dagExecutionStartTime time.Time, completionChMap map[string]chan bool, g *Graph) {
	cmdParts := []string{"bash", "-c", t.Command}
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	t.Status = Running

	// Create log directory if it doesn't exist
	logDir := fmt.Sprintf("logs/%s/%s", g.Name, dagExecutionStartTime.Format("2006-01-02_15-04-05"))
	os.MkdirAll(logDir, os.ModePerm)

	// Create log file
	logFile, err := os.Create(fmt.Sprintf("%s/%s.log", logDir, t.Name))
	if err != nil {
		log.Printf("Error creating log output file: %s", err)
		completionChMap[t.Name] <- false
		t.Status = Failed
		g.Fail()
		return
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Monitor the context for cancellation signals
	select {
	case <-g.Context.Done():
		// Cancellation has occurred
		log.Printf("%s execution cancelled at %s\n", g.Name, time.Now().Format("2006-01-02 15:04:05"))
		completionChMap[t.Name] <- false
		t.Status = Failed
		return
	default:
		// Continue with task execution
	}

	if err := cmd.Run(); err != nil {
		completionChMap[t.Name] <- false
		t.Status = Failed
		log.Printf("%s task execution failed at %s\n", t.Name, time.Now().Format("2006-01-02 15:04:05"))
		g.Fail()
	} else {
		completionChMap[t.Name] <- true
		t.Status = Success
	}
}
