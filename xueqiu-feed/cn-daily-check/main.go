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
	"strconv"
	"strings"
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
	open  float64
	close float64
	high  float64
	ts    uint64
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

func SendEmail(symbol, context string) {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "magineq6@126.com", "人造人六号")
	m.SetAddressHeader("To", "magineq@126.com", "人造人一号")
	// 等待连接建立
	time.Sleep(2 * time.Second)

	// 创建合约对象
	m.SetHeader("Subject", symbol+" 进入买点")
	m.SetBody("text/plain", context)

	d := gomail.NewDialer("smtp.126.com", 25, "magineq6@126.com", "EGEPQFJNPDSTODIV")

	if err := d.DialAndSend(m); err != nil {
		log.Println("send mail err:", err)
	}
}

func runCNTask(result map[string]int, pool map[string]*Point, cookie string) {
	now := time.Now()
	unixNano := now.UnixNano()
	// 将纳秒转换为毫秒
	unixMilli := unixNano / int64(time.Millisecond)

	// 配置ClickHouse连接参数
	options := &clickhouse.Options{
		Addr: []string{"localhost:19000"},
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer conn.Close()
	if len(pool) == 0 {
		// 将文件内容转换为字符串
		sqlStatements := fmt.Sprintf(`WITH RankedSymbols AS (
			SELECT *,
				ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) AS rn
			FROM cn_gap_records
			WHERE symbol in (SELECT DISTINCT symbol FROM cn_stock_daily where timestamp = (select max(timestamp) from cn_stock_daily)) and timestamp >= toUnixTimestamp(now()) * 1000 - 12* 24 * 3600 * 1000
			and  timestamp <= toUnixTimestamp(now()) * 1000 - 5* 24 * 3600 * 1000
		)
		SELECT symbol, open, close, timestamp
		FROM RankedSymbols

		WHERE rn = 1;`,
		)
		rows, err := conn.Query(context.Background(), sqlStatements)
		if err != nil {
			fmt.Println("数据库查询出错")
		}
		for rows.Next() {
			var (
				symbol string
				open   float64
				close  float64
				ts     uint64
			)
			temp := &Point{}
			err := rows.Scan(&symbol, &open, &close, &ts /* ... */)
			if err != nil {
				fmt.Println("数据绑定出错")
			}
			temp.open = open
			temp.close = close
			temp.ts = ts
			pool[symbol] = temp
		}
	}

	err = os.Remove("running-cn.txt")
	if err != nil {
		fmt.Println("删除文件出错")
	}
	file, err := os.OpenFile("running-cn.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	start := time.Now()
	fmt.Println(len(pool))
	for symbol, item := range pool {
		// 等待定时器触发
		// <-timer.C
		// url := "https://stock.xueqiu.com/v5/stock/quote.json?extend=detail&symbol=" + symbol
		url := "https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=" + symbol + "&begin=" + strconv.FormatInt(unixMilli, 10) + "&period=day&type=before&count=-20&indicator=kline"
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

		closePrice := make([]float64, 0)
		for i := 0; i < len(data); i++ {
			closePrice = append(closePrice, data[i][5].(float64))

			if i == len(data)-1 {
				avg := calculateMA(closePrice, 10)
				if data[i-1][5].(float64) < avg[len(avg)-2] {
					continue
				}
				if data[i][4].(float64) < item.open && data[i][5].(float64) > item.open {
					seconds := item.ts / 1000
					nanoseconds := (item.ts % 1000) * 1000000
					t := time.Unix(int64(seconds), int64(nanoseconds))
					formattedTime := t.Format("2006-01-02 15:04:05")

					str := fmt.Sprintf("%s\n        Gap点:%f(%s），10日平均：%f,开盘：%f, 最低：%f, 最高：%f, 现价：%f\n", symbol, item.open, formattedTime, avg[len(avg)-1], data[i][2].(float64), data[i][4].(float64), data[i][3].(float64), data[i][5].(float64))
					if _, ok := result[symbol]; !ok {
						result[symbol] = 1
						// SendEmail(symbol, str)
					}
					_, err = file.WriteString(str)
					if err != nil {
						fmt.Println("Error writing to file:", err)
						return
					}
				}
			}
		}
	}
	elapsed := time.Since(start) // 计算经过的时间
	fmt.Printf("耗时：%s\n", elapsed)
	loopLine := "======================================\n"
	_, err = file.WriteString(loopLine)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func runUsTask(result map[string]int, pool map[string]*Point, cookie string) {
	now := time.Now()
	unixNano := now.UnixNano()
	// 将纳秒转换为毫秒
	unixMilli := unixNano / int64(time.Millisecond)

	// 配置ClickHouse连接参数
	options := &clickhouse.Options{
		Addr: []string{"localhost:19000"},
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer conn.Close()
	if len(pool) == 0 {
		// 将文件内容转换为字符串
		sqlStatements := fmt.Sprintf(`WITH RankedSymbols AS (
			SELECT *,
				ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) AS rn
			FROM us_gap_records
			WHERE symbol in (SELECT DISTINCT symbol FROM us_stock_daily where timestamp = (select max(timestamp) from us_stock_daily) and market_capital > 5045411866) and timestamp >= toUnixTimestamp(now()) * 1000 - 12* 24 * 3600 * 1000
			and  timestamp <= toUnixTimestamp(now()) * 1000 - 5* 24 * 3600 * 1000
		)
		SELECT symbol, open, close, timestamp
		FROM RankedSymbols

		WHERE rn = 1;`,
		)
		rows, err := conn.Query(context.Background(), sqlStatements)
		if err != nil {
			fmt.Println("数据库查询出错")
		}
		for rows.Next() {
			var (
				symbol string
				open   float64
				close  float64
				ts     uint64
			)
			temp := &Point{}
			err := rows.Scan(&symbol, &open, &close, &ts /* ... */)
			if err != nil {
				fmt.Println("数据绑定出错")
			}
			temp.open = open
			temp.close = close
			temp.ts = ts
			pool[symbol] = temp
		}
	}
	fmt.Println(len(pool))
	err = os.Remove("running-us.txt")
	if err != nil {
		fmt.Println("删除文件出错")
	}
	file, err := os.OpenFile("running-us.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	start := time.Now()
	for symbol, item := range pool {
		// 等待定时器触发
		// <-timer.C

		// url := "https://stock.xueqiu.com/v5/stock/quote.json?extend=detail&symbol=" + symbol
		url := "https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=" + symbol + "&begin=" + strconv.FormatInt(unixMilli, 10) + "&period=day&type=before&count=-20&indicator=kline"
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

		closePrice := make([]float64, 0)
		for i := 0; i < len(data); i++ {
			closePrice = append(closePrice, data[i][5].(float64))

			if i == len(data)-1 {
				avg := calculateMA(closePrice, 10)
				if data[i-1][5].(float64) < avg[len(avg)-2] {
					continue
				}
				if data[i][4].(float64) < item.open && data[i][5].(float64) > item.open {
					seconds := item.ts / 1000
					nanoseconds := (item.ts % 1000) * 1000000
					t := time.Unix(int64(seconds), int64(nanoseconds))
					formattedTime := t.Format("2006-01-02 15:04:05")

					str := fmt.Sprintf("%s\n        Gap点:%f(%s），10日平均：%f,开盘：%f, 最低：%f, 最高：%f, 现价：%f\n", symbol, item.open, formattedTime, avg[len(avg)-1], data[i][2].(float64), data[i][4].(float64), data[i][3].(float64), data[i][5].(float64))
					if _, ok := result[symbol]; !ok {
						result[symbol] = 1
						//SendEmail(symbol, str)
					}
					_, err = file.WriteString(str)
					if err != nil {
						fmt.Println("Error writing to file:", err)
						return
					}
				}
			}
		}
	}
	elapsed := time.Since(start) // 计算经过的时间
	fmt.Printf("耗时：%s\n", elapsed)
	loopLine := "======================================\n"
	_, err = file.WriteString(loopLine)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
func main() {
	result := make(map[string]int, 0)
	pool := make(map[string]*Point, 0)
	// 定义时间段
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return
	}

	// Define time periods (Beijing time)
	morningCnStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 9, 30, 0, 0, loc)
	morningCnEnd := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 11, 30, 0, 0, loc)
	afternoonCnStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 13, 0, 0, 0, loc)
	afternoonCnEnd := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 15, 0, 0, 0, loc)

	clearCn := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 21, 15, 0, 0, loc)
	startUs := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 21, 30, 0, 0, loc)
	endUs := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, 4, 00, 0, 0, loc)
	clearUs := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, 9, 15, 0, 0, loc)
	// 创建一个 ticker，用于定时检查
	ticker := time.NewTicker(2 * time.Second) // 每分钟检查一次
	defer ticker.Stop()

	urlStr := "http://www.xueqiu.com"
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

	for {
		select {
		case t := <-ticker.C:
			// 当前时间
			now := t
			// 检查是否在定义的时间段内
			if (now.After(morningCnStart) && now.Before(morningCnEnd)) || (now.After(afternoonCnStart) && now.Before(afternoonCnEnd)) {
				runCNTask(result, pool, cookiesString)
			} else if now.After(clearCn) && now.Before(startUs) || (now.After(clearUs) && now.Before(morningCnStart)) {
				result = make(map[string]int, 0)
				pool = make(map[string]*Point, 0)

			} else if now.After(startUs) && now.Before(endUs) {
				runUsTask(result, pool, cookiesString)
			} else {

			}
		}
	}
}
