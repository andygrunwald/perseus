package config

// Provider represents a configuration provider.
// Valid provider that implement this interface
//
// * viper (https://github.com/spf13/viper)
type Provider interface {
	// Get returns key from the configuration and will not cast it
	Get(key string) interface{}
	// GetString returns key from the configuration as a casted String
	GetString(key string) string
}
