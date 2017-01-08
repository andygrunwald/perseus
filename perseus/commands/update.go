package commands

import (
	"fmt"

	"github.com/spf13/viper"
)

// UpdateCommand reflects the business logic and the Command interface to update all packages that were added or mirrored in the past.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type UpdateCommand struct {
	// Config is the main configuration
	Config *viper.Viper
}

// Run is the business logic of UpdateCommand.
func (c *UpdateCommand) Run() error {
	fmt.Println("Called: func(c *UpdateCommand) Run()")
	panic("Not implemented yet: bin/medusa update [config]")

	// TODO IMPLEMENT UpdateCommand

	return nil
}
