package lib

import (
	"fmt"
	"log"
	"math"
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
	Retries    int        `yaml:"retries"`
	RetryDelay int        `yaml:"retryDelay"`
	Status     NodeStatus
	Pid        int
}

// Define structure to match YAML format
type NodeYaml struct {
	Nodes []Node `yaml:"nodes"`
}

// execute a Node command
func (t *Node) execute(startTime time.Time) {
	var logFile *os.File
	var err error

	for attempt := 1; attempt <= int(math.Max(1, float64(t.Retries+1))); attempt++ {
		if attempt > 1 {
			t.Status = Pending
			G.StatusChannel <- NodeStatusMsg{NodeKey: t.Name, Status: t.Status, Attempt: fmt.Sprintf("%d/%d", attempt-1, t.Retries+1)}
			if t.RetryDelay > 0 {
				time.Sleep(time.Duration(t.RetryDelay) * time.Second)
			}
		}

		logFile, err = t.createLogFile(startTime, attempt)
		if err != nil {
			t.fail()
			return
		}

		cmd := exec.Command("bash", "-c", t.Command)
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		t.Status = Running
		if err := cmd.Start(); err != nil {
			log.Printf("Error starting command for node %s: %v\n", t.Name, err)
			logFile.Close()
			continue
		}

		t.Pid = cmd.Process.Pid
		G.StatusChannel <- NodeStatusMsg{NodeKey: t.Name, Status: t.Status, Pid: t.Pid, Attempt: fmt.Sprintf("%d/%d", attempt, t.Retries+1)}

		err = cmd.Wait()
		logFile.Close()

		if err == nil {
			t.succeed(attempt)
			return
		}
	}

	// If we reached here, all retries failed
	t.fail()
}

func (t *Node) createLogFile(startTime time.Time, attempt int) (*os.File, error) {
	logDir := fmt.Sprintf(".orca/%s", startTime.Format("2006-01-02_15-04-05"))
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(fmt.Sprintf("%s/%s_%d.log", logDir, t.Name, attempt))
}

func (t *Node) fail() {
	t.Status = Failed
	// Send final status update
	G.StatusChannel <- NodeStatusMsg{
		NodeKey: t.Name,
		Status:  Failed,
		Pid:     t.Pid,
		Attempt: fmt.Sprintf("%d/%d", t.Retries+1, t.Retries+1),
	}
	t.notifyChildren()
}

func (t *Node) succeed(attempt int) {
	t.Status = Success
	// Send final status update
	G.StatusChannel <- NodeStatusMsg{
		NodeKey: t.Name,
		Status:  Success,
		Pid:     t.Pid,
		Attempt: fmt.Sprintf("%d/%d", attempt, t.Retries+1),
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
