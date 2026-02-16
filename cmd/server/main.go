package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jasutiin/envlink/internal/server/auth"
	"github.com/jasutiin/envlink/internal/server/database"
	"github.com/jasutiin/envlink/internal/server/projects"
	"github.com/jasutiin/envlink/internal/server/pull"
	"github.com/jasutiin/envlink/internal/server/push"
)

func main() {
	server := gin.Default()
	api := server.Group("/api/v1")
	db := database.CreateDB()
	database.AutoMigrate(db) // creates tables if they don't exist

	auth.AuthRouter(api, db)
	push.PushRouter(api)
	pull.PullRouter(api)
	projects.ProjectsRouter(api)

	server.Run("0.0.0.0:8080")
};