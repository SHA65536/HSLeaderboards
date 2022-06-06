UPDATE wild
SET timestamp = ?
FROM (
    SELECT rowid
    FROM wild
    WHERE seasonId = ? AND region = ? AND name = ?
    ORDER BY timestamp desc
    LIMIT 1
) AS latest
WHERE wild.rowid = latest.rowid;