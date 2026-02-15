package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
  Use:   "login",
  Short: "Login to envlink.",
  Long:  `Login to envlink.`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Login command ran.")
  },
}