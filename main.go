package main

import (
	"lona"
	"net/http"

	"github.com/gin-gonic/gin"
)

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Ping test
	r.GET("/lona", func(c *gin.Context) {
		p := lona.LonaPrintReport()
		c.String(http.StatusOK, p)
	})

	return r
}

func main() {
	// lona.LonaPrintReport()
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")

}
