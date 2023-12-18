package lib

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
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
	Name    string
	Command string
	Status  TaskStatus
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

// ExecuteDAG orchestrates and executes the tasks given the DAG.
func ExecuteDAG(g *Graph) {
	// Create Channel Map for task completion
	completionChannels := make(map[string]chan bool)

	// Initialise channel for each task
	for taskName := range g.Tasks {
		completionChannels[taskName] = make(chan bool, 1)
	}

	// Use a WaitGroup to wait for all tasks to complete
	var waitGroup sync.WaitGroup

	// Create goroutines for each task
	for taskName := range g.Tasks {
		// Increment the wait group for each task
		g.Tasks[taskName].Status = Pending
		waitGroup.Add(1)
		// Start goroutines for each task that are blocked until all parents have sent successful completion message
		go func(taskName string, completionChMap map[string]chan bool) {
			// Wait for all parents to complete
			for parent := range g.Parents[taskName] {
				<-completionChMap[parent]
			}

			// Execute the task
			fmt.Println("executing ", taskName)
			executeTask(g.Tasks[taskName], completionChMap)

			close(completionChMap[taskName])
			waitGroup.Done()
		}(taskName, completionChannels)
	}

	// Wait for all tasks to complete before exiting
	waitGroup.Wait()
}
