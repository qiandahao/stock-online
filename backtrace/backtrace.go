package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

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
	ts      int
	close   float64
	low     float64
	high    float64
	trigger int
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

func SendEmail(symbol, context string) {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "magineq6@126.com", "人造人六号")
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
func main() {
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
	file, err := os.OpenFile("running.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	// 配置ClickHouse连接参数
	// symbols := []string{"SZ002332", "SH600837", "SZ300696", "SH601126", "SH600309", "SH600999"}
	symbols := []string{"NVDA", "VST", "AAPL", "META", "PDD", "MSFT", "DUOL", "TSLA", "LLY", "AMD", "NFLX", "AMZN"}
	downMap := make(map[string]*Point)

	//symbols := []string{"SH600398", "SH600674", "SH600760", "SH600765", "SH600999", "SH601126"}
	for {
		for _, symbol := range symbols {
			now := time.Now()
			unixNano := now.UnixNano()

			// 将纳秒转换为毫秒
			unixMilli := unixNano / int64(time.Millisecond)
			// url := "https://stock.xueqiu.com/v5/stock/quote.json?extend=detail&symbol=" + symbol
			url := "https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=" + symbol + "&begin=" + strconv.FormatInt(unixMilli, 10) + "&period=15m&type=before&count=-500&indicator=kline,pe,pb,ps,pcf,market_capital,agt,ggt,balance"

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
				vols = append(vols, item[1].(float64))
			}
			//fmt.Println(symbol, max_high_60_days_ago, timestamp)
			n8 := 5
			n20 := 20
			ema5 := calculateEMA(z, n8)
			ema20 := calculateEMA(z, n20)
			averageVols := calculateAverageVolume(vols, 5)
			str := ""
			up, down, downVol, upVol := 0.0, 0.0, 0.0, 0.0
			// Compare EMA5 and EMA20 and print where EMA5 is greater than EMA20
			fmt.Println(up)
			upline, downline, previousClose := 0.0, 0.0, 0.0
			downtimes, previousTimes := 0, 0
			downVal := 0.0
			for i, item := range data {
				if i < 21 {
					continue
				}

				ts := int64(item[0].(float64))
				unixTimestamp := int64(item[0].(float64)) / 1000
				normalTime := time.Unix(unixTimestamp, 0)
				if down > 0 {
					downtimes++
					if downMap[symbol] != nil {
						diff := math.Abs(downMap[symbol].low-item[3].(float64)) / item[5].(float64)
						if diff < 0.005 && downtimes < 4 {
							downMap[symbol].trigger = downMap[symbol].trigger + 1
						}
						if downtimes == 4 && downMap[symbol].trigger >= 3 {
							str += "进入买点不波动买点 " + fmt.Sprintf("%.2F", item[4].(float64)) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
							downMap[symbol].trigger = -1
							if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
								SendEmail(symbol, str)
							}
						} else if downMap[symbol].low > item[3].(float64) && downtimes > 4 && diff > 0.0095 && downMap[symbol].trigger >= -1 {
							str += "进入买点波动买点1 " + fmt.Sprintf("%f, %f", item[4].(float64), diff) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
							downMap[symbol].trigger = -2
							if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
								SendEmail(symbol, str)
							}
						} else if downMap[symbol].low > item[3].(float64) && downtimes > 4 && diff > 0.0195 && downMap[symbol].trigger == -2 {
							downMap[symbol].trigger = -3
							if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
								SendEmail(symbol, str)
							}
							str += "进入买点波动买点2 " + fmt.Sprintf("%f, %f", item[4].(float64), diff) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
						} else if downMap[symbol].low > item[3].(float64) && downtimes > 4 && diff > 0.0395 && downMap[symbol].trigger == -3 {
							downMap[symbol].trigger = -4
							str += "进入买点波动买点3 " + fmt.Sprintf("%f, %f", item[4].(float64), diff) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
							if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
								SendEmail(symbol, str)
							}
						}
					}

					if previousClose < item[5].(float64) {
						previousTimes++
					} else {
						previousTimes = 0
					}
					if previousTimes >= 3 {
						previousTimes = 0
						str += "\t 下线转强 " + fmt.Sprintf("%.2F", item[4].(float64)) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
					}
					if downline > item[4].(float64) {
						downline = item[4].(float64)
					}
				} else {
					if previousClose < item[5].(float64) {
						previousTimes++
					} else {
						previousTimes = 0
					}
					if previousTimes >= 3 {
						previousTimes = 0
						str += "\t 上线强 " + fmt.Sprintf("%.2F", item[4].(float64)) + normalTime.Format("2006-01-02 15:04:05 MST") + "\n"
					}
					if upline < item[3].(float64) {
						upline = item[3].(float64)
					}
				}
				previousClose = item[5].(float64)
				// i = 23 ema5[17] < ema20[]
				if ema5[i-5] < ema20[i-20] && ema5[i-4] > ema20[i-19] {
					//gap := item[0].(float64) - up
					up = item[0].(float64)
					upVol = averageVols[i-5]
					previousTimes = 0
					diff := math.Abs(downVal-downline) / item[4].(float64)

					if diff < 0.005 && downline > 0 && unixMilli-ts < 24*60*60*1000 {
						str = str + "yyyy"
					} else if diff < 0.005 && downline > 0 {
						str = str + "xxxx"
					}
					//fmt.Printf("Day %d: EMA5 = %.2f, EMA20 = %.2f (EMA5 < EMA20)\n", i+n20, ema5[i], ema20[i])
					str += "\t 下线：" + fmt.Sprintf("%.2f", downline) + "  " + fmt.Sprintf("%f", diff) + "\n"
					if unixMilli-ts < 2*60*60*1000 && upVol > 2*downVol && downVol != 0 {
						str = str + "###"
					}
					if unixMilli-ts >= 2*60*60*1000 && upVol > 2*downVol && downVol != 0 {
						str = str + "!!!"
					}

					if downtimes < 16 && unixMilli-ts < 24*60*60*1000 {
						str = str + "$$$"
					} else if downtimes < 16 {
						str = str + "@@@"
					}
					str += "上穿20:" + strconv.FormatFloat(item[5].(float64), 'f', 2, 64) + normalTime.Format("2006-01-02 15:04:05 MST") + " " + strconv.FormatFloat(averageVols[i-5], 'f', 2, 64)

					str = str + "\n"
					down = 0
					downtimes = 0
					//Points = append(Points, Point{i, item[4].(float64)})
					//str := "" + symbol + "于" + normalTime.Format("2006-01-02 15:04:05 MST") + "进入买点"
					//self.SendTextToFriend(test.First(), str)
					//playSound("./beiguozhichun.mp3")

					// fmt.Printf("Day %d: EMA5 = %.2f, EMA20 = %.2f (EMA5 > EMA20)\n", i+n20, ema5[i], ema20[i])
					if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
						SendEmail(symbol, str)
					}
				}
				if ema5[i-5] > ema20[i-20] && ema5[i+1-5] < ema20[i+1-20] {
					down = item[0].(float64)
					downtimes = 0
					downVal = item[5].(float64)
					downline = item[4].(float64)
					if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
						SendEmail(symbol, str)
					}
					str += "\t 上线：" + fmt.Sprintf("%.2f", upline) + "\n"
					str += "下穿20:" + strconv.FormatFloat(item[5].(float64), 'f', 2, 64) + normalTime.Format("2006-01-02 15:04:05 MST") + " " + strconv.FormatFloat(averageVols[i-5], 'f', 2, 64) + "\n"
					// if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
					// 	res := Point{int(ts), downVal, downline, item[3].(float64), 0}
					// 	downMap[symbol] = &res
					// }
					// if i == len(data)-1 && unixMilli-ts < 1*60*1000 {
					// SendEmail(symbol, str)
					// }
					res := Point{int(ts), downVal, downline, item[3].(float64), 0}
					downMap[symbol] = &res
					up = 0.0
					upline = 0.0
					//fmt.Printf("Day %d: EMA5 = %.2f, EMA20 = %.2f (EMA5 < EMA20)\n", i+n20, ema5[i], ema20[i])
				}
			}

			str += "http://xueqiu.com/s/" + symbol + "\n"
			//if start < end {
			//	slope := linearRegression(x[start:end+1], y[start:end+1])
			//	slope2 := linearRegression(x[end+1:], y[end+1:])
			//	fmt.Println("http://xueqiu.com/s/" + symbol)

			//	fmt.Printf("%.5f, %.5f\n", slope, slope2)
			//}
			_, err = file.WriteString(str)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
			// 处理每一行的数据
			//if response.Data.Quote.Current > max_high_60_days_ago {
			//}
		}
		fmt.Println("............")
		time.Sleep(1 * time.Minute)
	}

}
