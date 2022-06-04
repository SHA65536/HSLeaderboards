package hsleaderboards

import (
	_ "embed"
)

//go:embed queries/create_bg.sql
var battlegrounds_create string

type Battlegrounds struct {
	URL     string
	Regions []string
	Db      *Database
	Sc      *Scraper
}

type BGResponse struct {
	ID           string         `json:"leaderboard_id"`
	Season       int            `json:"seasonId"`
	Region       string         `json:"region"`
	Leaderboards BGLeaderboards `json:"leaderboard"`
}

type BGLeaderboards struct {
	Rows []BGRow `json:"rows"`
}

type BGRow struct {
	Name   string `json:"accountid"`
	Rank   int    `json:"rank"`
	Rating int    `json:"rating"`
}

func (b *Battlegrounds) Name() string {
	return "Battlegrounds"
}

func (b *Battlegrounds) Initialize(sc *Scraper, db *Database) error {
	b.Sc = sc
	b.Db = db
	_, err := db.Session.Exec(battlegrounds_create)
	return err
}
func (b *Battlegrounds) Scrape() error {
	return nil
}

func MakeBattlegrounds() Site {
	return &Battlegrounds{
		URL:     "https://playhearthstone.com/en-gb/api/community/leaderboardsData?region=%s&leaderboardId=BG",
		Regions: []string{"US", "EU", "AP"},
	}
}
