package routes

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// getDagFiles returns a list of file names in the specified directory.
func getDagFiles(directory string) ([]string, error) {
	files, err := os.ReadDir(directory)
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

func Dags(c *gin.Context) {
	dagFiles, err := getDagFiles("dags")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dagFiles)
}
