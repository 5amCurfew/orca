package lib

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Initalise Graph
func Init(filePath string) error {
	var err error

	// Check if the dag file exists before proceeding
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("file does not exist: " + filePath)
	}

	// Check that the file has a .yml extension
	if !strings.HasSuffix(filePath, ".yml") {
		return errors.New("file must be a valid yaml file: " + filePath)
	}

	G.File = filePath
	G.Name = strings.TrimSuffix(filepath.Base(filePath), ".yml")

	G.File = filePath
	G.Name = filePath[:strings.Index(filePath, ".yml")]
	G.Nodes = make(map[string]*Node)
	G.Parents = make(DepencyMap)
	G.Children = make(DepencyMap)

	err = G.parseNodes()
	if err != nil {
		return errors.New(err.Error())
	}

	err = G.parseEdges()
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
