package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
  Use:   "push",
  Short: "Pushes your project's .env to the database.",
  Long:  `Pushes your project's .env to the database. It will update the entry whether there are new changes or not.`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Push command ran.")
  },
}