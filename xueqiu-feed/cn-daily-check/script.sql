WITH RankedSymbols AS (
    SELECT *,
           ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) AS rn
    FROM cn_gap_records
    WHERE `timestamp` >= toUnixTimestamp(now()) * 1000 - 8* 24 * 3600 * 1000
    and  `timestamp` <= toUnixTimestamp(now()) * 1000 - 3* 24 * 3600 * 1000
)
SELECT symbol, open, close
FROM RankedSymbols
WHERE rn = 1;