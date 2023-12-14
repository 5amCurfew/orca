package lib

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type Task struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Status  string `json:"status"`
}

// ExecuteTask simulates the execution of a task
func executeTask(task *Task) {
	cmdParts := []string{"bash", "-c", task.Command}
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	task.Status = "running"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		task.Status = "failed"
	} else {
		task.Status = "success"
	}
}

// ExecuteTaskList executes the entire list of tasks
func ExecuteTasks(sortedTasks [][]string, tasks map[string]*Task) {
	// Iterate down sortedTasks
	for layerIndex, taskSet := range sortedTasks {
		fmt.Printf("Task layer %d is running\n", layerIndex)
		var taskWaitGroup sync.WaitGroup
		for _, taskName := range taskSet {
			taskWaitGroup.Add(1)
			go func(taskName string) {
				defer taskWaitGroup.Done()
				task := tasks[taskName]
				executeTask(task)
			}(taskName)
		}
		taskWaitGroup.Wait()
		fmt.Printf("Task layer %d is completed successfully\n", layerIndex)
	}
}
