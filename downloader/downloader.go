package downloader

import (
	"regexp"
)

type Downloader interface {
	Download(target string) error
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
