package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Data struct {
	Count int `json:"count"`
	List  []struct {
		Symbol                   string      `json:"symbol"`
		NetProfitCAGR            float64     `json:"net_profit_cagr"`
		NorthNetInflow           float64     `json:"north_net_inflow"`
		Ps                       interface{} `json:"ps"` // Assuming ps could be any type, represented as interface{}
		Type                     int         `json:"type"`
		Percent                  float64     `json:"percent"`
		HasFollow                bool        `json:"has_follow"`
		TickSize                 float64     `json:"tick_size"`
		PbTTM                    interface{} `json:"pb_ttm"` // Assuming pb_ttm could be any type, represented as interface{}
		FloatShares              int64       `json:"float_shares"`
		Current                  float64     `json:"current"`
		Amplitude                float64     `json:"amplitude"`
		Pcf                      interface{} `json:"pcf"` // Assuming pcf could be any type, represented as interface{}
		CurrentYearPercent       float64     `json:"current_year_percent"`
		FloatMarketCapital       float64     `json:"float_market_capital"`
		NorthNetInflowTime       int64       `json:"north_net_inflow_time"`
		MarketCapital            float64     `json:"market_capital"`
		DividendYield            float64     `json:"dividend_yield"`
		LotSize                  int         `json:"lot_size"`
		RoeTTM                   float64     `json:"roe_ttm"`
		TotalPercent             float64     `json:"total_percent"`
		Percent5m                float64     `json:"percent5m"`
		IncomeCAGR               float64     `json:"income_cagr"`
		Amount                   float64     `json:"amount"`
		Chg                      float64     `json:"chg"`
		IssueDateTS              int64       `json:"issue_date_ts"`
		EPS                      float64     `json:"eps"`
		MainNetInflows           float64     `json:"main_net_inflows"`
		Volume                   int         `json:"volume"`
		VolumeRatio              float64     `json:"volume_ratio"`
		PB                       float64     `json:"pb"`
		Followers                int         `json:"followers"`
		TurnoverRate             float64     `json:"turnover_rate"`
		MappingQuoteCurrent      interface{} `json:"mapping_quote_current"` // Assuming mapping_quote_current could be any type, represented as interface{}
		FirstPercent             float64     `json:"first_percent"`
		Name                     string      `json:"name"`
		PETTM                    float64     `json:"pe_ttm"`
		DualCounterMappingSymbol interface{} `json:"dual_counter_mapping_symbol"` // Assuming dual_counter_mapping_symbol could be any type, represented as interface{}
		TotalShares              int64       `json:"total_shares"`
		LimitupDays              int         `json:"limitup_days"`
	} `json:"list"`
}

