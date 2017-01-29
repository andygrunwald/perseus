package packagist

import (
	"net/http"
)

// ApiClient is the interface for API actions for packages
// Standard implementation is packagist.org
type ApiClient interface {
	// GetPackage will return information about package name
	GetPackage(name string) (*Package, *http.Response, error)
}
