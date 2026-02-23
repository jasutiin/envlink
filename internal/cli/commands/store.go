package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var StoreCmd = &cobra.Command{
  Use:   "store",
  Short: "Store your secret key.",
  Long:  `Store your secret key that was generated when you first registered.`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Store command ran.")
  },
}