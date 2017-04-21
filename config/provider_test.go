package config_test

// EmptyUnitTestProvider represents an 100% empty Provider implementation for unit testing.
type EmptyUnitTestProvider struct{}

func (p *EmptyUnitTestProvider) Get(key string) interface{} {
	return nil
}

func (p *EmptyUnitTestProvider) GetString(key string) string {
	return ""
}

func (p *EmptyUnitTestProvider) GetStringSlice(key string) []string {
	return []string{}
}

func (p *EmptyUnitTestProvider) GetContentMap() map[string]interface{} {
	return map[string]interface{}{}
}

// EmptyWithKeysUnitTestProvider represents an 100% empty Provider implementation for unit testing.
type EmptyWithKeysUnitTestProvider struct{}

func (p *EmptyWithKeysUnitTestProvider) Get(key string) interface{} {
	var m interface{}
	if key == "repositories" {
		m = []interface{}{}
	}

	return m
}

func (p *EmptyWithKeysUnitTestProvider) GetString(key string) string {
	return ""
}

func (p *EmptyWithKeysUnitTestProvider) GetStringSlice(key string) []string {
	return []string{}
}

func (p *EmptyWithKeysUnitTestProvider) GetContentMap() map[string]interface{} {
	return map[string]interface{}{}
}

// MedusaUnitTestProvider represents a Provider implementation to return medusa settings for unit testing.
type MedusaUnitTestProvider struct{}

func (p *MedusaUnitTestProvider) Get(key string) interface{} {
	var m interface{}

	if key == "repositories" {
		// This looks strange. Maybe it is.
		// But this is the structure which will be provided by viper.Get
		// when you parse the medusa.json configuration for repositories.
		// Checkout https://github.com/spf13/cast/issues/36 for details.
		m = []interface{}{
			map[string]interface{}{
				"name": "symfony/console",
				"url":  "git@github.com:symfony/console.git",
			},
			map[string]interface{}{
				"name": "symfony/polyfill",
				"url":  "https://github.com/symfony/polyfill.git",
			},
			map[string]interface{}{
				"name": "no/url",
			},
			map[string]interface{}{
				"name": "empty/url",
				"url":  "",
			},
			map[string]interface{}{
				"name": "invalid/url",
				"url":  "://github.com/invalid/url.git",
			},
		}
	}

	return m
}

func (p *MedusaUnitTestProvider) GetString(key string) string {
	var s string

	if key == "repodir" {
		s = "/var/perseus/git-mirror"
	}

	return s
}

func (p *MedusaUnitTestProvider) GetStringSlice(key string) []string {
	var s []string
	if key == "require" {
		s = []string{
			"symfony/symfony",
			"monolog/monolog",
			"swiftmailer/swiftmailer",
		}
	}
	return s
}

func (p *MedusaUnitTestProvider) GetContentMap() map[string]interface{} {
	m := map[string]interface{}{
		"require": 2,
	}

	return m
}
