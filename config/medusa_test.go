package config_test

import (
	"testing"

	. "github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/dependency"
)

func TestNewMedusa(t *testing.T) {
	v := &EmptyUnitTestProvider{}

	m, err := NewMedusa(v)
	if m == nil {
		t.Error("Medusa configuration object is nil.")
	}
	if err != nil {
		t.Errorf("NewMedusa(provider) throws error: %s", err)
	}
}

func TestNewMedusa_WithoutConfiguration(t *testing.T) {
	m, err := NewMedusa(nil)
	if m != nil {
		t.Errorf("Expected medusa configuration to be nil. Got: %+v", m)
	}
	if err == nil {
		t.Error("No error thrown by creating NewMedusa(nil).")
	}
}

func TestMedusa_GetRepositoryURLOfPackage_FaultyRepositories(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
	}{
		// No "repositories" configured at all
		{"symfony/console", &EmptyUnitTestProvider{}},
		// "repositories" key is there, but empty
		{"key/exists", &EmptyWithKeysUnitTestProvider{}},
		// Repositories configured, but not this one
		{"twig/twig", &MedusaUnitTestProvider{}},
		// Repositories configured, plus this one, but doesn't have an url key
		{"no/url", &MedusaUnitTestProvider{}},
		// Repositories configured, plus this one, but has an empty url key
		{"empty/url", &MedusaUnitTestProvider{}},
		// Repositories configured, plus this one, but has an invalid url key
		{"invalid/url", &MedusaUnitTestProvider{}},
	}

	for _, tt := range tests {
		m, err := NewMedusa(tt.provider)
		if err != nil {
			t.Errorf("NewMedusa(Provider) throws error: %s", err)
		}

		p := &dependency.Package{Name: tt.name}
		u, err := m.GetRepositoryURLOfPackage(p)
		if u != nil {
			t.Errorf("Package '%s': Expected url to be nil. Got: %+v", tt.name, u)
		}
		if err == nil {
			t.Errorf("Package '%s': No error thrown.", tt.name)
		}
	}
}

func TestMedusa_GetRepositoryURLOfPackage_CorrectRepositories(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
	}{
		// Everything is fine with a git ssh url (github)
		{"symfony/console", &MedusaUnitTestProvider{}},
		// Everything is fine with a https url (github)
		{"symfony/polyfill", &MedusaUnitTestProvider{}},
	}

	for _, tt := range tests {
		m, err := NewMedusa(tt.provider)
		if err != nil {
			t.Errorf("NewMedusa(Provider) throws error: %s", err)
		}

		p := &dependency.Package{Name: tt.name}
		u, err := m.GetRepositoryURLOfPackage(p)
		if u == nil {
			t.Errorf("Package '%s': No url returned. Expected one.", tt.name)

		}
		if err != nil {
			t.Errorf("Package '%s': Expected error to be nil. Got: %+v", tt.name, err)
		}
	}
}

func TestMedusa_GetNamesOfRepositories(t *testing.T) {
	provider := &MedusaUnitTestProvider{}
	numOfPackages := 2

	m, err := NewMedusa(provider)
	if err != nil {
		t.Errorf("NewMedusa(Provider) throws error: %s", err)
	}

	l, err := m.GetNamesOfRepositories()
	if err != nil {
		t.Errorf("Expected no error for provider %T. Got %s", provider, err)
	}
	if n := len(l); n != numOfPackages {
		t.Errorf("Expected a different number of packages. Got %d, expected %d for provider %T", n, numOfPackages, provider)
	}
}

func TestMedusa_GetNamesOfRepositories_EmptyRepositoryList(t *testing.T) {
	tests := []struct {
		provider Provider
		num      int
	}{
		// No "repositories" configured at all
		{&EmptyUnitTestProvider{}, 0},
		// "repositories" key is there, but empty
		{&EmptyWithKeysUnitTestProvider{}, 0},
	}

	for _, tt := range tests {
		m, err := NewMedusa(tt.provider)
		if err != nil {
			t.Errorf("NewMedusa(Provider) throws error: %s", err)
		}

		l, err := m.GetNamesOfRepositories()
		if n := len(l); n != tt.num {
			t.Errorf("Expected a different number of packages. Got %d, expected %d for provider %T", n, tt.num, tt.provider)
		}
		if err == nil {
			t.Errorf("No error thrown. Expected an error for provider %T", tt.provider)
		}
	}
}

func TestMedusa_GetRequire(t *testing.T) {
	tests := []struct {
		provider Provider
		num      int
	}{
		// Empty "require" configured
		{&EmptyUnitTestProvider{}, 0},
		// 3 "require"s configured
		{&MedusaUnitTestProvider{}, 3},
	}

	for _, tt := range tests {
		m, err := NewMedusa(tt.provider)
		if err != nil {
			t.Errorf("NewMedusa(Provider) throws error: %s", err)
		}

		l := m.GetRequire()
		if n := len(l); n != tt.num {
			t.Errorf("Expected a different number of packages. Got %d, expected %d for provider %T", n, tt.num, tt.provider)
		}
	}
}

func TestMedusa_GetString(t *testing.T) {
	provider := &MedusaUnitTestProvider{}
	key := "repodir"

	m, err := NewMedusa(provider)
	if err != nil {
		t.Errorf("NewMedusa(Provider) throws error: %s", err)
	}

	s := m.GetString(key)
	if len(s) == 0 {
		t.Errorf("Expected a non empty string. Got an empty string for key %s", key)
	}
}