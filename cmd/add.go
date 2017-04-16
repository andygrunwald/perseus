package cmd

import (
	"fmt"

	"log"
	"os"

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
  perseus add --with-deps "symfony/console"`,
	ValidArgs: []string{"package", "config"},
	RunE:      cmdAddRun,
}

func init() {
	// Original medusa command
	// 	medusa add [--with-deps] package [config]

	RootCmd.AddCommand(addCmd)

	addCmd.Flags().Bool("with-deps", false, "If set, the package dependencies will be downloaded, too")

	// Cobra is only able to define flags, but no arguments
	// If we were able to define arguments we would implement those:
	//
	// 	new InputArgument('package', InputArgument::REQUIRED, 'The name of a composer package', null),
	// 	new InputArgument('config', InputArgument::OPTIONAL, 'A config file', 'medusa.json')
	//
	// See https://github.com/spf13/cobra/issues/378 for details.
}

// cmdAddRun is the CLI interface for the "add" command
func cmdAddRun(cmd *cobra.Command, args []string) error {
	// Check first argument: package
	if len(args) == 0 {
		return fmt.Errorf("No argument applied. Please apply one argument: package")
	}
	packet := args[0]

	// Initialize logger
	// At the moment we run pretty standard golang logging to stderr
	// It works for us. We might consider to change this later.
	// If you have good reasons for this, feel free to talk to us.
	l := log.New(os.Stderr, "", log.LstdFlags)

	// Check if we got minimum 2 arguments.
	// We will only use the second argument here. The rest will be ignored.
	// Second argument is the configuration file, but it is optional.
	// When this is set, we have to overwrite the configuration that viper found before
	if len(args) >= 2 {
		configFileArg := args[1]
		if _, err := os.Stat(configFileArg); os.IsNotExist(err) {
			return fmt.Errorf("Configuration file %s applied as a configuration file, but don't exists", configFileArg)
		}
		viper.SetConfigFile(configFileArg)
	}
	l.Printf("Using configuration file %s", viper.ConfigFileUsed())

	// Check "with-deps" flag
	withDepsFlag, err := cmd.Flags().GetBool("with-deps")
	if err != nil {
		return fmt.Errorf("Couldn't determine \"with-deps\" flag: %s\n", err)
	}

	// Create viper based configuration provider for Medusa
	p, err := config.NewViperProvider(viper.GetViper())
	if err != nil {
		return fmt.Errorf("Couldn't create a viper configuration provider: %s\n", err)
	}

	m, err := config.NewMedusa(p)
	if err != nil {
		return fmt.Errorf("Couldn't create medusa configuration object: %s\n", err)
	}

	// Determine number of concurrent workers
	nOfWorkers, err := cmd.Flags().GetInt("numOfWorkers")
	if err != nil {
		return fmt.Errorf("Couldn't determine number of concurrent workers. Please control the 'numOfWorkers' flag. Error message: %s\n", err)
	}

	l.Printf("Running \"add\" command for package \"%s\"", packet)
	// Setup command and run it
	c := &commands.AddCommand{
		Package:          packet,
		WithDependencies: withDepsFlag,
		Config:           m,
		Log:              l,
		NumOfWorker:      nOfWorkers,
	}
	err = c.Run()
	if err != nil {
		return fmt.Errorf("Error during execution of \"add\" command: %s\n", err)
	}

	return nil
}
