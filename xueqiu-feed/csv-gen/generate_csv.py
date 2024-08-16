import requests
from requests.exceptions import ConnectionError
import time
import json
import pandas as pd
import os.path
from datetime import datetime
import os

def generate_stock_csv():
    with open('/app/data_volume/json/list.json', 'r') as file:
        data = json.load(file)
    symbols = [item['symbol'] for item in data['data']['list']]

    df = pd.DataFrame({'symbol': symbols})
    # 将数据转换为 DataFrame

    
    for index, row in df.iterrows():
        if not os.path.exists('/app/data_volume/json/'+ row['symbol'] + '.json'):
            continue
        print(index)
        json_data = pd.read_json('/app/data_volume/json/'+ row['symbol'] + '.json')
        
        df = pd.DataFrame(json_data['data']['item'], columns=json_data['data']['column'])
        
        df['symbol'] = row['symbol']
        # 将DataFrame保存为CSV文件
        df.to_csv('/app/csv/csv/'+row['symbol'] + '.csv', index=False)



generate_stock_csv()
    # 从 JSON 文件加载数据到 DataFrame
print("???")