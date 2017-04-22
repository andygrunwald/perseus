package dependency_test

import (
	"reflect"
	"testing"

	. "github.com/andygrunwald/perseus/dependency"
)

func TestNewPackage(t *testing.T) {
	tests := []struct {
		name string
		want *Package
	}{
		{"twig/twig", &Package{Name: "twig/twig"}},
		{"symfony/console", &Package{Name: "symfony/console"}},
		{"", nil},
	}

	for _, tt := range tests {
		if got, err := NewPackage(tt.name, ""); reflect.DeepEqual(got, tt.want) == false {
			t.Errorf("NewPackage(%s) = %+v; want %+v. Error: %s", tt.name, got, tt.want, err)
		}
	}
}
