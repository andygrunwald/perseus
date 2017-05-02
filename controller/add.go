package controller

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/dependency"
	"github.com/andygrunwald/perseus/dependency/repository"
	"github.com/andygrunwald/perseus/downloader"
)

// AddController reflects the business logic and the Command interface to add a new package.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type AddController struct {
	// WithDependencies decides if the dependencies of an external package needs to be mirrored as well
	WithDependencies bool
	// Package is the package to mirror
	Package string
	// Config is the main medusa configuration
	Config *config.Medusa
	// Log represents a logger to log messages
	Log logrus.FieldLogger
	// NumOfWorker is the number of worker used for concurrent actions (like resolving the dependency tree)
	NumOfWorker int
}

// downloadResult represents the result of a download
type downloadResult struct {
	Package string
	Error   error
}

// Run is the business logic of AddCommand.
func (c *AddController) Run() error {
	p, err := dependency.NewPackage(c.Package, "")
	if err != nil {
		return err
	}

	var satisRepositories []string
	downloadablePackages := []*dependency.Package{}

	// We don't respect the error here.
	// OH: "WTF? Why? You claim 'Serious error handling' in the README!"
	// Yep, you are right. And we still do.
	// In this case, it is okay, if p is not configured or no repositories are configured at all.
	// When this happen, we will ask Packagist fot the repository url.
	// If this package is not available on packagist, this will be shift to an error.
	p.Repository, _ = c.Config.GetRepositoryURLOfPackage(p)
	if p.Repository == nil {

		// Check if we should load the dependency also
		if c.WithDependencies {
			pUrl := "https://packagist.org/"
			c.Log.WithFields(logrus.Fields{
				"package": c.Package,
				"source":  pUrl,
			}).Info("Loading dependencies")

			packagistClient, err := repository.NewPackagist(pUrl, nil)
			if err != nil {
				return err
			}

			// Lets get a dependency resolver.
			// If we can't bootstrap one, we are lost anyway.
			// We set the queue length to the number of workers + 1. Why?
			// With this every worker has work, when the queue is filled.
			// During the add command, this is enough in most of the cases.
			d, err := dependency.NewComposerResolver(c.NumOfWorker, packagistClient)
			if err != nil {
				return err
			}
			results := d.GetResultStream()
			go d.Resolve([]*dependency.Package{p})

			dependencyNames := []string{}
			// Finally we collect all the results of the work.
			for v := range results {
				downloadablePackages = append(downloadablePackages, v.Package)
				dependencyNames = append(dependencyNames, v.Package.Name)
			}

			if l := len(dependencyNames); l == 0 {
				c.Log.WithFields(logrus.Fields{
					"amount":  l,
					"package": c.Package,
					"source":  pUrl,
				}).Info("No dependencies found")
			} else {
				c.Log.WithFields(logrus.Fields{
					"amount":       l,
					"package":      c.Package,
					"source":       pUrl,
					"dependencies": strings.Join(dependencyNames, ", "),
				}).Info("Dependencies found")
			}

		} else {
			// It seems to be that we don't have an URL for the package
			// Lets ask packagist for it
			p, err = c.getURLOfPackageFromPackagist(p)
			if err != nil {
				return err
			}
			downloadablePackages = append(downloadablePackages, p)
		}

	} else {
		c.Log.WithFields(logrus.Fields{
			"package":    p.Name,
			"repository": p.Repository,
		}).Info("Mirroring started")
		downloadablePackages = append(downloadablePackages, p)
	}

	// Okay, we have everything done here.
	// Resolved the dependencies (or not) and collected the packages.
	// I would say we can start with downloading them ....
	// Why we are talking? Lets do it!
	c.Log.WithFields(logrus.Fields{
		"amountPackages": len(downloadablePackages),
		"amountWorker":   c.NumOfWorker,
	}).Info("Start concurrent download process")
	d, err := downloader.NewGit(c.NumOfWorker, c.Config.GetString("repodir"))
	if err != nil {
		return err
	}

	results := d.GetResultStream()
	d.Download(downloadablePackages)

	for i := 1; i <= len(downloadablePackages); i++ {
		v := <-results
		if v.Error != nil {
			if os.IsExist(v.Error) {
				c.Log.WithFields(logrus.Fields{
					"package": v.Package.Name,
				}).Info("Package exists on disk. Try updating it instead. Skipping.")
			} else {
				c.Log.WithFields(logrus.Fields{
					"package": v.Package.Name,
				}).WithError(v.Error).Info("Error while mirroring package")
				// If we have an error, we don't need to add it to satis repositories
				continue
			}
		} else {
			c.Log.WithFields(logrus.Fields{
				"package": v.Package.Name,
			}).Info("Mirroring of package successful")
		}

		satisRepositories = append(satisRepositories, c.getLocalUrlForRepository(v.Package.Name))
	}
	d.Close()

	// And as a final step, write the satis configuration
	err = c.writeSatisConfig(satisRepositories...)
	return err
}

func (c *AddController) writeSatisConfig(satisRepositories ...string) error {
	// Write Satis file
	satisConfig := c.Config.GetString("satisconfig")
	if len(satisConfig) == 0 {
		c.Log.Info("No Satis configuration specified. Skipping to write a satis configuration.")
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

	c.Log.WithFields(logrus.Fields{
		"path": satisConfig,
	}).Info("Satis configuration successful written")
	return nil
}

func (c *AddController) getLocalUrlForRepository(p string) string {
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

func (c *AddController) getURLOfPackageFromPackagist(p *dependency.Package) (*dependency.Package, error) {
	packagistClient, err := repository.NewPackagist("https://packagist.org/", nil)
	if err != nil {
		return p, fmt.Errorf("Packagist client creation failed: %s", err)
	}

	packagistPackage, resp, err := packagistClient.GetPackageByName(p.Name)
	if err != nil {
		return p, fmt.Errorf("Failed to retrieve information about package \"%s\" from Packagist. Called %s. Error: %s", p.Name, resp.Request.URL.String(), err)
	}

	// Check if URL is empty
	if len(packagistPackage.Repository) == 0 {
		return p, fmt.Errorf("Received empty URL for package %s from Packagist", p.Name)
	}

	// Overwriting values from Packagist
	p.Name = packagistPackage.Name
	u, err := url.Parse(packagistPackage.Repository)
	if err != nil {
		return p, fmt.Errorf("URL conversion of %s to a net/url.URL object failed: %s", packagistPackage.Repository, err)
	}
	p.Repository = u

	return p, nil
}
