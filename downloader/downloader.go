package downloader

type Downloader interface {
	Download(target string) error
}