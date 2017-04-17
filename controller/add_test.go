package controller_test

import (
	"testing"

	. "github.com/andygrunwald/perseus/controller"
)

func TestAddController_Run_WithEmptyPackage(t *testing.T) {
	c := &AddController{
		Package: "",
	}

	err := c.Run()
	if err == nil {
		t.Fatal("Expected error while passing an empty package. Got none")
	}
}
