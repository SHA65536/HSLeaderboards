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

//go:embed queries/cl_create.sql
var classic_create string

//go:embed queries/cl_new.sql
var classic_new string

//go:embed queries/cl_update.sql
var classic_update string

type Classic struct {
	URL           string
	Regions       []string
	Retries       int
	Db            *Database
	Sc            *Scraper
	CurrSnapshots map[string]*CLResponse
	PrevSnapshots map[string]*CLResponse
	LatestSeason  int
	Logger        *log.Logger
}

func (b *Classic) Name() string {
	return "Classic"
}

func (b *Classic) Initialize(sc *Scraper, db *Database) error {
	b.Sc = sc
	b.Db = db
	b.Logger = sc.Logger
	now := time.Now()
	// Creating DB tables
	_, err := db.Session.Exec(classic_create)
	if err != nil {
		return err
	}
	// Getting latest season
	res, err := b.getResponse("US")
	if err != nil {
		return err
	}
	b.LatestSeason = res.CLMeta.Latest
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
	b.Logger.Printf("[Classic] Season: %d", b.LatestSeason)
	return err
}

// Scrape gets data from all regions and saves to database
func (b *Classic) Scrape() error {
	now := time.Now()
	for _, region := range b.Regions {
		start := time.Now()
		res, err := b.getResponse(region)
		if err != nil {
			b.Logger.Printf("[Classic] Failed to get region %s, %s", region, err)
			continue
		}
		if res.CLMeta.Latest != b.LatestSeason {
			b.Logger.Printf("[Classic] Season changed! %d -> %d", b.LatestSeason, res.CLMeta.Latest)
			b.LatestSeason = res.CLMeta.Latest
			continue
		}
		res.Timestamp = now.Unix()
		new, old := b.saveDifferences(res)
		b.PrevSnapshots[region] = b.CurrSnapshots[region]
		b.CurrSnapshots[region] = res
		b.Logger.Printf("[Classic] Saved region %s. New: %d, Old: %d | Took %s", region, new, old, time.Since(start))
	}
	return nil
}

func MakeClassic() Site {
	return &Classic{
		URL:           "https://playhearthstone.com/en-us/api/community/leaderboardsData?region=%s&leaderboardId=CLS&seasonId=%d",
		Regions:       []string{"US", "EU", "AP"},
		Retries:       3,
		CurrSnapshots: make(map[string]*CLResponse),
		PrevSnapshots: make(map[string]*CLResponse),
		LatestSeason:  104,
	}
}

// getResponse gets the data for the specified region
// handles retrie
func (b *Classic) getResponse(region string) (*CLResponse, error) {
	var err error
	var myClient = &http.Client{Timeout: 10 * time.Second}
	var response = &CLResponse{}
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
func (b *Classic) saveDifferences(res *CLResponse) (new, old int) {
	var ok bool
	for _, newR := range res.CLData.Rows {
		var curR, oldR CLRow

		// First scrape should always make new points
		if b.CurrSnapshots[res.Region].Timestamp == b.PrevSnapshots[res.Region].Timestamp {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// Getting info from last snapshot
		if curR, ok = b.CurrSnapshots[res.Region].CLData.Rows[newR.Name]; !ok {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}
		// Getting info from one before snapshot
		if oldR, ok = b.PrevSnapshots[res.Region].CLData.Rows[newR.Name]; !ok {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// Comparing rank
		if newR.Rank != curR.Rank || newR.Rank != oldR.Rank {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// If all failed, update the point
		b.updatePoint(&newR, res.Timestamp, res.Season, res.Region)
		old++
	}
	new = len(res.CLData.Rows) - old
	return
}

func (b *Classic) newPoint(p *CLRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(classic_new, t, season, region, p.Name, p.Rank)
	if err != nil {
		b.Logger.Fatal(err)
	}
}

func (b *Classic) updatePoint(p *CLRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(classic_update, t, season, region, p.Name)
	if err != nil {
		b.Logger.Fatal(err)
	}
}
