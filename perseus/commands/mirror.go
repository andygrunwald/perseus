package commands

import (
	"fmt"
	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/packagist"
	"github.com/andygrunwald/perseus/perseus"
	"log"
)

// MirrorCommand reflects the business logic and the Command interface to mirror all configured packages.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type MirrorCommand struct {
	// Config is the main medusa configuration
	Config *config.Medusa
	// Log represents a logger to log messages
	Log *log.Logger
}

// Run is the business logic of MirrorCommand.
func (c *MirrorCommand) Run() error {
	// TODO Make me concurrent

	repos, err := c.Config.GetNamesOfRepositories()
	if err != nil {
		// TODO Define own error type and handle it here
		c.Log.Println(err)
	}

	pUrl := "https://packagist.org/"
	require := c.Config.GetRequire()
	for _, v := range require {

		c.Log.Printf("Loading dependencies for package \"%s\" from %s", v, pUrl)

		packagistClient, err := packagist.New(pUrl, nil)
		if err != nil {
			return err
		}

		// TODO Okay, here we don't take error handling serious.
		//	Why? Easy. If an API request fails, we don't know it.
		//	Why? Easy. Which packages will be skipped? e.g. "php" ?
		//	We really have to refactor this. Checkout the articles / links
		//	That are mentioned IN the dependency resolver comments
		//	But you know. 1. Make it work. 2. Make it fast. 3. Make it beautiful
		// 	And this works for now.
		// TODO Make number of worker configurable
		d := perseus.NewDependencyResolver(v, 3, packagistClient)
		results := d.GetResults()
		go d.Start()

		dependencies := []string{}
		// Finally we collect all the results of the work.
		for v := range results {
			dependencies = append(dependencies, v.Package.Name)
		}

		// TODO List all deps here instead of the number
		c.Log.Printf("%d dependencies found for package \"%s\" on %s", len(dependencies), v, pUrl)

		for _, p := range dependencies {
			repos = append(repos, p)
		}
	}

	fmt.Printf("%+v\n", repos)

	fmt.Println("Called: func(c *MirrorCommand) Run()")
	panic("Not implemented yet: bin/medusa mirror [config]")

	// TODO IMPLEMENT MirrorCommand

	return nil
}
