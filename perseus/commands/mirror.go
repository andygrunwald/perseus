package commands

import (
	"fmt"
	"github.com/andygrunwald/perseus/config"
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
	fmt.Println("Called: func(c *MirrorCommand) Run()")
	panic("Not implemented yet: bin/medusa mirror [config]")

	// TODO IMPLEMENT MirrorCommand

	return nil
}
