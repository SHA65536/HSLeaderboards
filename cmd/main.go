package main

import (
	hs "hsleaderboards"
	"log"
)

func main() {
	cfg := hs.LoadConfig()
	l := log.Default()
	db := hs.MakeDatabase(l, cfg)
	sc := hs.MakeScraper(db, l, cfg)
	sc.Start()
}
