import requests
from requests.exceptions import ConnectionError
import time
import json
import pandas as pd
import os.path
from datetime import datetime
import os

def generate_stock_csv():
    today = datetime.now().strftime('%Y-%m-%d')  # 格式化为 'YYYY-MM-DD'
    
    # 构建文件夹路径
    folder_path = f'/app/json/{today}'
    
    # 确保目录存在
    if not os.path.exists(folder_path):
        return
    
    # 构建文件路径
    file_path = os.path.join(folder_path, 'list.json')
    print(file_path)
    with open(file_path, 'r') as file:
        data = json.load(file)
    symbols = [item['symbol'] for item in data['data']['list']]

    df = pd.DataFrame({'symbol': symbols})
    # 将数据转换为 DataFrame

    
    for index, row in df.iterrows():
        file_path = os.path.join(folder_path,  row['symbol'] + '.json')
        if not os.path.exists(file_path):
            print(file_path)
            continue
        print(index)
        json_data = pd.read_json(file_path)
        
        df = pd.DataFrame(json_data['data']['item'], columns=json_data['data']['column'])
        
        df['symbol'] = row['symbol']
        # 将DataFrame保存为CSV文件
        df.to_csv('/app/csv/csv/'+row['symbol'] + '.csv', index=False)



generate_stock_csv()
    # 从 JSON 文件加载数据到 DataFrame
print("???")