type Response struct {
	Data             Data   `json:"data"`
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

type DailyDataResponse struct {
	Data             DailyData `json:"data"`
	ErrorCode        int       `json:"error_code"`
	ErrorDescription string    `json:"error_description"`
}

type DailyData struct {
	Symbol string          `json:"symbol"`
	Column []string        `json:"column"`
	Item   [][]interface{} `json:"item"`
}

type KLine struct {
	Timestamp string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int
}

func main() {
	url := "https://stock.xueqiu.com/v5/stock/screener/quote/list.json?size=100&order=asc&order_by=symbol&market=US&type=sh_sz"
	cookie := "cookiesu=421719180263062; device_id=a6515f3041fb0ab4a40de82f413d8a7b; s=c118yzh5gw; HMACCOUNT=98EC87727C47C805; xq_is_login=1; u=5836728060; snbim_minify=true; bid=aa3d1df992f90ca8cdd6773895b2f006_lz8suhg2; xq_a_token=dd19a74be93d09e875bd102a365bec04201fd3d2; xqat=dd19a74be93d09e875bd102a365bec04201fd3d2; xq_id_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1aWQiOjU4MzY3MjgwNjAsImlzcyI6InVjIiwiZXhwIjoxNzI1ODQzMTYxLCJjdG0iOjE3MjMyNTExNjE1NzUsImNpZCI6ImQ5ZDBuNEFadXAifQ.YcAhvX12IuEaVB2QaSQ7nNdHkcBwTChfgy7ecmKoU2dkH6V5pDiaUXHWRDbX0s_dPJGnRrbYHm8Nh6Kwjte49KxOOWNd69J_rWIvJJcqve_Z_9RyOrr88oijJkoEsO9gFWR-HWZ3oZ7dgLDvsgflvsiVP_arnFdga_Vc5QBgXinxc6kjN6Gkr0VGN0ylQhcSrSQAOpIIBrPOoQ7rtWGUb-KRONntikpCkYjGzkm8czctdBTlniWYLtwvAodUR4_lG1khuwu4fFyjsKy6KfFtAaoEsHTo7ZOiWHR26NYxbC_6JTmp8XXsGgSW3ZhgqzZB5nOSK28deF16_79p4ye4GQ; xq_r_token=3d8b6ce104df5bfdc560803bb1aeead591d0647d; is_overseas=1; Hm_lvt_1db88642e346389874251b5a1eded6e3=1721180874,1721265714,1721834124; Hm_lpvt_1db88642e346389874251b5a1eded6e3=1723423218; ssxmod_itna=eqGxgQDtoCqiwkDl4Yq0P+p35Q37K5i33AmoWD/KDfr4AQDymD82ohA+G+pGp=aXNC7ZxG=Eb7nrfbWb2BqwWeARr4GLDmFjNkAmi4GGDBeGwD0eG+DD4DWYq03DoxGAg+x04kg92u9jHD0YDzqDgD7jH7qDEDG3D0+=5YD59D73Df4DAWw2yYDDlYGn1775Dbh6SnR0HieDS/AUxKG=DjqGgDBdF9pD9DDtzZpGZfk98r=PuYmiDtqD9FCU7in+ySMt2i64I3r+CDT4GKBqPQGRwCi4CCBwe/0Lo72Cba7vLRO449QvZPD=; ssxmod_itna2=eqGxgQDtoCqiwkDl4Yq0P+p35Q37K5i33AmExG9boDBdP7QHGcDewiD="

	now := time.Now()
	unixNano := now.UnixNano()

	dateFormat := "2006-01-02"

	// 格式化当前时间
	formattedDate := now.Format(dateFormat)

	// 打印结果
	fmt.Println("Formatted Date:", formattedDate)
	folder_path := "D:\\data_volume\\json\\" + formattedDate
	if _, err := os.Stat(folder_path); os.IsNotExist(err) {
		err := os.Mkdir(folder_path, 0755) // 0755 是文件夹的权限，类似于 rwxr-xr-x
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	} else {
		fmt.Println("folder exists")
	}

	checkSymbol := make(map[string]bool)
	symbols := make([]string, 0)
	list := make([]interface{}, 0)
	for i := 0; i < 40; i++ {
		temp := url + "&page=" + strconv.Itoa(i)
		req, _ := http.NewRequest("GET", temp, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("发送请求失败:", err)
			return
		}
		defer resp.Body.Close()

		// 读取响应体
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		var response Response

		// 解析JSON数据
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		for _, item := range response.Data.List {
			if _, ok := checkSymbol[item.Symbol]; !ok {
				checkSymbol[item.Symbol] = true
				symbols = append(symbols, item.Symbol)
				list = append(list, item)
			}
		}
		// fmt.Println(symbols)
	}

	url = "https://stock.xueqiu.com/v5/stock/screener/quote/list.json?size=100&order=desc&order_by=symbol&market=US&type=sh_sz"
	for i := 0; i < 40; i++ {
		temp := url + "&page=" + strconv.Itoa(i)
		req, _ := http.NewRequest("GET", temp, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("发送请求失败:", err)
			return
		}
		defer resp.Body.Close()

		// 读取响应体
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		var response Response

		// 解析JSON数据
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		for _, item := range response.Data.List {
			if _, ok := checkSymbol[item.Symbol]; !ok {
				checkSymbol[item.Symbol] = true
				symbols = append(symbols, item.Symbol)
				list = append(list, item)
			}
		}
		// fmt.Println(symbols)
	}

	mergedData := make(map[string]interface{})
	mergedData["data"] = map[string]interface{}{
		"count": 5000,
		"list":  list,
	}
	mergedFile, err := json.MarshalIndent(mergedData, "", "    ")
	if err != nil {
		fmt.Println("生成合并文件失败:", err)
		return
	}

	listFilePath := folder_path + "\\list.json"

	// 检查文件夹是否已存在
	if _, err := os.Stat(listFilePath); !os.IsNotExist(err) {
		// 文件夹不存在，创建它
		fmt.Println("list文件已存在")
	} else {
		err = ioutil.WriteFile(folder_path+"\\list.json", mergedFile, 0644)
		if err != nil {
			fmt.Println("写入合并文件失败:", err)
			return
		}
	}

	// 定义一个变量来存储解析后的数据

	// 将纳秒转换为毫秒
	unixMilli := unixNano / int64(time.Millisecond)
	for _, symbol := range symbols {

		filePath := folder_path + "\\" + symbol + ".json"

		// 检查文件夹是否已存在
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			// 文件夹不存在，创建它
			fmt.Println("文件已存在")
			continue
		}

		dailyDataUrl := "https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=" + symbol + "&begin=" + strconv.FormatInt(unixMilli, 10) + "&period=day&type=before&count=-1&indicator=kline,pe,pb,ps,pcf,market_capital,agt,ggt,balance"
		req, _ := http.NewRequest("GET", dailyDataUrl, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
		fmt.Println(dailyDataUrl)
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("发送请求失败:", err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		var response DailyDataResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		jsonDataOutput, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return
		}

		// 保存 JSON 数据到文件
		err = ioutil.WriteFile(filePath, jsonDataOutput, 0644)
		if err != nil {
			fmt.Println("Error writing JSON to file:", err)
			return
		}
	}

}
