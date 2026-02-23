package serverutils

import (
	"net/url"
	"testing"
	"time"
)

// TestIsAllowedCLICallback tests multiple URLs to see if they are allowed as a callback url.
func TestIsAllowedCLICallback(t *testing.T) {
	allowed := []string{
		"http://localhost:8080/cb",
		"http://127.0.0.1/cb",
		"http://[::1]/cb",
	}

	// all of the urls in allowed slice should return true
	for _, u := range allowed {
		if !IsAllowedCLICallback(u) {
			t.Errorf("expected allowed for %s", u)
		}
	}

	disallowed := []string{
		"https://localhost/cb",
		"http://example.com/cb",
		"notaurl",
	}

	// all of the urls in disallowed slice should return false
	for _, u := range disallowed {
		if IsAllowedCLICallback(u) {
			t.Errorf("expected disallowed for %s", u)
		}
	}
}

// TestBuildCLIRedirectURL tests if BuildCLIRedirectURL() correctly appends
// exchange_code and state to the callback's query string.
func TestBuildCLIRedirectURL(t *testing.T) {
	callback := "http://localhost:3000/cb?foo=bar"
	code := "abc123"
	state := "s1"

	out, err := BuildCLIRedirectURL(callback, code, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := url.Parse(out)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	q := parsed.Query()
	if q.Get("exchange_code") != code {
		t.Fatalf("expected exchange_code %s got %s", code, q.Get("exchange_code"))
	}
	if q.Get("state") != state {
		t.Fatalf("expected state %s got %s", state, q.Get("state"))
	}
}

// TestNewExchangeCode tests if NewExchangeCode() returns a non-error hex string
// of 48 characters or 24 encoded bytes.
func TestNewExchangeCode(t *testing.T) {
	code, err := NewExchangeCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != 48 {
		t.Fatalf("expected code length 48 got %d", len(code))
	}
}

// TestCLIExchangeStore_SaveConsume tests the cliExchangeStore's Save() and Consume()
// functions to see if they are working properly.
func TestCLIExchangeStore_SaveConsume(t *testing.T) {
	store := newCLIExchangeStore()

	store.Save("code1", "state1", "token1", time.Minute) // save a new entry
	token, found := store.Consume("code1", "state1")
	if !found || token != "token1" {
		t.Fatalf("expected to find token, got found=%v tok=%s", found, token)
	}

	// second consume should fail because we've already consumed it in the prev call
	_, found = store.Consume("code1", "state1")
	if found {
		t.Fatalf("expected second consume to fail")
	}

	// wrong state
	store.Save("code2", "state2", "token2", time.Minute)
	_, found = store.Consume("code2", "wrong")
	if found {
		t.Fatalf("expected consume with wrong state to fail")
	}

	// expired
	store.Save("code3", "state3", "token3", -time.Second)
	_, found = store.Consume("code3", "state3")
	if found {
		t.Fatalf("expected expired consume to fail")
	}
}
