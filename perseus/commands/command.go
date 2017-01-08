package commands

// Command reflects the interface for every command (like Add, Mirror or Update)
// which will be called by multiple human interfaces (CLI, HTTP, etc.)
type Command interface {
	// Run contains the business logic of the defined command.
	Run() error
}
