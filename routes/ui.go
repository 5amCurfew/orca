package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UI(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}
