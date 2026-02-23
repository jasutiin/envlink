package commands

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var LoginCmd = &cobra.Command{
  Use:   "login",
  Short: "Login to envlink.",
  Long:  `Login to envlink.`,
  Run: func(cmd *cobra.Command, args []string) {
    login()
  },
}

func login() {
  var email string;
  var password string;

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
  req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/auth/login", payload)

  if err != nil {
    fmt.Println("error on creating new POST req for login")
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