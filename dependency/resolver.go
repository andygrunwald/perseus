package dependency

import (
	"fmt"
	"sync"

	"github.com/andygrunwald/perseus/dependency/repository"
	"github.com/andygrunwald/perseus/types/set"
)

// Resolver is an interface to resolve package dependencies
type Resolver interface {
	Resolve(packageList []*Package)
	GetResultStream() <-chan *Result
}

// Result reflects a result of a dependency resolver process.
type Result struct {
	Package *Package
	Error   error
}

// NewComposerResolver will create a new instance of a Resolver.
// Standard implementation is the ComposerResolver.
func NewComposerResolver(numOfWorker int, p repository.Client) (Resolver, error) {
	if numOfWorker == 0 {
		return nil, fmt.Errorf("Starting a dependency resolver with zero worker is not possible")
	}
	if p == nil {
		return nil, fmt.Errorf("Starting a dependency resolver with an empty repository.Client is not possible")
	}

	d := &ComposerResolver{
		workerCount: numOfWorker,
		waitGroup:   sync.WaitGroup{},
		queue:       make(chan *Package, (numOfWorker + 1)),
		results:     make(chan *Result),
		resolved:    set.New(),
		queued:      set.New(),
		repository:  p,
		replacee:    getReplaceeMap(),
	}

	return d, nil
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
//
// TODO: This is packagist related. Move to packagist
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