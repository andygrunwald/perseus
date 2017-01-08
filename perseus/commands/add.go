package commands

import (
	"fmt"
	"os"

	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/downloader"
	"github.com/andygrunwald/perseus/packagist"
	"github.com/andygrunwald/perseus/perseus"
	"log"
	"net/url"
)

// AddCommand reflects the business logic and the Command interface to add a new package.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type AddCommand struct {
	// WithDependencies decides if the dependencies of an external package needs to be mirrored as well
	WithDependencies bool
	// Package is the package to mirror
	Package string
	// Config is the main medusa configuration
	Config *config.Medusa
	// Log represents a logger to log messages
	Log *log.Logger
}

// Run is the business logic of AddCommand.
func (c *AddCommand) Run() error {
	c.Log.Printf("Running Add command for package %s", c.Package)
	p, err := perseus.NewPackage(c.Package)
	if err != nil {
		return err
	}

	// We don't respect the error here.
	// OH: "WTF? Why? You claim 'Serious error handling' in the README!"
	// Yep, you are right. And we still do.
	// In this case, it is okay, if p is not configured or no repositories are configured at all.
	// When this happen, we will ask Packagist fot the repository url.
	// If this package is not available on packagist, this will be shift to an error.
	p.Repository, _ = c.Config.GetRepositoryURLOfPackage(p)
	if p.Repository == nil {

		dependencies := []string{p.Name}
		if c.WithDependencies {
			fmt.Println("with deps")
			packagistClient, err := packagist.New("https://packagist.org", nil)
			if err != nil {
				return err
			}

			// TODO Okay, here we don't take error handling serious.
			//	Why? Easy. If an API request fails, we don't know it.
			//	Why? Easy. Which packages will be skipped? e.g. "php" ?
			//	We really have to refactor this. Checkout the articles / links
			//	That are mentioned IN the depdency resolver comments
			//	But you know. 1. Make it work. 2. Make it fast. 3. Make it beautiful
			// 	And this works for now.
			d := perseus.NewDependencyResolver(p.Name, packagistClient)
			dependencies = d.Resolve()
		}

		for _, packet := range dependencies {
			// TODO: Add some kind of logging here: fmt.Printf(" - Mirroring <info>%s</info>\n", singlePacket)

			// TODO: Make this concurrent
			packetEntity, err := perseus.NewPackage(packet)
			if err != nil {
				return err
			}

			err = c.downloadPackage(packetEntity)
			if err != nil {
				return err
			}

			// TODO updateSatisConfig(packet) per package
		}

	} else {
		// TODO: Add some kind of logging here: fmt.Printf(" - Mirroring <info>%s</info>\n", packet)
		// TODO: downloadPackage will write to p (name + Repository url), we should test this with a package that is deprecated.
		// Afaik Packagist will forward you to the new one.
		// Facebook SDK is one of those
		err := c.downloadPackage(p)
		if err != nil {
			return err
		}

		// TODO updateSatisConfig(packet)
	}

	// TODO Implement everything and remove this
	fmt.Println("=============================")
	fmt.Println("Add command runs successful. Fuck Yeah!")
	fmt.Println("Important: This command is not complete yet. Write command of Satis configuration is missing.")
	fmt.Println("=============================")

	return nil
}

func (c *AddCommand) downloadPackage(p *perseus.Package) error {
	repoDir := c.Config.GetString("repodir")
	// TODO Path traversal in p.Name possible?
	targetDir := fmt.Sprintf("%s/%s.git", repoDir, p.Name)

	// Does targetDir already exist?
	if _, err := os.Stat(targetDir); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("The repository %s already exists in %s. Try updating it instead.", p.Name, targetDir)
		}
	}

	if p.Repository == nil {
		packagistClient, err := packagist.New("https://packagist.org/", nil)
		if err != nil {
			return fmt.Errorf("Packagist client creation failed: %s", err)
		}
		packagistPackage, resp, err := packagistClient.GetPackage(p.Name)
		if err != nil {
			return fmt.Errorf("Failed to retrieve information about package \"%s\" from Packagist. Called %s. Error: %s", p.Name, resp.Request.URL.String(), err)
		}

		// Check if URL is empty
		if len(packagistPackage.Repository) == 0 {
			// TODO What happens if Packagist rewrite the package? E.g. the facebook example? We should output here both names
			return fmt.Errorf("Received empty URL for package %s from Packagist", p.Name)
		}

		// Overwriting values from Packagist
		p.Name = packagistPackage.Name
		u, err := url.Parse(packagistPackage.Repository)
		if err != nil {
			return fmt.Errorf("URL conversion of %s to a net/url.URL object failed: %s", packagistPackage.Repository, err)
		}
		p.Repository = u
	}

	downloadClient, err := downloader.NewGit(p.Repository.String())
	if err != nil {
		return fmt.Errorf("Downloader client creation failed for package %s: %s", p.Name, err)
	}
	return downloadClient.Download(targetDir)
}
