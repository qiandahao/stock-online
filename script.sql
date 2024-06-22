WITH Last_60_Days_Max AS (
   SELECT
       symbol,
       MAX(high) AS max_high_60_days_ago
   FROM
       us_stock_daily csd
   WHERE
       `timestamp` >= toUnixTimestamp(now()) * 1000 - 120 * 24 * 3600 * 1000
       and `timestamp` <= toUnixTimestamp(now()) * 1000 - 5* 24 * 3600 * 1000
   GROUP BY
       symbol
)
SELECT
   usd.symbol,
   usd.high,
   usd.`open`,
   usd.volume,
   usd.timestamp,
   Last_60_Days_Max.max_high_60_days_ago
FROM
   us_stock_daily usd
LEFT JOIN
   Last_60_Days_Max ON usd.symbol = Last_60_Days_Max.symbol
WHERE
   `timestamp` = (SELECT MAX(`timestamp`) - 0 * 24 * 3600 * 1000 FROM us_stock_daily)
   AND usd.`close` > Last_60_Days_Max.max_high_60_days_ago * 0.96