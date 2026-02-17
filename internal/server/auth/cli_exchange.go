package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const cliExchangeTTL = 2 * time.Minute

type cliExchangeStore struct {
	mu      sync.Mutex
	entries map[string]cliExchangeEntry
}

type cliExchangeEntry struct {
	token     string
	state     string
	expiresAt time.Time
}

const (
	cliCallbackCookieName = "envlink_cli_callback"
	cliStateCookieName    = "envlink_cli_state"
	cliCookieTTLSeconds   = 300
)

func newCLIExchangeStore() *cliExchangeStore {
	return &cliExchangeStore{entries: make(map[string]cliExchangeEntry)}
}

func (store *cliExchangeStore) Save(exchangeCode, state, token string, ttl time.Duration) {
	if exchangeCode == "" || state == "" || token == "" {
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	store.entries[exchangeCode] = cliExchangeEntry{
		token:     token,
		state:     state,
		expiresAt: time.Now().Add(ttl),
	}
}

func (store *cliExchangeStore) Consume(exchangeCode, state string) (string, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	entry, found := store.entries[exchangeCode]
	if !found {
		return "", false
	}

	delete(store.entries, exchangeCode)

	if time.Now().After(entry.expiresAt) {
		return "", false
	}

	if entry.state != state {
		return "", false
	}

	return entry.token, true
}

var pendingCLIExchanges = newCLIExchangeStore()

/*
This function checks if the callback url is something valid that a user
initiated themselves. This prevents the server from returning a different
callback url that the user expects. If we did not check this, the user may
be taken to a malicious website.
*/
func isAllowedCLICallback(rawCallbackURL string) bool {
	parsedURL, err := url.Parse(rawCallbackURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" {
		return false
	}

	hostName := strings.ToLower(parsedURL.Hostname())
	return hostName == "localhost" || hostName == "127.0.0.1" || hostName == "::1"
}

/*
This function sets httpOnly cookies for the callback url and state separately, both with an expiration time.
It adds it to the Gin context object which adds it to the response the server sends back. From there,
the browser would be sending these cookies to the server upon each subsequent request.
*/
func writeCLIAuthContext(c *gin.Context, callbackURL, state string) {
	c.SetCookie(cliCallbackCookieName, url.QueryEscape(callbackURL), cliCookieTTLSeconds, "/", "", false, true)
	c.SetCookie(cliStateCookieName, state, cliCookieTTLSeconds, "/", "", false, true)
}

/*
This function checks if the caller has cookies storing the callback url and state.
*/
func readCLIAuthContext(c *gin.Context) (string, string, bool) {
	callbackCookie, callbackErr := c.Cookie(cliCallbackCookieName)
	stateCookie, stateErr := c.Cookie(cliStateCookieName)
	if callbackErr != nil || stateErr != nil {
		return "", "", false
	}

	decodedCallback, decodeErr := url.QueryUnescape(callbackCookie)
	if decodeErr != nil || !isAllowedCLICallback(decodedCallback) {
		return "", "", false
	}

	return decodedCallback, stateCookie, true
}

/*
This function clears cookies from the response, signalling that we have successfully
received the user's credentials.
*/
func clearCLIAuthContext(c *gin.Context) {
	c.SetCookie(cliCallbackCookieName, "", -1, "/", "", false, true)
	c.SetCookie(cliStateCookieName, "", -1, "/", "", false, true)
}

/*
This function creates a new exchange code that the browser will use to verify
against the server.
*/
func newExchangeCode() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func buildCLIRedirectURL(callbackURL, exchangeCode, state string) (string, error) {
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return "", err
	}

	queryValues := parsedURL.Query()
	queryValues.Set("exchange_code", exchangeCode)
	queryValues.Set("state", state)
	parsedURL.RawQuery = queryValues.Encode()

	return parsedURL.String(), nil
}