package cmd

import (
	"fmt"

	"github.com/andygrunwald/perseus/perseus/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	// Setup command and run it
	c := &commands.UpdateCommand{
		Config: viper.GetViper(),
	}
	err := c.Run()
	if err != nil {
		return fmt.Errorf("Error during execution of \"update\" command: %s\n", err)
	}

	return nil
}
