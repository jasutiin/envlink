package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

/*
This function is called to initialize the gothic package with the external
providers we will be using for OAuth.
*/
func NewAuth(port string, domain string, key string) error {
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientId == "" {
		return errors.New("Google Client Id was not provided!")
	}

	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientSecret == "" {
		return errors.New("Google Client Secret was not provided!")
	}

	store := sessions.NewCookieStore([]byte(key))

	var url string

	if domain == "" {
		url = fmt.Sprintf("http://localhost:%s/api/v1/auth/google/callback", port)
	} else {
		url = fmt.Sprintf("https://%s/api/v1/auth/google/callback", domain)
	}

	gothic.Store = store
	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, url),
	)

	return nil
}