package lib

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// Create Graph
var G Graph

// Initalise Graph
func (g *Graph) Init(filePath string) error {
	var err error

	g.File = filePath
	g.Name = filePath[:strings.Index(filePath, ".yml")]
	g.Tasks = make(map[string]*Task)
	g.Parents = make(DepencyMap)
	g.Children = make(DepencyMap)

	err = g.parseNodes()
	if err != nil {
		return errors.New(err.Error())
	}

	err = g.parseEdges()
	if err != nil {
		return errors.New(err.Error())
	}

	dirPath := ".orca"
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		log.Errorf("Error creating logs directory: %s\n", err)
	}

	return nil
}

// Parse nodes from file
func (g *Graph) parseNodes() error {
	graphYML, err := os.ReadFile(g.File)
	if err != nil {
		return err
	}

	// Define structure to match YAML format
	type TaskYaml struct {
		Tasks []Task `yaml:"tasks"`
	}

	// Unmarshal YAML data into TaskYaml struct
	var taskYaml TaskYaml
	err = yaml.Unmarshal(graphYML, &taskYaml)
	if err != nil {
		return err
	}

	// Create a map of tasks from the parsed YAML
	tasks := make(map[string]*Task)
	for _, task := range taskYaml.Tasks {
		taskCopy := task          // Create a copy of the task
		taskCopy.Status = Pending // Initialize status
		if taskCopy.ParentRule == "" {
			taskCopy.ParentRule = AllSuccess // Default parentRule if not set
		}
		tasks[task.Name] = &taskCopy
	}

	g.Tasks = tasks

	return nil
}

// Parse edges from File
func (g *Graph) parseEdges() error {
	// Open and read the YAML file
	graphYML, err := os.ReadFile(g.File)
	if err != nil {
		return err
	}

	// Define structure to match YAML format
	type DependencyYaml struct {
		Dependencies map[string][]string `yaml:"dependencies"`
	}

	// Unmarshal YAML data into DependencyYaml struct
	var dependencyYaml DependencyYaml
	err = yaml.Unmarshal(graphYML, &dependencyYaml)
	if err != nil {
		return err
	}

	// Loop through the dependencies and add them to the graph
	for task, parents := range dependencyYaml.Dependencies {
		for _, parent := range parents {
			err := g.addDependency(task, parent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
