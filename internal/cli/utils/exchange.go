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
// original login attempt and helps prevent CSRF/callback-injection attacks.
func NewCLISessionID() (string, error) {
	b := make([]byte, googleAuthStateBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// BuildServerGoogleAuthURL builds the API auth endpoint URL for CLI login.
// It attaches the local callback URL and CLI state so the server can redirect
// back to the CLI listener and preserve request integrity across the flow.
func BuildServerGoogleAuthURL(callbackURL, state string) string {
	baseURL := "http://localhost:8080/api/v1/auth/google"
	values := url.Values{}
	values.Set("cli_callback", callbackURL)
	values.Set("cli_state", state)

	return baseURL + "?" + values.Encode()
}

/*
This function creates a local server on the machine. This is used for listening to the browser's callback
function, which will be called if the server returns successfully.
*/
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

/*
ExchangeServerCode sends the one-time exchange code from the local OAuth callback to the API.

The API validates that:
1) the exchange code exists and is still valid,
2) the state value matches what was originally issued,
3) the code has not already been consumed.

If validation succeeds, the server returns an auth token for the CLI session.
If validation fails (expired/invalid code, state mismatch, or server rejection),
this function returns an error so the user can retry login.
*/
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

// OpenInBrowser opens a targetURL in a browser.
func OpenInBrowser(targetURL string) error {
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL).Start(); err == nil {
		return nil
	}

	return exec.Command("cmd", "/c", "start", "", fmt.Sprintf("\"%s\"", targetURL)).Start()
}
