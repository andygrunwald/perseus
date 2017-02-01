package perseus

import (
	"fmt"
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

// Start will kick of the dependency resolver process.
func (d *PackagistDependencyResolver) Start() {
	// Boot up worker routines
	for w := 1; w <= d.workerCount; w++ {
		go d.worker(w, d.queue, d.results)
	}

	// Add the root package to the queue
	d.waitGroup.Add(1)
	p, _ := NewPackage(d.Package)
	d.queue <- p

	// Wait until all packages are resolved and close everything
	d.waitGroup.Wait()
	close(d.queue)
	close(d.results)
}

// worker is a single worker routine. This worker will be launched multiple times to work on
// the queue as efficient as possible.
// id the a id per worker (only for logging/debugging purpose).
// jobs is the jobs channel (the worker needs to be able to add more jobs to the queue as well).
// results is the channel where all results will be stored once they are resolved.
func (d *PackagistDependencyResolver) worker(id int, jobs chan<- *Package, results chan<- *Result) {
	// Worker has started. Lets do the hard work. Gimme the jobs.
	for j := range d.queue {
		packageName := j.Name

		// We don't need to process system packages.
		// System packages (like php or ext-curl) needs to be fulfilled by the system.
		// Not by the ApiClient
		if d.isSystemPackage(packageName) {
			d.waitGroup.Done()
			continue
		}

		// Overwrite a package here
		// TODO Fix this dirty hack here. Medusa does it exactly like this. PS: Why this is necessary at all?
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

		// TODO Respect response here
		// Get information about the package from ApiClient
		p, _, err := d.packagist.GetPackage(packageName)
		if err != nil {
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
		// Check if we got information from Packagist.
		// Maybe no error was thrown, but no package comes with the payload.
		// For us, as a dependency resolver, this is equal an error.
		if p == nil {
			// API Call error here. No package received from Packagist
			// TODO Add response code
			r := &Result{
				Package: j,
				Error:   fmt.Errorf("API Call to Packagist successful, but no package received"),
			}
			results <- r
			d.waitGroup.Done()
			continue
		}

		// Now we got the package.
		// Let us determine all requirements / dependencies from all versions,
		// because those packages needs to be resolved as well
		for _, version := range p.Versions {
			// If we don` have required packaged, we can handle the next one
			if len(version.Require) == 0 {
				continue
			}

			// Handle dependency per dependency
			for dependency, _ := range version.Require {
				// TODO Add a global check via Set is it a member
				// We check if this dependency was already queued.
				// It is typical that many different versions of one package don't
				// change dependencies so often. So we would queue one package
				// multiple times. With this small check we save a lot of work here.
				if d.shouldPackageBeQueued(dependency) {
					d.markAsQueued(dependency)

					packageToResolve, _ := NewPackage(dependency)
					// We add two additional waitgroup entries here.
					// You might ask why? Reguarly we add a new entry when we have a new package.
					// Here we add two, because of a) the new package and b) the new queue
					// entry of the package. We queue the package in a new go routine to
					// avoid a blocking state here. But we need to know when this go routine
					// is finished. So we observice this "Add package to queue" go routine
					// with the same waitgroup.
					d.waitGroup.Add(2)
					go func() {
						jobs <- packageToResolve
						d.waitGroup.Done()
					}()
				}
			}
		}

		// Package was resolved. Lets do everything which is necessary to change this package to a result.
		resolvedPackage, _ := NewPackage(p.Name)
		r := &Result{
			Package: resolvedPackage,
			Error:   nil,
		}
		results <- r
		d.waitGroup.Done()
		d.markAsResolved(p.Name)
	}
}

// markAsResolved will mark package p as resolved.
func (d *PackagistDependencyResolver) markAsResolved(p string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.resolved = append(d.resolved, p)
}

// markAsQueued will mark package p as queued.
func (d *PackagistDependencyResolver) markAsQueued(p string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.queued = append(d.queued, p)
}

// isPackageAlreadyResolved returns true if package p was already resolved.
// False otherwise.
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

// isPackageAlreadyQueued returns true if package p was already queued.
// False otherwise.
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

// shouldPackageBeQueued will return true if package p should be queued.
// False otherwise.
// A package should be queued if
// - it is not a system package
// - was not already queued
// - was not already resolved
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

// isSystemPackage returns true if p is a system package. False otherwise.
//
// A system package is a package that is not part of your package repository
// and and it needs to be fulfilled by the system.
// Examples: php, ext-curl
func (d *PackagistDependencyResolver) isSystemPackage(p string) bool {
	// If the package name don't contain a "/" we will skip it here.
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
