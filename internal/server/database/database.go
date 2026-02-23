package database

import (
	"fmt"
	"os"

	"github.com/jasutiin/envlink/internal/server/api/auth"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CreateDB creates a GORM DB connection configured from environment variables.
// It attempts to load a .env file and reads DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, and DB_PORT; errors are printed and not returned, and the function may return nil if the connection cannot be opened.
func CreateDB() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require", host, user, password, name, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Println("failed to open db")
	}

	return db
}

// AutoMigrate applies schema migrations for the auth.User model to the provided database.
// If migration fails, the error is printed to standard output and no error is returned.
func AutoMigrate(db *gorm.DB) {
	if err := db.AutoMigrate(&auth.User{}); err != nil {
		fmt.Println("migrate failed:", err)
	}
}