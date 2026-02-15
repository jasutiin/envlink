package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
  Use:   "pull",
  Short: "Pulls the project's latest changes to the .env file.",
  Long:  `Pulls the project's latest changes to the .env file. It will update your local .env whether it is new or not.`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Pull command ran.")
  },
}