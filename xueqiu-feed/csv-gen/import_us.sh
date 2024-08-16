#!/bin/bash

for file in *.csv; do
		clickhouse-client -q "insert into us_stock_daily format CSV" --input_format_allow_errors_ratio=0 < "$file"
done
