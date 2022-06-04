package hsleaderboards

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	Cfg     *Config
	Session *sql.DB
	Logger  *log.Logger
}

func MakeDatabase(logger *log.Logger, cfg *Config) (*Database, error) {
	db, err := sql.Open("sqlite3", cfg.DBPath)
	return &Database{
		Cfg:     cfg,
		Session: db,
		Logger:  logger,
	}, err
}
