package routes

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UI(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":      "Orca",
		"graphHTML":  template.HTML("<div class=\"placeholder\">Nothing selected</div>"),
		"statusHTML": template.HTML("<div class=\"placeholder\">Nothing selected</div>"),
	})
}
