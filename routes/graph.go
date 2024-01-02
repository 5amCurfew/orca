package routes

import (
	"encoding/json"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_path required"})
		return
	}

	g, err := lib.NewGraph(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to parse DAG: %s", err)})
		return
	}

	jsonRepresentation, _ := json.Marshal(g)

	c.JSON(http.StatusOK, gin.H{
		"graph":   json.RawMessage(jsonRepresentation),
		"message": fmt.Sprintf("DAG %s graph created", filePath),
	})
}
