package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ProjectsCmd = &cobra.Command{
  Use:   "projects",
  Short: "Lists all the .envs you have stored.",
  Long:  `Lists all the .envs you have stored.`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Projects command ran.")
  },
}