package downloader

import (
	"io"

	"github.com/andygrunwald/perseus/dependency"
)

// Downloader will take care about everything related to downloads / initial mirror.
type Downloader interface {
	io.Closer

	Download(packages []*dependency.Package)
	GetResultStream() <-chan *Result
}
