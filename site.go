package hsleaderboards

type Site interface {
	Name() string
	Initialize(*Scraper) error
	Scrape() error
}
