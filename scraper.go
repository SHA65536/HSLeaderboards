package hsleaderboards

import "log"

type Scraper struct {
	Sites  []Site
	Logger *log.Logger
}

func MakeScraper(db *Database, logger *log.Logger, cfg *Config) *Scraper {
	return &Scraper{}
}

func (sc *Scraper) AddSite(site Site) {
	err := site.Initialize(sc)
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
