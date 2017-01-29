package perseus

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/andygrunwald/perseus/packagist"
)

// DependencyResolver is an interface to resolve package dependencies
type DependencyResolver interface {
	Start()
	GetResultStream() <-chan *Result
}

// PackagistDependencyResolver is an implementation of DependencyResolver
// that will resolve the dependencies of a package with the help of https://packagist.org/
type PackagistDependencyResolver struct {
	// Package contains the package name like "twig/twig" or "symfony/console"
	Package string
	// packagist is a Client to talk to a packagist instance
	packagist packagist.ApiClient

	// workerCount is the number of worker that will be started
	workerCount int
	waitGroup   sync.WaitGroup
	lock        sync.RWMutex
	// queue is the queue channel where all jobs are stored that needs to be processed by the worker
	queue chan *Package
	// results is the channel where all resolved dependencies will be streamed
	results chan *Result
	// resolved is a storage to track which packages are already resolved
	resolved []string
	// queued is a storage to track which packages were already queued
	queued []string
}

// Result reflects a result of a dependency resolver process.
type Result struct {
	Package *Package
	Error   error
}

// NewDependencyResolver will create a new instance of a DependencyResolver.
// Standard implementation is the PackagistDependencyResolver.
func NewDependencyResolver(packageName string, numOfWorker int, p packagist.ApiClient) (DependencyResolver, error) {
	if len(packageName) == 0 {
		return nil, fmt.Errorf("No package name given.")
	}
	if numOfWorker == 0 {
		return nil, fmt.Errorf("Starting a dependency resolver with zero worker is not possible")
	}
	if p == nil {
		return nil, fmt.Errorf("Starting a dependency resolver with an empty ApiClient is not possible")
	}

	d := &PackagistDependencyResolver{
		workerCount: numOfWorker,
		waitGroup:   sync.WaitGroup{},
		lock:        sync.RWMutex{},
		queue:       make(chan *Package, 4),
		results:     make(chan *Result),
		resolved:    []string{},
		Package:     packageName,
		packagist:   p,
	}

	return d, nil
}

// GetResultStream will return the results stream.
// During the process of resolving dependencies, this channel will be filled
// with the results. Those can be processed next to the resolve process.
func (d *PackagistDependencyResolver) GetResultStream() <-chan *Result {
	return d.results
}

func (d *PackagistDependencyResolver) Start() {
	for w := 1; w <= d.workerCount; w++ {
		go d.worker(w, d.queue, d.results)
	}

	d.waitGroup.Add(1)
	p, _ := NewPackage(d.Package)
	d.queue <- p

	d.waitGroup.Wait()
	close(d.queue)
	close(d.results)
}

func (d *PackagistDependencyResolver) worker(id int, jobs chan *Package, results chan<- *Result) {
	log.Printf("Worker %d: Started", id)
	for j := range d.queue {
		packageName := j.Name
		log.Printf("Worker %d: With job %s", id, packageName)
		if d.isSystemPackage(packageName) {
			d.waitGroup.Done()
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

		p, _, err := d.packagist.GetPackage(packageName)
		log.Printf("Worker %d: Got packagist response for %s", id, packageName)
		if err != nil {
			log.Printf("Worker %d: Failed result #1 sent for package %s", id, packageName)
			// API Call error here. Request to Packagist failed
			// TODO Maybe a little bit more information? Return code?
			r := &Result{
				Package: j,
				Error:   err,
			}
			results <- r
			d.waitGroup.Done()
			continue
		}
		if p == nil {
			log.Printf("Worker %d: Failed result #2 sent for package %s", id, packageName)
			// API Call error here. No package received from Packagist
			r := &Result{
				Package: j,
				Error:   fmt.Errorf("API Call to Packagist successful, but o package received"),
			}
			results <- r
			d.waitGroup.Done()
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
					d.waitGroup.Add(2)
					log.Printf("Worker %d: New package queued (before) %s -> %s", id, packageName, dependency)

					go func() {
						jobs <- packageToResolve
						d.waitGroup.Done()
					}()
					log.Printf("Worker %d: New package queued (after) %s -> %s", id, packageName, dependency)
				}
			}
		}

		log.Printf("Worker %d: Package resolved %s", id, p.Name)
		resolvedPackage, _ := NewPackage(p.Name)
		r := &Result{
			Package: resolvedPackage,
			Error:   nil,
		}
		results <- r
		d.waitGroup.Done()
		d.markAsResolved(p.Name)
	}
	log.Printf("Worker %d: done", id)
}

func (d *PackagistDependencyResolver) markAsResolved(p string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.resolved = append(d.resolved, p)
}

func (d *PackagistDependencyResolver) markAsQueued(p string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.queued = append(d.queued, p)
}

func (d *PackagistDependencyResolver) isPackageAlreadyResolved(p string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	for _, b := range d.resolved {
		if b == p {
			return true
		}
	}
	return false
}

func (d *PackagistDependencyResolver) isPackageAlreadyQueued(p string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	for _, b := range d.queued {
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
