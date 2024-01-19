package lib

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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

func parseSchedule(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "schedule") {
			fields := strings.Split(line, "=")
			return strings.TrimSpace(fields[1]), nil
		}
	}
	return "", nil
}

func parseDependencies(filePath string, g *Graph) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.Contains(line, ">>") {
			// Split the line at ">>" to separate dependencies and the dependent task
			parts := strings.Split(line, ">>")
			dependencies := strings.TrimSpace(parts[0])
			dependentTask := strings.TrimSpace(parts[1])

			// Parse dependencies and add edges to the graph
			if strings.HasPrefix(dependencies, "[") && strings.HasSuffix(dependencies, "]") {
				// Case where dependencies are enclosed in square brackets
				dependencies = strings.TrimSuffix(strings.TrimPrefix(dependencies, "["), "]")
				dependencyList := strings.Split(dependencies, ",")

				for i := range dependencyList {
					dependencyList[i] = strings.TrimSpace(dependencyList[i])
				}

				for _, dependency := range dependencyList {
					err := g.dependOn(dependentTask, dependency)
					if err != nil {
						fmt.Println(err.Error())
						return err
					}
				}
			} else {
				// Case where there is a single dependency
				err := g.dependOn(dependentTask, dependencies)
				if err != nil {
					return err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
