package lib

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

// ParentRule controls how a node behaves when one or more of its parents
// did not succeed.
type ParentRule string

// NodeStatus represents the current lifecycle state of a node.
type NodeStatus string

const (
	Pending     NodeStatus = "pending"    // waiting for parents to complete
	Running     NodeStatus = "running"    // bash command is executing
	Success     NodeStatus = "success"    // command exited with code 0
	Skipped     NodeStatus = "skipped"    // not run because a parent failed/was skipped
	Failed      NodeStatus = "failed"     // command exited with a non-zero code
	UpForRetry  NodeStatus = "upforretry" // failed attempt, waiting before next retry
	AllComplete ParentRule = "complete"   // run regardless of parent outcome
	AllSuccess  ParentRule = "success"    // run only if all parents succeeded (default)
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

// execute runs the node's bash command, retrying up to t.Retries additional
// times on failure. stdout and stderr are written to a per-attempt log file
// under .orca/<run-timestamp>/. Progress is reported via g.StatusChannel.
func (t *Node) execute(g *Graph, startTime time.Time) {
	var logFile *os.File
	var err error

	for attempt := 1; attempt <= int(math.Max(1, float64(t.Retries+1))); attempt++ {
		if attempt > 1 {
			t.Status = UpForRetry
			g.statusChannel <- NodeStatusMsg{NodeKey: t.Name, Status: t.Status, Attempt: fmt.Sprintf("%d/%d", attempt-1, t.Retries+1)}
			if t.RetryDelay > 0 {
				time.Sleep(time.Duration(t.RetryDelay) * time.Second)
			}
		}

		logFile, err = t.createLogFile(startTime, attempt)
		if err != nil {
			t.fail(g)
			return
		}

		cmd := exec.Command("bash", "-c", t.Command)
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		t.Status = Running
		if err := cmd.Start(); err != nil {
			log.Errorf("Error starting command for node %s: %v", t.Name, err)
			logFile.Close()
			continue
		}

		t.Pid = cmd.Process.Pid
		g.statusChannel <- NodeStatusMsg{NodeKey: t.Name, Status: t.Status, Pid: t.Pid, Attempt: fmt.Sprintf("%d/%d", attempt, t.Retries+1)}

		err = cmd.Wait()
		logFile.Close()

		if err == nil {
			t.succeed(g, attempt)
			return
		}
	}

	t.fail(g)
}

// createLogFile opens (or creates) the log file for the given attempt under
// .orca/<run-timestamp>/<node-name>_<attempt>.log.
func (t *Node) createLogFile(startTime time.Time, attempt int) (*os.File, error) {
	logDir := fmt.Sprintf(".orca/%s", startTime.Format("2006-01-02_15-04-05"))
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(fmt.Sprintf("%s/%s_%d.log", logDir, t.Name, attempt))
}

// fail marks the node as Failed, records the failure on the graph, sends a
// final status update to the TUI, and notifies downstream nodes.
func (t *Node) fail(g *Graph) {
	g.anyFailed.Store(true)
	t.Status = Failed
	g.statusChannel <- NodeStatusMsg{
		NodeKey: t.Name,
		Status:  Failed,
		Pid:     t.Pid,
		Attempt: fmt.Sprintf("%d/%d", t.Retries+1, t.Retries+1),
	}
	t.notifyChildren(g)
}

// succeed marks the node as Success, sends a final status update to the TUI,
// and notifies downstream nodes.
func (t *Node) succeed(g *Graph, attempt int) {
	t.Status = Success
	g.statusChannel <- NodeStatusMsg{
		NodeKey: t.Name,
		Status:  Success,
		Pid:     t.Pid,
		Attempt: fmt.Sprintf("%d/%d", attempt, t.Retries+1),
	}
	t.notifyChildren(g)
}

// notifyChildren sends the node's final status to the relay channel of each
// direct child, then closes that channel. Each send runs in its own goroutine
// tracked by g.notifyWG.
func (t *Node) notifyChildren(g *Graph) {
	for child := range g.Children[t.Name] {
		g.notifyWG.Add(1)
		go func(child string) {
			defer g.notifyWG.Done()
			key := edgeKey(t.Name, child)
			g.nodeRelay[key] <- t.Status
			close(g.nodeRelay[key])
		}(child)
	}
}
