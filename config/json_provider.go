package config

import (
	"encoding/json"
)

type JSONProvider struct {
	content map[string]*json.RawMessage
}

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
func (j *JSONProvider) Get(key string) interface{} {
	return interface{}(j.content[key])
}

// GetContentMap returns the complete content of the provider data source as a map
func (j *JSONProvider) GetContentMap() map[string]interface{} {
	m := make(map[string]interface{}, len(j.content))
	for k, v := range j.content {
		m[k] = interface{}(*v)
	}
	return m
}

// GetString returns key from the configuration as a casted String
func (j *JSONProvider) GetString(key string) string {
	var s string

	// Yep. We skip error handling here. Dirty? Yep.
	// Alternative? Break the interface.
	// In an error case we will return an empty string here.
	// TODO extend GetString with an error
	json.Unmarshal(*j.content[key], &s)
	return s
}
