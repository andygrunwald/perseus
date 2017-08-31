package dependency

import (
	"fmt"
	"strings"
	"sync"

	"github.com/andygrunwald/perseus/dependency/repository"
	"github.com/andygrunwald/perseus/types/set"
)

// ComposerResolver is an implementation of Resolver for Composer (PHP)
type ComposerResolver struct {
	// repository is the Client to talk to a specific endpoint (e.g. Packagist)
	repository repository.Client

	// workerCount is the number of worker that will be started
	workerCount int
	waitGroup   sync.WaitGroup
	// queue is the channel where all jobs are stored that needs to be processed by the worker
	queue chan *Package
	// results is the channel where all resolved dependencies will be streamed
	results chan *Result
	// resolved is a storage to track which packages are already resolved
	resolved *set.Set
	// queued is a storage to track which packages were already queued
	queued *set.Set
	// replacee is a hashmap to replace old/renamed/obsolete packages that would throw an error otherwise
	replacee map[string]string
}

// GetResultStream will return the channel for results.
// During the process of resolving dependencies, this channel will be filled
// with the results. Those can be processed next to the resolve process.
func (d *ComposerResolver) GetResultStream() <-chan *Result {
	return d.results
}

// Resolve will start of the dependency resolver process.
func (d *ComposerResolver) Resolve(packageList []*Package) {
	d.startWorker()

	// Queue packages
	for _, p := range packageList {
		d.queuePackage(p)
	}

	// Wait until all packages are resolved and close everything
	d.waitGroup.Wait()
	close(d.queue)
	close(d.results)
}

// QueuePackage adds package p to the queue
func (d *ComposerResolver) queuePackage(p *Package) {
	d.waitGroup.Add(1)
	d.markAsQueued(p.Name)
	d.queue <- p
}

// startWorker will boot up the worker routines
func (d *ComposerResolver) startWorker() {
	for w := 1; w <= d.workerCount; w++ {
		go d.worker(w, d.queue, d.results)
	}
}

// worker is a single worker routine. This worker will be launched multiple times to work on
// the queue as efficient as possible.
// id is a unique number assigned per worker (only for logging/debugging purpose).
// jobs is the jobs channel. The worker needs to be able to add more jobs to the queue as well.
// results is the channel where all results will be stored once they are resolved.
func (d *ComposerResolver) worker(id int, queue chan<- *Package, results chan<- *Result) {
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

		// We overwrite specific packages, because they are added as dependencies to some older tags.
		// And those was renamed (for some reasons). But we are scanning all tags / branches.
		if r, ok := d.replacee[packageName]; ok {
			packageName = r
		}

		// Get information about the package from ApiClient
		p, resp, err := d.repository.GetPackageByName(packageName)
		if err != nil {
			// API Call error here. Request to Packagist failed
			r := &Result{
				Package: j,
				Response: resp,
				Error:   fmt.Errorf("API returned status code %d: %s", resp.StatusCode, err),
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
			r := &Result{
				Package: j,
				Response: resp,
				Error:   fmt.Errorf("API Call to Packagist successful (Status code %d), but no package received", resp.StatusCode),
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
			for dependency := range version.Require {
				// We check if this dependency was already queued.
				// It is typical that many different versions of one package don't
				// change dependencies so often. So we would queue one package
				// multiple times. With this small check we save a lot of work here.
				if d.shouldPackageBeQueued(dependency) {
					d.markAsQueued(dependency)

					packageToResolve, _ := NewPackage(dependency, "")
					// We add two additional waitgroup entries here.
					// You might ask why? Regularly we add a new entry when we have a new package.
					// Here we add two, because of a) the new package and b) the new queue
					// entry of the package. We queue the package in a new go routine to
					// avoid a blocking state here. But we need to know when this go routine
					// is finished. So we observice this "Add package to queue" go routine
					// with the same waitgroup.
					d.waitGroup.Add(2)
					go func() {
						queue <- packageToResolve
						d.waitGroup.Done()
					}()
				}
			}
		}

		// Package was resolved. Lets do everything which is necessary to change this package to a result.
		resolvedPackage, err := NewPackage(p.Name, p.Repository)
		r := &Result{
			Package: resolvedPackage,
			Response: resp,
			Error:   err,
		}
		results <- r
		d.waitGroup.Done()
		d.markAsResolved(p.Name)
	}
}

// markAsResolved will mark package p as resolved.
func (d *ComposerResolver) markAsResolved(p string) {
	d.resolved.Add(p)
}

// markAsQueued will mark package p as queued.
func (d *ComposerResolver) markAsQueued(p string) {
	d.queued.Add(p)
}

// shouldPackageBeQueued will return true if package p should be queued.
// False otherwise.
// A package should be queued if
// - it is not a system package
// - was not already queued
// - was not already resolved
func (d *ComposerResolver) shouldPackageBeQueued(p string) bool {
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

// isPackageAlreadyResolved returns true if package p was already resolved.
// False otherwise.
func (d *ComposerResolver) isPackageAlreadyResolved(p string) bool {
	return d.resolved.Exists(p)
}

// isPackageAlreadyQueued returns true if package p was already queued.
// False otherwise.
func (d *ComposerResolver) isPackageAlreadyQueued(p string) bool {
	return d.queued.Exists(p)
}

// isSystemPackage returns true if p is a system package. False otherwise.
//
// A system package is a package that is not part of your package repository
// and and it needs to be fulfilled by the system.
// Examples: php, ext-curl
func (d *ComposerResolver) isSystemPackage(p string) bool {
	// If the package name don't contain a "/" we will skip it here.
	// In a composer.json in the require / require-dev part you normally add packages
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
