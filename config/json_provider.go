package config

import (
	"encoding/json"
)

// JSONProvider provides the data structure for a configuration
// file that is defined in JSON
type JSONProvider struct {
	content map[string]*json.RawMessage
}

// NewJSONProvider will create a new provider to work
// with JSON content c in a convenient way
func NewJSONProvider(c []byte) (*JSONProvider, error) {
	b := make(map[string]*json.RawMessage)
	err := json.Unmarshal(c, &b)
	if err != nil {
		return nil, err
	}

	j := &JSONProvider{
		content: b,
	}

	return j, nil
}

// Get returns key from the configuration and will not cast it
func (p *JSONProvider) Get(key string) interface{} {
	return interface{}(p.content[key])
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (p *JSONProvider) GetStringSlice(key string) []string {
	var l []string
	if v, ok := p.content[key]; ok {
		json.Unmarshal(*v, &l)
	}

	return l
}

// GetContentMap returns the complete content of the provider data source as a map
func (p *JSONProvider) GetContentMap() map[string]interface{} {
	m := make(map[string]interface{}, len(p.content))
	for k, v := range p.content {
		m[k] = interface{}(*v)
	}
	return m
}

// GetString returns key from the configuration as a casted String
func (p *JSONProvider) GetString(key string) string {
	var s string

	// We check if the key exists.
	// If we wouldn't do it and it would fail with SIGSEGV (invalid memory access).
	// Instead we just return an empty string.
	if v, ok := p.content[key]; ok {
		json.Unmarshal(*v, &s)
	}

	return s
}
