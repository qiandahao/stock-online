import pandas as pd
import matplotlib.pyplot as plt
import mplfinance as mpf
import clickhouse_connect

# 读取symbol列表文件
def read_symbols(file_path):
    with open(file_path, 'r') as file:
        symbols = [line.strip() for line in file]
    return symbols

def calculate_rolling_high(idx, series, window=5):
    # 获取窗口的起始点
    start_idx = max(0, idx - window)
    # 计算窗口内最大值，不包括当前行的值
    return series.iloc[start_idx:idx].max()

# 从ClickHouse数据库中读取历史记录
def fetch_historical_data(symbol, client):
    query = f"""
    Select timestamp AS date, open, high, low, close, volume
    FROM cn_stock_daily
    WHERE symbol = '{symbol}' ORDER BY date desc LIMIT 20
    """
    df = client.query_df(query)
    df.set_index('date', inplace=True)
    df = df.sort_index(ascending=True)
    
    df['RollingHigh'] = df['high'].rolling(window=5).max()

    # 找到满足条件的日期

    df['Special'] = [calculate_rolling_high(i, df['high']) for i in range(len(df))]
    # 计算均线

    # 找到 open 大于 high 的记录
    special_records = df[df['open'] > df['Special']]

    # 将这些记录插入到 ClickHouse 数据库中
    # 假设 `special_records` 是要插入的目标表
    insert_query_template = """
        INSERT INTO cn_gap_records (timestamp, open, high, low, close, volume, symbol) VALUES
    """

    # 生成插入语句的值部分
    values = []
    for index, row in special_records.iterrows():
        values.append(f"({index}, {row['open']}, {row['high']}, {row['low']}, {row['close']}, {row['volume']}, '{symbol}')")

    # 将生成的值部分拼接到插入语句中
    insert_query = insert_query_template + ", ".join(values)
    print(insert_query)
    client.query(insert_query)
    return df

# 主程序
def main():
    print("gg")
    client = clickhouse_connect.get_client(host='localhost', port=18123, database='default')
    query = "SELECT distinct symbol FROM cn_stock_daily where timestamp =  (select max(timestamp) from cn_stock_daily)"
    result = client.query(query)

    # 遍历查询结果并处理每个 symbol
    for row in result.result_rows:
        symbol = row[0]
        print(f"Processing symbol: {symbol}")

        fetch_historical_data(symbol, client)

if __name__ == "__main__":
    main()
