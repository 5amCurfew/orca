package main

import (
	"fmt"
	"os"

	"github.com/5amCurfew/orca/routes"
	"github.com/gin-gonic/gin"
)

// curl http://localhost:8080/ping

// curl -X POST http://localhost:8080/execute -H 'Content-Type: application/json' -d '{"file_path": "dags/test.orca"}'

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("ui/*.html")

	router.GET("/ping", routes.Ping)
	router.GET("/dags", routes.Dags)
	router.GET("/ui", routes.UI)
	router.POST("/execute", routes.Execute)

	if err := router.Run(":8080"); err != nil {
		fmt.Fprintf(os.Stderr, "error starting Gin server: %s", err)
		os.Exit(1)
	}
}
