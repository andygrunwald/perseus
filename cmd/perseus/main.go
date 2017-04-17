package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/controller"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cfgFile contains the path and name of the configuration file.
var cfgFile string

// numOfWorkers reflects the number of workers used for concurrent processes
var numOfWorkers int

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

// mirrorCmd represents the "mirror" command for the CLI interface.
var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Mirrors all repositories that are specified in the configuration file",
	Long: `The mirror command reads the given configuration file and mirrors the git repository for each package (including dependencies), so they can be used locally.

Both package lists form the configuration file (repositories and require) will be taken into account.
Dependencies will be only resolved from the packages entered in the require section.
Repositories entered in the repositories section will be mirrors as is without resolving the dependencies.
`,
	Example:   "  perseus mirror",
	ValidArgs: []string{"config"},
	RunE:      cmdMirrorRun,
}

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
	Example:   "  perseus update",
	ValidArgs: []string{"config"},
	RunE:      cmdUpdateRun,
}

// init kicks of cobra (our CLI interface) and defines all global flags that can be used across all commands.
func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "medusa.json", "Medusa configuration file")
	RootCmd.PersistentFlags().IntVar(&numOfWorkers, "numOfWorkers", runtime.GOMAXPROCS(0), "Number of worker used for concurrent operations (e.g. resolving a dependency tree or downloads)")

	// Original medusa command
	// 	medusa add [--with-deps] package [config]
	RootCmd.AddCommand(addCmd)
	addCmd.Flags().Bool("with-deps", false, "If set, the package dependencies will be downloaded, too")

	// Original medusa command
	// 	medusa mirror [config]
	RootCmd.AddCommand(mirrorCmd)

	// Original medusa command
	// 	medusa update [config]
	RootCmd.AddCommand(updateCmd)

	// Cobra is only able to define flags, but no arguments
	// If we were able to define arguments we would implement those:
	//
	// 	new InputArgument('package', InputArgument::REQUIRED, 'The name of a composer package', null),
	// 	new InputArgument('config', InputArgument::OPTIONAL, 'A config file', 'medusa.json')
	//
	// See https://github.com/spf13/cobra/issues/378 for details.
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
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
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Configuration file is missing and required.")
		fmt.Println("Please checkout https://github.com/andygrunwald/perseus#configuration for further details.")
		os.Exit(1)
	}
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
	c := &controller.AddController{
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

// cmdMirrorRun is the CLI interface for the "mirror" command
func cmdMirrorRun(cmd *cobra.Command, args []string) error {
	// Initialize logger
	// At the moment we run pretty standard golang logging to stderr
	// It works for us. We might consider to change this later.
	// If you have good reasons for this, feel free to talk to us.
	l := log.New(os.Stderr, "", log.LstdFlags)

	// Check if we got minimum 1 argument.
	// We will only use the first argument here. The rest will be ignored.
	// First argument is the configuration file, but it is optional.
	// When this is set, we have to overwrite the configuration that viper found before
	if len(args) >= 1 {
		configFileArg := args[0]
		if _, err := os.Stat(configFileArg); os.IsNotExist(err) {
			return fmt.Errorf("Configuration file %s applied as a configuration file, but don't exists", configFileArg)
		}
		viper.SetConfigFile(configFileArg)
	}
	l.Printf("Using configuration file %s", viper.ConfigFileUsed())

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

	l.Println("Running \"mirror\" command")
	// Setup command and run it
	c := &controller.MirrorController{
		Config:      m,
		Log:         l,
		NumOfWorker: nOfWorkers,
	}
	err = c.Run()
	if err != nil {
		return fmt.Errorf("Error during execution of \"mirror\" command: %s\n", err)
	}

	return nil
}

// cmdUpdateRun is the CLI interface for the "update" command
func cmdUpdateRun(cmd *cobra.Command, args []string) error {
	// Initialize logger
	// At the moment we run pretty standard golang logging to stderr
	// It works for us. We might consider to change this later.
	// If you have good reasons for this, feel free to talk to us.
	l := log.New(os.Stderr, "", log.LstdFlags)

	// Check if we got minimum 1 argument.
	// We will only use the first argument here. The rest will be ignored.
	// First argument is the configuration file, but it is optional.
	// When this is set, we have to overwrite the configuration that viper found before
	if len(args) >= 1 {
		configFileArg := args[0]
		if _, err := os.Stat(configFileArg); os.IsNotExist(err) {
			return fmt.Errorf("Configuration file %s applied as a configuration file, but don't exists", configFileArg)
		}
		viper.SetConfigFile(configFileArg)
	}
	l.Printf("Using configuration file %s", viper.ConfigFileUsed())

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

	l.Println("Running \"update\" command")
	// Setup command and run it
	c := &controller.UpdateController{
		Config:      m,
		Log:         l,
		NumOfWorker: nOfWorkers,
	}
	err = c.Run()
	if err != nil {
		return fmt.Errorf("Error during execution of \"update\" command: %s\n", err)
	}

	return nil
}
