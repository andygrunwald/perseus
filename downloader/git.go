package downloader

import (
	"fmt"
	"os"
	"os/exec"
)

type GitDownloader struct {
	url string
}

func (d *GitDownloader) Download(target string) error {
	if _, err := os.Stat(target); err != nil {
		if os.IsExist(err) {
			return nil
		}
	}

	// Initial clone
	cmd := exec.Command("git", "clone", "--mirror", d.url, target)
	stdOut, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("Error during cmd \"%+v\". Process state: %s. stdOut: %s. stdErr: %s", cmd.Args, ee.String(), stdOut, ee.Stderr)
		}
		return fmt.Errorf("Error during cmd \"%+v\". stdOut: %s", cmd.Args, stdOut)
	}

	// TODO Do we need this? Medusa implemented this, but i don't know why yet
	cmd = exec.Command("git", "update-server-info", "-f")
	cmd.Dir = target
	stdOut, err = cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("Error during cmd \"%+v\". Process state: %s. stdOut: %s. stdErr: %s", cmd.Args, ee.String(), stdOut, ee.Stderr)
		}
		return fmt.Errorf("Error during cmd \"%+v\". stdOut: %s", cmd.Args, stdOut)
	}

	// TODO Lets have a deeper look what it does, what it means, and why this was implemented at all by Medusa
	cmd = exec.Command("git", "fsck")
	cmd.Dir = target
	stdOut, err = cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("Error during cmd \"%+v\". Process state: %s. stdOut: %s. stdErr: %s", cmd.Args, ee.String(), stdOut, ee.Stderr)
		}
		return fmt.Errorf("Error during cmd \"%+v\". stdOut: %s", cmd.Args, stdOut)
	}

	return nil
}
