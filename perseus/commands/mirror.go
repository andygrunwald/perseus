package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/downloader"
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
	repos := types.NewSet()

	// Get list of manual entered repositories
	// and add them to the set
	repoList, err := c.Config.GetNamesOfRepositories()
	if err != nil {
		if config.IsNoRepositories(err) {
			c.Log.Printf("Config: %s", err)
		} else {
			c.Log.Println(err)
		}
	}

	for _, r := range repoList {
		repos.Add(r)
	}

	// Get all required repositories and resolve those dependencies
	pURL := "https://packagist.org/"
	packagistClient, err := packagist.New(pURL, nil)
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

	require := c.Config.GetRequire()
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

		repos.Add(p.Package)
	}

	c.Log.Printf("Start concurrent download process for %d packages with %d worker", repos.Len(), c.NumOfWorker)
	loader, err := downloader.NewGit(c.NumOfWorker, c.Config.GetString("repodir"))
	if err != nil {
		return err
	}

	loaderResults := loader.GetResultStream()
	flatten := repos.Flatten()
	loaderList := make([]*perseus.Package, 0, len(flatten))
	for _, item := range repos.Flatten() {
		loaderList = append(loaderList, item.(*perseus.Package))
	}
	loader.Download(loaderList)

	var satisRepositories []string
	for i := 1; i <= int(repos.Len()); i++ {
		v := <-loaderResults
		if v.Error != nil {
			if os.IsExist(v.Error) {
				c.Log.Printf("Package \"%s\" exists on disk. Try updating it instead. Skipping.", v.Package.Name)
			} else {
				c.Log.Printf("Error while mirroring package \"%s\": %s", v.Package.Name, v.Error)
				// If we have an error, we don't need to add it to satis repositories
				continue
			}
		} else {
			c.Log.Printf("Mirroring of package \"%s\" successful", v.Package.Name)
		}

		satisRepositories = append(satisRepositories, c.getLocalURLForRepository(v.Package.Name))
	}
	loader.Close()

	// And as a final step, write the satis configuration
	err = c.writeSatisConfig(satisRepositories...)
	return err
}

func (c *MirrorCommand) getLocalURLForRepository(p string) string {
	var r string

	satisURL := c.Config.GetString("satisurl")
	repoDir := c.Config.GetString("repodir")

	if len(satisURL) > 0 {
		r = fmt.Sprintf("%s/%s.git", satisURL, p)
	} else {
		t := fmt.Sprintf("%s/%s.git", repoDir, p)
		t = strings.TrimLeft(filepath.Clean(t), "/")
		r = fmt.Sprintf("file:///%s", t)
	}

	return r
}

func (c *MirrorCommand) writeSatisConfig(satisRepositories ...string) error {
	// Write Satis file
	satisConfig := c.Config.GetString("satisconfig")
	if len(satisConfig) == 0 {
		c.Log.Print("No Satis configuration specified. Skipping to write a satis configuration.")
		return nil
	}

	satisContent, err := ioutil.ReadFile(satisConfig)
	if err != nil {
		return fmt.Errorf("Can't read Satis configuration %s: %s", satisConfig, err)
	}

	j, err := config.NewJSONProvider(satisContent)
	if err != nil {
		return fmt.Errorf("Error while creating JSONProvider: %s", err)
	}

	s, err := config.NewSatis(j)
	if err != nil {
		return fmt.Errorf("Error while creating Satis object: %s", err)
	}

	s.AddRepositories(satisRepositories...)
	err = s.WriteFile(satisConfig, 0644)
	if err != nil {
		return fmt.Errorf("Writing Satis configuration to %s failed: %s", satisConfig, err)
	}

	c.Log.Printf("Satis configuration successful written to %s", satisConfig)
	return nil
}
