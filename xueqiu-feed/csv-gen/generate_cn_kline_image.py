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
    Select fromUnixTimestamp(toInt64(timestamp / 1000)) AS date, open, high, low, close, volume
    FROM cn_stock_daily
    WHERE symbol = '{symbol}' ORDER BY date desc LIMIT 180
    """
    df = client.query_df(query)
    df['date'] = pd.to_datetime(df['date'])
    df.set_index('date', inplace=True)
    df = df.sort_index(ascending=True)
    
    df['RollingHigh'] = df['high'].rolling(window=5).max()

    # 找到满足条件的日期

    df['Special'] = [df['open'].iloc[i] > calculate_rolling_high(i, df['high']) for i in range(len(df))]
    # 计算均线
    df['ma_5'] = df['close'].rolling(window=5).mean()
    df['ma_10'] = df['close'].rolling(window=10).mean()
    df['ma_20'] = df['close'].rolling(window=20).mean()
    df['ma_30'] = df['close'].rolling(window=30).mean()
    df['ma_60'] = df['close'].rolling(window=60).mean()
    
    return df

# 生成蜡烛图
def plot_candlestick(data, symbol):
    # Define additional plots (moving averages)
    add_plots = [
        mpf.make_addplot(data['ma_5'], color='blue', title='5-Day MA'),
        mpf.make_addplot(data['ma_10'], color='green', title='10-Day MA'),
        mpf.make_addplot(data['ma_20'], color='red', title='20-Day MA'),
        mpf.make_addplot(data['ma_30'], color='orange', title='30-Day MA'),
        mpf.make_addplot(data['ma_60'], color='black', title='60-Day MA'),
        mpf.make_addplot(data['Special'], color='grey', title='Special')
    ]
    
    # Plot the candlestick chart with moving averages
    mpf.plot(data, type='candle', style='charles', figsize=(30, 20), title=symbol,
             ylabel='Price', volume=True, addplot=add_plots,
             savefig=f"us/{symbol}_candlestick.png")

# 主程序
def main(symbols_file):
    client = clickhouse_connect.get_client(host='localhost', port=18123, database='default')
    symbols = read_symbols(symbols_file)
    for symbol in symbols:
        print(f"Processing symbol: {symbol}")

        data = fetch_historical_data(symbol, client)
        print(data.size)
        if not data.empty:
            if data.size != 2160:
                continue
            plot_candlestick(data, symbol)
        else:
            print(f"No data found for symbol: {symbol}")

if __name__ == "__main__":
    main('symbols.txt')
