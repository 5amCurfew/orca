package lib

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type ParentRule string
type NodeStatus string

const (
	Pending     NodeStatus = "pending"
	Running     NodeStatus = "running"
	Success     NodeStatus = "success"
	Skipped     NodeStatus = "skipped"
	Failed      NodeStatus = "failed"
	AllComplete ParentRule = "complete"
	AllSuccess  ParentRule = "success"
)

// Node represents a Node in the DAG
type Node struct {
	Name       string     `yaml:"name"`
	Desc       string     `yaml:"desc"`
	Command    string     `yaml:"cmd"`
	ParentRule ParentRule `yaml:"parentRule"`
	Status     NodeStatus
	Pid        int
}

// Define structure to match YAML format
type NodeYaml struct {
	Nodes []Node `yaml:"nodes"`
}

// execute a Node command
func (t *Node) execute(startTime time.Time) {
	t.Status = Running
	G.StatusChannel <- NodeStatusMsg{NodeKey: t.Name, Status: Running}

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
	G.StatusChannel <- NodeStatusMsg{NodeKey: t.Name, Status: Running, Pid: t.Pid}

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		t.fail()
	} else {
		t.succeed()
	}
}

func (t *Node) createLogFile(startTime time.Time) (*os.File, error) {
	logDir := fmt.Sprintf(".orca/%s", startTime.Format("2006-01-02_15-04-05"))
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(fmt.Sprintf("%s/%s.log", logDir, t.Name))
}

func (t *Node) fail() {
	t.Status = Failed
	// Send final status update
	G.StatusChannel <- NodeStatusMsg{
		NodeKey: t.Name,
		Status:  Failed,
		Pid:     t.Pid,
	}
	t.notifyChildren()
}

func (t *Node) succeed() {
	t.Status = Success
	// Send final status update
	G.StatusChannel <- NodeStatusMsg{
		NodeKey: t.Name,
		Status:  Success,
		Pid:     t.Pid,
	}
	t.notifyChildren()
}

func (t *Node) notifyChildren() {
	for child := range G.Children[t.Name] {
		notifyWG.Add(1)
		go func(child string) {
			defer notifyWG.Done()
			key := edgeKey(t.Name, child)
			NodeRelay[key] <- t.Status
			close(NodeRelay[key])
		}(child)
	}
}
