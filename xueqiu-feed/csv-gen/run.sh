docker run -d -v D:\\data_volume\\clickhouse: -v D:\\ -p 18123:8123 -p19000:9000 --name some-clickhouse-server --ulimit nofile=262144:262144 clickhouse/clickhouse-server
docker run -v csv:/app/csv -v D:\\data_volume\\json\\us\\daily:/app/json -it csv-gen /bin/bash
rm /app/csv/csv/*.csv
rm /app/csv/csv/*.*.csv
rm /app/csv/csv/*-*.csv
rm /app/csv/csv/*+*.csv
docker run -d -p 18123:8123 -p 19000:9000 -v  D:\\data_volume\\clickhouse:/var/lib/clickhouse/ -v D:\\data_volume\\json\\cn\\daily\\2024-08-08\\csv:/var/lib/csv_data --name some-clickhouse-server --ulimit nofile=262144:262144 clickhouse/clickhouse-server


docker run -d -p 18123:8123 -p 19000:9000 -v  csv:/var/lib/csv -v clickhouse:/var/lib/clickhouse --name some-clickhouse-server --ulimit nofile=262144:262144 clickhouse/clickhouse-server