package commands

import (
	"fmt"

	"github.com/spf13/viper"
)

// AddCommand reflects the business logic and the Command interface to add a new package.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type AddCommand struct {
	// WithDependencies decides if the dependencies of an external package needs to be mirrored as well
	WithDependencies bool
	// Package is the package to mirror
	Package string
	// Config is the main configuration
	Config *viper.Viper
}

// Run is the business logic of AddCommand.
func (c *AddCommand) Run() error {
	fmt.Println("Called: func(c *AddCommand) Run()")
	panic("Not implemented yet: bin/medusa add [--with-deps] package [config]")

	// TODO IMPLEMENT AddCommand

	return nil
	/*
		// We need a package as a first argument of the add command.
		if len(args) == 0 {
			fmt.Println("No argument applied. Please apply one argument: package")
			cmd.Usage()
			return
		}

		packet := args[0]
		var err error
		if repositoryUrl := getRepositoryUrlFromConfig(packet); len(repositoryUrl) == 0 {
			withDeps, err := cmd.Flags().GetBool("with-deps")
			if err != nil {
				panic(err)
			}
			err = mirrorPackagistAndRepositories(withDeps, packet)
		} else {
			err = mirrorRepositoryOnly(packet, repositoryUrl)
		}

		if err != nil {
			panic(err)
		}
		fmt.Println("add called =====")
	*/
}

/*

func getGitRepo(packet string, repositoryUrl string) error {
	outputDir := viper.GetString("repodir")
	dir := fmt.Sprintf("%s/%s.git", outputDir, packet)

	if _, err := os.Stat(dir); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("The repository %s already exists. Try updating it instead.", packet)
		}
	}

	if len(repositoryUrl) == 0 {
		packagistClient, err := packagist.New("https://packagist.org/", nil)
		if err != nil {
			return fmt.Errorf("Packagist client creation failed: %s", err)
		}
		packagistPackage, _, err := packagistClient.GetPackage(packet)
		if err != nil {
			return fmt.Errorf("Failed to retrieve information about package %s from Packagist: %s", packet, err)
		}
		// Overwriting values from packagist
		packet = packagistPackage.Name
		repositoryUrl = packagistPackage.Repository
	}

	downloadClient, err := downloader.NewGit(packet, repositoryUrl)
	if err != nil {
		return fmt.Errorf("Downloader client creation failed for package %s: %s", packet, err)
	}
	return downloadClient.Download(outputDir)
}

func mirrorRepositoryOnly(packet string, repositoryUrl string) (error) {
	fmt.Printf(" - Mirroring <info>%s</info>\n", packet)
	err := getGitRepo(packet, repositoryUrl)
	if err != nil {
		return err
	}

	return updateSatisConfig(packet)
}

func mirrorPackagistAndRepositories(withDeps bool, packet string) (error) {
	deps := []string{packet}

	if withDeps {
		p, err := packagist.New("https://packagist.org", nil)
		if err != nil {
			return err
		}

		d := perseus.NewDependencyResolver(packet, p)
		deps = d.Resolve()
	}

	for _, singlePacket := range deps {
		fmt.Printf(" - Mirroring <info>%s</info>\n", singlePacket)
		err := getGitRepo(singlePacket, "")
		if err != nil {
			return err
		}

		return updateSatisConfig(packet)
	}

	return nil
}

func getRepositoryUrlFromConfig(repo string) string {
	// TODO Is there a better solution? We cast here and cast and cast ...
	// Yep, checkout https://github.com/spf13/viper#getting-values-from-viper
	repositories := viper.Get("repositories")

	repositoriesSlice := repositories.([]interface{})
	if (len(repositoriesSlice) == 0) {
		return ""
	}

	for _, repoEntry := range repositoriesSlice {
		repoEntryMap := repoEntry.(map[string]interface{})
		if val, ok := repoEntryMap["name"]; !ok {
			if val.(string) == repo {
				// TODO: Check if key "url" exists
				return repoEntryMap["url"].(string)
			}
		}
	}

	return ""
}

func updateSatisConfig(pack string) (error) {
	panic("updateSatisConfig is not implemented yet. Do it lazy boy!")

	/*
	satisConfig := viper.GetString("satisconfig")
	satisUrl := viper.GetString("satisurl")

	if !len(satisConfig) {
		return fmt.Errorf("No satisconfig set in your medusa.json configuration file.")
	}
*/

/*
	PHP Code of the original implementation

	protected function updateSatisConfig($package)
	    {
		$satisConfig = $this->config->satisconfig;
		$satisUrl = $this->config->satisurl;

		if ($satisConfig) {
		    $file = new JsonFile($satisConfig);
		    $config = $file->read();

		    if ($satisUrl) {
			$url = $package.'.git';
			$repo = array(
			    'type' => 'git',
			    'url' => $satisUrl . '/' . $url
			);
		    } else {
			$url = ltrim(realpath($this->config->repodir.'/'.$package.'.git'), '/');
			$repo = array(
			    'type' => 'git',
			    'url' => 'file:///' . $url
			);
		    }

		    $config['repositories'][] = $repo;
		    $config['repositories'] = $this->deduplicate($config['repositories']);
		    $file->write($config);
		}
	    }

}
*/
