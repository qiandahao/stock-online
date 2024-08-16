package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"gopkg.in/gomail.v2"
)

type Market struct {
	StatusID     int    `json:"status_id"`
	Region       string `json:"region"`
	Status       string `json:"status"`
	TimeZone     string `json:"time_zone"`
	TimeZoneDesc string `json:"time_zone_desc"`
	DelayTag     int    `json:"delay_tag"`
}

type Quote struct {
	CurrentExt         *float64 `json:"current_ext"`
	Symbol             string   `json:"symbol"`
	VolumeExt          *int     `json:"volume_ext"`
	High52w            float64  `json:"high52w"`
	Delayed            int      `json:"delayed"`
	Type               int      `json:"type"`
	TickSize           float64  `json:"tick_size"`
	FloatShares        int64    `json:"float_shares"`
	LimitDown          float64  `json:"limit_down"`
	NoProfit           string   `json:"no_profit"`
	High               float64  `json:"high"`
	FloatMarketCapital float64  `json:"float_market_capital"`
	TimestampExt       int64    `json:"timestamp_ext"`
	LotSize            int      `json:"lot_size"`
	//LockSet                  *string  `json:"lock_set"`
	WeightedVotingRights     string   `json:"weighted_voting_rights"`
	Chg                      float64  `json:"chg"`
	Eps                      float64  `json:"eps"`
	LastClose                float64  `json:"last_close"`
	ProfitFour               float64  `json:"profit_four"`
	Volume                   int64    `json:"volume"`
	VolumeRatio              float64  `json:"volume_ratio"`
	ProfitForecast           float64  `json:"profit_forecast"`
	TurnoverRate             float64  `json:"turnover_rate"`
	Low52w                   float64  `json:"low52w"`
	Name                     string   `json:"name"`
	Exchange                 string   `json:"exchange"`
	PeForecast               float64  `json:"pe_forecast"`
	TotalShares              int64    `json:"total_shares"`
	Status                   int      `json:"status"`
	IsVieDesc                string   `json:"is_vie_desc"`
	SecurityStatus           *string  `json:"security_status"`
	Code                     string   `json:"code"`
	GoodwillInNetAssets      *float64 `json:"goodwill_in_net_assets"`
	AvgPrice                 float64  `json:"avg_price"`
	Percent                  float64  `json:"percent"`
	WeightedVotingRightsDesc string   `json:"weighted_voting_rights_desc"`
	Amplitude                float64  `json:"amplitude"`
	Current                  float64  `json:"current"`
	IsVie                    string   `json:"is_vie"`
	CurrentYearPercent       float64  `json:"current_year_percent"`
	IssueDate                int64    `json:"issue_date"`
	SubType                  string   `json:"sub_type"`
	Low                      float64  `json:"low"`
	IsRegistrationDesc       string   `json:"is_registration_desc"`
	NoProfitDesc             string   `json:"no_profit_desc"`
	MarketCapital            float64  `json:"market_capital"`
	Dividend                 float64  `json:"dividend"`
	DividendYield            float64  `json:"dividend_yield"`
	Currency                 string   `json:"currency"`
	Navps                    float64  `json:"navps"`
	Profit                   float64  `json:"profit"`
	Timestamp                int64    `json:"timestamp"`
	PeLyr                    float64  `json:"pe_lyr"`
	Amount                   float64  `json:"amount"`
	PledgeRatio              *float64 `json:"pledge_ratio"`
	TradedAmountExt          *float64 `json:"traded_amount_ext"`
	IsRegistration           string   `json:"is_registration"`
	Pb                       float64  `json:"pb"`
	LimitUp                  float64  `json:"limit_up"`
	PeTtm                    float64  `json:"pe_ttm"`
	Time                     int64    `json:"time"`
	Open                     float64  `json:"open"`
}

type Tag struct {
	Description string `json:"description"`
	Value       int    `json:"value"`
}

type Others struct {
	PankouRatio float64 `json:"pankou_ratio"`
	CybSwitch   bool    `json:"cyb_switch"`
}

type Data struct {
	Market Market `json:"market"`
	Quote  Quote  `json:"quote"`
	Others Others `json:"others"`
	Tags   []Tag  `json:"tags"`
}

// 定义嵌套结构体
type Data15m struct {
	Symbol string          `json:"symbol"`
	Column []string        `json:"column"`
	Item   [][]interface{} `json:"item"`
}

