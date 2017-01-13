package cmd

import (
	"fmt"

	"github.com/andygrunwald/perseus/perseus/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/andygrunwald/perseus/config"
	"log"
	"os"
)

// updateCmd represents the "update" command for the CLI interface.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Fetch latest updates for each mirrored package",
	Long: `The update command reads the given configuration file file and updates each mirrored package.

This command will not reflect the configured repositories from the configuration.
It will only reflect the packages that were already mirrored in the past.

If you add a new package to the configuration you need either call the "add" command with the package as an argument.
Or you add the new package to the configuration and call the "mirror" command.

The update command is useful to ensure that every branch, tag or change in the configured packages is mirrors downstream.
Otherwise you would stuck with the version from the time you added the package.`,
	Example: "  perseus update",
	// TODO Write a bash completion for the package arg. Checkout https://github.com/spf13/cobra/blob/master/bash_completions.md
	ValidArgs: []string{"config"},
	RunE:      cmdUpdateRun,
}

func init() {
	// Original medusa command
	// 	medusa update [config]

	RootCmd.AddCommand(updateCmd)

	// Cobra is only able to define flags, but no arguments
	// If we are able to define arguments we would implement those:
	//
	// 	new InputArgument('config', InputArgument::OPTIONAL, 'A config file', 'medusa.json')
	//
	// See https://github.com/spf13/cobra/issues/378 for details.
}

// cmdUpdateRun is the CLI interface for the "mirror" comment
func cmdUpdateRun(cmd *cobra.Command, args []string) error {
	// TODO If the first argument is given, it is the configuration file.
	// This file needs to be read in and used for further operations

	// Create viper based configuration provider for Medusa
	p, err := config.NewViperProvider(viper.GetViper())
	if err != nil {
		return fmt.Errorf("Couldn't create a viper configuration provider: %s\n", err)
	}

	m, err := config.NewMedusa(p)
	if err != nil {
		return fmt.Errorf("Couldn't create medusa configuration object: %s\n", err)
	}

	// Initialize logger
	// At the moment we run pretty standard golang logging to stderr
	// It works for us. We might consider to change this later.
	// If you have good reasons for this, feel free to talk to us.
	l := log.New(os.Stderr, "", log.LstdFlags)

	l.Println("Running \"update\" command")
	// Setup command and run it
	c := &commands.UpdateCommand{
		Config:           m,
		Log:              l,
	}
	err = c.Run()
	if err != nil {
		return fmt.Errorf("Error during execution of \"update\" command: %s\n", err)
	}

	return nil
}
