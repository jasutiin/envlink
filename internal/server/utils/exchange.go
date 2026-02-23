package serverutils

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const CLIExchangeTTL = 2 * time.Minute

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

// The store's mutex is left as the zero value (unlocked).
func newCLIExchangeStore() *cliExchangeStore {
	return &cliExchangeStore{entries: make(map[string]cliExchangeEntry)}
}

/*
Save saves a new cliExchangeEntry to the cliExchangeStore.
*/
func (store *cliExchangeStore) Save(exchangeCode, state, token string, ttl time.Duration) {
	if exchangeCode == "" || state == "" || token == "" {
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock() // store.mu.Unlock() is called before Save() returns

	store.entries[exchangeCode] = cliExchangeEntry{
		token:     token,
		state:     state,
		expiresAt: time.Now().Add(ttl),
	}
}

/*
Consume consumes a cliExchangeEntry given an exchange code
*/
func (store *cliExchangeStore) Consume(exchangeCode, state string) (string, bool) {
	store.mu.Lock()
	defer store.mu.Unlock() // store.mu.Unlock() is called before Consume() returns

	entry, found := store.entries[exchangeCode]
	if !found {
		return "", false
	}

	delete(store.entries, exchangeCode) // deletes the entry with exchangeCode as its key from the store.entries map

	if time.Now().After(entry.expiresAt) {
		return "", false
	}

	if entry.state != state {
		return "", false
	}

	return entry.token, true
}

var PendingCLIExchanges = newCLIExchangeStore()

// IsAllowedCLICallback reports whether rawCallbackURL is an allowed local HTTP callback URL.
// It accepts only URLs with scheme "http" and hostname "localhost", "127.0.0.1", or "::1".
func IsAllowedCLICallback(rawCallbackURL string) bool {
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

// WriteCLIAuthContext sets HTTP-only cookies on the provided Gin context to persist a URL-escaped
// callback URL and the state value for a CLI authentication flow. The cookies are set for path "/",
// use the package TTL (cliCookieTTLSeconds), and are not marked as secure. The callback value is URL-escaped
// before being stored.
func WriteCLIAuthContext(c *gin.Context, callbackURL, state string) {
	c.SetCookie(cliCallbackCookieName, url.QueryEscape(callbackURL), cliCookieTTLSeconds, "/", "", false, true)
	c.SetCookie(cliStateCookieName, state, cliCookieTTLSeconds, "/", "", false, true)
}

// ReadCLIAuthContext reads the CLI callback URL and state from HTTP cookies, validates and URL-decodes the callback, and returns them when valid.
// 
// It expects cookies named "envlink_cli_callback" and "envlink_cli_state". The callback value is URL-decoded and must be an allowed local HTTP callback (for example, localhost, 127.0.0.1, or ::1).
//
// Returns the decoded callback URL, the state value, and `true` on success; otherwise returns empty strings and `false`.
func ReadCLIAuthContext(c *gin.Context) (string, string, bool) {
	callbackCookie, callbackErr := c.Cookie(cliCallbackCookieName)
	stateCookie, stateErr := c.Cookie(cliStateCookieName)
	if callbackErr != nil || stateErr != nil {
		return "", "", false
	}

	decodedCallback, decodeErr := url.QueryUnescape(callbackCookie)
	if decodeErr != nil || !IsAllowedCLICallback(decodedCallback) {
		return "", "", false
	}

	return decodedCallback, stateCookie, true
}

// ClearCLIAuthContext clears the CLI authentication cookies on the response, removing the stored callback URL and state by expiring their cookies.
func ClearCLIAuthContext(c *gin.Context) {
	c.SetCookie(cliCallbackCookieName, "", -1, "/", "", false, true)
	c.SetCookie(cliStateCookieName, "", -1, "/", "", false, true)
}

// NewExchangeCode returns a new random exchange code encoded as a 48-character hex string.
// It reads 24 cryptographically secure random bytes and returns their hex encoding, or an error if random generation fails.
func NewExchangeCode() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// BuildCLIRedirectURL constructs a redirect URL by adding the "exchange_code" and "state"
// query parameters to the provided callbackURL.
// It returns the resulting URL string, or an error if callbackURL cannot be parsed.
func BuildCLIRedirectURL(callbackURL, exchangeCode, state string) (string, error) {
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