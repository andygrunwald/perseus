package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cfgFile contains the path and name of the configuration file.
var cfgFile string

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "perseus",
	Short: "Local git mirror for your PHP (composer) project dependencies that works together with Satis",
	Long: `perseus is a tool that works together with Satis to create a local git mirror for your PHP (composer) project dependencies.

Every modern PHP project is managed by composer.
To save development time, external packages will be used to focus on your business logic.
Most external packages are downloaded from Packagist, Github or other places every time you hit composer install or update.
To speed up your development workflow, minimize network traffic and being independent from other 3rd party services for building and deploying your apps, a local mirror in your office make sense.

perseus will create a mirror of all your project dependencies and lets you fetch everything from there rather than fetching the whole source from the internet (e.g. Github or Packagist).
Each dependency is entirely mirrored, meaning you'll have all versions, tags, and branches on your local machine or server.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// init kicks of cobra (our CLI interface) and defines all global flags that can be used across all commands.
func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "medusa.json", "Medusa configuration file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("medusa")
	viper.AddConfigPath(".")

	// Prefix env vars
	viper.SetEnvPrefix("PERSEUS")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	// If an error happen, quit.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Configuration file is missing and required.")
		fmt.Println("Please checkout https://github.com/andygrunwald/perseus#configuration for further details.")
		os.Exit(1)
	}
}
