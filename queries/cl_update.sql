UPDATE classic
SET timestamp = ?
FROM (
    SELECT rowid
    FROM classic
    WHERE seasonId = ? AND region = ? AND name = ?
    ORDER BY timestamp desc
    LIMIT 1
) AS latest
WHERE classic.rowid = latest.rowid;