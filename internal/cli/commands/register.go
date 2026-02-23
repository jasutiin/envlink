package commands

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	cliutils "github.com/jasutiin/envlink/internal/cli/utils"
	"github.com/spf13/cobra"
)

var RegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register to envlink.",
	Long:  `Register to envlink.`,
	Run: func(cmd *cobra.Command, args []string) {
		register()
	},
}

func register() {
	var choice string
	fmt.Println("1) Email/Password")
	fmt.Println("2) Google")

	fmt.Print("Which auth provider would you like to use? ")
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		registerUsingEmailPassword()
	case "2":
		registerUsingGoogle()
	default:
		fmt.Println("Cancelled.")
	}
}

func registerUsingEmailPassword() {
	var email string
	var password string

	fmt.Printf("Email: ")
	fmt.Scanln(&email)

	fmt.Printf("Password: ")
	fmt.Scanln(&password)

	if email != "" {
		fmt.Println("email provided")
	} else {
		fmt.Println("email not provided")
	}

	if password != "" {
		fmt.Println("password provided")
	} else {
		fmt.Println("password not provided")
	}

	jsonStr := []byte(fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password))
	payload := bytes.NewBuffer(jsonStr)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/auth/register", payload)

	if err != nil {
		fmt.Println("error on creating new POST req for register")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error performing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("request successful with status code: %d\n", resp.StatusCode)
	} else {
		fmt.Printf("request failed with status code: %d\n", resp.StatusCode)
	}
}

func registerUsingGoogle() {
	state, err := cliutils.NewCLISessionID()
	if err != nil {
		fmt.Println("failed to create auth state")
		return
	}

	// listen on localhost with a port assigned by OS
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("failed to start local callback listener")
		return
	}
	defer listener.Close() // this will run when registerUsingGoogle() function returns

	// callback path on the CLI's temporary HTTP server
	callbackPath := "/oauth/google/callback"
	callbackURL := fmt.Sprintf("http://%s%s", listener.Addr().String(), callbackPath)

	// make the actual server url that the CLI will call to
	authURL := cliutils.BuildServerGoogleAuthURL(callbackURL, state)
	resultChan := make(chan cliutils.CallbackResult, 1)
	server := cliutils.CreateLocalServer(callbackPath, state, resultChan)

	// run a local server for the CLI
	go func() {
		_ = server.Serve(listener)
	}()

	fmt.Println("Opening browser to:", authURL)
	if err := cliutils.OpenInBrowser(authURL); err != nil {
		fmt.Println("failed to open browser automatically. Open this URL manually:")
		fmt.Println(authURL)
		_ = server.Close()
		return
	}

	waitTimer := time.NewTimer(2 * time.Minute)
	defer waitTimer.Stop()

	var callback cliutils.CallbackResult

	// either wait until the waitTimer runs out, or if a result was returned from the local server
	select {
	case callback = <-resultChan:
		if callback.Err != nil {
			fmt.Printf("google auth did not complete: %v\n", callback.Err)
			_ = server.Close()
			return
		}
	case <-waitTimer.C:
		fmt.Println("timed out waiting for google authentication")
		_ = server.Close()
		return
	}

	token, err := cliutils.ExchangeServerCode(callback.ExchangeCode, callback.State)
	if err != nil {
		fmt.Printf("token exchange failed: %v\n", err)
		_ = server.Close()
		return
	}

	_ = server.Close()

	fmt.Println("Google authentication successful.")
	fmt.Println("Token:", token.AccessToken)
}