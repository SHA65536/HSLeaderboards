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

//go:embed queries/bg_create.sql
var battlegrounds_create string

//go:embed queries/bg_new.sql
var battlegrounds_new string

//go:embed queries/bg_update.sql
var battlegrounds_update string

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
	var res *BGResponse
	var err error
	b.Sc = sc
	b.Db = db
	b.Logger = sc.Logger
	now := time.Now()
	// Creating DB tables
	_, err = db.Session.Exec(battlegrounds_create)
	if err != nil {
		return err
	}
	// Getting snapshots for comparison
	for _, region := range b.Regions {
		res, err = b.getResponse(region)
		if err != nil {
			return err
		}
		res.Timestamp = now.Unix()
		b.CurrSnapshots[region] = res
		b.PrevSnapshots[region] = res
	}
	b.Logger.Printf("[Battlegrounds] Season: %d", res.Season)
	return err
}

// Scrape gets data from all regions and saves to database
func (b *Battlegrounds) Scrape() error {
	now := time.Now()
	for _, region := range b.Regions {
		start := time.Now()
		res, err := b.getResponse(region)
		if err != nil {
			b.Logger.Printf("[Battlegrounds] Failed to get region %s, %s", region, err)
			continue
		}
		res.Timestamp = now.Unix()
		new, old := b.saveDifferences(res)
		b.PrevSnapshots[region] = b.CurrSnapshots[region]
		b.CurrSnapshots[region] = res
		b.Logger.Printf("[Battlegrounds] Saved region %s. New: %d, Old: %d | Took %s", region, new, old, time.Since(start))
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
		r, err := myClient.Get(fmt.Sprintf(b.URL, region))
		if err != nil {
			continue
		}
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

// saveDifferences compares a snapshot to the last snapshot
// and saves the differences into database
func (b *Battlegrounds) saveDifferences(res *BGResponse) (new, old int) {
	var ok bool
	for _, newR := range res.BGData.Rows {
		var curR, oldR BGRow

		// First scrape should always make new points
		if b.CurrSnapshots[res.Region].Timestamp == b.PrevSnapshots[res.Region].Timestamp {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// Getting info from last snapshot
		if curR, ok = b.CurrSnapshots[res.Region].BGData.Rows[newR.Name]; !ok {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}
		// Getting info from one before snapshot
		if oldR, ok = b.PrevSnapshots[res.Region].BGData.Rows[newR.Name]; !ok {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// Comparing rank and rating
		if newR.Rank != curR.Rank || newR.Rank != oldR.Rank {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}
		if newR.Rating != curR.Rating || newR.Rating != oldR.Rating {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// If all failed, update the point
		b.updatePoint(&newR, res.Timestamp, res.Season, res.Region)
		old++
	}
	new = len(res.BGData.Rows) - old
	return
}

func (b *Battlegrounds) newPoint(p *BGRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(battlegrounds_new, t, season, region, p.Name, p.Rank, p.Rating)
	if err != nil {
		b.Logger.Fatal(err)
	}
}

func (b *Battlegrounds) updatePoint(p *BGRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(battlegrounds_update, t, season, region, p.Name)
	if err != nil {
		b.Logger.Fatal(err)
	}
}
