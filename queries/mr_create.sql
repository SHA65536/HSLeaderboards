CREATE TABLE IF NOT EXISTS "merceneries" (
	"rowid" INTEGER PRIMARY KEY AUTOINCREMENT,
	"timestamp"	INTEGER NOT NULL,
	"seasonId"	INTEGER NOT NULL,
	"region"    TEXT NOT NULL,
    "name"      TEXT NOT NULL,
    "rank" 	    INTEGER NOT NULL,
    "rating" 	INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS "ix_mrc_name_timestamp"
ON merceneries(seasonId, name, timestamp DESC);