type Response15m struct {
	Data             Data15m `json:"data"`
	ErrorCode        int     `json:"error_code"`
	ErrorDescription string  `json:"error_description"`
}

type Response struct {
	Data             Data   `json:"data"`
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

type KLine struct {
	Timestamp string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int
}

type Point struct {
	index   int
	close   float64
	low     float64
	high    float64
	trigger int
	ts      int64
}

// linearRegression computes the slope of the linear regression line for the given data points
func linearRegression(x, y []float64) float64 {
	if len(x) != len(y) {
		panic("x and y must have the same length")
	}
	n := float64(len(x))
	var sumX, sumY, sumXY, sumXX float64
	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumXX += x[i] * x[i]
	}

	// Calculating the slope (m)
	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	return slope
}

func calculateMA(prices []float64, period int) []float64 {
	var ma []float64
	for i := 0; i <= len(prices)-period; i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += prices[i+j]
		}
		ma = append(ma, sum/float64(period))
	}
	return ma
}

// Calculate BIAS based on prices and moving averages (MA)
func calculateBIAS(prices, ma []float64, period int) []float64 {
	var bias []float64
	for i := 0; i < len(ma); i++ {
		biasValue := (prices[i+period-1] - ma[i]) / ma[i] * 100
		bias = append(bias, biasValue)
	}
	return bias
}

func calculateEMA(prices []float64, n int) []float64 {
	var emaValues []float64
	k := 2.0 / float64(n+1)

	// 初始 EMA 用简单移动平均线 (SMA)
	sma := calculateSMA(prices[:n])
	emaValues = append(emaValues, sma)

	for i := n; i < len(prices); i++ {
		ema := prices[i]*k + emaValues[i-n]*(1-k)
		emaValues = append(emaValues, ema)
	}

	return emaValues
}

func calculateAverageVolume(prices []float64, n int) []float64 {
	var emaValues []float64

	// 初始 EMA 用简单移动平均线 (SMA)
	sma := calculateSMA(prices[:n])
	emaValues = append(emaValues, sma)

	for i := n; i < len(prices); i++ {
		temp := calculateSMA(prices[i-n : i+1])
		emaValues = append(emaValues, temp)
	}

	return emaValues
}

func calculateSMA(prices []float64) float64 {
	sum := 0.0
	for _, price := range prices {
		sum += price
	}

	return sum / float64(len(prices))
}

