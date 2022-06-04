package hsleaderboards

import (
	"log"
	"time"
)

type Scraper struct {
	Sites    []Site
	Db       *Database
	Schedule *time.Ticker
	Cfg      *Config
	Logger   *log.Logger
}

type Site interface {
	Name() string
	Initialize(*Scraper, *Database) error
	Scrape() error
}

func MakeScraper(db *Database, logger *log.Logger, cfg *Config) *Scraper {
	return &Scraper{
		Sites:    make([]Site, 0),
		Db:       db,
		Schedule: time.NewTicker(time.Duration(cfg.Interval) * time.Second),
		Cfg:      cfg,
		Logger:   logger,
	}
}

func (sc *Scraper) AddSite(site Site) {
	sc.Sites = append(sc.Sites, site)
}

func (sc *Scraper) initialize() {
	for _, site := range sc.Sites {
		err := site.Initialize(sc, sc.Db)
		if err != nil {
			sc.Logger.Fatalf("[Scraper] Failed Initializing %s, %s", site.Name(), err)
		}
		sc.Logger.Printf("[Scraper] Initialized %s", site.Name())
	}
}

func (sc *Scraper) Start() error {
	sc.initialize()
	sc.Logger.Println("[Scraper] Scraper Started.")
	for range sc.Schedule.C {
		for _, site := range sc.Sites {
			sc.Logger.Printf("[Scraper] Started Scraping %s", site.Name())
			err := site.Scrape()
			if err != nil {
				sc.Logger.Fatalf("[Scraper] Failed Scraping %s, %s", site.Name(), err)
			}
		}
	}
	return nil
}

func (sc *Scraper) Stop() {
	sc.Schedule.Stop()
}
