package config_test

import (
	"testing"
	. "github.com/andygrunwald/perseus/config"
	"bytes"
	"encoding/json"
	"reflect"
)

func unitTestJSONContent() []byte {
	b := bytes.NewBufferString(`{
    "archive": {
        "directory": "dist",
        "format": "tar",
        "prefix-url": "http://my.url.com/packages/",
        "skip-dev": true
    },
    "homepage": "http://my.url.com/packages/",
    "name": "My private php package repositories",
    "providers": true,
    "repositories": [
        {
            "type": "git",
            "url": "http://my.url.com/git-mirror/twig/twig.git"
        },
        {
            "type": "git",
            "url": "http://my.url.comgit-mirror/mre/phpench.git"
        },
        {
            "type": "git",
            "url": "http://my.url.com/git-mirror/symfony/console.git"
        }
    ],
    "dummy-list": [
    	"https://www.google.com/",
    	"https://www.github.com/",
    	"https://gobot.io/"
    ],
    "require-all": true
}`)
	return b.Bytes()
}

func TestNewJSONProvider(t *testing.T) {
	p, err := NewJSONProvider(unitTestJSONContent())
	if err != nil {
		t.Fatalf("Got error: %s", err)
	}
	if p == nil {
		t.Fatal("Got an empty JSON provider. Expected a valid one.")
	}
}

func TestNewJSONProvider_EmptyContent(t *testing.T) {
	p, err := NewJSONProvider([]byte{})
	if err == nil {
		t.Fatal("Expected an error. Got none")
	}
	if p != nil {
		t.Fatal("Got a valid JSON provider. Expected a n empty one.")
	}
}

func TestJSONProvider_GetString(t *testing.T) {
	expected := "http://my.url.com/packages/"
	p, err := NewJSONProvider(unitTestJSONContent())
	if err != nil {
		t.Fatalf("Got error: %s", err)
	}

	got := p.GetString("homepage")
	if got != expected {
		t.Fatalf("Got different value than expected. Expected %s, got %s", expected, got)
	}
}

func TestJSONProvider_GetString_NotExists(t *testing.T) {
	expected := "http://my.url.com/packages/"
	p, err := NewJSONProvider(unitTestJSONContent())
	if err != nil {
		t.Fatalf("Got error: %s", err)
	}

	got := p.GetString("not-exists")
	if got != "" {
		t.Fatalf("Got different value than expected. Expected an empty string, got %s", expected, got)
	}
}

func TestJSONProvider_Get(t *testing.T) {
	c := unitTestJSONContent()
	b := make(map[string]*json.RawMessage)
	err := json.Unmarshal(c, &b)
	if err != nil {
		t.Fatalf("Got error while creating json.RawMessage dummy: %s", err)
	}

	expected := b["homepage"]
	p, err := NewJSONProvider(c)
	if err != nil {
		t.Fatalf("Got error while creating new JSON provier: %s", err)
	}

	got := p.Get("homepage")
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Got different value than expected. Expected %+v, got %+v", expected, got)
	}
}

func TestJSONProvider_Get_NotExists(t *testing.T) {
	p, err := NewJSONProvider(unitTestJSONContent())
	if err != nil {
		t.Fatalf("Got error while creating new JSON provier: %s", err)
	}

	got := p.Get("not-exists")
	if got.(*json.RawMessage) != nil {
		t.Fatalf("Got different value than expected. Expected nil, got %+v", got.(*json.RawMessage))
	}
}

func TestJSONProvider_GetContentMap(t *testing.T) {
	c := unitTestJSONContent()
	b := make(map[string]*json.RawMessage)
	err := json.Unmarshal(c, &b)
	if err != nil {
		t.Fatalf("Got error while creating json.RawMessage dummy: %s", err)
	}

	p, err := NewJSONProvider(c)
	if err != nil {
		t.Fatalf("Got error while creating new JSON provier: %s", err)
	}

	m := p.GetContentMap()

	if !reflect.DeepEqual(*b["homepage"], m["homepage"]) {
		t.Fatalf("Homepage -> Got different value than expected. Expected %+v, got %+v", b["homepage"], m["homepage"])
	}
	if !reflect.DeepEqual(*b["archive"], m["archive"]) {
		t.Fatalf("Archive -> Got different value than expected. Expected %+v, got %+v", b["archive"], m["archive"])
	}
	if !reflect.DeepEqual(*b["repositories"], m["repositories"]) {
		t.Fatalf("Repositories -> Got different value than expected. Expected %+v, got %+v", b["repositories"], m["repositories"])
	}
}

func TestJSONProvider_GetStringSlice(t *testing.T) {
	expected := []string{
		"https://www.google.com/",
		"https://www.github.com/",
		"https://gobot.io/",
	}

	p, err := NewJSONProvider(unitTestJSONContent())
	if err != nil {
		t.Fatalf("Got error while creating new JSON provier: %s", err)
	}

	got := p.GetStringSlice("dummy-list")
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Got different value than expected. Expected %+v, got %+v", expected, got)
	}
}

func TestJSONProvider_GetStringSlice_NotExists(t *testing.T) {
	var expected []string

	p, err := NewJSONProvider(unitTestJSONContent())
	if err != nil {
		t.Fatalf("Got error while creating new JSON provier: %s", err)
	}

	got := p.GetStringSlice("not-exists")
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Got different value than expected. Expected %+v, got %+v", expected, got)
	}
}