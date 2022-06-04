package hsleaderboards

import (
	"log"
	"time"
)

// Scraper is the main scraper struct
// it handles scheduling the different
// site scraping functions
type Scraper struct {
	Sites    []Site
	Db       *Database
	Schedule *time.Ticker
	Cfg      *Config
	Logger   *log.Logger
}

// Site is the interface every different game mode implements
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

// AddSite adds a gamemode scraper to the list
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

// Start starts scraping the different sites
// This is blocking so call this in a goroutine
func (sc *Scraper) Start() error {
	sc.initialize()
	sc.Logger.Println("[Scraper] Scraper Started")
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

// Stop stops the scheduler
// active tasks may take some time stopping
func (sc *Scraper) Stop() {
	sc.Logger.Println("[Scraper] Scraper Stopping...")
	sc.Schedule.Stop()
}
