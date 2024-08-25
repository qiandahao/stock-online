package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
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
	country := "cn"
	talbeName := country + "_stock_daily"
	talbeGap := country + "_gap_records"
	// tableName := "cn_stock_daily"
	file, err := os.OpenFile("running.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	// 配置ClickHouse连接参数
	options := &clickhouse.Options{
		Addr: []string{"localhost:19000"},
	}
	conn, err := clickhouse.Open(options)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer conn.Close()

	// 将文件内容转换为字符串

	err = os.Remove("running.txt")
	if err != nil {
		fmt.Println("删除文件出错")
	}

	symbolsQuery := `SELECT DISTINCT symbol FROM ` + talbeName + ` where timestamp = (select max(timestamp) from ` + talbeName + `) and market_capital > 10045411866`
	rows, err := conn.Query(context.Background(), symbolsQuery)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		symbols = append(symbols, symbol)
	}
	index := -40

	for index < -2 {
		// 创建 time.Time 对象
		startTime := now.AddDate(0, 0, index).Unix() * 1000 // 8天前的时间戳（毫秒）
		maTime := now.AddDate(0, 0, index-14).Unix() * 1000 // 8天前的时间戳（毫秒）
		seconds := startTime / 1000
		nanoseconds := (startTime % 1000) * 1000000
		t := time.Unix(int64(seconds), int64(nanoseconds))
		formattedTime := t.Format("2006-01-02 15:04:05")
		_, err = file.WriteString(formattedTime + "\n")
		wins := 0
		loses := 0
		points := 0.0
		for _, symbol := range symbols {
			frontTime := now.AddDate(0, 0, index-5).Unix() * 1000 // 5天前的时间戳（毫秒）
			backTime := now.AddDate(0, 0, index-12).Unix() * 1000 // 10天前的时间戳（毫秒）

			// 使用 fmt.Sprintf 构建 SQL 查询
			sqlStatements := fmt.Sprintf(`
			WITH ranked_records AS (
				SELECT symbol, open, close, high, low, timestamp, AVG(close) OVER (
					PARTITION BY symbol
					ORDER BY timestamp
					ROWS BETWEEN 9 PRECEDING AND CURRENT ROW
				) as avg, market_capital AS mc
				FROM %s
				WHERE timestamp > %d and symbol = '%s'
			)
				SELECT symbol, open, close, high, timestamp, avg, low
				FROM ranked_records
				WHERE timestamp > %d
				ORDER BY timestamp ASC 
				LIMIT 2`,
				talbeName,
				maTime,
				symbol,
				startTime,
			)

			rows, err := conn.Query(context.Background(), sqlStatements)
			if err != nil {
				log.Fatalf("Failed to execute query: %v", err)
			}
			defer rows.Close()
			var str string
			var op, cl, hi, lo float64
			var tss uint64
			buy := 0.0
			for rows.Next() {
				var (
					symbol string
					open   float64
					close  float64
					high   float64
					ts     uint64
					avg    float64
					low    float64
				)

				err := rows.Scan(&symbol, &open, &close, &high, &ts, &avg, &low)
				if err != nil {
					log.Fatalf("Failed to scan row: %v", err)
				}

				if str != "." && len(str) > 0 {
					if open-buy > 0 {
						wins++
					} else {
						loses++
					}
					points += (open - buy) / buy
					str += fmt.Sprintf("    	卖点 %f, 盈利 %f\n", open, (open-buy)/buy)
					continue
				} else if str == "." {
					continue
				}

				sql := fmt.Sprintf(`
					WITH RankedSymbols AS (
						SELECT *,
							ROW_NUMBER() OVER (PARTITION BY symbol ORDER BY timestamp DESC) AS rn
						FROM %s
						WHERE timestamp >= %d
						AND timestamp <= %d
						AND symbol = '%s'
					)
					SELECT symbol, open, close, high,low, timestamp
					FROM RankedSymbols
					WHERE rn = 1;`,
					talbeGap,
					backTime,
					frontTime,
					symbol,
				)

				gap, err := conn.Query(context.Background(), sql)
				if err != nil {
					log.Fatalf("Failed to execute query: %v", err)
				}
				defer gap.Close()

				for gap.Next() {
					err := gap.Scan(&symbol, &op, &cl, &hi, &lo, &tss)
					if err != nil {
						fmt.Println("gg")
						return
					}
					seconds := ts / 1000
					nanoseconds := (ts % 1000) * 1000000

					// 创建 time.Time 对象
					t := time.Unix(int64(seconds), int64(nanoseconds))
					ft := t.Format("2006-01-02 15:04:05")
					seconds = tss / 1000
					nanoseconds = (tss % 1000) * 1000000
					tt := time.Unix(int64(seconds), int64(nanoseconds))
					ftt := tt.Format("2006-01-02 15:04:05")
					if low < avg && op > avg && op < high {
						if open > op {
							buy = open
						} else {
							buy = op
						}
						str = fmt.Sprintf("%s\n        Gap点:%f(%s），10日平均：%f,买点：%f(%s), 收盘：%f\n", symbol, cl, ftt, avg, open, ft, close)
					} else {
						str = "."
					}
				}
			}

			if str != "." && len(str) > 0 {
				_, err = file.WriteString(str)
				if err != nil {
					fmt.Println("Error writing to file:", err)
					return
				}
			}
		}
		fmt.Printf("%d, 赚了：%d, 亏了：%d, 点数：%f\n", index, wins, loses, points/float64(wins+loses))
		loopLine := "======================================\n"
		_, err = file.WriteString(loopLine)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
		index = index + 1
	}

	file.Close()
}
