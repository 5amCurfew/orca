package routes

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func ExecutionTaskLog(c *gin.Context) {
	var requestData map[string]interface{}

	// Parse JSON request body
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract orca file path from the request
	logPath, ok := requestData["path"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}

	logContent, err := os.ReadFile(logPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log := string(logContent)

	c.JSON(http.StatusOK, gin.H{
		"log":     log,
		"message": logPath,
	})
}
