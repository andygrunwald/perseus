package packagist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"path/filepath"
)

// Client represents a client to communicate with a Packagist instance
type Client struct {
	url        *url.URL
	httpClient *http.Client
}

// packageResponse represents a typical response of a Packagist API call.
// It is an unexported struct, because we right now we are only interested
// in single packages. We offer dedicated functions to get information
// about the single package in a more uncomplicated way.
type packageResponse struct {
	Package Package `json:"package"`
}

// Package represents a single package from the Packagist perspective.
// This struct might represent all information what Packagist offers.
// Only those information we are interested in (right now).
// Checkout the Packagist API at https://packagist.org/apidoc for more details
// what information are available.
type Package struct {
	// Name of the package
	Name string `json:"name"`
	// TODO Should be a net/url.URL (with the UnmashalJSON(b []byte) error interface)
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

// New will create a new Packagist client
func New(instance string, httpClient *http.Client) (ApiClient, error) {
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

	c := &Client{
		url:        u,
		httpClient: httpClient,
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	return c, nil
}

// GetPackage will retrieve information about package name from
// a given packagist instance.
func (c *Client) GetPackage(name string) (*Package, *http.Response, error) {
	u := fmt.Sprintf("%s/packages%s.json", c.url.String(), filepath.Clean("/"+name))
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, resp, err
	}
	defer resp.Body.Close()

	// Check the status codes
	if c := resp.StatusCode; c < 200 || c > 299 {
		return nil, resp, fmt.Errorf("Expected a return code within 2xx for package \"%s\". Got %d", name, c)
	}

	b, err := ioutil.ReadAll(resp.Body)
	var p packageResponse
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, resp, err
	}

	return &p.Package, resp, err
}
