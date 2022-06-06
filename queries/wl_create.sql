CREATE TABLE IF NOT EXISTS "wild" (
	"rowid" INTEGER PRIMARY KEY AUTOINCREMENT,
	"timestamp"	INTEGER NOT NULL,
	"seasonId"	INTEGER NOT NULL,
	"region"    TEXT NOT NULL,
    "name"      TEXT NOT NULL,
    "rank" 	    INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS "ix_wld_name_timestamp"
ON wild(seasonId, name, timestamp DESC);