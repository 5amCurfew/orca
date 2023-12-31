package main

import (
	"fmt"
	"os"

	"github.com/5amCurfew/orca/routes"
	"github.com/gin-gonic/gin"
)

func main() {

	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		fmt.Println("Error creating logs directory file:", err)
	}

	router := gin.Default()

	router.Static("/ui", "./ui")
	router.LoadHTMLGlob("ui/*.html")

	router.GET("/ping", routes.Ping)
	router.GET("/ui", routes.UI)
	router.GET("/pulse", routes.Pulse)
	router.POST("/graph", routes.Graph)
	router.POST("/status", routes.Status)
	router.POST("/execute", routes.Execute)

	if err := router.Run(":8080"); err != nil {
		fmt.Fprintf(os.Stderr, "error starting Gin server: %s", err)
		os.Exit(1)
	}
}
