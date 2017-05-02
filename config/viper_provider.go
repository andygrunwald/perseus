package config

import (
	"github.com/spf13/viper"
)

// ViperProvider represents the structure to manage
// configurations based on github.com/spf13/viper
type ViperProvider struct {
	v *viper.Viper
}

// NewViperProvider creates a new configuration provider
// based on github.com/spf13/viper
func NewViperProvider(v *viper.Viper) (*ViperProvider, error) {
	p := &ViperProvider{
		v: v,
	}

	return p, nil
}

// Get returns key from the configuration and will not cast it
func (p *ViperProvider) Get(key string) interface{} {
	return p.v.Get(key)
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (p *ViperProvider) GetStringSlice(key string) []string {
	return p.v.GetStringSlice(key)
}

// GetContentMap returns the complete content of the provider data source as a map
func (p *ViperProvider) GetContentMap() map[string]interface{} {
	// TODO Implement
	return map[string]interface{}{}
}

// GetString returns key from the configuration as a casted String
func (p *ViperProvider) GetString(key string) string {
	return p.v.GetString(key)
}
