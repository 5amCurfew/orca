package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/5amCurfew/orca/lib"
	"github.com/gin-gonic/gin"
)

func Execute(c *gin.Context) {
	var requestData map[string]interface{}

	// Parse JSON request body
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract orca file path from the request
	d, ok := requestData["path"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}
	filePath := fmt.Sprintf("%s.orca", d)

	g, err := lib.NewGraph(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to parse DAG: %s", err)})
		return
	}

	g.Execute(time.Now())

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("DAG %s execution completed", filePath),
	})
}
