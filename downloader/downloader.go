package downloader

import (
	"io"

	"github.com/andygrunwald/perseus/dependency"
)

type Downloader interface {
	io.Closer

	Download(packages []*dependency.Package)
	GetResultStream() <-chan *Result
}
