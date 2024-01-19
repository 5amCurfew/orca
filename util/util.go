package util

import "os"

type DepencyMap map[string]map[string]struct{}

func AddEdge(dm DepencyMap, from, to string) {
	nodes, ok := dm[from]
	if !ok {
		nodes = make(map[string]struct{})
		dm[from] = nodes
	}
	nodes[to] = struct{}{}
}

// getDagFiles returns a list of file names in the specified directory.
func ListFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		if file.IsDir() {
			// Skip directories, include only files
			continue
		}
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}

// getDagFiles returns a list of file names in the specified directory.
func ListDirs(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirNames []string
	for _, file := range files {
		if file.IsDir() {
			dirNames = append(dirNames, file.Name())
		}
	}

	return dirNames, nil
}
