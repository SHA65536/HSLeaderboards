package hsleaderboards

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Interval int
	DBPath   string
}

func LoadConfig() *Config {
	godotenv.Load()
	var interval = 600
	var dbpath = "hearthstone.db"
	val, err := strconv.Atoi(os.Getenv("INTERVAL"))
	if err == nil && val != 0 {
		interval = val
	}
	if val := os.Getenv("DB_PATH"); val != "" {
		dbpath = val
	}
	return &Config{
		DBPath:   dbpath,
		Interval: interval,
	}
}
