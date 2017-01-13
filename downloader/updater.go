package downloader

type Updater interface {
	Update(target string) error
}
