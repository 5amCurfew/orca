package routes

import (
	"net/http"

	"github.com/5amCurfew/orca/util"
	"github.com/gin-gonic/gin"
)

func ExecutionLogs(c *gin.Context) {
	var requestData map[string]interface{}

	// Parse JSON request body
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract orca file path from the request
	logsPath, ok := requestData["path"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}

	logDirectories, _ := util.ListDirs(logsPath)
	c.JSON(http.StatusOK, gin.H{
		"logList": logDirectories,
		"message": logsPath,
	})
}
