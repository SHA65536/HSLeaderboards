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

//go:embed queries/st_create.sql
var standard_create string

//go:embed queries/st_new.sql
var standard_new string

//go:embed queries/st_update.sql
var standard_update string

type Standard struct {
	URL           string
	Regions       []string
	Retries       int
	Db            *Database
	Sc            *Scraper
	CurrSnapshots map[string]*STResponse
	PrevSnapshots map[string]*STResponse
	LatestSeason  int
	Logger        *log.Logger
}

func (b *Standard) Name() string {
	return "Standard"
}

func (b *Standard) Initialize(sc *Scraper, db *Database) error {
	b.Sc = sc
	b.Db = db
	b.Logger = sc.Logger
	now := time.Now()
	// Creating DB tables
	_, err := db.Session.Exec(standard_create)
	if err != nil {
		return err
	}
	// Getting latest season
	res, err := b.getResponse("US")
	if err != nil {
		return err
	}
	b.LatestSeason = res.STMeta.Latest
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
	b.Logger.Printf("[Standard] Season: %d", b.LatestSeason)
	return err
}

// Scrape gets data from all regions and saves to database
func (b *Standard) Scrape() error {
	now := time.Now()
	for _, region := range b.Regions {
		start := time.Now()
		res, err := b.getResponse(region)
		if err != nil {
			b.Logger.Printf("[Standard] Failed to get region %s, %s", region, err)
			continue
		}
		if res.STMeta.Latest != b.LatestSeason {
			b.Logger.Printf("[Standard] Season changed! %d -> %d", b.LatestSeason, res.STMeta.Latest)
			b.LatestSeason = res.STMeta.Latest
			continue
		}
		res.Timestamp = now.Unix()
		new, old := b.saveDifferences(res)
		b.PrevSnapshots[region] = b.CurrSnapshots[region]
		b.CurrSnapshots[region] = res
		b.Logger.Printf("[Standard] Saved region %s. New: %d, Old: %d | Took %s", region, new, old, time.Since(start))
	}
	return nil
}

func MakeStandard() Site {
	return &Standard{
		URL:           "https://playhearthstone.com/en-us/api/community/leaderboardsData?region=%s&leaderboardId=STD&seasonId=%d",
		Regions:       []string{"US", "EU", "AP"},
		Retries:       3,
		CurrSnapshots: make(map[string]*STResponse),
		PrevSnapshots: make(map[string]*STResponse),
		LatestSeason:  104,
	}
}

// getResponse gets the data for the specified region
// handles retrie
func (b *Standard) getResponse(region string) (*STResponse, error) {
	var err error
	var myClient = &http.Client{Timeout: 10 * time.Second}
	var response = &STResponse{}
	var url = fmt.Sprintf(b.URL, region, b.LatestSeason)
	for i := 0; i < b.Retries; i++ {
		r, _ := myClient.Get(url)
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
func (b *Standard) saveDifferences(res *STResponse) (new, old int) {
	var ok bool
	for _, newR := range res.STData.Rows {
		var curR, oldR STRow

		// First scrape should always make new points
		if b.CurrSnapshots[res.Region].Timestamp == b.PrevSnapshots[res.Region].Timestamp {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// Getting info from last snapshot
		if curR, ok = b.CurrSnapshots[res.Region].STData.Rows[newR.Name]; !ok {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}
		// Getting info from one before snapshot
		if oldR, ok = b.PrevSnapshots[res.Region].STData.Rows[newR.Name]; !ok {
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
	new = len(res.STData.Rows) - old
	return
}

func (b *Standard) newPoint(p *STRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(standard_new, t, season, region, p.Name, p.Rank)
	if err != nil {
		b.Logger.Fatal(err)
	}
}

func (b *Standard) updatePoint(p *STRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(standard_update, t, season, region, p.Name)
	if err != nil {
		b.Logger.Fatal(err)
	}
}
