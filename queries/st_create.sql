CREATE TABLE IF NOT EXISTS "standard" (
	"rowid" INTEGER PRIMARY KEY AUTOINCREMENT,
	"timestamp"	INTEGER NOT NULL,
	"seasonId"	INTEGER NOT NULL,
	"region"    TEXT NOT NULL,
    "name"      TEXT NOT NULL,
    "rank" 	    INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS "ix_std_name_timestamp"
ON standard(seasonId, name, timestamp DESC);