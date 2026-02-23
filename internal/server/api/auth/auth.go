package auth

import (
	"errors"
	"fmt"
	"net/http"
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
func NewAuth(port string, domain string, key string, isProd bool) error {
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientId == "" {
		return errors.New("Google Client Id was not provided!")
	}

	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientSecret == "" {
		return errors.New("Google Client Secret was not provided!")
	}

	store := sessions.NewCookieStore([]byte(key))
	store.Options = &sessions.Options{
		Path:     "/", // cookie is valid for all paths on the host
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	}

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