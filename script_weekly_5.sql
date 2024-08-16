WITH LatestStocks AS (
    SELECT
        symbol,
        open,
        high,
        volume, 
        ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) AS rn
    FROM
        cn_stock_weekly
)
SELECT
    l.symbol,
    l.high AS latest_high,
    l_prev.open AS prev_open,
    l.rn
FROM
    LatestStocks l
JOIN
    LatestStocks l_prev ON l.symbol = l_prev.symbol AND l.rn = l_prev.rn + 1
WHERE
    l.high < l_prev.open and l.rn < 4 ORDER BY l.volume desc;