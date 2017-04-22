package repository

import (
	"net/http"
)

// Client is the interface for actions to talk to a package repository.
// The main need is to get information about a package.
// Typical implementations are Packagist (for PHP) or PyPI (Python)
type Client interface {
	// GetPackageByName returns a package by name
	GetPackageByName(name string) (*PackagistPackage, *http.Response, error)
}
