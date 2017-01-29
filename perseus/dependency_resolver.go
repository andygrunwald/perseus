package perseus

import (
	"strings"
	"sync"
	"fmt"
	"log"

	"github.com/andygrunwald/perseus/packagist"
)

type DependencyResolver interface {
	Start()
	GetResults() chan *ResolverResult
}

type PackagistDependencyResolver struct {
	NumWorker int
	WaitGroup sync.WaitGroup
	lock      sync.RWMutex
	Jobs      chan *Package
	results   chan *ResolverResult
	Resolved  []string
	Queued    []string

	// Package contains the package name like "twig/twig" or "symfony/console"
	Package   string

	// Packagist is a Client to talk to a packagist instance
	Packagist *packagist.Client
}

type ResolverResult struct {
	Package *Package
	Error error
}

func NewDependencyResolver(packet string, worker int, p *packagist.Client) DependencyResolver {
	d := &PackagistDependencyResolver{
		NumWorker: worker,
		WaitGroup: sync.WaitGroup{},
		lock: sync.RWMutex{},
		Jobs: make(chan *Package, 4),
		results: make(chan *ResolverResult),
		Resolved: []string{},
		Package:   packet,
		Packagist: p,
	}

	return d
}

func (d *PackagistDependencyResolver) GetResults() chan *ResolverResult {
	return d.results
}

func (d *PackagistDependencyResolver) Start() {
	for w := 1; w <= d.NumWorker; w++ {
		go d.worker(w, d.Jobs, d.results)
	}

	d.WaitGroup.Add(1)
	p, _ := NewPackage(d.Package)
	d.Jobs <- p

	d.WaitGroup.Wait()
	close(d.Jobs)
	close(d.results)
}

func (d *PackagistDependencyResolver) worker(id int, jobs chan *Package, results chan<- *ResolverResult) {
	//time.Sleep(5 * time.Second)
	log.Printf("Worker %d: Started", id)
	for j := range d.Jobs {
		packageName := j.Name
		log.Printf("Worker %d: With job %s", id, packageName)
		if d.isSystemPackage(packageName) {
			d.WaitGroup.Done()
			log.Printf("Worker %d: With job %s skipped", id, packageName)
			continue
		}

		// Overwrite a package here
		// TODO Fix this dirty hack here. Medusa does it exactly like this.
		// We overwrite packages, because they are added as dependencies to some
		// Maybe we should just skip it
		if packageName == "symfony/translator" {
			packageName = "symfony/translation"
		}
		if packageName == "symfony/doctrine-bundle" {
			packageName = "doctrine/doctrine-bundle"
		}
		if packageName == "metadata/metadata" {
			packageName = "jms/metadata"
		}
		if packageName == "zendframework/zend-registry" {
			packageName = "zf1/zend-registry"
		}

		p, _, err := d.Packagist.GetPackage(packageName)
		log.Printf("Worker %d: Got packagist response for %s", id, packageName)
		if err != nil {
			log.Printf("Worker %d: Failed result #1 sent for package %s", id, packageName)
			// API Call error here. Request to Packagist failed
			// TODO Maybe a little bit more information? Return code?
			r := &ResolverResult{
				Package: j,
				Error: err,
			}
			results <- r
			d.WaitGroup.Done()
			continue
		}
		if p == nil {
			log.Printf("Worker %d: Failed result #2 sent for package %s", id, packageName)
			// API Call error here. No package received from Packagist
			r := &ResolverResult{
				Package: j,
				Error: fmt.Errorf("API Call to Packagist successful, but o package received"),
			}
			results <- r
			d.WaitGroup.Done()
			continue
		}

		// Loop over versions
		for _, version := range p.Versions {
			// If we don` have required packaged, we can handle the next one
			if len(version.Require) == 0 {
				log.Printf("Worker %d: Nothing to require for package %s", id, packageName)
				continue
			}

			for dependency, _ := range version.Require {
				// TODO Add a global check via Set is it a member
				if d.shouldPackageBeQueued(dependency) {
					log.Printf("Worker %d: New package queued %s -> %s", id, packageName, dependency)
					d.markAsQueued(dependency)

					packageToResolve, _ := NewPackage(dependency)
					// 2 wegen einmal dem Package und einmal der neuen Go-Routine
					d.WaitGroup.Add(2)
					log.Printf("Worker %d: New package queued (before) %s -> %s", id, packageName, dependency)

					go func(){
						jobs <- packageToResolve
						d.WaitGroup.Done()
					}()
					log.Printf("Worker %d: New package queued (after) %s -> %s", id, packageName, dependency)
					/*
					d.WaitGroup.Add(1)
					log.Printf("Worker %d: New package queued (before) %s -> %s", id, packageName, dependency)
					// This can block if the channel has not a big buffer
					// Reason ist einfach: Alle adden packete und blocken hier.
					jobs <- packageToResolve
					log.Printf("Worker %d: New package queued (after) %s -> %s", id, packageName, dependency)
					 */
				}
			}
		}

		log.Printf("Worker %d: Package resolved %s", id, p.Name)
		resolvedPackage, _ := NewPackage(p.Name)
		r := &ResolverResult{
			Package: resolvedPackage,
			Error: nil,
		}
		results <- r
		d.WaitGroup.Done()
		d.markAsResolved(p.Name)
	}
	log.Printf("Worker %d: done", id)
}

func (d *PackagistDependencyResolver) markAsResolved(p string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.Resolved = append(d.Resolved, p)
}

func (d *PackagistDependencyResolver) markAsQueued(p string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.Queued = append(d.Queued, p)
}

func (d *PackagistDependencyResolver) isPackageAlreadyResolved(p string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	for _, b := range d.Resolved {
		if b == p {
			return true
		}
	}
	return false
}

func (d *PackagistDependencyResolver) isPackageAlreadyQueued(p string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	for _, b := range d.Queued {
		if b == p {
			return true
		}
	}
	return false
}

func (d *PackagistDependencyResolver) shouldPackageBeQueued(p string) bool {
	if d.isSystemPackage(p) {
		return false
	}

	if d.isPackageAlreadyQueued(p) {
		return false
	}

	if d.isPackageAlreadyResolved(p) {
		return false
	}

	return true
}

func (d *PackagistDependencyResolver) isSystemPackage(p string) bool {
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
	return !strings.Contains(p, "/")
}