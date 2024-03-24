package lib

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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
	log.Infof("%s task execution started", t.Name)

	cmdParts := []string{"bash", "-c", t.Command}
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	t.Status = Running

	// Create log directory if it doesn't exist
	logDir := fmt.Sprintf(".orca/%s/%s", g.Name, dagExecutionStartTime.Format("2006-01-02_15-04-05"))
	os.MkdirAll(logDir, os.ModePerm)

	// Create log file
	logFile, err := os.Create(fmt.Sprintf("%s/%s.log", logDir, t.Name))
	if err != nil {
		log.Warnf("Error creating log output file: %s", err)
		completionChMap[t.Name] <- false
		t.Status = Failed
		return
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		completionChMap[t.Name] <- false
		t.Status = Failed
		log.Warnf("%s task execution failed at %s\n", t.Name, time.Now().Format("2006-01-02 15:04:05"))
	} else {
		log.Infof("%s task execution completed sucessfully", t.Name)
		completionChMap[t.Name] <- true
		t.Status = Success
	}
}

func parseTasks(filePath string) (map[string]*Task, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tasks := make(map[string]*Task)
	var currentTask *Task
	var currentField string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "task {"):
			currentTask = &Task{Status: Pending}
		case strings.HasPrefix(line, "name"):
			fields := strings.Split(line, "=")
			currentTask.Name = strings.TrimSpace(fields[1])
		case strings.HasPrefix(line, "desc"):
			fields := strings.Split(line, "=")
			currentField = "desc"
			currentTask.Desc = strings.TrimSpace(fields[1])
		case strings.HasPrefix(line, "cmd"):
			fields := strings.Split(line, "=")
			currentField = "cmd"
			currentTask.Command = strings.TrimSpace(fields[1])
		case line == "}" && currentTask != nil:
			tasks[currentTask.Name] = currentTask
			currentTask = nil
			currentField = ""
		default:
			if currentField == "desc" {
				currentTask.Desc += " " + strings.TrimSpace(line)
			} else if currentField == "cmd" {
				currentTask.Command += " " + strings.TrimSpace(line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
