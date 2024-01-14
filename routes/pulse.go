package routes

import (
	"fmt"
	"net/http"

	"github.com/5amCurfew/orca/util"
	"github.com/gin-gonic/gin"
)

func Pulse(c *gin.Context) {
	dagFiles, err := util.GetDagFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dagList": dagFiles,
		"message": fmt.Sprintf("DAGs %s", dagFiles),
	})
}
