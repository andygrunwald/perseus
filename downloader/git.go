package downloader

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/andygrunwald/perseus/dependency"
)

// Git represents an Updater and Downloader for the git protocol
type Git struct {
	// workerCount is the number of worker that will be started
	workerCount int

	// Directory where to download the data into
	dir string

	// queue is the queue channel where all jobs are stored that needs to be processed by the worker
	queue chan *dependency.Package
	// results is the channel where all resolved dependencies will be streamed
	results chan *Result
}

// Result reflects a result of a concurrent download process.
type Result struct {
	Package *dependency.Package
	Error   error
}

// NewGitDownloader creates a new downloader based on the git protocol.
// numOfWorker initiates the number of workers we should spawn to work concurrent.
// dir is the base directory where the downloads will be mirrored, too.
func NewGitDownloader(numOfWorker int, dir string) (Downloader, error) {
	if numOfWorker == 0 {
		return nil, fmt.Errorf("Starting a concurrent git downloader with zero worker is not possible")
	}

	c := &Git{
		workerCount: numOfWorker,
		dir:         dir,
		queue:       make(chan *dependency.Package, (numOfWorker + 1)),
		results:     make(chan *Result),
	}
	return c, nil
}

// GetResultStream will return the results stream.
// During the process of downloading git repositories, this channel will be filled
// with the results.
func (d *Git) GetResultStream() <-chan *Result {
	return d.results
}

// Close will close the process
func (d *Git) Close() error {
	close(d.results)
	return nil
}

// Download will start the concurrent download process
func (d *Git) Download(packages []*dependency.Package) {
	// Start the worker
	for w := 1; w <= d.workerCount; w++ {
		go d.worker(w, d.queue, d.results)
	}

	// Queue the downloads
	for _, p := range packages {
		d.queue <- p
	}
	close(d.queue)
}

// worker is a single worker routine. This worker will be launched multiple times to work on
// the queue as efficient as possible.
// id the a id per worker (only for logging/debugging purpose).
// jobs is the jobs channel (the worker needs to be able to read the jobs).
// results is the channel where all results will be stored once they are resolved.
func (d *Git) worker(id int, jobs <-chan *dependency.Package, results chan<- *Result) {
	for j := range jobs {
		targetDir := fmt.Sprintf("%s/%s.git", d.dir, j.Name)

		// Check if directory already exists
		_, err := os.Stat(targetDir)
		if err == nil {
			// Directory exists
			r := &Result{
				Package: j,
				Error:   os.ErrExist,
			}
			results <- r
			continue
		}

		// Initial clone
		err = d.clone(j.Repository.String(), targetDir)
		if err != nil {
			r := &Result{
				Package: j,
				Error:   err,
			}
			results <- r
			continue
		}

		err = d.updateServerInfo(targetDir)
		if err != nil {
			r := &Result{
				Package: j,
				Error:   err,
			}
			results <- r
			continue
		}

		err = d.fsck(targetDir)
		if err != nil {
			r := &Result{
				Package: j,
				Error:   err,
			}
			results <- r
			continue
		}

		// Everything successful downloaded
		r := &Result{
			Package: j,
			Error:   nil,
		}
		results <- r
	}
}

func (d *Git) clone(repository, target string) error {
	cmd := exec.Command("git", "clone", "--mirror", repository, target)
	stdOut, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("Error during cmd \"%+v\". Process state: %s. stdOut: %s. stdErr: %s", cmd.Args, ee.String(), stdOut, ee.Stderr)
		}
		return fmt.Errorf("Error during cmd \"%+v\". stdOut: %s", cmd.Args, stdOut)
	}

	return nil
}

func (d *Git) fsck(target string) error {
	// Firing a git file system check.
	// This was originally introduced, because on of the KDE git mirrors has problems.
	// See https://github.com/instaclick/medusa/issues/6
	cmd := exec.Command("git", "fsck")
	cmd.Dir = target
	stdOut, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("Error during cmd \"%+v\". Process state: %s. stdOut: %s. stdErr: %s", cmd.Args, ee.String(), stdOut, ee.Stderr)
		}
		return fmt.Errorf("Error during cmd \"%+v\". stdOut: %s", cmd.Args, stdOut)
	}

	return nil
}

func (d *Git) updateServerInfo(target string) error {
	// Lets be save and fire a update-server-info
	// This is useful if the remote server don`t support on-the-fly pack generations.
	// See `git help update-server-info`
	// See https://github.com/instaclick/medusa/commit/ff4270f56afacf0a788b8b192e76180fbe32452e#diff-74b630cd9501803fdde532d1e2128e2f
	cmd := exec.Command("git", "update-server-info", "-f")
	cmd.Dir = target
	stdOut, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("Error during cmd \"%+v\". Process state: %s. stdOut: %s. stdErr: %s", cmd.Args, ee.String(), stdOut, ee.Stderr)
		}
		return fmt.Errorf("Error during cmd \"%+v\". stdOut: %s", cmd.Args, stdOut)
	}

	return nil
}

// NewGitUpdater created a new updater based on the git protocol.
//
// TODO Make me concurrent
func NewGitUpdater() (Updater, error) {
	client := &Git{}
	return client, nil
}

// Update updates target with a simple `git fetch`.
func (d *Git) Update(target string) error {
	err := d.fetch(target)
	if err != nil {
		return err
	}

	err = d.updateServerInfo(target)
	if err != nil {
		return err
	}

	return nil
}

func (d *Git) fetch(target string) error {
	cmd := exec.Command("git", "fetch", "--prune")
	cmd.Dir = target
	stdOut, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("Error during cmd \"%+v\". Process state: %s. stdOut: %s. stdErr: %s", cmd.Args, ee.String(), stdOut, ee.Stderr)
		}
		return fmt.Errorf("Error during cmd \"%+v\". stdOut: %s", cmd.Args, stdOut)
	}

	return nil
}
