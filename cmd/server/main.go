package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jasutiin/envlink/internal/server/auth"
)

func main() {
	server := gin.Default()
	api := server.Group("/api/v1")
	auth.AuthRouter(api)
	server.Run("localhost:8080")
};

func store(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "store")
}

func push(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "push")
}

func pull(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "pull")
}

func projects(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "projects")
}