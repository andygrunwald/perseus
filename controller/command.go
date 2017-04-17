package controller

// Controller reflects the interface for every controller (like Add, Mirror or Update)
// which will be called by multiple human interfaces (CLI, HTTP, etc.)
type Controller interface {
	// Run contains the business logic of the defined command.
	Run() error
}
