package perseus_test

import (
	"testing"

	"github.com/andygrunwald/perseus/packagist"
	. "github.com/andygrunwald/perseus/perseus"
	"net/http"
)

// testApiClient is a dummy implementation for packagist.ApiClient
// for unit test purpse
type testApiClient struct{}

// GetPackage will return information about package name.
// This is a dummy implementation only for unit test purpose.
// It is required that this return the exact same results at every unit test run.
func (c *testApiClient) GetPackage(name string) (*packagist.Package, *http.Response, error) {
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
