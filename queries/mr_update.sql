UPDATE merceneries
SET timestamp = ?
FROM (
    SELECT rowid
    FROM merceneries
    WHERE seasonId = ? AND region = ? AND name = ?
    ORDER BY timestamp desc
    LIMIT 1
) AS latest
WHERE merceneries.rowid = latest.rowid;