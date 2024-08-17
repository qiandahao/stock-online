package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
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

type DownPeriod struct {
	ts         int
	downClimb  int
	downUpdate int
	downTouch  int
	downtimes  int
	downline   float64
	downRatio  float64
	upLine     float64
	enterLine  float64
	avgVal     float64
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

func updateCookies(url string, jar *cookiejar.Jar) error {
	client := &http.Client{
		Jar: jar,
	}

	// 发送GET请求以获取最新的Cookies
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("无法获取URL: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("无法更新cookie，状态码: %d", resp.StatusCode)
	}

	// 这里可以添加保存Cookies到本地逻辑
	// 例如: 保存到文件或其他持久化存储

	fmt.Println("Cookies更新成功")
	return nil
}

func main() {
	now := time.Now()
	unixNano := now.UnixNano()
	// 将纳秒转换为毫秒
	unixMilli := unixNano / int64(time.Millisecond)
	urlStr := "http://www.xueqiu.com"
	fmt.Println(urlStr)
	// 创建CookieJar来存储Cookies
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("创建CookieJar失败: %v", err)
	}

	// 更新Cookies
	if err := updateCookies(urlStr, jar); err != nil {
		log.Fatalf("更新Cookies失败: %v", err)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Fatalf("解析URL失败: %v", err)
	}
	cookies := jar.Cookies(parsedURL)
	var sb strings.Builder
	for _, cookie := range cookies {
		sb.WriteString(fmt.Sprintf("%s=%s; ", cookie.Name, cookie.Value))
	}

	// 打印合并后的Cookies字符串
	cookiesString := sb.String()
	if len(cookiesString) > 2 {
		cookiesString = cookiesString[:len(cookiesString)-2] // 去掉最后的 "; "
	}
	// 配置ClickHouse连接参数
	options := &clickhouse.Options{
		Addr: []string{"localhost:19000"},
	}
	cookie := "cookiesu=421719180263062; device_id=a6515f3041fb0ab4a40de82f413d8a7b; s=c118yzh5gw; xq_is_login=1; bid=aa3d1df992f90ca8cdd6773895b2f006_lz8suhg2; xq_a_token=dd19a74be93d09e875bd102a365bec04201fd3d2; xqat=dd19a74be93d09e875bd102a365bec04201fd3d2; xq_id_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1aWQiOjU4MzY3MjgwNjAsImlzcyI6InVjIiwiZXhwIjoxNzI1ODQzMTYxLCJjdG0iOjE3MjMyNTExNjE1NzUsImNpZCI6ImQ5ZDBuNEFadXAifQ.YcAhvX12IuEaVB2QaSQ7nNdHkcBwTChfgy7ecmKoU2dkH6V5pDiaUXHWRDbX0s_dPJGnRrbYHm8Nh6Kwjte49KxOOWNd69J_rWIvJJcqve_Z_9RyOrr88oijJkoEsO9gFWR-HWZ3oZ7dgLDvsgflvsiVP_arnFdga_Vc5QBgXinxc6kjN6Gkr0VGN0ylQhcSrSQAOpIIBrPOoQ7rtWGUb-KRONntikpCkYjGzkm8czctdBTlniWYLtwvAodUR4_lG1khuwu4fFyjsKy6KfFtAaoEsHTo7ZOiWHR26NYxbC_6JTmp8XXsGgSW3ZhgqzZB5nOSK28deF16_79p4ye4GQ; xq_r_token=3d8b6ce104df5bfdc560803bb1aeead591d0647d; u=5836728060; Hm_lvt_1db88642e346389874251b5a1eded6e3=1721834124,1723444989,1723650182,1723803179; Hm_lpvt_1db88642e346389874251b5a1eded6e3=1723803179; HMACCOUNT=0CBDE8C10CD55194; is_overseas=1; ssxmod_itna=Qqjx0Q0=i=D=DtEK0dGQDHQySeeTf8BxOAxY5wiNND/SFIDnqD=GFDK40EE8YqLWdQj7xdhqljADL5e87Cr4Fbz0RGadTDCPGnDB9DtazQxiigDCeDIDWeDiDGbtDFxYoDeaXQDFCT5XzUhKDpxGrDlKDRx07qSKDbxDaDGpk=70YUx0WDWPDi29w2ODDBO0EEmi3Dm+ky2a+qUYDn=011nhkD75Dux0HdBLxUxDCVKjxZ1v6BU3dA1BhDCKDjg71z6BP2ZHzp6fPNaxPVQ9ePc2D1lGoHFDPaii4oWR5102DIWxUEVGVLlwtQxqBEseixD=; ssxmod_itna2=Qqjx0Q0=i=D=DtEK0dGQDHQySeeTf8BxOAxY5wiNG9F=DBkP7Q7GcDeuiD=="
	// 创建ClickHouse连接
	conn, err := clickhouse.Open(options)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer conn.Close()

	// 读取SQL脚本文件内容
	sqlFile, err := ioutil.ReadFile("./script.sql")
	// sqlFile, err := ioutil.ReadFile("./script_30_days_gap_exists.sql")
	if err != nil {
		log.Fatalf("Failed to read SQL script file: %v", err)
	}

	// 将文件内容转换为字符串
	sqlStatements := string(sqlFile)
	err = os.Remove("running.txt")
	if err != nil {
		fmt.Println("删除文件出错")
	}
	for {
		file, err := os.OpenFile("running.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
		start := time.Now() // 记录开始时间

		nextRun := time.Date(start.Year(), start.Month(), start.Day(), 9, 30, 5, 0, start.Location())

		// 如果当前时间已经过了今天9点30分，则将下次运行时间设定为明天的9点30分
		if start.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		// 计算下次运行时间与当前时间的间隔
		duration := nextRun.Sub(start)

		// 创建定时器，在间隔时间后执行任务
		timer := time.NewTimer(duration)
		defer timer.Stop()

		fmt.Printf("下次运行时间：%s\n", nextRun)

		// 等待定时器触发
		// <-timer.C
		fmt.Println("开始运行\n")
		result := make([]string, 0)
		for rows.Next() {
			index++
			var (
				symbol string
				open   float64
				close  float64
				// ... 定义其他列的类型
			)

			err := rows.Scan(&symbol, &open, &close /* ... */)
			fmt.Println(symbol)
			if err != nil {
				log.Fatalf("Failed to scan row: %v", err)
			}

			// url := "https://stock.xueqiu.com/v5/stock/quote.json?extend=detail&symbol=" + symbol
			url := "https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=" + symbol + "&begin=" + strconv.FormatInt(unixMilli, 10) + "&period=day&type=before&count=-3&indicator=kline"
			fmt.Println(url)
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", cookie)
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")
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
			for i := 0; i < len(data); i++ {
				if data[i][2].(float64) < open && data[i][5].(float64) > open {
					result = append(result, symbol)
					str := fmt.Sprintf("%s,%f,%f\n", symbol, open, data[i][5].(float64))
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
			}

			//fmt.Printf("Column 1: %s, Column 2: %d\n", symbol, max_high_60_days_ago /* ... */)
			//fmt.Printf("Quote Information:\n")
			//fmt.Printf("Symbol: %s\n", response.Data.Quote.Symbol)
			//fmt.Printf("Current: %.2f\n", response.Data.Quote.Current)
		}
		str := ""
		res := make([]Quote, 0)
		for _, item := range result {
			url := "https://stock.xueqiu.com/v5/stock/quote.json?extend=detail&symbol=" + item
			// url := "https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=" + symbol + "&begin=" + strconv.FormatInt(unixMilli, 10) + "&period=day&type=before&count=-60&indicator=kline"
			// fmt.Println(url)
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", cookie)
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
			var response Response

			// 解析JSON数据
			err = json.Unmarshal(body, &response)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			res = append(res, response.Data.Quote)
		}
		sort.Slice(res, func(i, j int) bool {
			return res[i].MarketCapital > res[j].MarketCapital
		})
		for _, item := range res {
			str += item.Symbol + "   " + item.Name + "  " + fmt.Sprintf("%f", item.MarketCapital) + "\n"
		}
		_, err = file.WriteString(str)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
		elapsed := time.Since(start) // 计算经过的时间
		fmt.Printf("耗时：%s\n", elapsed)
		loopLine := "======================================\n"
		_, err = file.WriteString(loopLine)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
		fmt.Println(index)
		time.Sleep(1 * time.Minute)
		file.Close()
		err = os.Remove("running.txt")
		if err != nil {
			fmt.Println("删除文件出错")
		}
	}
}
