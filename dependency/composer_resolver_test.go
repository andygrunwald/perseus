package dependency_test

import (
	"testing"

	. "github.com/andygrunwald/perseus/dependency"
	"github.com/andygrunwald/perseus/dependency/repository"
	"fmt"
)

// resolvePackages is a small helper function to avoid code duplication
// It will start the dependency resolver for packageName
func resolvePackages(t testError, packageName string) []*Result {
	apiClient := &testApiClient{}
	d, err := NewComposerResolver(3, apiClient)
	if err != nil {
		t.Errorf("Didn't expected an error. Got %s", err)
	}
	results := d.GetResultStream()
	p, _ := NewPackage(packageName, "")
	go d.Resolve([]*Package{p})

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

func TestComposerResolver_SystemPackage(t *testing.T) {
	got := resolvePackages(t, "php")

	if len(got) > 0 {
		t.Errorf("Didn't expected results. Got %+v", got)
	}
}

func TestComposerResolver_ApiClientError(t *testing.T) {
	p := "api/error"
	got := resolvePackages(t, p)

	if got[0].Error == nil {
		t.Errorf("Expected an error for package %s to emulate an API error. Got nothing", p)
	}
}

func TestComposerResolver_EmptyPackageFromApiClient(t *testing.T) {
	p := "api/empty"
	got := resolvePackages(t, p)

	if got[0].Error == nil {
		t.Errorf("Expected an error for package %s to emulate an empty package from API. Got nothing", p)
	}
}

func TestComposerResolver_SuccessSymfonyConsole(t *testing.T) {
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

func BenchmarkComposerResolver_SuccessSymfonyConsole(b *testing.B) {
	p := "symfony/console"
	for n := 0; n < b.N; n++ {
		resolvePackages(b, p)
	}
}

func TestComposerResolver_ReplacedPackageNames(t *testing.T) {
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

func ExampleComposerResolver() {
	u := "https://packagist.org/"
	packageName := "symfony/console"
	numOfWorker := 3

	// Create a new package
	p, err := NewPackage(packageName, "")
	if err != nil {
		panic(err)
	}

	// Create a packagist client (PHP)
	packagistClient, err := repository.NewPackagist(u, nil)
	if err != nil {
		panic(err)
	}

	// Create a composer resolver and inject the packagist client
	resolver, err := NewComposerResolver(numOfWorker, packagistClient)
	if err != nil {
		panic(err)
	}

	results := resolver.GetResultStream()
	go resolver.Resolve([]*Package{p})

	dependencies := []string{}
	// Finally we collect all the results of the work.
	for v := range results {
		dependencies = append(dependencies, v.Package.Name)
	}

	fmt.Printf("%d dependencies found for package \"%s\" on %s", len(dependencies), p.Name, u)
	// Output: 4 dependencies found for package "symfony/console" on https://packagist.org/
}