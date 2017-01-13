package config

// Provider represents a configuration provider.
// Valid provider that implement this interface
//
// * ViperProvider (viper - https://github.com/spf13/viper)
// * JSONProvider
type Provider interface {
	// Get returns key from the configuration and will not cast it
	Get(key string) interface{}
	// GetString returns key from the configuration as a casted String
	GetString(key string) string
	// GetStringSlice returns the value associated with the key as a slice of strings.
	GetStringSlice(key string) []string
	// GetContentMap returns the complete content of the provider data source as a map
	GetContentMap() map[string]interface{}
}
