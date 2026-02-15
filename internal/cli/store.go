package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
  Use:   "store",
  Short: "Store your secret key.",
  Long:  `Store your secret key that was generated when you first registered.`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Store command ran.")
  },
}