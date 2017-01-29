package packagist_test

import (
	"net/http"
	"testing"

	"fmt"
	. "github.com/andygrunwald/perseus/packagist"
	"net/http/httptest"
	"strings"
)

var (
	// testMux is the HTTP request multiplexer used with the test server.
	testMux *http.ServeMux

	// testClient is the Packagist client being tested.
	testClient ApiClient

	// testServer is a test HTTP server used to provide mock API responses.
	testServer *httptest.Server
)

func TestNew(t *testing.T) {
	tests := []struct {
		instanceURL string
		httpClient  *http.Client
	}{
		{"https://packagist.org/", nil},
		{"https://packagist.org/", http.DefaultClient},
	}

	for _, tt := range tests {
		got, err := New(tt.instanceURL, tt.httpClient)
		if err != nil {
			t.Errorf("New(instanceURL, httpClient) throws an error: %s", err)
		}

		if got == nil {
			t.Errorf("No packagist client created. Got nil. Expected a valid client for instance %s", tt.instanceURL)
		}
	}
}

func TestNew_InvalidInstances(t *testing.T) {
	tests := []struct {
		instanceURL string
		httpClient  *http.Client
	}{
		{"", nil},
		{"://packagist.org/", http.DefaultClient},
	}

	for _, tt := range tests {
		got, err := New(tt.instanceURL, tt.httpClient)
		if err == nil {
			t.Errorf("New(instanceURL, httpClient) throws no error. Expected one for instance: \"%s\"", tt.instanceURL)
		}

		if got != nil {
			t.Errorf("Packagist client created. Expected nothing. Got %+v", got)
		}
	}
}

// setup sets up a test HTTP server along with a packagist.Client that is configured to talk to that test server.
// Tests should register handlers on mux which provide mock responses for the API method being tested.
func setup() {
	// Test server
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)

	// Packagist client configured to use test server
	testClient, _ = New(testServer.URL, nil)
}

