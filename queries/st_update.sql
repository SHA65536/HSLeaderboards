UPDATE standard
SET timestamp = ?
FROM (
    SELECT rowid
    FROM standard
    WHERE seasonId = ? AND region = ? AND name = ?
    ORDER BY timestamp desc
    LIMIT 1
) AS latest
WHERE standard.rowid = latest.rowid;