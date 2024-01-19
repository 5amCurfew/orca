package routes

import (
	"fmt"
	"net/http"

	"github.com/5amCurfew/orca/lib"
	"github.com/5amCurfew/orca/util"
	"github.com/gin-gonic/gin"
)

func Refresh(c *gin.Context) {
	dagFiles, err := util.ListFiles("dags")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go lib.UpdateSchedule()

	c.JSON(http.StatusOK, gin.H{
		"dagList": dagFiles,
		"message": fmt.Sprintf("DAGs %s", dagFiles),
	})
}
