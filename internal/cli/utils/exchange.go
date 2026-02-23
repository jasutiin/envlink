package cliutils

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

const googleAuthStateBytes = 24

type tokenExchangeRequest struct {
	ExchangeCode string `json:"exchange_code"`
	State        string `json:"state"`
}

type tokenExchangeResponse struct {
	Token string `json:"token"`
}

type CallbackResult struct {
	ExchangeCode string
	State        string
	Err          error
}

type GoogleTokenResult struct {
	AccessToken string
}

// NewCLISessionID creates a random state value for a CLI OAuth session.
//
// The CLI includes this state in the browser auth request and validates that
// the same value is returned on callback. This binds the callback to the
// NewCLISessionID generates a cryptographically secure random hex-encoded string to use as the CLI OAuth session state.
// The value encodes googleAuthStateBytes random bytes (hex length 2*googleAuthStateBytes) and is suitable for mitigating CSRF and callback-injection attacks.
// It returns the encoded state or an error if random byte generation fails.
func NewCLISessionID() (string, error) {
	b := make([]byte, googleAuthStateBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// BuildServerGoogleAuthURL builds the API auth endpoint URL for CLI login.
// It attaches the local callback URL and CLI state so the server can redirect
// BuildServerGoogleAuthURL constructs the local server URL that initiates Google OAuth for the CLI.
// The returned URL targets the local API auth endpoint and includes `cli_callback` and `cli_state` query parameters set to the provided callbackURL and state respectively.
func BuildServerGoogleAuthURL(callbackURL, state string) string {
	baseURL := "http://localhost:8080/api/v1/auth/google"
	values := url.Values{}
	values.Set("cli_callback", callbackURL)
	values.Set("cli_state", state)

	return baseURL + "?" + values.Encode()
}

// CreateLocalServer returns an *http.Server configured to handle the OAuth browser callback at callbackPath.
// It validates that the returned state matches expectedState and that an exchange code is present.
// On a successful callback it writes a success HTML response and sends a CallbackResult with ExchangeCode and State to resultChan.
// On error or validation failure it writes an error HTML response and sends a CallbackResult containing an Err to resultChan.
// The returned server is configured with the handler but is not started (caller must call ListenAndServe or Start).
func CreateLocalServer(callbackPath, expectedState string, resultChan chan<- CallbackResult) *http.Server {
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// if there is an OAuth error, display "Authentication failed" in the browser
		if oauthErr := strings.TrimSpace(query.Get("error")); oauthErr != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("<h1>Authentication failed</h1><p>Return to your CLI and try again.</p>"))
			resultChan <- CallbackResult{Err: fmt.Errorf("oauth error: %s", oauthErr)}
			return
		}

		// if invalid state, display "State validation failed." in the browser
		returnedState := strings.TrimSpace(query.Get("state"))
		exchangeCode := strings.TrimSpace(query.Get("exchange_code"))
		if returnedState == "" || exchangeCode == "" || returnedState != expectedState {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("<h1>Invalid callback</h1><p>State validation failed.</p>"))
			resultChan <- CallbackResult{Err: fmt.Errorf("invalid oauth callback state")}
			return
		}

		// if all checks pass, then authentication was successful. show it in browser
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<h1>Authentication complete</h1><p>You can close this window and return to the CLI.</p>"))
		resultChan <- CallbackResult{ExchangeCode: exchangeCode, State: returnedState}
	})

	return server
}

// ExchangeServerCode posts the one-time exchange code and state to the local API and returns the CLI access token.
// It sends the code and state to the local exchange endpoint, validates the server response, and returns a GoogleTokenResult
// containing the access token, or an error if the server rejects the exchange or the response is invalid.
func ExchangeServerCode(exchangeCode, state string) (*GoogleTokenResult, error) {
	// build the payload the API expects for code/state validation
	payload := tokenExchangeRequest{ExchangeCode: exchangeCode, State: state}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// create a bounded-time HTTP client to avoid hanging CLI auth
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/auth/cli/exchange", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	// send the exchange request to the local API
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read response body for both success and error branches
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// non-200 means the server rejected or could not validate the exchange
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	// parse successful response and ensure token is present
	var exchangeResp tokenExchangeResponse
	if err := json.Unmarshal(body, &exchangeResp); err != nil {
		return nil, err
	}

	if strings.TrimSpace(exchangeResp.Token) == "" {
		return nil, fmt.Errorf("empty token response")
	}

	return &GoogleTokenResult{AccessToken: exchangeResp.Token}, nil
}

// OpenInBrowser opens targetURL in the system default web browser on Windows.
// It returns any error encountered while attempting to launch the browser.
func OpenInBrowser(targetURL string) error {
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL).Start(); err == nil {
		return nil
	}

	return exec.Command("cmd", "/c", "start", "", fmt.Sprintf("\"%s\"", targetURL)).Start()
}