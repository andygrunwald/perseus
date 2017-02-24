package commands

import (
	"fmt"

	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/downloader"
	"log"
	"path/filepath"
)

// UpdateCommand reflects the business logic and the Command interface to update all packages that were added or mirrored in the past.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type UpdateCommand struct {
	// Config is the main medusa configuration
	Config *config.Medusa
	// Log represents a logger to log messages
	Log *log.Logger
}

// Run is the business logic of UpdateCommand.
func (c *UpdateCommand) Run() error {
	repoDir := c.Config.GetString("repodir")

	p := fmt.Sprintf("%s/*/*.git", repoDir)
	matches, err := filepath.Glob(p)
	if err != nil {
		fmt.Errorf("Error while determining folders for updating: %s", err)
	}

	// If no repositories were found, we will exit here
	if len(matches) == 0 {
		c.Log.Printf("No repositories found in %s", p)
		return nil
	}

	// TODO Make me concurrent
	for _, v := range matches {
		updateClient, err := downloader.NewGitUpdater()
		if err != nil {
			return fmt.Errorf("Updater client creation failed for package %s: %s", v, err)
		}
		err = updateClient.Update(v)
		if err != nil {
			c.Log.Printf("Error while updating %s: %s", v, err)
		} else {
			c.Log.Printf("Update of %s successful", v)
		}

	}

	return nil
}
