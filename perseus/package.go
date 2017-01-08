package perseus

import "errors"

// Package represents a single package.
// This can be seen as the main unit of perseus.
type Package struct {
	// Name is the name of the package (e.g. "twig/twig" or "symfony/console")
	Name string
}

// NewPackage will create a new Package
func NewPackage(name string) (*Package, error) {
	if len(name) == 0 {
		return nil, errors.New("NewPackage failed. Name attribute required. Empty string given.")
	}

	p := &Package{
		Name: name,
	}

	return p, nil
}
