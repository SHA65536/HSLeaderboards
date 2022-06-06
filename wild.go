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

//go:embed queries/wl_create.sql
var wild_create string

//go:embed queries/wl_new.sql
var wild_new string

//go:embed queries/wl_update.sql
var wild_update string

type Wild struct {
	URL           string
	Regions       []string
	Retries       int
	Db            *Database
	Sc            *Scraper
	CurrSnapshots map[string]*WLResponse
	PrevSnapshots map[string]*WLResponse
	LatestSeason  int
	Logger        *log.Logger
}

func (b *Wild) Name() string {
	return "Wild"
}

func (b *Wild) Initialize(sc *Scraper, db *Database) error {
	b.Sc = sc
	b.Db = db
	b.Logger = sc.Logger
	now := time.Now()
	// Creating DB tables
	_, err := db.Session.Exec(wild_create)
	if err != nil {
		return err
	}
	// Getting latest season
	res, err := b.getResponse("US")
	if err != nil {
		return err
	}
	b.LatestSeason = res.WLMeta.Latest
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
	b.Logger.Printf("[Wild] Season: %d", b.LatestSeason)
	return err
}

// Scrape gets data from all regions and saves to database
func (b *Wild) Scrape() error {
	now := time.Now()
	for _, region := range b.Regions {
		start := time.Now()
		res, err := b.getResponse(region)
		if err != nil {
			b.Logger.Printf("[Wild] Failed to get region %s, %s", region, err)
			continue
		}
		if res.WLMeta.Latest != b.LatestSeason {
			b.Logger.Printf("[Wild] Season changed! %d -> %d", b.LatestSeason, res.WLMeta.Latest)
			b.LatestSeason = res.WLMeta.Latest
			continue
		}
		res.Timestamp = now.Unix()
		new, old := b.saveDifferences(res)
		b.PrevSnapshots[region] = b.CurrSnapshots[region]
		b.CurrSnapshots[region] = res
		b.Logger.Printf("[Wild] Saved region %s. New: %d, Old: %d | Took %s", region, new, old, time.Since(start))
	}
	return nil
}

func MakeWild() Site {
	return &Wild{
		URL:           "https://playhearthstone.com/en-us/api/community/leaderboardsData?region=%s&leaderboardId=WLD&seasonId=%d",
		Regions:       []string{"US", "EU", "AP"},
		Retries:       3,
		CurrSnapshots: make(map[string]*WLResponse),
		PrevSnapshots: make(map[string]*WLResponse),
		LatestSeason:  104,
	}
}

// getResponse gets the data for the specified region
// handles retrie
func (b *Wild) getResponse(region string) (*WLResponse, error) {
	var err error
	var myClient = &http.Client{Timeout: 10 * time.Second}
	var response = &WLResponse{}
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
func (b *Wild) saveDifferences(res *WLResponse) (new, old int) {
	var ok bool
	for _, newR := range res.WLData.Rows {
		var curR, oldR WLRow

		// First scrape should always make new points
		if b.CurrSnapshots[res.Region].Timestamp == b.PrevSnapshots[res.Region].Timestamp {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}

		// Getting info from last snapshot
		if curR, ok = b.CurrSnapshots[res.Region].WLData.Rows[newR.Name]; !ok {
			b.newPoint(&newR, res.Timestamp, res.Season, res.Region)
			continue
		}
		// Getting info from one before snapshot
		if oldR, ok = b.PrevSnapshots[res.Region].WLData.Rows[newR.Name]; !ok {
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
	new = len(res.WLData.Rows) - old
	return
}

func (b *Wild) newPoint(p *WLRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(wild_new, t, season, region, p.Name, p.Rank)
	if err != nil {
		b.Logger.Fatal(err)
	}
}

func (b *Wild) updatePoint(p *WLRow, t int64, season int, region string) {
	_, err := b.Db.Session.Exec(wild_update, t, season, region, p.Name)
	if err != nil {
		b.Logger.Fatal(err)
	}
}
