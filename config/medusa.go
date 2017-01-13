package config

import (
	"errors"
	"fmt"
	"github.com/andygrunwald/perseus/perseus"
	"net/url"
)

// Medusa reflects the original Medusa configuration file.
type Medusa struct {
	// config is the configuration provider object that has read the medusa configuration file
	config Provider
}

// NewMedusa will create a new medusa configuration object.
// If no configuration is given, an error will be returned.
func NewMedusa(c Provider) (*Medusa, error) {
	if c == nil {
		return nil, errors.New("No conifguration provider applied")
	}

	m := &Medusa{
		config: c,
	}
	return m, nil
}

// GetRepositoryURLOfPackage will determine if package p is part of the configuration.
// If p is part and a url is configured and this url is valid, this url will be returned.
// Otherwise an error.
func (m *Medusa) GetRepositoryURLOfPackage(p *perseus.Package) (*url.URL, error) {
	// TODO Is there a better solution? We cast here and cast and cast ...
	// Yep, checkout https://github.com/spf13/viper#getting-values-from-viper
	// Sadly they don't support a []map[string]string which is the "repositories" section (yet).
	// And i don't know if it make sense to implement there.
	// I raised this question here: https://github.com/spf13/cast/issues/36
	// Lets wait for feedback.
	repositories := m.config.Get("repositories")
	if repositories == nil {
		return nil, errors.New("No repositories configured.")
	}

	repositoriesSlice := repositories.([]interface{})
	if len(repositoriesSlice) == 0 {
		return nil, errors.New("No repositories configured.")
	}

	for _, repoEntry := range repositoriesSlice {
		repoEntryMap := repoEntry.(map[string]interface{})
		if val, ok := repoEntryMap["name"]; ok {
			if r := val.(string); r == p.Name {
				if v, ok := repoEntryMap["url"]; ok {
					// Check if the url is empty
					if u := v.(string); len(u) > 0 {
						return url.Parse(u)
					}

				}
			}
		}
	}

	return nil, fmt.Errorf("No repository url found for package %s", p.Name)
}

// GetString returns key from the Medusa configuration as a casted String
func (m *Medusa) GetString(key string) string {
	return m.config.GetString(key)
}
