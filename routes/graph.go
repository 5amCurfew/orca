package routes

import (
	"fmt"
	"net/http"

	"github.com/5amCurfew/orca/lib"
	"github.com/gin-gonic/gin"
)

func Graph(c *gin.Context) {
	var requestData map[string]interface{}

	// Parse JSON request body
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract orca file path from the request
	filePath, ok := requestData["file_path"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing orca file_path"})
		return
	}

	tasks, err := lib.ParseTasks(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to parse tasks: %s", err)})
		return
	}

	g := lib.NewGraph(tasks)
	g.AddNodes()
	lib.ParseDependencies(filePath, g)
	g.CreateTopologicalLayers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to marshal JSON: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"html":    fmt.Sprintf("<div class=\"placeholder\">%s selected</div>", filePath),
		"message": fmt.Sprintf("DAG %s graph created", filePath),
	})
}
