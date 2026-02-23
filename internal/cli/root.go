package cli

import (
	"fmt"
	"os"

	"github.com/jasutiin/envlink/internal/cli/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "envlink",
		Short: "CLI application that keeps track of all your .env files!",
		Long: `envlink is an application that keeps track your projects' .env files. This removes the need to
			share or store your secrets in other apps.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes the root CLI command: registers initConfig to run on startup, defines persistent flags (config, author, license, viper), binds those flags to Viper keys and sets Viper defaults, and attaches exported subcommands from the commands package.
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	viper.SetDefault("license", "apache")

	rootCmd.AddCommand(commands.LoginCmd)
	rootCmd.AddCommand(commands.RegisterCmd)
	rootCmd.AddCommand(commands.PushCmd)
	rootCmd.AddCommand(commands.PullCmd)
	rootCmd.AddCommand(commands.ProjectsCmd)
	rootCmd.AddCommand(commands.StoreCmd)
}

// initConfig initializes Viper configuration for the CLI.
// If cfgFile is set, that file is used; otherwise the function configures Viper
// to look for a YAML config named .cobra in the user's home directory.
// It enables automatic environment variable support and, if a config file is found
// and read successfully, prints the path of the used config file.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}