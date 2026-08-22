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
	Queued      NodeStatus = "queued"     // parents done, waiting for a concurrency slot
	Running     NodeStatus = "running"    // bash command is executing
	Success     NodeStatus = "success"    // command exited with code 0
	Skipped     NodeStatus = "skipped"    // not run because a parent failed/was skipped
	Failed      NodeStatus = "failed"     // command exited with a non-zero code
	UpForRetry  NodeStatus = "upforretry" // failed attempt, waiting before next retry
	AllComplete ParentRule = "complete"   // run regardless of parent outcome
	AllSuccess  ParentRule = "success"    // run only if all parents succeeded (default)
)

// NodeSpec contains the static YAML-parsed configuration of a node.
type NodeSpec struct {
	Name       string     `yaml:"name"`
	Desc       string     `yaml:"desc"`
	Command    string     `yaml:"cmd"`
	ParentRule ParentRule `yaml:"parentRule"`
	Retries    int        `yaml:"retries"`
	RetryDelay int        `yaml:"retryDelay"`
}

// NodeState holds the runtime execution state of a node.
type NodeState struct {
	Status NodeStatus
	Pid    int
}

// Node represents a node in the DAG, combining its static spec with its
// runtime state.
type Node struct {
	Spec  NodeSpec
	State NodeState
}

// execute runs the node's bash command, retrying up to t.Spec.Retries additional
// times on failure. stdout and stderr are written to a per-attempt log file
// under .orca/<run-timestamp>/. Progress is reported via g.StatusChannel.
func (t *Node) execute(g *Graph, startTime time.Time) {
	var logFile *os.File
	var err error

	for attempt := 1; attempt <= int(math.Max(1, float64(t.Spec.Retries+1))); attempt++ {
		if attempt > 1 {
			t.retry(g, attempt)
			if t.Spec.RetryDelay > 0 {
				time.Sleep(time.Duration(t.Spec.RetryDelay) * time.Second)
			}
		}

		// Create log file for this attempt
		logFile, err = t.createLogFile(startTime, attempt)
		if err != nil {
			log.Errorf("Error creating log file for node %s: %v", t.Spec.Name, err)
			logFile.Close()
			return
		}

		cmd := exec.Command("bash", "-c", t.Spec.Command)
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		t.State.Status = Running
		if err := cmd.Start(); err != nil {
			log.Errorf("Error starting command for node %s: %v", t.Spec.Name, err)
			logFile.Close()
			continue
		}

		t.State.Pid = cmd.Process.Pid
		g.exec.statusChannel <- NodeStatusMsg{NodeKey: t.Spec.Name, Status: t.State.Status, Pid: t.State.Pid, Attempt: fmt.Sprintf("%d/%d", attempt, t.Spec.Retries+1)}

		err = cmd.Wait()
		logFile.Close()

		if err == nil {
			t.succeed(g, attempt)
			return
		}
	}

	// All attempts have failed
	t.fail(g)
}

// createLogFile opens (or creates) the log file for the given attempt under
// .orca/<run-timestamp>/<node-name>_<attempt>.log.
func (t *Node) createLogFile(startTime time.Time, attempt int) (*os.File, error) {
	logDir := fmt.Sprintf(".orca/%s", startTime.Format("2006-01-02_15-04-05.000000000"))
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(fmt.Sprintf("%s/%s_%d.log", logDir, t.Spec.Name, attempt))
}

// fail marks the node as Failed, records the failure on the graph, sends a
// final status update to the TUI, and notifies downstream nodes.
func (t *Node) fail(g *Graph) {
	g.exec.anyFailed.Store(true)
	t.State.Status = Failed
	g.exec.statusChannel <- NodeStatusMsg{
		NodeKey: t.Spec.Name,
		Status:  t.State.Status,
		Pid:     t.State.Pid,
		Attempt: fmt.Sprintf("%d/%d", t.Spec.Retries+1, t.Spec.Retries+1),
	}
	t.notifyChildren(g)
}

// succeed marks the node as Success, sends a final status update to the TUI,
// and notifies downstream nodes.
func (t *Node) succeed(g *Graph, attempt int) {
	t.State.Status = Success
	g.exec.statusChannel <- NodeStatusMsg{
		NodeKey: t.Spec.Name,
		Status:  t.State.Status,
		Pid:     t.State.Pid,
		Attempt: fmt.Sprintf("%d/%d", attempt, t.Spec.Retries+1),
	}
	t.notifyChildren(g)
}

// retry marks the node as UpForRetry, sends a status update to the TUI,
// and notifies downstream nodes.
func (t *Node) retry(g *Graph, attempt int) {
	t.State.Status = UpForRetry
	g.exec.statusChannel <- NodeStatusMsg{
		NodeKey: t.Spec.Name,
		Status:  t.State.Status,
		Pid:     t.State.Pid,
		Attempt: fmt.Sprintf("%d/%d", attempt, t.Spec.Retries+1),
	}
}

// notifyChildren sends the node's final status to the relay channel of each
// direct child, then closes that channel. Each send runs in its own goroutine
// tracked by g.notifyWG.
func (t *Node) notifyChildren(g *Graph) {
	for child := range g.Children[t.Spec.Name] {
		g.exec.notifyWG.Add(1)
		go func(child string) {
			defer g.exec.notifyWG.Done()
			key := edgeKey(t.Spec.Name, child)
			g.exec.nodeRelay[key] <- t.State.Status
			close(g.exec.nodeRelay[key])
		}(child)
	}
}
