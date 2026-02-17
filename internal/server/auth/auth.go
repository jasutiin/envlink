package auth

import (
	"fmt"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	key = "randomString"
	MaxAge = 86400 * 30
	IsProd = false
)

func NewAuth() {
	err := godotenv.Load()
	if err != nil{
		fmt.Println("problem loading .env file")
	}

	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	store := sessions.NewCookieStore([]byte(key))

	var url string
	domain := os.Getenv("RAILWAY_PUBLIC_DOMAIN")
	port := os.Getenv("PORT")
	if domain == "" {
		url = fmt.Sprintf("http://localhost:%s/api/v1/auth/google/callback", port)
	} else {
		url = fmt.Sprintf("https://%s/api/v1/auth/google/callback", domain)
	}

	gothic.Store = store
	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, url),
	)
}