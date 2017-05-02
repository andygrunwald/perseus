package controller

import (
	"fmt"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/perseus/config"
	"github.com/andygrunwald/perseus/downloader"
)

// UpdateController reflects the business logic and the Command interface to update all packages that were added or mirrored in the past.
// This command is independent from an human interface (CLI, HTTP, etc.)
// The human interfaces will interact with this command.
type UpdateController struct {
	// Config is the main medusa configuration
	Config *config.Medusa
	// Log represents a logger to log messages
	Log logrus.FieldLogger
	// NumOfWorker is the number of worker used for concurrent actions (like updating git repositories)
	NumOfWorker int
}

// updateResult is the result of an update process of a single repository
type updateResult struct {
	// Path reflects the file path of the repository to update like /tmp/perseus/git-mirror/symfony/console.git
	Path string
	// Err contains an error once there was one during the update process
	Err error
}

// Run is the business logic of UpdateCommand.
func (c *UpdateController) Run() error {
	repoDir := c.Config.GetString("repodir")

	p := fmt.Sprintf("%s/*/*.git", repoDir)
	matches, err := filepath.Glob(p)
	if err != nil {
		return fmt.Errorf("Error while determining folders for updating: %s", err)
	}

	// If no repositories were found, we will exit here
	if len(matches) == 0 {
		c.Log.WithFields(logrus.Fields{
			"path": p,
		}).Info("No repositories found")
		return nil
	}

	// We run the update process concurrent.
	// We will boot up a small worker pool and adding all repositories that we want to update.
	// Let the show begin
	jobs := make(chan string, len(matches))
	results := make(chan updateResult, len(matches))
	for w := 1; w <= c.NumOfWorker; w++ {
		go c.worker(w, jobs, results)
	}

	for _, v := range matches {
		jobs <- v
	}
	close(jobs)

	// Now lets have a look at all results and log them.
	for a := 1; a <= len(matches); a++ {
		r := <-results
		if r.Err != nil {
			c.Log.WithFields(logrus.Fields{
				"path": r.Path,
			}).WithError(r.Err).Info("Error while updating")
		} else {
			c.Log.WithFields(logrus.Fields{
				"path": r.Path,
			}).Info("Update successful")
		}
	}

	return nil
}

// worker is a single worker of the UpdateCommand.
// Workers job is to update a bunch of repositories on disk.
func (c *UpdateController) worker(id int, jobs <-chan string, results chan<- updateResult) {
	for j := range jobs {
		updateClient, err := downloader.NewGitUpdater()
		if err != nil {
			results <- updateResult{Path: j, Err: fmt.Errorf("Updater client creation failed for package %s: %s", j, err)}
			continue
		}
		err = updateClient.Update(j)
		if err != nil {
			results <- updateResult{Path: j, Err: err}
		} else {
			results <- updateResult{Path: j, Err: nil}
		}
	}
}
