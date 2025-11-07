package lib

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// Create Graph
var G Graph

// Initalise Graph
func (g *Graph) Init(filePath string) error {
	var err error

	// Check if the dag file exists before proceeding
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("file does not exist: " + filePath)
	}

	// Check that the file has a .yml extension
	if !strings.HasSuffix(filePath, ".yml") {
		return errors.New("file must be a valid yaml file: " + filePath)
	}

	g.File = filePath
	g.Name = strings.TrimSuffix(filepath.Base(filePath), ".yml")

	g.File = filePath
	g.Name = filePath[:strings.Index(filePath, ".yml")]
	g.Nodes = make(map[string]*Node)
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

	// Unmarshal YAML data into NodeYaml struct
	var nodeYaml NodeYaml
	err = yaml.Unmarshal(graphYML, &nodeYaml)
	if err != nil {
		return err
	}

	// Create a map of Nodes from the parsed YAML
	Nodes := make(map[string]*Node)
	for _, Node := range nodeYaml.Nodes {
		NodeCopy := Node          // Create a copy of the Node
		NodeCopy.Status = Pending // Initialize status
		if NodeCopy.ParentRule == "" {
			NodeCopy.ParentRule = AllSuccess // Default parentRule if not set
		}
		Nodes[Node.Name] = &NodeCopy
	}

	g.Nodes = Nodes

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
	for Node, parents := range dependencyYaml.Dependencies {
		for _, parent := range parents {
			err := g.addDependency(Node, parent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
