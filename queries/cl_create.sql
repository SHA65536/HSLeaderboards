CREATE TABLE IF NOT EXISTS "classic" (
	"rowid" INTEGER PRIMARY KEY AUTOINCREMENT,
	"timestamp"	INTEGER NOT NULL,
	"seasonId"	INTEGER NOT NULL,
	"region"    TEXT NOT NULL,
    "name"      TEXT NOT NULL,
    "rank" 	    INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS "ix_cls_name_timestamp"
ON classic(seasonId, name, timestamp DESC);