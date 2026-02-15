package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jasutiin/envlink/internal/server/auth"
	"github.com/jasutiin/envlink/internal/server/projects"
	"github.com/jasutiin/envlink/internal/server/pull"
	"github.com/jasutiin/envlink/internal/server/push"
)

func main() {
	server := gin.Default()
	api := server.Group("/api/v1")
	auth.AuthRouter(api)
	push.PushRouter(api)
	pull.PullRouter(api)
	projects.ProjectsRouter(api)
	server.Run("localhost:8080")
};