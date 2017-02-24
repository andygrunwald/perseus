package perseus

import (
	"fmt"
	"strings"
	"sync"

	"github.com/andygrunwald/perseus/packagist"
	"github.com/andygrunwald/perseus/types"
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
	resolved *types.Set
	// queued is a storage to track which packages were already queued
	queued *types.Set
	// replacee is a hashmap to replace old/renamed/obsolete packages that would throw an error otherwise
	replacee map[string]string
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
		resolved:    types.NewSet(),
		queued:      types.NewSet(),
		Package:     packageName,
		packagist:   p,
		replacee:    getReplaceeMap(),
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

		// We overwrite specific packages, because they are added as dependencies to some older tags.
		// And those was renamed (for some reasons). But we are scanning all tags / branches.
		if r, ok := d.replacee[packageName]; ok {
			packageName = r
		}

		// Get information about the package from ApiClient
		p, resp, err := d.packagist.GetPackage(packageName)
		if err != nil {
			// API Call error here. Request to Packagist failed
			r := &Result{
				Package: j,
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
			for dependency, _ := range version.Require {
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
	d.resolved.Add(p)
	/*
		d.lock.Lock()
		defer d.lock.Unlock()
		d.resolved = append(d.resolved, p)
	*/
}

// markAsQueued will mark package p as queued.
func (d *PackagistDependencyResolver) markAsQueued(p string) {
	d.queued.Add(p)
	/*
		d.lock.Lock()
		defer d.lock.Unlock()
		d.queued = append(d.queued, p)
	*/
}

// isPackageAlreadyResolved returns true if package p was already resolved.
// False otherwise.
func (d *PackagistDependencyResolver) isPackageAlreadyResolved(p string) bool {
	return d.resolved.Exists(p)
}

// isPackageAlreadyQueued returns true if package p was already queued.
// False otherwise.
func (d *PackagistDependencyResolver) isPackageAlreadyQueued(p string) bool {
	return d.queued.Exists(p)
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

// getReplaceeMap will return a map of package names that needs to be replaced with some other package names.
// We need to do this, because those packages was renamed a long time ago.
// But the original packages are still part of older tags and branches.
// I guess, those packages were renamed before even packagist exists.
// That might explain the reason why there is no "obsolete"/"deprecated"-reference on packagist.
// Anyway. We need to deal with it.
// I tried so find some information why or when this packages were renamed.
// If you have more information about this, please ping me or make a PR and add this information
// as a simple comment. Might be useful for someone who is using this at a later stage.
func getReplaceeMap() map[string]string {
	m := map[string]string{
		"symfony/translator": "symfony/translation",

		// On January 2, 2012 they moved packages around.
		// One of them was the symfony/doctrine-bundle => doctrine/doctrine-bundle move.
		// Checkout the blogpost at https://symfony.com/blog/symfony-2-1-the-doctrine-bundle-has-moved-to-the-doctrine-organization
		//
		// Here are a few changes in packages / apps:
		//	- Symfony standard: https://github.com/symfony/symfony-standard/commit/5dee24eb280452fe46e76f99706c21ab417462ac
		//	- GeneratorBundle: https://github.com/sensiolabs/SensioGeneratorBundle/commit/11c9b68bd7f67a9cf3429d48af2d1a817dbd58cb#diff-b5d0ee8c97c7abd7e3fa29b9a27d1780
		//
		// Older tags still reference "symfony/doctrine-bundle".
		"symfony/doctrine-bundle":     "doctrine/doctrine-bundle",
		"metadata/metadata":           "jms/metadata",
		"zendframework/zend-registry": "zf1/zend-registry",
	}
	return m
}
