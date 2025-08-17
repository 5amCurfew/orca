package lib

import (
	"fmt"
	"os"
	"os/exec"
	"time"
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
	Name       string     `yaml:"name"`
	Desc       string     `yaml:"desc"`
	Command    string     `yaml:"cmd"`
	ParentRule ParentRule `yaml:"parentRule"`
	Status     TaskStatus
	Pid        int
}

// execute a Node Task command
func (t *Task) execute(startTime time.Time) {
	t.Status = Running
	G.StatusChannel <- TaskStatusMsg{TaskKey: t.Name, Status: Running}

	logFile, err := t.createLogFile(startTime)
	if err != nil {
		t.fail()
		return
	}
	defer logFile.Close()

	cmd := exec.Command("bash", "-c", t.Command)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Start the process first
	if err := cmd.Start(); err != nil {
		t.fail()
		return
	}

	// Now we can safely get the PID
	t.Pid = cmd.Process.Pid
	G.StatusChannel <- TaskStatusMsg{TaskKey: t.Name, Status: Running, Pid: t.Pid}

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		t.fail()
	} else {
		t.succeed()
	}
}

func (t *Task) createLogFile(startTime time.Time) (*os.File, error) {
	logDir := fmt.Sprintf(".orca/%s", startTime.Format("2006-01-02_15-04-05"))
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(fmt.Sprintf("%s/%s.log", logDir, t.Name))
}

func (t *Task) fail() {
	t.Status = Failed
	// Send final status update
	G.StatusChannel <- TaskStatusMsg{
		TaskKey: t.Name,
		Status:  Failed,
		Pid:     t.Pid,
	}
	t.notifyChildren()
}

func (t *Task) succeed() {
	t.Status = Success
	// Send final status update
	G.StatusChannel <- TaskStatusMsg{
		TaskKey: t.Name,
		Status:  Success,
		Pid:     t.Pid,
	}
	t.notifyChildren()
}

func (t *Task) notifyChildren() {
	for child := range G.Children[t.Name] {
		notifyWG.Add(1)
		go func(child string) {
			defer notifyWG.Done()
			key := edgeKey(t.Name, child)
			taskRelay[key] <- t.Status
			close(taskRelay[key])
		}(child)
	}
}
