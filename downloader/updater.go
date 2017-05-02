package downloader

import (
	"io"
)

// Updater will take care about everything related to updates.
type Updater interface {
	io.Closer

	Update(target string) error
}
