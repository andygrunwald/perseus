package config

import "errors"

var (
	// ErrNoRepositories reflects an own error dedicated to the situation
	// that there are no repositories configured / defined.
	ErrNoRepositories = errors.New("No repositories defined/configured.")
)

// IsNoRepositories returns a boolean indicating whether the error is known to report
// that there are no repositories.
func IsNoRepositories(err error) bool {
	return err == ErrNoRepositories
}
