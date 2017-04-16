package perseus

import (
	"errors"
	"net/url"
	"regexp"
)

// Package represents a single package.
// This can be seen as the main unit of perseus.
type Package struct {
	// Name is the name of the package (e.g. "twig/twig" or "symfony/console")
	Name       string
	Repository *url.URL
}

// NewPackage will create a new Package
func NewPackage(name, repository string) (*Package, error) {
	if len(name) == 0 {
		return nil, errors.New("NewPackage failed. Name attribute required. Empty string given.")
	}

	p := &Package{
		Name: name,
	}

	if len(repository) == 0 {
		return p, nil
	}

	reg, err := regexp.Compile("^git@github.com:")
	if err != nil {
		return p, err
	}

	safeURL := reg.ReplaceAllString(repository, "git://github.com/")
	u, err := url.Parse(safeURL)
	if err != nil {
		return p, err
	}

	p.Repository = u

	return p, nil
}
