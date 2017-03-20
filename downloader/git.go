package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

type GitDownloader struct {
	url string
}

// TODO We should think about it to make repositoryUrl a net/url.URL
func NewGit(repositoryUrl string) (Downloader, error) {
	reg, err := regexp.Compile("^git@github.com:")
	if err != nil {
		return nil, err
	}

	safeUrl := reg.ReplaceAllString(repositoryUrl, "git://github.com/")
	client := &GitDownloader{
		url: safeUrl,
	}
	return client, nil
}

func NewGitUpdater() (Updater, error) {
	client := &GitDownloader{}
	return client, nil
}

func (d *GitDownloader) Download(target string) error {
	if _, err := os.Stat(target); err != nil {
		if os.IsExist(err) {
			return nil
		}
	}

	// Initial clone
	err := d.clone(target)
	if err != nil {
		return err
	}

	err = d.updateServerInfo(target)
	if err != nil {
		return err
	}

	err = d.fsck(target)
	if err != nil {
		return err
	}

	return nil
}

func (d *GitDownloader) Update(target string) error {
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

func (d *GitDownloader) clone(target string) error {
	cmd := exec.Command("git", "clone", "--mirror", d.url, target)
	stdOut, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("Error during cmd \"%+v\". Process state: %s. stdOut: %s. stdErr: %s", cmd.Args, ee.String(), stdOut, ee.Stderr)
		}
		return fmt.Errorf("Error during cmd \"%+v\". stdOut: %s", cmd.Args, stdOut)
	}

	return nil
}

func (d *GitDownloader) fsck(target string) error {
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

func (d *GitDownloader) updateServerInfo(target string) error {
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

func (d *GitDownloader) fetch(target string) error {
	cmd := exec.Command("git", "fetch")
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
