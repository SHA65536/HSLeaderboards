package main

import (
	hs "hsleaderboards"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := hs.LoadConfig()
	l := log.Default()
	db, _ := hs.MakeDatabase(l, cfg)
	sc := hs.MakeScraper(db, l, cfg)

	sc.AddSite(hs.MakeStandard())
	//sc.AddSite(hs.MakeBattlegrounds())

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go sc.Start()
	defer sc.Stop()
	<-done
	l.Println("Exiting...")
}
