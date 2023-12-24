package main

import (
	"fmt"
	"os"

	"github.com/5amCurfew/orca/routes"
	"github.com/gin-gonic/gin"
)

// curl http://localhost:8080/ping

// curl -X POST \
// http://localhost:8080/execute \
// -H 'Content-Type: application/json' \
// -d '{"file_path": "dags/test.orca"}'

func main() {
	rest := gin.Default()

	rest.GET("/ping", func(c *gin.Context) {
		routes.Ping(c)
	})

	rest.POST("/execute", func(c *gin.Context) {
		routes.Execute(c)
	})

	if err := rest.Run(":8080"); err != nil {
		fmt.Fprintf(os.Stderr, "error starting Gin server: %s", err)
		os.Exit(1)
	}
}
