package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
  Use:   "register",
  Short: "Register to envlink.",
  Long:  `Register to envlink.`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Register command ran.")
  },
}