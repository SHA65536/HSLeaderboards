package hsleaderboards

import "log"

type Database struct {
	Cfg    *Config
	Logger *log.Logger
}

func MakeDatabase(logger *log.Logger, cfg *Config) *Database {
	return &Database{
		Cfg:    cfg,
		Logger: logger,
	}
}
