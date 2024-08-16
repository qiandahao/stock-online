SELECT
    distinct symbol
FROM cn_stock_daily
where timestamp =  (select max(timestamp) from cn_stock_daily)