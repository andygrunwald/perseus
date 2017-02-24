package commands

import (
	"fmt"
	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/packagist"
	"github.com/andygrunwald/perseus/perseus"
	"github.com/andygrunwald/perseus/types"
	"log"
	"strings"
)

// MirrorCommand reflects the business logic and the Command interface to mirror all configured packages.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type MirrorCommand struct {
	// Config is the main medusa configuration
	Config *config.Medusa
	// Log represents a logger to log messages
	Log *log.Logger
	// NumOfWorker is the number of worker used for concurrent actions (like resolving the dependency tree)
	NumOfWorker int
}

// Run is the business logic of MirrorCommand.
func (c *MirrorCommand) Run() error {
	// TODO Make me concurrent

	repoList, err := c.Config.GetNamesOfRepositories()
	if err != nil {
		if config.IsNoRepositories(err) {
			c.Log.Printf("Config: %s", err)
		} else {
			c.Log.Println(err)
		}
	}

	repos := types.NewSet(repoList...)

	// TODO Make me way faster. We do a lot of duplicate work here like
	// init a new packagist client every time, a new dependency resolver every time
	// booting up new channels, etc.
	// Why not applying here the worker principle again?

	pUrl := "https://packagist.org/"
	require := c.Config.GetRequire()
	for _, v := range require {

		c.Log.Printf("Loading dependencies for package \"%s\" from %s", v, pUrl)
		packagistClient, err := packagist.New(pUrl, nil)
		if err != nil {
			return err
		}

		// Lets get a dependency resolver.
		// If we can't bootstrap one, we are lost anyway.
		d, err := perseus.NewDependencyResolver(v, c.NumOfWorker, packagistClient)
		if err != nil {
			return err
		}
		results := d.GetResultStream()
		go d.Start()

		dependencyNames := []string{}
		// Finally we collect all the results of the work.
		for r := range results {
			if r.Package.Name == v {
				continue
			}
			repos.Add(r.Package.Name)
			dependencyNames = append(dependencyNames, r.Package.Name)
		}

		if l := len(dependencyNames); l == 0 {
			c.Log.Printf("%d dependencies found for package \"%s\" on %s", len(dependencyNames), v, pUrl)
		} else {
			c.Log.Printf("%d dependencies found for package \"%s\" on %s: %s", len(dependencyNames), v, pUrl, strings.Join(dependencyNames, ", "))
		}
	}

	// Current results

	// 2017/02/24 20:58:17 Using configuration file /.../Go/src/github.com/andygrunwald/perseus/medusa.json
	// 2017/02/24 20:58:45 11 dependencies found for pa....
	// Found 163 entries: [psr/container jms/di-extra-bundle react/promise
	// ./pers mirror  6.84s user 0.96s system 28% cpu 27.600 total

	fmt.Printf("Found %d entries: %+v\n", repos.Len(), repos.Flatten())

	fmt.Println("Called: func(c *MirrorCommand) Run()")
	panic("Not implemented yet: bin/medusa mirror [config]")

	// TODO IMPLEMENT MirrorCommand

	return nil
}
