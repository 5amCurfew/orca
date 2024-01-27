package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Status(c *gin.Context) {
	var requestData map[string]interface{}

	// Parse JSON request body
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract orca file path from the request
	filePath, ok := requestData["path"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("status route called for %s", filePath),
	})
}
