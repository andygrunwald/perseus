package downloader

import (
	"io"

	"github.com/andygrunwald/perseus/perseus"
)

type Downloader interface {
	io.Closer

	Download(packages []*perseus.Package)
	GetResultStream() <-chan *Result
}
