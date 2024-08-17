
  WITH ranked_data AS (
    SELECT 
        symbol,
        timestamp,
        open,
        high,
        low,
        close,
        volume,
        -- 计算过去5天的最高价
        MAX(high) OVER (
            PARTITION BY symbol 
            ORDER BY timestamp 
            ROWS BETWEEN 5 PRECEDING AND 1 PRECEDING
        ) AS max_past_5_days
    FROM 
        cn_stock_daily
    WHERE 
        timestamp BETWEEN toUnixTimestamp(now() - INTERVAL 20 DAY) * 1000 AND toUnixTimestamp(now()) * 1000
)
SELECT
    distinct symbol
FROM 
    ranked_data
WHERE 
    open > max_past_5_days
    AND max_past_5_days != 0
    AND  timestamp BETWEEN toUnixTimestamp(now() - INTERVAL 10 DAY) * 1000 AND toUnixTimestamp(now()) * 1000
ORDER BY 
    symbol,
    timestamp DESC;