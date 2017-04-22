package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

// PackagistClient represents a client to communicate with a Packagist instance
type PackagistClient struct {
	url        *url.URL
	httpClient *http.Client
}

// packagistResponse represents a typical response of a Packagist API call.
// It is an unexported struct, because we right now we are only interested
// in single packages. We offer dedicated functions to get information
// about the single package in a more uncomplicated way.
type packagistResponse struct {
	Package PackagistPackage `json:"package"`
}

// PackagistPackage represents a single package from the Packagist perspective.
// This struct might represent all information what Packagist offers.
// Only those information we are interested in (right now).
// Checkout the Packagist API at https://packagist.org/apidoc for more details
// what information are available.
// TODO: Replace PackagistPackage with "Package", when we return it
type PackagistPackage struct {
	// Name of the package (e.g. symfony/symfony)
	Name string `json:"name"`
	// Repository URL of the Package
	Repository string `json:"repository"`
	// Available and released versions of this package
	Versions map[string]Composer `json:"versions"`
}

// Composer represents a composer.json definition of single package (received by the Packagist API).
// This struct might represent all information what the according composer.json offers.
// Only those information we are interested in (right now).
// Checkout the Composer docs at https://getcomposer.org/ for more details
// what information are available.
type Composer struct {
	// Require are a map of other packages incl. the version constraint that package depends on
	Require map[string]string `json:"require"`
}

// NewPackagist will create a new PackagistClient.
// Instance should be a URL (e.g. https://packagist.org).
func NewPackagist(instance string, httpClient *http.Client) (*PackagistClient, error) {
	if len(instance) == 0 {
		return nil, errors.New("Instance URL is empty")
	}

	// Remove trailing "/"
	if strings.HasSuffix(instance, "/") {
		instance = instance[0 : len(instance)-1]
	}

	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	c := &PackagistClient{
		url:        u,
		httpClient: httpClient,
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	return c, nil
}

// GetPackageByName returns a package by a given name
func (c *PackagistClient) GetPackageByName(name string) (*PackagistPackage, *http.Response, error) {
	u := fmt.Sprintf("%s/packages%s.json", c.url.String(), filepath.Clean("/"+name))
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, resp, err
	}
	defer resp.Body.Close()

	if c := resp.StatusCode; c < 200 || c > 299 {
		return nil, resp, fmt.Errorf("Expected a return code within 2xx for package \"%s\". Got %d", name, c)
	}

	b, err := ioutil.ReadAll(resp.Body)
	var p packagistResponse
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, resp, err
	}

	// TODO Return a normal package here, not a packagist package
	return &p.Package, resp, err
}
