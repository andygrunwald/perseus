package perseus_test

import (
	"testing"
	"net/http"

	"github.com/andygrunwald/perseus/packagist"
	. "github.com/andygrunwald/perseus/perseus"
	"fmt"
)

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
		return nil, nil, fmt.Errorf("API returns an error")
	// Simulate: API returns nothing for the package
	case "api/empty":
		return nil, nil, nil
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

func resolvePackages(t *testing.T, packageName string) []*Result {
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