package commands

import (
	"fmt"
	"os"
	"log"
	"net/url"

	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/downloader"
	"github.com/andygrunwald/perseus/packagist"
	"github.com/andygrunwald/perseus/perseus"
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

// downloadResult represents the result of a download
type downloadResult struct {
	Package string
	Error   error
}

// Run is the business logic of AddCommand.
func (c *AddCommand) Run() error {
	c.Log.Printf("Running \"add\" command for package \"%s\"", c.Package)
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

		dependencies := []*perseus.Package{p}
		if c.WithDependencies {
			pUrl := "https://packagist.org"
			c.Log.Printf("Loading dependencies for package \"%s\" from %s", c.Package, pUrl)
			packagistClient, err := packagist.New(pUrl, nil)
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
			c.Log.Printf("%d dependencies found for package \"%s\" on %s", len(dependencies), c.Package, pUrl)
		}

		// Download package incl. dependencies concurrent
		dependencyCount := len(dependencies)
		downloadsChan := make(chan downloadResult, dependencyCount)
		c.startConcurrentDownloads(dependencies, downloadsChan)

		// Check which dependencies where download successful and which not
		c.processFinishedDownloads(downloadsChan, dependencyCount)
		close(downloadsChan)

	} else {
		c.Log.Printf("Mirroring of package \"%s\" from repository \"%s\" started", p.Name, p.Repository)
		// TODO: downloadPackage will write to p (name + Repository url), we should test this with a package that is deprecated.
		// Afaik Packagist will forward you to the new one.
		// Facebook SDK is one of those
		err := c.downloadPackage(p)
		if err != nil {
			return err
		}
		c.Log.Printf("Mirroring of package \"%s\" successful", p.Name)

		// TODO updateSatisConfig(packet)
	}

	// TODO Implement everything and remove this
	c.Log.Println("=============================")
	c.Log.Println("Add command runs successful. Fuck Yeah!")
	c.Log.Println("Important: This command is not complete yet. Write command of Satis configuration is missing.")
	c.Log.Println("=============================")

	return nil
}

func (c *AddCommand) downloadPackage(p *perseus.Package) error {
	repoDir := c.Config.GetString("repodir")
	// TODO Path traversal in p.Name possible?
	targetDir := fmt.Sprintf("%s/%s.git", repoDir, p.Name)

	// Does targetDir already exist?
	if _, err := os.Stat(targetDir); err != nil {
		return err
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		return os.ErrExist
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

func (c *AddCommand) startConcurrentDownloads(dependencies []*perseus.Package, downloadChan chan<- downloadResult) {
	// Loop over all dependencies and download them concurrent
	for _, packet := range dependencies {
		c.Log.Printf("Mirroring of package \"%s\" started", packet.Name)

		go func(singlePacket *perseus.Package, ch chan<- downloadResult) {
			err := c.downloadPackage(singlePacket)
			if err != nil {
				ch <- downloadResult{
					Package: singlePacket.Name,
					Error:   err,
				}
				return
			}

			// Successful result
			ch <- downloadResult{
				Package: singlePacket.Name,
				Error:   nil,
			}
			// TODO updateSatisConfig(packet) per package
		}(packet, downloadChan)
	}
}

func (c *AddCommand) processFinishedDownloads(ch <-chan downloadResult, dependencyCount int) {
	for i:= 0; i < dependencyCount; i++ {
		download := <-ch
		if download.Error == nil {
			c.Log.Printf("Mirroring of package \"%s\" successful", download.Package)
		} else {
			if os.IsExist(download.Error) {
				c.Log.Printf("Package \"%s\" exists on disk. Try updating it instead. Skipping.", download.Package)
			} else {
				c.Log.Printf("Error while mirroring package \"%s\": %s", download.Package, download.Error)
			}
		}
	}
}