// teardown closes the test HTTP server.
func teardown() {
	testServer.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testRequestURL(t *testing.T, r *http.Request, want string) {
	if got := r.URL.String(); !strings.HasPrefix(got, want) {
		t.Errorf("Request URL: %v, want %v", got, want)
	}
}

func TestGetPackage(t *testing.T) {
	pName := "symfony/polyfill"
	setup()
	defer teardown()
	testMux.HandleFunc("/packages/symfony/polyfill.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testRequestURL(t, r, "/packages/symfony/polyfill.json")

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"package":{"name":"symfony\/polyfill","description":"Symfony polyfills backporting features to lower PHP versions","time":"2015-11-04T21:15:52+00:00","maintainers":[{"name":"fabpot","avatar_url":"https:\/\/www.gravatar.com\/avatar\/9a22d09f92d50fa3d2a16766d0ba52f8?d=identicon"}],"versions":{"dev-master":{"name":"symfony\/polyfill","description":"Symfony polyfills backporting features to lower PHP versions","keywords":["compatibility","compat","polyfill","shim"],"homepage":"https:\/\/symfony.com","version":"dev-master","version_normalized":"9999999-dev","license":["MIT"],"authors":[{"name":"Nicolas Grekas","email":"p@tchwork.com"},{"name":"Symfony Community","homepage":"https:\/\/symfony.com\/contributors"}],"source":{"type":"git","url":"https:\/\/github.com\/symfony\/polyfill.git","reference":"385d033a8e1d8778446d699ecbd886480716eba7"},"dist":{"type":"zip","url":"https:\/\/api.github.com\/repos\/symfony\/polyfill\/zipball\/385d033a8e1d8778446d699ecbd886480716eba7","reference":"385d033a8e1d8778446d699ecbd886480716eba7","shasum":""},"type":"library","time":"2016-11-14T01:15:23+00:00","autoload":{"psr-4":{"Symfony\\Polyfill\\":"src\/"},"files":["src\/Apcu\/bootstrap.php","src\/Php54\/bootstrap.php","src\/Php55\/bootstrap.php","src\/Php56\/bootstrap.php","src\/Php70\/bootstrap.php","src\/Php71\/bootstrap.php","src\/Iconv\/bootstrap.php","src\/Intl\/Grapheme\/bootstrap.php","src\/Intl\/Icu\/bootstrap.php","src\/Intl\/Normalizer\/bootstrap.php","src\/Mbstring\/bootstrap.php","src\/Xml\/bootstrap.php"],"classmap":["src\/Intl\/Normalizer\/Resources\/stubs","src\/Php70\/Resources\/stubs","src\/Php54\/Resources\/stubs"]},"extra":{"branch-alias":{"dev-master":"1.3-dev"}},"require":{"php":"\u003E=5.3.3","ircmaxell\/password-compat":"~1.0","symfony\/intl":"~2.3|~3.0","paragonie\/random_compat":"~1.0|~2.0"},"replace":{"symfony\/polyfill-php54":"self.version","symfony\/polyfill-php55":"self.version","symfony\/polyfill-php56":"self.version","symfony\/polyfill-php70":"self.version","symfony\/polyfill-iconv":"self.version","symfony\/polyfill-intl-grapheme":"self.version","symfony\/polyfill-intl-icu":"self.version","symfony\/polyfill-intl-normalizer":"self.version","symfony\/polyfill-mbstring":"self.version","symfony\/polyfill-util":"self.version","symfony\/polyfill-xml":"self.version","symfony\/polyfill-apcu":"self.version","symfony\/polyfill-php71":"self.version"}},"v1.3.0":{"name":"symfony\/polyfill","description":"Symfony polyfills backporting features to lower PHP versions","keywords":["compatibility","compat","polyfill","shim"],"homepage":"https:\/\/symfony.com","version":"v1.3.0","version_normalized":"1.3.0.0","license":["MIT"],"authors":[{"name":"Nicolas Grekas","email":"p@tchwork.com"},{"name":"Symfony Community","homepage":"https:\/\/symfony.com\/contributors"}],"source":{"type":"git","url":"https:\/\/github.com\/symfony\/polyfill.git","reference":"385d033a8e1d8778446d699ecbd886480716eba7"},"dist":{"type":"zip","url":"https:\/\/api.github.com\/repos\/symfony\/polyfill\/zipball\/385d033a8e1d8778446d699ecbd886480716eba7","reference":"385d033a8e1d8778446d699ecbd886480716eba7","shasum":""},"type":"library","time":"2016-11-14T01:15:23+00:00","autoload":{"psr-4":{"Symfony\\Polyfill\\":"src\/"},"files":["src\/Apcu\/bootstrap.php","src\/Php54\/bootstrap.php","src\/Php55\/bootstrap.php","src\/Php56\/bootstrap.php","src\/Php70\/bootstrap.php","src\/Php71\/bootstrap.php","src\/Iconv\/bootstrap.php","src\/Intl\/Grapheme\/bootstrap.php","src\/Intl\/Icu\/bootstrap.php","src\/Intl\/Normalizer\/bootstrap.php","src\/Mbstring\/bootstrap.php","src\/Xml\/bootstrap.php"],"classmap":["src\/Intl\/Normalizer\/Resources\/stubs","src\/Php70\/Resources\/stubs","src\/Php54\/Resources\/stubs"]},"extra":{"branch-alias":{"dev-master":"1.3-dev"}},"require":{"php":"\u003E=5.3.3","ircmaxell\/password-compat":"~1.0","paragonie\/random_compat":"~1.0|~2.0","symfony\/intl":"~2.3|~3.0"},"replace":{"symfony\/polyfill-apcu":"self.version","symfony\/polyfill-php54":"self.version","symfony\/polyfill-php55":"self.version","symfony\/polyfill-php56":"self.version","symfony\/polyfill-php70":"self.version","symfony\/polyfill-php71":"self.version","symfony\/polyfill-iconv":"self.version","symfony\/polyfill-intl-grapheme":"self.version","symfony\/polyfill-intl-icu":"self.version","symfony\/polyfill-intl-normalizer":"self.version","symfony\/polyfill-mbstring":"self.version","symfony\/polyfill-util":"self.version","symfony\/polyfill-xml":"self.version"}},"v1.2.0":{"name":"symfony\/polyfill","description":"Symfony polyfills backporting features to lower PHP versions","keywords":["compatibility","compat","polyfill","shim"],"homepage":"https:\/\/symfony.com","version":"v1.2.0","version_normalized":"1.2.0.0","license":["MIT"],"authors":[{"name":"Nicolas Grekas","email":"p@tchwork.com"},{"name":"Symfony Community","homepage":"https:\/\/symfony.com\/contributors"}],"source":{"type":"git","url":"https:\/\/github.com\/symfony\/polyfill.git","reference":"ee2c9c2576fdd4a42b024260a1906a9888770c34"},"dist":{"type":"zip","url":"https:\/\/api.github.com\/repos\/symfony\/polyfill\/zipball\/ee2c9c2576fdd4a42b024260a1906a9888770c34","reference":"ee2c9c2576fdd4a42b024260a1906a9888770c34","shasum":""},"type":"library","time":"2016-05-18T14:27:53+00:00","autoload":{"psr-4":{"Symfony\\Polyfill\\":"src\/"},"files":["src\/Apcu\/bootstrap.php","src\/Php54\/bootstrap.php","src\/Php55\/bootstrap.php","src\/Php56\/bootstrap.php","src\/Php70\/bootstrap.php","src\/Iconv\/bootstrap.php","src\/Intl\/Grapheme\/bootstrap.php","src\/Intl\/Icu\/bootstrap.php","src\/Intl\/Normalizer\/bootstrap.php","src\/Mbstring\/bootstrap.php","src\/Xml\/bootstrap.php"],"classmap":["src\/Intl\/Normalizer\/Resources\/stubs","src\/Php70\/Resources\/stubs","src\/Php54\/Resources\/stubs"]},"extra":{"branch-alias":{"dev-master":"1.2-dev"}},"require":{"php":"\u003E=5.3.3","ircmaxell\/password-compat":"~1.0","paragonie\/random_compat":"~1.0|~2.0","symfony\/intl":"~2.3|~3.0"},"replace":{"symfony\/polyfill-apcu":"self.version","symfony\/polyfill-php54":"self.version","symfony\/polyfill-php55":"self.version","symfony\/polyfill-php56":"self.version","symfony\/polyfill-php70":"self.version","symfony\/polyfill-iconv":"self.version","symfony\/polyfill-intl-grapheme":"self.version","symfony\/polyfill-intl-icu":"self.version","symfony\/polyfill-intl-normalizer":"self.version","symfony\/polyfill-mbstring":"self.version","symfony\/polyfill-util":"self.version","symfony\/polyfill-xml":"self.version"}},"v1.1.1":{"name":"symfony\/polyfill","description":"Symfony polyfills backporting features to lower PHP versions","keywords":["compatibility","compat","polyfill","shim"],"homepage":"https:\/\/symfony.com","version":"v1.1.1","version_normalized":"1.1.1.0","license":["MIT"],"authors":[{"name":"Nicolas Grekas","email":"p@tchwork.com"},{"name":"Symfony Community","homepage":"https:\/\/symfony.com\/contributors"}],"source":{"type":"git","url":"https:\/\/github.com\/symfony\/polyfill.git","reference":"3dc21aeff3e1f8cb708421ed02cf1a8901d7b535"},"dist":{"type":"zip","url":"https:\/\/api.github.com\/repos\/symfony\/polyfill\/zipball\/3dc21aeff3e1f8cb708421ed02cf1a8901d7b535","reference":"3dc21aeff3e1f8cb708421ed02cf1a8901d7b535","shasum":""},"type":"library","time":"2016-03-03T16:58:13+00:00","autoload":{"psr-4":{"Symfony\\Polyfill\\":"src\/"},"files":["src\/Apcu\/bootstrap.php","src\/Php54\/bootstrap.php","src\/Php55\/bootstrap.php","src\/Php56\/bootstrap.php","src\/Php70\/bootstrap.php","src\/Iconv\/bootstrap.php","src\/Intl\/Grapheme\/bootstrap.php","src\/Intl\/Icu\/bootstrap.php","src\/Intl\/Normalizer\/bootstrap.php","src\/Mbstring\/bootstrap.php","src\/Xml\/bootstrap.php"],"classmap":["src\/Intl\/Normalizer\/Resources\/stubs","src\/Php70\/Resources\/stubs","src\/Php54\/Resources\/stubs"]},"extra":{"branch-alias":{"dev-master":"1.1-dev"}},"require":{"php":"\u003E=5.3.3","ircmaxell\/password-compat":"~1.0","paragonie\/random_compat":"~1.0","symfony\/intl":"~2.3|~3.0"},"replace":{"symfony\/polyfill-apcu":"self.version","symfony\/polyfill-php54":"self.version","symfony\/polyfill-php55":"self.version","symfony\/polyfill-php56":"self.version","symfony\/polyfill-php70":"self.version","symfony\/polyfill-iconv":"self.version","symfony\/polyfill-intl-grapheme":"self.version","symfony\/polyfill-intl-icu":"self.version","symfony\/polyfill-intl-normalizer":"self.version","symfony\/polyfill-mbstring":"self.version","symfony\/polyfill-util":"self.version","symfony\/polyfill-xml":"self.version"}},"v1.1.0":{"name":"symfony\/polyfill","description":"Symfony polyfills backporting features to lower PHP versions","keywords":["compatibility","compat","polyfill","shim"],"homepage":"https:\/\/symfony.com","version":"v1.1.0","version_normalized":"1.1.0.0","license":["MIT"],"authors":[{"name":"Nicolas Grekas","email":"p@tchwork.com"},{"name":"Symfony Community","homepage":"https:\/\/symfony.com\/contributors"}],"source":{"type":"git","url":"https:\/\/github.com\/symfony\/polyfill.git","reference":"ceffa85c57f023a816f5c511ad35081e7c67d7cd"},"dist":{"type":"zip","url":"https:\/\/api.github.com\/repos\/symfony\/polyfill\/zipball\/ceffa85c57f023a816f5c511ad35081e7c67d7cd","reference":"ceffa85c57f023a816f5c511ad35081e7c67d7cd","shasum":""},"type":"library","time":"2016-01-25T08:44:42+00:00","autoload":{"psr-4":{"Symfony\\Polyfill\\":"src\/"},"files":["src\/Apcu\/bootstrap.php","src\/Php54\/bootstrap.php","src\/Php55\/bootstrap.php","src\/Php56\/bootstrap.php","src\/Php70\/bootstrap.php","src\/Iconv\/bootstrap.php","src\/Intl\/Grapheme\/bootstrap.php","src\/Intl\/Icu\/bootstrap.php","src\/Intl\/Normalizer\/bootstrap.php","src\/Mbstring\/bootstrap.php","src\/Xml\/bootstrap.php"],"classmap":["src\/Apcu\/Resources\/stubs","src\/Intl\/Normalizer\/Resources\/stubs","src\/Php70\/Resources\/stubs","src\/Php54\/Resources\/stubs"]},"extra":{"branch-alias":{"dev-master":"1.1-dev"}},"require":{"php":"\u003E=5.3.3","ircmaxell\/password-compat":"~1.0","paragonie\/random_compat":"~1.0","symfony\/intl":"~2.3|~3.0"},"replace":{"symfony\/polyfill-apcu":"self.version","symfony\/polyfill-php54":"self.version","symfony\/polyfill-php55":"self.version","symfony\/polyfill-php56":"self.version","symfony\/polyfill-php70":"self.version","symfony\/polyfill-iconv":"self.version","symfony\/polyfill-intl-grapheme":"self.version","symfony\/polyfill-intl-icu":"self.version","symfony\/polyfill-intl-normalizer":"self.version","symfony\/polyfill-mbstring":"self.version","symfony\/polyfill-util":"self.version","symfony\/polyfill-xml":"self.version"}},"v1.0.1":{"name":"symfony\/polyfill","description":"Symfony polyfills backporting features to lower PHP versions","keywords":["compatibility","compat","polyfill","shim"],"homepage":"https:\/\/symfony.com","version":"v1.0.1","version_normalized":"1.0.1.0","license":["MIT"],"authors":[{"name":"Nicolas Grekas","email":"p@tchwork.com"},{"name":"Symfony Community","homepage":"https:\/\/symfony.com\/contributors"}],"source":{"type":"git","url":"https:\/\/github.com\/symfony\/polyfill.git","reference":"dd9db1dc4013821a63f7afbd8340dd57939fe674"},"dist":{"type":"zip","url":"https:\/\/api.github.com\/repos\/symfony\/polyfill\/zipball\/dd9db1dc4013821a63f7afbd8340dd57939fe674","reference":"dd9db1dc4013821a63f7afbd8340dd57939fe674","shasum":""},"type":"library","time":"2015-12-18T15:10:25+00:00","autoload":{"psr-4":{"Symfony\\Polyfill\\":"src\/"},"files":["src\/Php54\/bootstrap.php","src\/Php55\/bootstrap.php","src\/Php56\/bootstrap.php","src\/Php70\/bootstrap.php","src\/Iconv\/bootstrap.php","src\/Intl\/Grapheme\/bootstrap.php","src\/Intl\/Icu\/bootstrap.php","src\/Intl\/Normalizer\/bootstrap.php","src\/Mbstring\/bootstrap.php","src\/Xml\/bootstrap.php"],"classmap":["src\/Intl\/Normalizer\/Resources\/stubs","src\/Php70\/Resources\/stubs","src\/Php54\/Resources\/stubs"]},"extra":{"branch-alias":{"dev-master":"1.0-dev"}},"require":{"php":"\u003E=5.3.3","ircmaxell\/password-compat":"~1.0","paragonie\/random_compat":"~1.0","symfony\/intl":"~2.3|~3.0"},"replace":{"symfony\/polyfill-php54":"self.version","symfony\/polyfill-php55":"self.version","symfony\/polyfill-php56":"self.version","symfony\/polyfill-php70":"self.version","symfony\/polyfill-iconv":"self.version","symfony\/polyfill-intl-grapheme":"self.version","symfony\/polyfill-intl-icu":"self.version","symfony\/polyfill-intl-normalizer":"self.version","symfony\/polyfill-mbstring":"self.version","symfony\/polyfill-util":"self.version","symfony\/polyfill-xml":"self.version"}},"v1.0.0":{"name":"symfony\/polyfill","description":"Symfony polyfills backporting features to lower PHP versions","keywords":["compatibility","compat","polyfill","shim"],"homepage":"https:\/\/symfony.com","version":"v1.0.0","version_normalized":"1.0.0.0","license":["MIT"],"authors":[{"name":"Nicolas Grekas","email":"p@tchwork.com"},{"name":"Symfony Community","homepage":"https:\/\/symfony.com\/contributors"}],"source":{"type":"git","url":"https:\/\/github.com\/symfony\/polyfill.git","reference":"fef21adc706d3bb8f31d37c503ded2160c76c64a"},"dist":{"type":"zip","url":"https:\/\/api.github.com\/repos\/symfony\/polyfill\/zipball\/fef21adc706d3bb8f31d37c503ded2160c76c64a","reference":"fef21adc706d3bb8f31d37c503ded2160c76c64a","shasum":""},"type":"library","time":"2015-11-04T20:29:00+00:00","autoload":{"psr-4":{"Symfony\\Polyfill\\":"src\/"},"files":["src\/Php54\/bootstrap.php","src\/Php55\/bootstrap.php","src\/Php56\/bootstrap.php","src\/Php70\/bootstrap.php","src\/Iconv\/bootstrap.php","src\/Intl\/Grapheme\/bootstrap.php","src\/Intl\/Icu\/bootstrap.php","src\/Intl\/Normalizer\/bootstrap.php","src\/Mbstring\/bootstrap.php","src\/Xml\/bootstrap.php"],"classmap":["src\/Intl\/Normalizer\/Resources\/stubs","src\/Php70\/Resources\/stubs","src\/Php54\/Resources\/stubs"]},"extra":{"branch-alias":{"dev-master":"1.0-dev"}},"require":{"php":"\u003E=5.3.3","ircmaxell\/password-compat":"~1.0","paragonie\/random_compat":"~1.0","symfony\/intl":"~2.3|~3.0"},"replace":{"symfony\/polyfill-php54":"self.version","symfony\/polyfill-php55":"self.version","symfony\/polyfill-php56":"self.version","symfony\/polyfill-php70":"self.version","symfony\/polyfill-iconv":"self.version","symfony\/polyfill-intl-grapheme":"self.version","symfony\/polyfill-intl-icu":"self.version","symfony\/polyfill-intl-normalizer":"self.version","symfony\/polyfill-mbstring":"self.version","symfony\/polyfill-util":"self.version","symfony\/polyfill-xml":"self.version"}}},"type":"library","repository":"https:\/\/github.com\/symfony\/polyfill","github_stars":212,"github_watchers":28,"github_forks":33,"github_open_issues":8,"language":"PHP","dependents":4,"suggesters":0,"downloads":{"total":47730,"monthly":6148,"daily":97},"favers":212}}`)
	})

	p, _, err := testClient.GetPackage(pName)
	if p == nil {
		t.Error("Expected a valid package. Package is nil")
	}
	if err != nil {
		t.Errorf("Didn't expected an error. Got: %s", err)
	}
	if pName != p.Name {
		t.Errorf("Expected package name is %s. Got %s", pName, p.Name)
	}
}

func TestGetPackage_InvalidPackage(t *testing.T) {
	setup()
	defer teardown()
	testMux.HandleFunc("/packages/invalid/package.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testRequestURL(t, r, "/packages/invalid/package.json")

		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"status":"error","message":"Package not found"}`)
	})

	p, _, err := testClient.GetPackage("invalid/package")
	if p != nil {
		t.Errorf("Expected an empty package. Got: %+v", p)
	}
	if err == nil {
		t.Error("Expected an error. Got nothing")
	}
}
