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
	val, err := strconv.Atoi(os.Getenv("INTERVAL"))
	if err == nil && val != 0 {
		interval = val
	}
	return &Config{
		DBPath:   os.Getenv("DB_PATH"),
		Interval: interval,
	}
}
