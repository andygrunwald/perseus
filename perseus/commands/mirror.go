package commands

import (
	"fmt"
	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/packagist"
	"github.com/andygrunwald/perseus/perseus"
	"github.com/andygrunwald/perseus/types"
	"log"
	"strings"
	"sync"
)

// MirrorCommand reflects the business logic and the Command interface to mirror all configured packages.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type MirrorCommand struct {
	// Config is the main medusa configuration
	Config *config.Medusa
	// Log represents a logger to log messages
	Log *log.Logger
	// NumOfWorker is the number of worker used for concurrent actions (like resolving the dependency tree)
	NumOfWorker int

	wg sync.WaitGroup
}

// resolveResult is the result of an dependency resolve process of a single repository
type resolveResult struct {
	// Path reflects the file path of the repository to update like /tmp/perseus/git-mirror/symfony/console.git
	Package string
	// Err contains an error once there was one during the update process
	Err error
}

// Run is the business logic of MirrorCommand.
func (c *MirrorCommand) Run() error {
	c.wg = sync.WaitGroup{}

	repoList, err := c.Config.GetNamesOfRepositories()
	if err != nil {
		if config.IsNoRepositories(err) {
			c.Log.Printf("Config: %s", err)
		} else {
			c.Log.Println(err)
		}
	}

	repos := types.NewSet(repoList...)

	require := c.Config.GetRequire()
	jobs := make(chan string, len(require))
	results := make(chan resolveResult, 100)

	// We follow here a kind of typical queue / worker model
	// to resolve the dependencies.
	// Bootstrap the worker
	for w := 1; w <= c.NumOfWorker; w++ {
		c.wg.Add(1)
		go c.resolveWorker(w, jobs, results)
	}

	// The problem at this usecase for a typical worker / queue model is
	// that we don't know how many results we will get.
	// That way it is tricky to know when to close the results channel.
	// We solve it this way:
	// We bootstrap n worker (see above).
	// We combine those workers in a waitgroup.
	// Each worker will mark the process as Done in the waitgroup when the worker is finish.
	// We know how many jobs we put into the queue. So we can close the job channel respective.
	// When the job queue is empty AND all workers are Done, we can close the results channel :)
	go func() {
		c.wg.Wait()
		close(results)
	}()

	// Add all jobs to the worker queue
	for _, v := range require {
		jobs <- v
	}
	close(jobs)

	// Lets catch all results
	for p := range results {
		if p.Err != nil {
			c.Log.Println(p.Err)
			continue
		}

		repos.Add(p.Package)
	}

	fmt.Printf("Found %d entries: %+v\n", repos.Len(), repos.Flatten())

	fmt.Println("Called: func(c *MirrorCommand) Run()")
	panic("Not implemented yet: bin/medusa mirror [config]")

	/*

		No we have everything and we need to call the add command

		$output->writeln('<info>Create mirror repositories</info>');

	        foreach ($repos as $repo) {
	            $command = $this->getApplication()->find('add');

	            $arguments = array(
	                'command'     => 'add',
	                'package'     => $repo,
	                'config'      => $medusaConfig,
	            );

	            $input = new ArrayInput($arguments);
	            $returnCode = $command->run($input, $output);
	        }
	*/
	// TODO IMPLEMENT MirrorCommand

	return nil
}

func (c *MirrorCommand) resolveWorker(id int, jobs <-chan string, jobResults chan<- resolveResult) {
	defer c.wg.Done()

	pUrl := "https://packagist.org/"
	packagistClient, err := packagist.New(pUrl, nil)
	if err != nil {
		jobResults <- resolveResult{"", fmt.Errorf("Can't create packagist client for worker %d: %s", id, err)}
		return
	}

	var d perseus.DependencyResolver

	for j := range jobs {
		c.Log.Printf("Loading dependencies for package \"%s\" from %s", j, pUrl)

		// Lets get a dependency resolver.
		// If we can't bootstrap one, we are lost anyway.
		if d == nil {
			d, err = perseus.NewDependencyResolver(j, c.NumOfWorker, packagistClient)
			if err != nil {
				jobResults <- resolveResult{"", fmt.Errorf("Can't create dependency resolver for worker %d with package \"%s\": %s", id, j, err)}
				return
			}
		} else {
			d.SetPackage(j)
		}

		results := d.GetResultStream()
		go d.Start()

		dependencyNames := []string{}
		// Finally we collect all the results of the work.
		for r := range results {
			jobResults <- resolveResult{r.Package.Name, r.Error}
			if r.Package.Name == j {
				continue
			}

			dependencyNames = append(dependencyNames, r.Package.Name)
		}

		if l := len(dependencyNames); l == 0 {
			c.Log.Printf("%d dependencies found for package \"%s\" on %s", len(dependencyNames), j, pUrl)
		} else {
			c.Log.Printf("%d dependencies found for package \"%s\" on %s: %s", len(dependencyNames), j, pUrl, strings.Join(dependencyNames, ", "))
		}
	}
}
