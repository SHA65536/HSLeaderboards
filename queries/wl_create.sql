CREATE TABLE IF NOT EXISTS "wild" (
	"rowid" INTEGER PRIMARY KEY AUTOINCREMENT,
	"timestamp"	INTEGER NOT NULL,
	"seasonId"	INTEGER NOT NULL,
	"region"    TEXT NOT NULL,
    "name"      TEXT NOT NULL,
    "rank" 	    INTEGER NOT NULL
);