SELECT
       symbol,
       MAX(high) AS max_high_5_days
       -- count(*)
   FROM
       cn_stock_daily csd
   WHERE
       `timestamp` > toUnixTimestamp(now()) * 1000 - 8* 24 * 3600 * 1000
   GROUP BY
       symbol