func main() {
	ticker := time.NewTicker(15 * time.Minute)

	// 启动一个 goroutine 来执行定时任务
	go func() {
		index := 1
		for {
			select {
			case <-ticker.C:
				err := copyFile("running-days.txt", "destination"+strconv.Itoa(index)+".txt")
				if err != nil {
					fmt.Println("Error copying file:", err)
				} else {
					fmt.Println("File copied successfully at", time.Now().Format(time.RFC3339))
				}
			}
		}
	}()
	// // 注册登陆二维码回调
	// bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// // 登陆
	// if err := bot.Login(); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// // 获取登陆的用户
	// self, err := bot.GetCurrentUser()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// // 获取所有的好友
	// friends, err := self.Friends()
	// fmt.Println(friends, err)

	// if err != nil {
	// 	return
	// }
	// test := friends.Search(1, func(friend *openwechat.Friend) bool { return strings.Contains(friend.User.NickName, "德明") })

	// 配置ClickHouse连接参数
	options := &clickhouse.Options{
		Addr: []string{"192.168.3.10:19000"},
	}

	// 创建ClickHouse连接
	conn, err := clickhouse.Open(options)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer conn.Close()

	// 读取SQL脚本文件内容
	sqlFile, err := ioutil.ReadFile("./script.sql")
	if err != nil {
		log.Fatalf("Failed to read SQL script file: %v", err)
	}

	// 将文件内容转换为字符串
	sqlStatements := string(sqlFile)
	printedSymbols := make(map[string]bool)
	err = os.Remove("running-days.txt")
	if err != nil {
		fmt.Println("删除文件出错")
	}
	for {
		file, err := os.OpenFile("running-days.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}

		rows, err := conn.Query(context.Background(), sqlStatements)
		if err != nil {
			log.Fatalf("Failed to execute query: %v", err)
		}
		defer rows.Close()
		// 遍历查询结果
		index := 0
		//symbols := []string{"NVDA", "VST", "AAPL", "META", "PDD", "MSFT", "DUOL", "TSLA", "LLY", "AMD", "NFLX", "BABA"}
		downMap := make(map[string]*Point)
		for rows.Next() {
			index++
			var (
				symbol               string
				high                 float64
				open                 float64
				volume               uint64
				timestamp            float64
				max_high_60_days_ago float64
				// ... 定义其他列的类型
			)

			err := rows.Scan(&symbol, &high, &open, &volume, &timestamp, &max_high_60_days_ago /* ... */)
			if err != nil {
				log.Fatalf("Failed to scan row: %v", err)
			}
			if open*float64(volume) < 100000000.0 {
				continue
			}
			now := time.Now()
			unixNano := now.UnixNano()

			// 将纳秒转换为毫秒
			unixMilli := unixNano / int64(time.Millisecond)
			// url := "https://stock.xueqiu.com/v5/stock/quote.json?extend=detail&symbol=" + symbol
			url := "https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=" + symbol + "&begin=" + strconv.FormatInt(unixMilli, 10) + "&period=day&type=before&count=-300&indicator=kline,pe,pb,ps,pcf,market_capital,agt,ggt,balance"

			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", "cookiesu=931702263491225; s=bg11zw7opj; snbim_minify=true; bid=aa3d1df992f90ca8cdd6773895b2f006_lus3wyz6; u=931702263491225; Hm_lvt_1db88642e346389874251b5a1eded6e3=1716359813; device_id=1fa1304aabcac0d1a5db98dfddfee983; xq_a_token=483932c5fb313ca4c93e7165a31f179fb71e1804; xqat=483932c5fb313ca4c93e7165a31f179fb71e1804; xq_r_token=f3a274495a9b8053787677ff7ed85d1639c6e3e0; xq_id_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1aWQiOi0xLCJpc3MiOiJ1YyIsImV4cCI6MTcyMTQzNjcyOSwiY3RtIjoxNzE4ODQ3MTkzNDczLCJjaWQiOiJkOWQwbjRBWnVwIn0.qh129FV_Bo8_33CthG-kAjewrfCyPxvgltfbnn-yfTXygxQqlT1lGfeAAZta0IrF-OYAhA1eWgxuwhRSUN_Got2rdESYk2tLIpLIZ-yP3SrYYwYozCaXepFM4y8n1y8lkg45ng846NMvwCa1oSQj0Mjj8Y72HqHP146Fod1zwlxiMb0PAeIylLoe4XKQegjNP7uZWVVjnwd275y14HPsyQCDq_8wNGqV_RAOO8gf9SmIjUAFkdDIMO3nZzqNh9Zr9zlQKJORzbSYir-vRY6YsUKJ4qaCnE1IW9ru2cfBRI_FuozoXH9eeytfy3avfLYy_IV1dvk3JRGDT46nfukvDQ; Hm_lpvt_1db88642e346389874251b5a1eded6e3=1718905782; is_overseas=0")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")

			//fmt.Println(url)
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

			// 定义一个变量来存储解析后的数据
			var response Response15m

			// 解析JSON数据
			err = json.Unmarshal(body, &response)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			data := response.Data.Item
			if len(data) == 0 {
				log.Fatal("No data available")
			}

			// var highest, lowest float64
			// highest = math.Inf(-1)
			// lowest = math.Inf(1)

			var z, vols []float64
			var avgVol float64

			var lines []float64
			var downDiff []float64
			var downDiffMax float64
			// var start, end int

			// for idx, item := range data {
			// 	// timestamp := item[0].(float64)
			// 	high := item[5].(float64)
			// 	low := item[4].(float64)

			// 	if high > highest {
			// 		highest = high
			// 		start = idx
			// 	}
			// 	if low < lowest {
			// 		lowest = low
			// 		end = idx
			// 	}
			// }

			str := ""
			for _, item := range data {
				// // 使用 Unix 函数将时间戳转换为 Time 对象

				// // 输出时分秒格式的时间

				// if idx == start {
				// 	//fmt.Println("Unix时间戳转换后的正常时间:", normalTime.Format("2006-01-02 15:04:05 MST"))
				// } else if idx == end {
				// 	//fmt.Println("Unix时间戳转换后的正常时间:", normalTime.Format("2006-01-02 15:04:05 MST"))
				// }

				// value := (item[5].(float64) - lowest) * 10 / (highest - lowest)
				// x = append(x, float64(idx))
				// y = append(y, value)
				z = append(z, item[5].(float64))
				avgVol += item[1].(float64)
				vols = append(vols, avgVol)
			}
			avgVol = avgVol / float64(len(data)-1)
			//fmt.Println(symbol, max_high_60_days_ago, timestamp)
			n8 := 5
			n20 := 20
			ema5 := calculateEMA(z, n8)
			ema20 := calculateEMA(z, n20)
			ma20 := calculateMA(z, n20)
			up, down, downVol, upVol := 0.0, 0.0, 0.0, 0.0
			// Compare EMA5 and EMA20 and print where EMA5 is greater than EMA20

			// Compare EMA5 and EMA20 and print where EMA5 is greater than EMA20

			upline, downline, previousClose := 0.0, 0.0, 0.0
			uptimes, downtimes, previousTimes := 0, 0, 0
			//downZero, downOne, downTwo, downThree := 0, 0, 0, 0
			downVal := 0.0

			for i, item := range data {
				if i < 21 {
					continue
				}
				//"timestamp","volume","open","high","low","close","chg","percent","turnoverrate","amount","volume_post","amount_post"
				ts := int64(item[0].(float64))
				unixTimestamp := int64(item[0].(float64)) / 1000
				normalTime := time.Unix(unixTimestamp, 0)
				avgVol = vols[i] / float64(i+1)
				if down > 0 {
					downVol += item[1].(float64)
					downtimes++
					//if downMap[symbol] != nil {
					//lowDiff := math.Abs(downMap[symbol].low-item[3].(float64)) / item[5].(float64)
					// if lowDiff < 0.005 && downtimes < 4 {
					// 	downMap[symbol].trigger = downMap[symbol].trigger + 1
					// }
					// if downtimes == 4 && downMap[symbol].trigger >= 3 {
					// 	diff := math.Abs(downMap[symbol].close - item[5].(float64))
					// 	if diff < 0.005 {
					// 		downratio := (vols[i] - vols[downMap[symbol].index]) / float64(i-index) / avgVol
					// 		downZero = i
					// 		str += "进入小波动买点 " + fmt.Sprintf("low price: %f, close price: %f, vol ration: %f", item[4].(float64), item[5].(float64), downratio) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"

					// 		if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
					// 			SendEmail(symbol, str)
					// 		}
					// 		downMap[symbol].trigger = -1
					// 	}

					// } else if downMap[symbol].low > item[3].(float64) && downtimes > 4 && lowDiff >= 0.0095 && lowDiff < 0.0195 && downMap[symbol].trigger >= -1 {
					// 	downOne = i
					// 	downratio := (vols[i] - vols[downZero]) / float64(i-downZero) / avgVol
					// 	str += "进入波动买点1 " + fmt.Sprintf("low price: %f, close price: %f, vol ration: %f", item[4].(float64), item[5].(float64), downratio) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
					// 	downMap[symbol].trigger = -2
					// } else if downMap[symbol].low > item[3].(float64) && downtimes > 4 && lowDiff >= 0.0195 && lowDiff < 0.0395 && downMap[symbol].trigger >= -2 {
					// 	downTwo = i
					// 	downratio := (vols[i] - vols[downOne]) / float64(i-downOne) / avgVol
					// 	downMap[symbol].trigger = -3
					// 	str += "进入波动买点2 " + fmt.Sprintf("low price: %f, close price: %f, vol ration: %f", item[4].(float64), item[5].(float64), downratio) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
					// 	downMap[symbol].trigger = -3
					// } else if downMap[symbol].low > item[3].(float64) && downtimes > 4 && lowDiff >= 0.0395 && lowDiff < 0.0795 && downMap[symbol].trigger >= -3 {
					// 	downThree = i
					// 	downMap[symbol].trigger = -4
					// 	downratio := (vols[i] - vols[downTwo]) / float64(i-downTwo) / avgVol
					// 	str += "进入波动买点3 " + fmt.Sprintf("low price: %f, close price: %f, vol ration: %f", item[4].(float64), item[5].(float64), downratio) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
					// } else if downMap[symbol].low > item[3].(float64) && downtimes > 4 && lowDiff >= 0.0795 && lowDiff < 0.15395 && downMap[symbol].trigger >= -4 {
					// 	downMap[symbol].trigger = -5
					// 	downratio := (vols[i] - vols[downThree]) / float64(i-downThree) / avgVol
					// 	str += "进入波动买点4 " + fmt.Sprintf("low price: %f, close price: %f, vol ration: %f", item[4].(float64), item[5].(float64), downratio) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
					// }
					//}

					if previousClose < item[5].(float64) {
						previousTimes++
					} else {
						previousTimes = 0
					}
					if previousTimes >= 3 {
						previousTimes = 0
						downratio := (vols[i] - vols[i-3]) / 3 / avgVol
						str += "\t 下线转强 " + fmt.Sprintf("low price: %f, close price: %f, vol ration: %f", item[4].(float64), item[5].(float64), downratio) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
						if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
							SendEmail(symbol, str)
						}
					}
					if downline > item[4].(float64) {
						downline = item[4].(float64)
						str += "\t 更新底部下线 " + fmt.Sprintf("low price: %f, close price: %f", item[4].(float64), item[5].(float64)) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
						downDiffMax = 0
					} else {
						diff := (downline - item[4].(float64)) / downline
						diff = diff * diff * 10000
						if diff < 0.0001 {
							str += "\t 接触底部下线 " + fmt.Sprintf("low price: %f, close price: %f", item[4].(float64), item[5].(float64)) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
							downDiffMax = 0
						} else if diff > downDiffMax {
							str += "\t MA20: " + fmt.Sprintf("%f ", ma20[i-19])
							downDiffMax = diff
							str += "增加底部上线 " + fmt.Sprintf("low price: %f, close price: %f, maxdiff: %f", item[4].(float64), item[5].(float64), downDiffMax) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
						} else {
							str += "趋近底部下线 " + fmt.Sprintf("low price: %f, close price: %f, maxdiff: %f", item[4].(float64), item[5].(float64), downDiffMax) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
						}
						downDiff = append(downDiff, diff)
					}
				} else {
					uptimes++
					upVol += item[1].(float64)
					if previousClose < item[5].(float64) {
						previousTimes++
					} else {
						previousTimes = 0
					}
					if previousTimes >= 3 {
						previousTimes = 0
						upratio := (vols[i] - vols[i-3]) / 3 / avgVol
						str += "\t 上线强 " + fmt.Sprintf("low price: %f, close price: %f, vol ration: %f", item[4].(float64), item[5].(float64), upratio) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
					}
					if upline < item[3].(float64) {
						upline = item[3].(float64)
					}
				}
				previousClose = item[5].(float64)
				// i = 23 ema5[17] < ema20[]
				if ema5[i-5] < ema20[i-20] && ema5[i-4] > ema20[i-19] {
					str += "diff array: " + str
					for _, data := range downDiff {
						str += fmt.Sprintf("%f,", data)
					}
					str += "\n"
					lines = append(lines, downline)

					//gap := item[0].(float64) - up
					up = item[0].(float64)
					upVol = item[1].(float64)
					upline = item[3].(float64)
					previousTimes = 0
					//fmt.Printf("Day %d: EMA5 = %.2f, EMA20 = %.2f (EMA5 < EMA20)\n", i+n20, ema5[i], ema20[i])
					str += "\t 下线：" + fmt.Sprintf("%.2f", downline) + "\n"
					// diff := math.Abs(downVal-downline) / item[4].(float64)

					// if diff < 0.005 && downline > 0 && unixMilli-ts < 24*60*60*1000 {
					// 	str = str + "yyyy"
					// } else if diff < 0.005 && downline > 0 {
					// 	str = str + "xxxx"
					// }
					// if unixMilli-ts < 2*60*60*1000 && upVol > 2*downVol && downVol != 0 {
					// 	str = str + "###"F
					// }
					// if unixMilli-ts >= 2*60*60*1000 && upVol > 2*downVol && downVol != 0 {
					// 	str = str + "!!!"
					// }

					// if downtimes < 16 && unixMilli-ts < 24*60*60*1000 {
					// 	str = str + "$$$"
					// } else if downtimes < 16 {
					// 	str = str + "@@@"
					// }
					// diff := (downVal - downline) / downVal

					if downtimes > 5 && downtimes < 20 && downVol/float64(downtimes)/avgVol < 1.5 && downVol/float64(downtimes)/avgVol > 1.1 {
						str += "@@@@@"
						str += "\t MA20: " + fmt.Sprintf("%f ", ma20[i-19])
					}
					if downtimes > 30 {
						str += "ccccc"
					}
					str += "上穿: " + fmt.Sprintf("low price: %f, close price: %f, high price: %f", item[4].(float64), item[5].(float64), item[3].(float64)) + "  " + fmt.Sprintf("%d %f", downtimes, downVol/float64(downtimes)/avgVol) + "  " + normalTime.Format("2006-01-02 15:04:05 MST")
					downVol = 0
					str = str + "\n"

					down = 0
					uptimes = 0
					if !printedSymbols[symbol] && down > 1718553600000-4*24*60*60*1000 && up > down && upVol >= downVol*2 {
						str1 := time.Now().Format("2006-01-02 15:04:05 MST") + " " + "http://xueqiu.com/s/" + symbol + " ema5上穿20: " + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
						fmt.Println(str1)
						printedSymbols[symbol] = true
					}
					//str := "" + symbol + "于" + normalTime.Format("2006-01-02 15:04:05 MST") + "进入买点"
					//self.SendTextToFriend(test.First(), str)
					//playSound("./beiguozhichun.mp3")

					// fmt.Printf("Day %d: EMA5 = %.2f, EMA20 = %.2f (EMA5 > EMA20)\n", i+n20, ema5[i], ema20[i])
				}
				if ema5[i-5] > ema20[i-20] && ema5[i+1-5] < ema20[i+1-20] {
					if len(lines) > 1 {
						temp := upline - lines[len(lines)-1]
						if temp/upline > 0.03 {
							str += "+++++++++++++++++\n"
						}
					}

					lines = append(lines, upline)
					down = item[0].(float64)
					downtimes = 0
					downVal = item[5].(float64)
					downline = item[4].(float64)
					downVol = item[1].(float64)
					downDiff = make([]float64, 0)
					downDiffMax = 0
					str += "\t 上线：" + fmt.Sprintf("%.2f", upline) + "\n"
					//fmt.Println(upVol, uptimes, avgVol, upVol/float64(uptimes)/avgVol)
					tempVol := upVol / float64(uptimes) / avgVol
					if tempVol > 3 {
						str += "$$$$"
					}
					str += "下穿: " + strconv.FormatFloat(item[5].(float64), 'f', 2, 64) + "  " + fmt.Sprintf("%d %f", uptimes, tempVol) + "  " + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
					upVol = 0
					up = 0
					// if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
					// 	res := Point{int(ts), downVal, downline, item[3].(float64), 0}
					// 	downMap[symbol] = &res
					// }

					res := Point{i, downVal, downline, item[3].(float64), 0, ts}
					downMap[symbol] = &res
					//fmt.Printf("Day %d: EMA5 = %.2f, EMA20 = %.2f (EMA5 < EMA20)\n", i+n20, ema5[i], ema20[i])
				}
			}
			str += "====================" + symbol + "==================\n"
			_, err = file.WriteString(str)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
			//if start < end {
			//	slope := linearRegression(x[start:end+1], y[start:end+1])
			//	slope2 := linearRegression(x[end+1:], y[end+1:])
			//	fmt.Println("http://xueqiu.com/s/" + symbol)

			//	fmt.Printf("%.5f, %.5f\n", slope, slope2)
			//}

			// 处理每一行的数据
			//if response.Data.Quote.Current > max_high_60_days_ago {
			//}
		}
		fmt.Println(index)
		fmt.Println("............")
		time.Sleep(1 * time.Minute)
		file.Close()
		err = os.Remove("running-days.txt")
		if err != nil {
			fmt.Println("删除文件出错")
		}
		//fmt.Printf("Column 1: %s, Column 2: %d\n", symbol, max_high_60_days_ago /* ... */)
		//fmt.Printf("Quote Information:\n")
		//fmt.Printf("Symbol: %s\n", response.Data.Quote.Symbol)
		//fmt.Printf("Current: %.2f\n", response.Data.Quote.Current)
	}
}

func SendEmail(symbol, context string) {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "magineq6@126.com", "人造人六号买买买")
	m.SetAddressHeader("To", "magineq@126.com", "人造人一号")
	// 等待连接建立
	time.Sleep(2 * time.Second)

	// 创建合约对象
	m.SetHeader("Subject", symbol+" 进入通道")
	m.SetBody("text/plain", context)

	d := gomail.NewDialer("smtp.126.com", 25, "magineq6@126.com", "EGEPQFJNPDSTODIV")

	if err := d.DialAndSend(m); err != nil {
		log.Println("send mail err:", err)
	}
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}
