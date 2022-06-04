package main

import (
	hs "hsleaderboards"
	"log"
)

func main() {
	cfg := hs.LoadConfig()
	l := log.Default()
	db, _ := hs.MakeDatabase(l, cfg)
	sc := hs.MakeScraper(db, l, cfg)
	sc.Start()
}
