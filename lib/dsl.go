package lib

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// //////////////////////////////
// Create Graph from File Path
// //////////////////////////////
func NewGraph(filePath string) (*Graph, error) {
	tasks, err := parseTasks(filePath)
	if err != nil {
		return &Graph{}, errors.New(err.Error())
	}

	g := &Graph{
		File:     filePath,
		Name:     filePath[:strings.Index(filePath, ".orca")],
		Tasks:    tasks,
		Parents:  make(DepencyMap),
		Children: make(DepencyMap),
	}

	err = g.parseDependencies()
	if err != nil {
		return &Graph{}, errors.New(err.Error())
	}

	dirPath := fmt.Sprintf(".orca/%s", g.Name)
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		log.Errorf("Error creating logs directory: %s\n", err)
	}

	return g, nil
}

// //////////////////////////////
// Parse Dependency Edges from File
// //////////////////////////////
func (g *Graph) parseDependencies() error {
	file, _ := os.Open(g.File)
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
					err := g.addDependency(dependentTask, dependency)
					if err != nil {
						fmt.Println(err.Error())
						return err
					}
				}
			} else {
				// Case where there is a single dependency
				err := g.addDependency(dependentTask, dependencies)
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
