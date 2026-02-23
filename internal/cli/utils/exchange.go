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

type callbackResult struct {
	exchangeCode string
	state        string
	err          error
}

type tokenExchangeRequest struct {
	ExchangeCode string `json:"exchange_code"`
	State        string `json:"state"`
}

type tokenExchangeResponse struct {
	Token string `json:"token"`
}

type googleTokenResult struct {
	accessToken string
}

type CallbackResult struct {
	ExchangeCode string
	State        string
	Err          error
}

type GoogleTokenResult struct {
	AccessToken string
}

func NewCLISessionID() (string, error) {
	b := make([]byte, googleAuthStateBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

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

		if oauthErr := strings.TrimSpace(query.Get("error")); oauthErr != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("<h1>Authentication failed</h1><p>Return to your CLI and try again.</p>"))
			resultChan <- CallbackResult{Err: fmt.Errorf("oauth error: %s", oauthErr)}
			return
		}

		returnedState := strings.TrimSpace(query.Get("state"))
		exchangeCode := strings.TrimSpace(query.Get("exchange_code"))
		if returnedState == "" || exchangeCode == "" || returnedState != expectedState {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("<h1>Invalid callback</h1><p>State validation failed.</p>"))
			resultChan <- CallbackResult{Err: fmt.Errorf("invalid oauth callback state")}
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<h1>Authentication complete</h1><p>You can close this window and return to the CLI.</p>"))
		resultChan <- CallbackResult{ExchangeCode: exchangeCode, State: returnedState}
	})

	return server
}

/*
This function takes the exchange code and checks if it exists in the server. If it doesn't, then it may have expired and
the user would have to redo the auth process. If it does, return an authentication token
*/
func ExchangeServerCode(exchangeCode, state string) (*GoogleTokenResult, error) {
	payload := tokenExchangeRequest{ExchangeCode: exchangeCode, State: state}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/auth/cli/exchange", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var exchangeResp tokenExchangeResponse
	if err := json.Unmarshal(body, &exchangeResp); err != nil {
		return nil, err
	}

	if strings.TrimSpace(exchangeResp.Token) == "" {
		return nil, fmt.Errorf("empty token response")
	}

	return &GoogleTokenResult{AccessToken: exchangeResp.Token}, nil
}

func OpenInBrowser(targetURL string) error {
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL).Start(); err == nil {
		return nil
	}

	return exec.Command("cmd", "/c", "start", "", fmt.Sprintf("\"%s\"", targetURL)).Start()
}
