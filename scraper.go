package hsleaderboards

import "log"

type Scraper struct {
	Sites  []Site
	Db     *Database
	Logger *log.Logger
}

type Site interface {
	Name() string
	Initialize(*Scraper, *Database) error
	Scrape() error
}

func MakeScraper(db *Database, logger *log.Logger, cfg *Config) *Scraper {
	return &Scraper{
		Sites:  make([]Site, 0),
		Db:     db,
		Logger: logger,
	}
}

func (sc *Scraper) AddSite(site Site) {
	err := site.Initialize(sc, sc.Db)
	if err != nil {
		sc.Logger.Fatal(err)
	}
	sc.Sites = append(sc.Sites, site)
}

func (sc *Scraper) Start() error {
	return nil
}

func (sc *Scraper) Stop() {
}
