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

//go:embed queries/mr_create.sql
var merc_create string

//go:embed queries/mr_new.sql
var merc_new string

//go:embed queries/mr_update.sql
var merc_update string

type Merceneries struct {
	URL           string
	Regions       []string
	Retries       int
	Db            *Database
	Sc            *Scraper
	CurrSnapshots map[string]*MRResponse
	PrevSnapshots map[string]*MRResponse
	LatestSeason  int
	Logger        *log.Logger
}

func (b *Merceneries) Name() string {
	return "Merceneries"
}

func (b *Merceneries) Initialize(sc *Scraper, db *Database) error {
	b.Sc = sc
	b.Db = db
	b.Logger = sc.Logger
	now := time.Now()
	// Creating DB tables
	_, err := db.Session.Exec(merc_create)
	if err != nil {
		return err
	}
	// Getting latest season
	res, err := b.getResponse("US")
	if err != nil {
		return err
	}
	b.LatestSeason = res.MRMeta.Latest
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
	b.Logger.Printf("[Merceneries] Season: %d", b.LatestSeason)
	return err
}

// Scrape gets data from all regions and saves to database
func (b *Merceneries) Scrape() error {
	now := time.Now()
	for _, region := range b.Regions {
		start := time.Now()
		res, err := b.getResponse(region)
		if err != nil {
			b.Logger.Printf("[Merceneries] Failed to get region %s, %s", region, err)
			continue
		}
		if res.MRMeta.Latest != b.LatestSeason {
			b.Logger.Printf("[Merceneries] Season changed! %d -> %d", b.LatestSeason, res.MRMeta.Latest)
			b.LatestSeason = res.MRMeta.Latest
			continue
		}
		res.Timestamp = now.Unix()
		new, old := b.saveDifferences(res)
		b.PrevSnapshots[region] = b.CurrSnapshots[region]
		b.CurrSnapshots[region] = res
		b.Logger.Printf("[Merceneries] Saved region %s. New: %d, Old: %d | Took %s", region, new, old, time.Since(start))
	}
	return nil
}

func MakeMerceneries() Site {
	return &Merceneries{
		URL:           "https://playhearthstone.com/en-us/api/community/leaderboardsData?region=%s&leaderboardId=MRC&seasonId=%d",
		Regions:       []string{"US", "EU", "AP"},
		Retries:       3,
		CurrSnapshots: make(map[string]*MRResponse),
		PrevSnapshots: make(map[string]*MRResponse),
		LatestSeason:  7,
	}
}

// getResponse gets the data for the specified region
// handles retrie
func (b *Merceneries) getResponse(region string) (*MRResponse, error) {
	var err error
	var myClient = &http.Client{Timeout: 10 * time.Second}
	var response = &MRResponse{}
	var url = fmt.Sprintf(b.URL, region, b.LatestSeason)
	for i := 0; i < b.Retries; i++ {
		r, err := myClient.Get(url)
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
func (b *Merceneries) saveDifferences(res *MRResponse) (new, old int) {
	var ok bool
	for _, newR := range res.MRData.Rows {
		var curR, oldR MRRow

		// First scrape should always make new points
		if b.CurrSnapshots[res.Region].Timestamp == b.PrevSnapshots[res.Region].Timestamp {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// Getting info from last snapshot
		if curR, ok = b.CurrSnapshots[res.Region].MRData.Rows[newR.Name]; !ok {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}
		// Getting info from one before snapshot
		if oldR, ok = b.PrevSnapshots[res.Region].MRData.Rows[newR.Name]; !ok {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// Comparing rank
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
	new = len(res.MRData.Rows) - old
	return
}

func (b *Merceneries) newPoint(p *MRRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(merc_new, t, season, region, p.Name, p.Rank, p.Rating)
	if err != nil {
		b.Logger.Fatal(err)
	}
}

func (b *Merceneries) updatePoint(p *MRRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(merc_update, t, season, region, p.Name)
	if err != nil {
		b.Logger.Fatal(err)
	}
}
