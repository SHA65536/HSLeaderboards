CREATE TABLE IF NOT EXISTS "battlegrounds" (
	"rowid" INTEGER PRIMARY KEY AUTOINCREMENT,
	"timestamp"	INTEGER NOT NULL,
	"seasonId"	INTEGER NOT NULL,
	"region"    TEXT NOT NULL,
    "name"      TEXT NOT NULL,
    "rank" 	    INTEGER NOT NULL,
    "rating" 	INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS "ix_bgs_name_timestamp"
ON battlegrounds(seasonId, name, timestamp DESC);