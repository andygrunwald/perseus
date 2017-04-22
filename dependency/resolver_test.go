package dependency_test

import (
	"fmt"
	"net/http"
	"testing"

	. "github.com/andygrunwald/perseus/dependency"
	"github.com/andygrunwald/perseus/dependency/repository"
)

// testError is an interface to be able to handle
// t *testing.T and b *testing.B in the same way.
// Checkout the usage.
type testError interface {
	Errorf(format string, args ...interface{})
}

// testApiClient is a dummy implementation for packagist.ApiClient
// for unit test purpse
type testApiClient struct{}

// GetPackageByName will return information about package name.
// This is a dummy implementation only for unit test purpose.
// It is required that this return the exact same results at every unit test run.
func (c *testApiClient) GetPackageByName(name string) (*repository.PackagistPackage, *http.Response, error) {
	switch name {
	// Simulate: API returns an error
	case "api/error":
		return nil, &http.Response{StatusCode: http.StatusBadGateway}, fmt.Errorf("API returns an error")
	// Simulate: API returns nothing for the package
	case "api/empty":
		return nil, &http.Response{StatusCode: http.StatusOK}, nil
	// Simulate: API returns valid content for symfony/console
	case "symfony/console":
		p := &repository.PackagistPackage{
			Name: name,
			Versions: map[string]repository.Composer{
				"3.2.2": {
					Require: map[string]string{
						"php": ">=5.5.9",
						"symfony/polyfill-mbstring": " ~1.0",
						"symfony/debug":             "~2.8|~3.0",
					},
				},
				"2.8.12": {
					Require: map[string]string{
						"php": ">=5.3.9",
						"symfony/polyfill-mbstring": " ~1.0",
						"symfony/debug":             "~2.7,>=2.7.2|~3.0.0",
					},
				},
				"2.0.4": {
					Require: map[string]string{
						"php": ">=5.3.2",
					},
				},
			},
		}
		return p, &http.Response{StatusCode: http.StatusOK}, nil
	// Simulate: API returns valid content for symfony/debug
	case "symfony/debug":
		p := &repository.PackagistPackage{
			Name: name,
			Versions: map[string]repository.Composer{
				"3.2.1": {
					Require: map[string]string{
						"php":     ">=5.5.9",
						"psr/log": "~1.0",
					},
				},
				"2.8.7": {},
			},
		}
		return p, &http.Response{StatusCode: http.StatusOK}, nil
	// Simulate: API returns valid content for symfony/polyfill-mbstring
	case "symfony/polyfill-mbstring":
		p := &repository.PackagistPackage{
			Name: name,
			Versions: map[string]repository.Composer{
				"1.3.0": {
					Require: map[string]string{
						"php": ">=5.3.3",
					},
				},
			},
		}
		return p, &http.Response{StatusCode: http.StatusOK}, nil
	// Simulate: API returns valid content for psr/log
	case "psr/log":
		p := &repository.PackagistPackage{
			Name: name,
			Versions: map[string]repository.Composer{
				"1.0.2": {
					Require: map[string]string{
						"php": ">=5.3.0",
					},
				},
				"1.0.1": {
					Require: map[string]string{
						// With this entry we test two things:
						// 1. A true return of isPackageAlreadyResolved (because this package is resolved at first)
						// 2. If we can deal with circular dependency
						// Tricky and neat :)
						"symfony/console": "~3.0",
					},
				},
			},
		}
		return p, &http.Response{StatusCode: http.StatusOK}, nil
	// Simulate: API returns valid content for symfony/translation
	case "symfony/translation":
		fallthrough
	case "doctrine/doctrine-bundle":
		fallthrough
	case "jms/metadata":
		fallthrough
	case "zf1/zend-registry":
		p := &repository.PackagistPackage{
			Name: name,
		}
		return p, &http.Response{StatusCode: http.StatusOK}, nil
	}

	return nil, nil, nil
}

func TestNewComposerResolver(t *testing.T) {
	d, err := NewComposerResolver(10, &testApiClient{})
	if err != nil {
		t.Errorf("Error while creating a new dependency resolver: %s", err)
	}
	if d == nil {
		t.Error("Got an empty dependency resolver. Expected a valid one")
	}
}

func TestNewComposerResolver_Error(t *testing.T) {
	tests := []struct {
		numOfWorker     int
		packagistClient repository.Client
	}{
		{0, &testApiClient{}},
		{9, nil},
		{0, nil},
	}

	for _, tt := range tests {
		if got, err := NewComposerResolver(tt.numOfWorker, tt.packagistClient); err == nil {
			t.Errorf("No error while creating a new dependency resolver. Got: %+v", got)
		}
	}
}