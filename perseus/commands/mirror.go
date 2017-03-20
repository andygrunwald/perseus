package commands

import (
	"fmt"
	"log"
	"sync"

	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/packagist"
	"github.com/andygrunwald/perseus/perseus"
	"github.com/andygrunwald/perseus/types"
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

	pUrl := "https://packagist.org/"
	packagistClient, err := packagist.New(pUrl, nil)
	if err != nil {
		c.Log.Println(err)
	}

	// Lets get a dependency resolver.
	// If we can't bootstrap one, we are lost anyway.
	// We set the queue length to the number of workers + 1. Why?
	// With this every worker has work, when the queue is filled.
	// During the add command, this is enough in most of the cases.
	d, err := perseus.NewDependencyResolver(c.NumOfWorker, packagistClient)
	if err != nil {
		return err
	}
	results := d.GetResultStream()

	// Loop over the packages and add them
	l := []*perseus.Package{}
	for _, r := range require {
		p, _ := perseus.NewPackage(r, "")
		l = append(l, p)
	}

	go d.Resolve(l)

	// Finally we collect all the results of the work.
	for p := range results {
		if p.Error != nil {
			c.Log.Println(p.Error)
			continue
		}

		repos.Add(p.Package.Name)
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

		// TODO IMPLEMENT MirrorCommand
		// TODO And make satis write file at the end
		// TODO and make a worker / queue implementation

		// Okay, now it gets wired, but we will call the Add command for every package.
		//
		for _, packet := range repos.Flatten() {
			c := &AddCommand{
				Package:          packet,
				// We don't need dependencies here, because we had resolved them already
				WithDependencies: false,
				Config:           c.Config,
				Log:              c.Log,
				NumOfWorker:      c.NumOfWorker,
			}
			err = c.Run()
			if err != nil {
				return fmt.Errorf("Error during execution of \"add\" command: %s\n", err)
			}
		}
	*/
	return nil
}
