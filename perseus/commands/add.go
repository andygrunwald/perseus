package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
	p, err := perseus.NewPackage(c.Package)
	if err != nil {
		return err
	}

	var satisRepositories []string

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
			pUrl := "https://packagist.org/"
			c.Log.Printf("Loading dependencies for package \"%s\" from %s", c.Package, pUrl)
			packagistClient, err := packagist.New(pUrl, nil)
			if err != nil {
				return err
			}

			// Lets get a dependency resolver.
			// If we can't bootstrap one, we are lost anyway.
			// TODO Make number of worker configurable
			d, err := perseus.NewDependencyResolver(p.Name, 3, packagistClient)
			if err != nil {
				return err
			}
			results := d.GetResultStream()
			go d.Start()

			dependencies := []string{}
			// Finally we collect all the results of the work.
			for v := range results {
				dependencies = append(dependencies, v.Package.Name)
			}

			c.Log.Printf("%d dependencies found for package \"%s\" on %s: %s", len(dependencies), c.Package, pUrl, strings.Join(dependencies, ", "))
		}

		// Download package incl. dependencies concurrent
		dependencyCount := len(dependencies)
		downloadsChan := make(chan downloadResult, dependencyCount)
		defer close(downloadsChan)
		c.startConcurrentDownloads(dependencies, downloadsChan)

		// Check which dependencies where download successful and which not
		satisRepositories = c.processFinishedDownloads(downloadsChan, dependencyCount)

	} else {
		c.Log.Printf("Mirroring of package \"%s\" from repository \"%s\" started", p.Name, p.Repository)
		// TODO: downloadPackage will write to p (name + Repository url), we should test this with a package that is deprecated.
		// Afaik Packagist will forward you to the new one.
		// Facebook SDK is one of those
		err := c.downloadPackage(p)
		if err != nil {
			if os.IsExist(err) {
				c.Log.Printf("Package \"%s\" exists on disk. Try updating it instead. Skipping.", p.Name)
			} else {
				return err
			}
		} else {
			c.Log.Printf("Mirroring of package \"%s\" successful", p.Name)
		}

		satisRepositories = append(satisRepositories, c.getLocalUrlForRepository(p.Name))
	}

	// Write Satis file
	satisConfig := c.Config.GetString("satisconfig")
	if len(satisConfig) == 0 {
		c.Log.Print("No Satis configuration specified. Skipping to write a satis configuration.")
		return nil
	}

	satisContent, err := ioutil.ReadFile(satisConfig)
	if err != nil {
		return fmt.Errorf("Can't read Satis configuration %s: %s", satisConfig, err)
	}

	j, err := config.NewJSONProvider(satisContent)
	if err != nil {
		return fmt.Errorf("Error while creating JSONProvider: %s", err)
	}

	s, err := config.NewSatis(j)
	if err != nil {
		return fmt.Errorf("Error while creating Satis object: %s", err)
	}

	s.AddRepositories(satisRepositories...)
	err = s.WriteFile(satisConfig, 0644)
	if err != nil {
		return fmt.Errorf("Writing Satis configuration to %s failed: %s", satisConfig, err)
	}

	c.Log.Printf("Satis configuration successful written to %s", satisConfig)

	return nil
}

func (c *AddCommand) getLocalUrlForRepository(p string) string {
	var r string

	satisURL := c.Config.GetString("satisurl")
	repoDir := c.Config.GetString("repodir")

	if len(satisURL) > 0 {
		r = fmt.Sprintf("%s/%s.git", satisURL, p)
	} else {
		t := fmt.Sprintf("%s/%s.git", repoDir, p)
		t = strings.TrimLeft(filepath.Clean(t), "/")
		r = fmt.Sprintf("file:///%s", t)
	}

	return r
}

func (c *AddCommand) downloadPackage(p *perseus.Package) error {
	repoDir := c.Config.GetString("repodir")
	// TODO Path traversal in p.Name possible?
	targetDir := fmt.Sprintf("%s/%s.git", repoDir, p.Name)

	// Does targetDir already exist?
	if _, err := os.Stat(targetDir); err != nil {
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
			ch <- downloadResult{
				Package: singlePacket.Name,
				Error:   err,
			}
		}(packet, downloadChan)
	}
}

func (c *AddCommand) processFinishedDownloads(ch <-chan downloadResult, dependencyCount int) []string {
	var success []string
	for i := 0; i < dependencyCount; i++ {
		download := <-ch
		if download.Error == nil {
			c.Log.Printf("Mirroring of package \"%s\" successful", download.Package)
			success = append(success, c.getLocalUrlForRepository(download.Package))
		} else {
			if os.IsExist(download.Error) {
				c.Log.Printf("Package \"%s\" exists on disk. Try updating it instead. Skipping.", download.Package)
			} else {
				c.Log.Printf("Error while mirroring package \"%s\": %s", download.Package, download.Error)
			}
		}
	}

	return success
}
