package hsleaderboards

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//go:embed queries/create_bg.sql
var battlegrounds_create string

type Battlegrounds struct {
	URL           string
	Regions       []string
	Retries       int
	Db            *Database
	Sc            *Scraper
	CurrSnapshots map[string]*BGResponse
	PrevSnapshots map[string]*BGResponse
	Logger        *log.Logger
}

func (b *Battlegrounds) Name() string {
	return "Battlegrounds"
}

func (b *Battlegrounds) Initialize(sc *Scraper, db *Database) error {
	b.Sc = sc
	b.Db = db
	b.Logger = sc.Logger
	now := time.Now()
	// Creating DB tables
	_, err := db.Session.Exec(battlegrounds_create)
	if err != nil {
		return err
	}
	// Getting snapshots for comparison
	for _, region := range b.Regions {
		res, err := b.getResponse(region)
		if err != nil {
			return err
		}
		res.Timestamp = now.Unix()
		b.CurrSnapshots[region] = res
		b.PrevSnapshots[region] = res
	}
	return err
}

// Scrape gets data from all regions and saves to database
func (b *Battlegrounds) Scrape() error {
	now := time.Now()
	for _, region := range b.Regions {
		res, err := b.getResponse(region)
		if err != nil {
			b.Logger.Printf("[Battlegrounds] Failed to get region %s, %s", region, err)
			continue
		}
		res.Timestamp = now.Unix()
		new, old := b.saveDifferences(res)
		b.PrevSnapshots[region] = b.CurrSnapshots[region]
		b.CurrSnapshots[region] = res
		b.Logger.Printf("[Battlegrounds] Saved region %s. New: %d, Old: %d", region, new, old)
	}
	return nil
}

func MakeBattlegrounds() Site {
	return &Battlegrounds{
		URL:           "https://playhearthstone.com/en-gb/api/community/leaderboardsData?region=%s&leaderboardId=BG",
		Regions:       []string{"US", "EU", "AP"},
		Retries:       3,
		CurrSnapshots: make(map[string]*BGResponse),
		PrevSnapshots: make(map[string]*BGResponse),
	}
}

// getResponse gets the data for the specified region
// handles retrie
func (b *Battlegrounds) getResponse(region string) (*BGResponse, error) {
	var err error
	var myClient = &http.Client{Timeout: 10 * time.Second}
	var response = &BGResponse{}
	for i := 0; i < b.Retries; i++ {
		r, _ := myClient.Get(fmt.Sprintf(b.URL, region))
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			continue
		}
		defer r.Body.Close()
		err = json.Unmarshal(body, response)
		return response, err
	}
	return response, err
}

// TODO: implement comparison and saving
// saveDifferences compares a snapshot to the last snapshot
// and saves the differences into database
func (b *Battlegrounds) saveDifferences(res *BGResponse) (new, old int) {
	return
}
