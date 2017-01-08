package cmd

import (
	"fmt"

	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/perseus/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addCmd represents the "add" command for the CLI interface.
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Mirrors one given package and adds it to Satis",
	Long: `Mirrors one given package and adds it to Satis.

If the package is available in the medusa.json configuration file and contains a URL, the URL from the configuration file will be used.
Otherwise perseus will request the URL from packagist.

When "with-deps" is given, dependencies of the package will be mirrored as well.
Dependencies will be determined through API requests to packagist.org.
`,
	Example: `  perseus add "twig/twig"
  perseus add --width-deps "symfony/console"`,
	// TODO Write a bash completion for the package arg. Checkout https://github.com/spf13/cobra/blob/master/bash_completions.md
	ValidArgs: []string{"package", "config"},
	RunE:      cmdAddRun,
}

func init() {
	// Original medusa command
	// 	medusa add [--with-deps] package [config]

	RootCmd.AddCommand(addCmd)

	addCmd.Flags().Bool("with-deps", false, "If set, the package dependencies will be downloaded, too")

	// Cobra is only able to define flags, but no arguments
	// If we are able to define arguments we would implement those:
	//
	// 	new InputArgument('package', InputArgument::REQUIRED, 'The name of a composer package', null),
	// 	new InputArgument('config', InputArgument::OPTIONAL, 'A config file', 'medusa.json')
	//
	// See https://github.com/spf13/cobra/issues/378 for details.
}

// cmdAddRun is the CLI interface for the "add" comment
func cmdAddRun(cmd *cobra.Command, args []string) error {
	// Check first argument: package
	if len(args) == 0 {
		return fmt.Errorf("No argument applied. Please apply one argument: package")
	}
	packet := args[0]

	// TODO If the second argument is given, it is the configuration file.
	// This file needs to be read in and used for further operations

	// Check "with-deps" flag
	withDepsFlag, err := cmd.Flags().GetBool("with-deps")
	if err != nil {
		return fmt.Errorf("Couldn't determine \"with-deps\" flag: %s\n", err)
	}

	m, err := config.NewMedusa(viper.GetViper())
	if err != nil {
		return fmt.Errorf("Couldn't create medusa configuration object: %s\n", err)
	}

	// Setup command and run it
	c := &commands.AddCommand{
		Package:          packet,
		WithDependencies: withDepsFlag,
		Config:           m,
	}
	err = c.Run()
	if err != nil {
		return fmt.Errorf("Error during execution of \"add\" command: %s\n", err)
	}

	return nil
}
