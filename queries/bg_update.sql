UPDATE battlegrounds
SET timestamp = ?
FROM (
    SELECT rowid
    FROM battlegrounds
    WHERE seasonId = ? AND region = ? AND name = ?
    ORDER BY timestamp desc
    LIMIT 1
) AS latest
WHERE battlegrounds.rowid = latest.rowid;