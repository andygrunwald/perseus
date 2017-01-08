package perseus

import (
	"strings"

	"github.com/andygrunwald/perseus/packagist"
)

type DependencyResolver interface {
	Resolve() []*Package
}

// DependencyResolver reflects the main structure of an Dependency Resolver :)
type PackagistDependencyResolver struct {
	// Package contains the package name like "twig/twig" or "symfony/console"
	Package string

	// Packagist is a Client to talk to a packagist instance
	Packagist *packagist.Client
}

// NewDependencyResolver will create a new DependencyResolver to resolve all `required`
// dependencies for packet
func NewDependencyResolver(packet string, p *packagist.Client) DependencyResolver {
	d := &PackagistDependencyResolver{
		Package:   packet,
		Packagist: p,
	}

	return d
}

func (d *PackagistDependencyResolver) Resolve() []*Package {
	// TODO This method can be speed up by using maps instead of the stringInSlice lookup

	// TODO This can be speed up by concurrency, but i don`t know if it is worth the effort + complexity
	// 	When it comes to concurrency it could be that we request a single package twice.
	//	If it good? If its bad? If it worth the effort / speed?
	//	I don`t know yet. But here are a few good ideas how to build in concurrency here:
	//
	//	- https://matt.aimonetti.net/posts/2012/11/27/real-life-concurrency-in-go/
	//	- http://blog.narenarya.in/concurrent-http-in-go.html
	//
	//	I like the idea of the result struct. But all of them forget to close the channel

	deps := []string{d.Package}
	resolved := []*Package{}
	var packet string

	for len(deps) > 0 {
		// Get the last element
		packet, deps = deps[len(deps)-1], deps[:len(deps)-1]

		// If the package name don`t contain a "/" we will skip it here.
		// In a composer.json in the require / require-dev part you normally add packaged
		// you depend on. A package name follows the format "vendor/package".
		// E.g. symfony/console
		// You can put other dependencies in here as well like `php` or `ext-zip`.
		// Those dependencies will be skipped (because they don`t have a vendor ;)).
		// The reason is simple: If you try to request the package "php" at packagist
		// you won`t get a JSON response with information we expect.
		// You will get valid HTML of the packagist search.
		// To avoid those errors and to save API calls we skip dependencies without a vendor.
		//
		// This follows the documentation as well:
		//
		// 	The package name consists of a vendor name and the project's name.
		// 	Often these will be identical - the vendor name just exists to prevent naming clashes.
		//	Source: https://getcomposer.org/doc/01-basic-usage.md
		if !strings.Contains(packet, "/") {
			continue
		}

		p, _, err := d.Packagist.GetPackage(packet)
		if err != nil {
			// TODO What to do when we have an api call error here? fix it
			panic(err)
		}
		if p == nil {
			// TODO What to do when nothing comes back from packagist? fix it
			panic("DependencyResolver -> Resolve: Nothing from packagist")
		}

		// Loop over versions
		for _, version := range p.Versions {
			// If we don` have required packaged, we can handle the next one
			if len(version.Require) == 0 {
				continue
			}

			for dependency, _ := range version.Require {
				if !stringInPacketlist(dependency, resolved) && !stringInSlice(dependency, deps) {
					deps = append(deps, dependency)
					// $deps = array_unique($deps);
				}
			}
		}

		resolvedPackage, err := NewPackage(p.Name)
		resolved = append(resolved, resolvedPackage)
	}

	return resolved
}

func stringInPacketlist(a string, list []*Package) bool {
	for _, b := range list {
		if b.Name == a {
			return true
		}
	}
	return false
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
