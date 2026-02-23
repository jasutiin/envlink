package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jasutiin/envlink/internal/server/auth"
	"github.com/jasutiin/envlink/internal/server/database"
	"github.com/jasutiin/envlink/internal/server/projects"
	"github.com/jasutiin/envlink/internal/server/pull"
	"github.com/jasutiin/envlink/internal/server/push"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil{
		log.Fatalf("Could not load environment variables!")
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("Port was not provided!")
	}

	server := gin.Default()
	api := server.Group("/api/v1")
	db := database.CreateDB()
	database.AutoMigrate(db) // creates tables if they don't exist

	// empty RAILWAY_ENVIRONMENT_NAME means dev environment, otherwise production
	isProd := os.Getenv("RAILWAY_ENVIRONMENT_NAME") != ""

	key := os.Getenv("COOKIE_SESSION_KEY")
	if key == "" {
		log.Fatalf("COOKIE_SESSION_KEY is required")
	}

	domain := os.Getenv("RAILWAY_PUBLIC_DOMAIN")

	err = auth.NewAuth(port, domain, key, isProd)
	if err != nil {
		log.Fatalf("Failed to initialize auth: %s", err)
	}

	auth.AuthRouter(api, db)
	push.PushRouter(api)
	pull.PullRouter(api)
	projects.ProjectsRouter(api)

	fmt.Printf("listening on port %s", port)
	server.Run("0.0.0.0:" + port)
};