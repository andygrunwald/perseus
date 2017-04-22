package config_test

import (
	"testing"

	. "github.com/andygrunwald/perseus/config"
)

func TestNewSatis_NoProvider(t *testing.T) {
	_, err := NewSatis(nil)
	if err == nil {
		t.Error("Expected an error. Got none.")
	}
}