package commands_test

import (
	. "github.com/andygrunwald/perseus/perseus/commands"
	"testing"
)

func TestAddCommand_Run_WithEmptyPackage(t *testing.T) {
	c := &AddCommand{
		Package: "",
	}

	err := c.Run()
	if err == nil {
		t.Fatal("Expected error while passing an empty package. Got none")
	}
}
