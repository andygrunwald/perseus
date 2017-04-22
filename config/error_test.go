package config_test

import (
	"errors"
	"testing"

	. "github.com/andygrunwald/perseus/config"
)

func TestIsNoRepositories(t *testing.T) {
	tests := []struct {
		err    error
		result bool
	}{
		{ErrNoRepositories, true},
		{errors.New("Dummy error"), false},
	}

	for _, tt := range tests {
		if res := IsNoRepositories(tt.err); res != tt.result {
			t.Errorf("Expected IsNoRepositories(%+v) to be %+v. Got %+v.", tt.err, tt.result, res)
		}
	}
}
