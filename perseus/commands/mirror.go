package commands

import (
	"fmt"

	"github.com/spf13/viper"
)

// MirrorCommand reflects the business logic and the Command interface to mirror all configured packages.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type MirrorCommand struct {
	// Config is the main configuration
	Config *viper.Viper
}

// Run is the business logic of MirrorCommand.
func (c *MirrorCommand) Run() error {
	fmt.Println("Called: func(c *MirrorCommand) Run()")
	panic("Not implemented yet: bin/medusa mirror [config]")

	// TODO IMPLEMENT MirrorCommand

	return nil
}
