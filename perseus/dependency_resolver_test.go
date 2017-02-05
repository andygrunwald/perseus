package perseus_test

import (
	"net/http"
	"testing"

	"fmt"
	"github.com/andygrunwald/perseus/packagist"
	. "github.com/andygrunwald/perseus/perseus"
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

// GetPackage will return information about package name.
// This is a dummy implementation only for unit test purpose.
// It is required that this return the exact same results at every unit test run.
func (c *testApiClient) GetPackage(name string) (*packagist.Package, *http.Response, error) {
	switch name {
	// Simulate: API returns an error
	case "api/error":
		return nil, &http.Response{StatusCode: http.StatusBadGateway}, fmt.Errorf("API returns an error")
	// Simulate: API returns nothing for the package
	case "api/empty":
		return nil, &http.Response{StatusCode: http.StatusOK}, nil
	// Simulate: API returns valid content for symfony/console
	case "symfony/console":
		p := &packagist.Package{
			Name: name,
			Versions: map[string]packagist.Composer{
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
		p := &packagist.Package{
			Name: name,
			Versions: map[string]packagist.Composer{
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
		p := &packagist.Package{
			Name: name,
			Versions: map[string]packagist.Composer{
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
		p := &packagist.Package{
			Name: name,
			Versions: map[string]packagist.Composer{
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
		p := &packagist.Package{
			Name: name,
		}
		return p, &http.Response{StatusCode: http.StatusOK}, nil
	}

	return nil, nil, nil
}

func TestNewDependencyResolver(t *testing.T) {
	d, err := NewDependencyResolver("symfony/console", 10, &testApiClient{})
	if err != nil {
		t.Errorf("Error while creating a new dependency resolver: %s", err)
	}
	if d == nil {
		t.Error("Got an empty dependency resolver. Expected a valid one")
	}
}

func TestNewDependencyResolver_Error(t *testing.T) {
	tests := []struct {
		name            string
		numOfWorker     int
		packagistClient packagist.ApiClient
	}{
		{"", 5, &testApiClient{}},
		{"twig/twig", 0, &testApiClient{}},
		{"psr/log", 9, nil},
		{"", 0, nil},
	}

	for _, tt := range tests {
		if got, err := NewDependencyResolver(tt.name, tt.numOfWorker, tt.packagistClient); err == nil {
			t.Errorf("No error while creating a new dependency resolver. Got: %+v", got)
		}
	}
}

// resolvePackages is a small helper function to avoid code duplication
// It will start the dependency resolver for packageName
func resolvePackages(t testError, packageName string) []*Result {
	apiClient := &testApiClient{}
	d, err := NewDependencyResolver(packageName, 3, apiClient)
	if err != nil {
		t.Errorf("Didn't expected an error. Got %s", err)
	}
	results := d.GetResultStream()
	go d.Start()

	r := []*Result{}
	// Finally we collect all the results of the work.
	for v := range results {
		r = append(r, v)
	}

	return r
}

// isStringInResult returns true once needle was found in haystack
func isStringInResult(needle string, haystack []*Result) bool {
	for _, b := range haystack {
		if b.Package == nil {
			return false
		}
		if b.Package.Name == needle {
			return true
		}
	}
	return false
}

func TestPackagistDependencyResolver_SystemPackage(t *testing.T) {
	got := resolvePackages(t, "php")

	if len(got) > 0 {
		t.Errorf("Didn't expected results. Got %+v", got)
	}
}

func TestPackagistDependencyResolver_ApiClientError(t *testing.T) {
	p := "api/error"
	got := resolvePackages(t, p)

	if got[0].Error == nil {
		t.Errorf("Expected an error for package %s to emulate an API error. Got nothing", p)
	}
}

func TestPackagistDependencyResolver_EmptyPackageFromApiClient(t *testing.T) {
	p := "api/empty"
	got := resolvePackages(t, p)

	if got[0].Error == nil {
		t.Errorf("Expected an error for package %s to emulate an empty package from API. Got nothing", p)
	}
}

func TestPackagistDependencyResolver_SuccessSymfonyConsole(t *testing.T) {
	p := "symfony/console"
	got := resolvePackages(t, p)

	if n := len(got); n != 4 {
		t.Errorf("Expected four resolved dependencies. Got %d: %+v", n, got)
	}

	if isStringInResult(p, got) == false {
		t.Errorf("Expected package %s in resultset. Not found.", p)
	}

	p = "symfony/polyfill-mbstring"
	if isStringInResult(p, got) == false {
		t.Errorf("Expected package %s in resultset. Not found.", p)
	}

	p = "symfony/debug"
	if isStringInResult(p, got) == false {
		t.Errorf("Expected package %s in resultset. Not found.", p)
	}

	p = "psr/log"
	if isStringInResult(p, got) == false {
		t.Errorf("Expected package %s in resultset. Not found.", p)
	}
}

func BenchmarkPackagistDependencyResolver_SuccessSymfonyConsole(b *testing.B) {
	p := "symfony/console"
	for n := 0; n < b.N; n++ {
		resolvePackages(b, p)
	}
}

func TestPackagistDependencyResolver_ReplacedPackageNames(t *testing.T) {
	tests := []struct {
		packageName         string
		replacedPackageName string
	}{
		{"symfony/translator", "symfony/translation"},
		{"symfony/doctrine-bundle", "doctrine/doctrine-bundle"},
		{"metadata/metadata", "jms/metadata"},
		{"zendframework/zend-registry", "zf1/zend-registry"},
	}

	for _, tt := range tests {
		if got := resolvePackages(t, tt.packageName); got[0].Package.Name != tt.replacedPackageName {
			t.Errorf("Package %s was not replaced as expected. Expected: %s, got: %s", tt.packageName, tt.replacedPackageName, got[0].Package.Name)
		}
	}
}
