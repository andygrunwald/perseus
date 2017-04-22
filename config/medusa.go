package config

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/andygrunwald/perseus/dependency"
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
		return nil, errors.New("No configurations provider applied")
	}

	m := &Medusa{
		config: c,
	}
	return m, nil
}

// GetRepositoryURLOfPackage will determine if package p is part of the configuration.
// If p is part and a url is configured and this url is valid, this url will be returned.
// Otherwise an error.
func (m *Medusa) GetRepositoryURLOfPackage(p *dependency.Package) (*url.URL, error) {
	// TODO Is there a better solution? We cast here and cast and cast ...
	// Yep, checkout https://github.com/spf13/viper#getting-values-from-viper
	// Sadly they don't support a []map[string]string which is the "repositories" section (yet).
	// And i don't know if it make sense to implement there.
	// I raised this question here: https://github.com/spf13/cast/issues/36
	// Lets wait for feedback.
	repositoriesSlice, err := m.getRepositories()
	if err != nil {
		return nil, ErrNoRepositories
	}

	for _, repoEntry := range repositoriesSlice {
		repoEntryMap := repoEntry.(map[string]interface{})
		if val, ok := repoEntryMap["name"]; ok {
			if r := val.(string); r == p.Name {
				if v, ok := repoEntryMap["url"]; ok {
					// Check if the url is empty
					if u := v.(string); len(u) > 0 {

						// Sanitize URL
						// Not the best part here, i know
						reg, err := regexp.Compile("^git@github.com:")
						if err != nil {
							return nil, err
						}

						u = reg.ReplaceAllString(u, "git://github.com/")
						return url.Parse(u)
					}

				}
			}
		}
	}

	return nil, fmt.Errorf("No repository url found for package %s", p.Name)
}

// GetNamesOfRepositories returns all Packages from the configuration
// key "repositories". A repository will only be returned when
// it is complete (means a name and an url exists)
func (m *Medusa) GetNamesOfRepositories() ([]*dependency.Package, error) {
	repositoriesSlice, err := m.getRepositories()
	if err != nil {
		return nil, ErrNoRepositories
	}

	r := []*dependency.Package{}
	for _, repoEntry := range repositoriesSlice {
		repoEntryMap := repoEntry.(map[string]interface{})
		if val, ok := repoEntryMap["name"]; ok {
			name := val.(string)

			if v, ok := repoEntryMap["url"]; ok {
				if u := v.(string); len(u) > 0 {
					pack, err := dependency.NewPackage(name, u)
					if err != nil {
						continue
					}
					r = append(r, pack)
				}
			}
		}
	}

	return r, nil
}

func (m *Medusa) getRepositories() ([]interface{}, error) {
	repositories := m.config.Get("repositories")
	if repositories == nil {
		return nil, ErrNoRepositories
	}

	repositoriesSlice := repositories.([]interface{})
	if len(repositoriesSlice) == 0 {
		return nil, ErrNoRepositories
	}

	return repositoriesSlice, nil
}


// GetRequire returns all the configuration key "require"
func (m *Medusa) GetRequire() []string {
	return m.config.GetStringSlice("require")
}

// GetString returns key from the Medusa configuration as a casted String
func (m *Medusa) GetString(key string) string {
	return m.config.GetString(key)
}
