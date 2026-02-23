package cliutils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestNewCLISessionID verifies the generated session ID is a 24-byte hex string.
func TestNewCLISessionID(t *testing.T) {
	// generate a new random CLI session id
	sessionID, err := NewCLISessionID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 24 random bytes should be hex-encoded into 48 characters
	if len(sessionID) != 48 {
		t.Fatalf("expected session id length 48, got %d", len(sessionID))
	}

	// ensure every character is valid lowercase hex
	for _, char := range sessionID {
		if !strings.ContainsRune("0123456789abcdef", char) {
			t.Fatalf("expected hex string, got %q", sessionID)
		}
	}
}

// TestBuildServerGoogleAuthURL verifies callback and state are encoded into the auth URL.
func TestBuildServerGoogleAuthURL(t *testing.T) {
	callbackURL := "http://127.0.0.1:54001/oauth/google/callback"
	state := "test-state"

	// build URL and parse it back for query validation
	out := BuildServerGoogleAuthURL(callbackURL, state)
	parsed, err := url.Parse(out)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	if parsed.Scheme != "http" || parsed.Host != "localhost:8080" {
		t.Fatalf("unexpected base URL: %s", out)
	}

	query := parsed.Query()
	if query.Get("cli_callback") != callbackURL {
		t.Fatalf("expected cli_callback %q, got %q", callbackURL, query.Get("cli_callback"))
	}

	if query.Get("cli_state") != state {
		t.Fatalf("expected cli_state %q, got %q", state, query.Get("cli_state"))
	}
}

// TestCreateLocalServer_OAuthError verifies OAuth error responses are surfaced to the CLI.
func TestCreateLocalServer_OAuthError(t *testing.T) {
	// create handler and simulate callback with OAuth error
	resultChan := make(chan CallbackResult, 1)
	server := CreateLocalServer("/oauth/google/callback", "expected-state", resultChan)

	req := httptest.NewRequest("GET", "http://localhost/oauth/google/callback?error=access_denied", nil)
	recorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	// handler should report the error to result channel
	select {
	case result := <-resultChan:
		if result.Err == nil {
			t.Fatalf("expected callback error")
		}
		if !strings.Contains(result.Err.Error(), "oauth error") {
			t.Fatalf("expected oauth error message, got %v", result.Err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for callback result")
	}
}

// TestCreateLocalServer_InvalidState verifies state mismatches are rejected.
func TestCreateLocalServer_InvalidState(t *testing.T) {
	// create handler and simulate callback with mismatched state
	resultChan := make(chan CallbackResult, 1)
	server := CreateLocalServer("/oauth/google/callback", "expected-state", resultChan)

	req := httptest.NewRequest("GET", "http://localhost/oauth/google/callback?exchange_code=abc&state=wrong-state", nil)
	recorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	// handler should return state validation error through channel
	select {
	case result := <-resultChan:
		if result.Err == nil {
			t.Fatalf("expected callback error")
		}
		if !strings.Contains(result.Err.Error(), "invalid oauth callback state") {
			t.Fatalf("expected invalid state error, got %v", result.Err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for callback result")
	}
}

// TestCreateLocalServer_Success verifies a valid callback returns exchange code and state.
func TestCreateLocalServer_Success(t *testing.T) {
	// create handler and simulate a valid callback payload
	resultChan := make(chan CallbackResult, 1)
	expectedState := "expected-state"
	expectedCode := "exchange-code"
	server := CreateLocalServer("/oauth/google/callback", expectedState, resultChan)

	req := httptest.NewRequest("GET", "http://localhost/oauth/google/callback?exchange_code="+expectedCode+"&state="+expectedState, nil)
	recorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	// handler should pass code/state back to the CLI via channel
	select {
	case result := <-resultChan:
		if result.Err != nil {
			t.Fatalf("expected no error, got %v", result.Err)
		}
		if result.ExchangeCode != expectedCode {
			t.Fatalf("expected exchange code %q, got %q", expectedCode, result.ExchangeCode)
		}
		if result.State != expectedState {
			t.Fatalf("expected state %q, got %q", expectedState, result.State)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for callback result")
	}
}

// TestExchangeServerCode_Success verifies a valid server response returns an access token.
func TestExchangeServerCode_Success(t *testing.T) {
	// replace default transport so no real HTTP call is made
	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// validate outgoing request shape before returning mocked response
		if req.URL.String() != "http://localhost:8080/api/v1/auth/cli/exchange" {
			t.Fatalf("unexpected request URL: %s", req.URL.String())
		}

		if req.Method != http.MethodPost {
			t.Fatalf("expected method POST, got %s", req.Method)
		}

		if req.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content type, got %s", req.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		if !bytes.Contains(body, []byte(`"exchange_code":"code123"`)) || !bytes.Contains(body, []byte(`"state":"state123"`)) {
			t.Fatalf("unexpected request body: %s", string(body))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"token":"jwt-token"}`)),
			Header:     make(http.Header),
		}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	// function should parse token from mocked 200 response
	result, err := ExchangeServerCode("code123", "state123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatalf("expected token result")
	}

	if result.AccessToken != "jwt-token" {
		t.Fatalf("expected token jwt-token, got %s", result.AccessToken)
	}
}

// TestExchangeServerCode_NonOKResponse verifies non-200 server responses return an error.
func TestExchangeServerCode_NonOKResponse(t *testing.T) {
	// mock unauthorized response from API
	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader(`{"error":"invalid"}`)),
			Header:     make(http.Header),
		}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	// function should surface status code in returned error
	_, err := ExchangeServerCode("bad-code", "state123")
	if err == nil {
		t.Fatalf("expected error for non-200 response")
	}

	if !strings.Contains(err.Error(), "server exchange failed with status 401") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestExchangeServerCode_EmptyToken verifies empty/blank token payloads are rejected.
func TestExchangeServerCode_EmptyToken(t *testing.T) {
	// mock success status with blank token payload
	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"token":"   "}`)),
			Header:     make(http.Header),
		}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	// function should reject empty token responses
	_, err := ExchangeServerCode("code123", "state123")
	if err == nil {
		t.Fatalf("expected error for empty token")
	}

	if !strings.Contains(err.Error(), "empty token response") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestExchangeServerCode_InvalidJSON verifies malformed JSON responses return an error.
func TestExchangeServerCode_InvalidJSON(t *testing.T) {
	// mock malformed JSON body from API
	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`not-json`)),
			Header:     make(http.Header),
		}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	// function should return JSON parse error
	_, err := ExchangeServerCode("code123", "state123")
	if err == nil {
		t.Fatalf("expected JSON unmarshal error")
	}
}

// TestExchangeServerCode_DoError verifies transport-level HTTP failures are propagated.
func TestExchangeServerCode_DoError(t *testing.T) {
	// mock transport-level failure before any response is received
	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("transport failure")
	})
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	// function should propagate the transport error
	_, err := ExchangeServerCode("code123", "state123")
	if err == nil {
		t.Fatalf("expected transport error")
	}

	if !strings.Contains(err.Error(), "transport failure") {
		t.Fatalf("unexpected error: %v", err)
	}
}