package database

import (
	"fmt"
	"os"

	"github.com/jasutiin/envlink/internal/server/api/auth"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

func AutoMigrate(db *gorm.DB) {
	if err := db.AutoMigrate(&auth.User{}); err != nil {
		fmt.Println("migrate failed:", err)
	}
}
