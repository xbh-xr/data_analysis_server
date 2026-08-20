package service

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"github.com/pkg/errors"

	"go-admin/app/report/service/dto"
)

const (
	maxDailyFiatRangeDays = 366
	defaultDailyFiatDays  = 30
	zeroAmount            = "0"
)

type DailyFiat struct {
	service.Service
}

type dailyFiatAggRow struct {
	GroupDate           string `gorm:"column:group_date"`
	FiatCurrency        string `gorm:"column:fiat_currency"`
	CompletedOrderCount int64  `gorm:"column:completed_order_count"`
	VolumeCompleted     string `gorm:"column:volume_completed"`
	HedgedOrderCount    int64  `gorm:"column:hedged_order_count"`
	HedgedVolumeUsdt    string `gorm:"column:hedged_volume_usdt"`
	ActualProfit        string `gorm:"column:actual_profit"`
	EstimatedProfit     string `gorm:"column:estimated_profit"`
}

// Get 查询并透视成前端报表宽表
func (e *DailyFiat) Get(req *dto.DailyFiatGetReq) (*dto.DailyFiatGetResp, error) {
	begin, endExclusive, err := parseDailyFiatRange(req.BeginTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	currencies := splitCurrencies(req.FiatCurrency)
	side, err := parseSide(req.Side)
	if err != nil {
		return nil, err
	}

	rows, err := e.queryAgg(begin, endExclusive, currencies, side)
	if err != nil {
		e.Log.Errorf("DailyFiat query error: %s", err)
		return nil, errors.New("查询失败")
	}

	endInclusive := endExclusive.Add(-time.Second)
	resp := pivotDailyFiat(begin, endInclusive, currencies, rows)
	return resp, nil
}

func (e *DailyFiat) queryAgg(begin, endExclusive time.Time, currencies []string, side int) ([]dailyFiatAggRow, error) {
	// 完成单：source_status=6，含已结算(flow_status=19)与对冲中(12)。
	// 对冲完成(hedges.status=3)只计 actual_profit，否则只计 estimated_profit。
	sql := `
SELECT
  DATE_FORMAT(DATE(o.source_created_at), '%Y-%m-%d') AS group_date,
  o.fiat_currency,
  COUNT(1) AS completed_order_count,
  CAST(CAST(COALESCE(SUM(o.fiat_amount / NULLIF(o.rate, 0)), 0) AS DECIMAL(20, 8)) AS CHAR) AS volume_completed,
  CAST(SUM(IF(IFNULL(h.hedge_cnt, 0) > 0, 1, 0)) AS SIGNED) AS hedged_order_count,
  CAST(CAST(COALESCE(SUM(
    CASE
      WHEN IFNULL(h.hedged_volume_usdt, 0) > 0 THEN h.hedged_volume_usdt
      WHEN IFNULL(h.hedge_cnt, 0) > 0 THEN o.fiat_amount / NULLIF(o.rate, 0)
      ELSE 0
    END
  ), 0) AS DECIMAL(20, 8)) AS CHAR) AS hedged_volume_usdt,
  CAST(CAST(COALESCE(SUM(IF(IFNULL(h.successful_cnt, 0) > 0, CAST(NULLIF(JSON_VALUE(o.actual_profit, '$.fiat'), '') AS DECIMAL(20, 8)), 0)), 0) AS DECIMAL(20, 8)) AS CHAR) AS actual_profit,
  CAST(CAST(COALESCE(SUM(
    IF(IFNULL(h.successful_cnt, 0) > 0, 0,
      CASE o.side
        WHEN 1 THEN CAST(NULLIF(JSON_VALUE(o.buy_check_point, '$.estimated_profit.fiat'), '') AS DECIMAL(20, 8))
        WHEN 2 THEN CAST(NULLIF(JSON_VALUE(o.sell_check_point, '$.estimated_profit.fiat'), '') AS DECIMAL(20, 8))
        ELSE 0
      END
    )
  ), 0) AS DECIMAL(20, 8)) AS CHAR) AS estimated_profit
FROM orders o
LEFT JOIN (
  SELECT
    order_number,
    COUNT(1) AS hedge_cnt,
    SUM(IF(status = 3, 1, 0)) AS successful_cnt,
    SUM(IF(hedged_rate > 0, hedged_fiat_amount / hedged_rate, 0)) AS hedged_volume_usdt
  FROM hedges
  WHERE order_created_at >= ?
    AND order_created_at < ?
  GROUP BY order_number
) h ON h.order_number = o.order_number
WHERE o.source_created_at >= ?
  AND o.source_created_at < ?
  AND o.source_status = 6
  AND o.flow_status IN (12, 19)
`
	args := []interface{}{begin, endExclusive, begin, endExclusive}
	if len(currencies) > 0 {
		placeholders := strings.Repeat("?,", len(currencies))
		sql += "  AND o.fiat_currency IN (" + placeholders[:len(placeholders)-1] + ")\n"
		for _, ccy := range currencies {
			args = append(args, ccy)
		}
	}
	if side > 0 {
		sql += "  AND o.side = ?\n"
		args = append(args, side)
	}
	sql += "GROUP BY DATE(o.source_created_at), o.fiat_currency\nORDER BY group_date, o.fiat_currency"

	var rows []dailyFiatAggRow
	if err := e.Orm.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func pivotDailyFiat(begin, endInclusive time.Time, filterCurrencies []string, rows []dailyFiatAggRow) *dto.DailyFiatGetResp {
	index := make(map[string]map[string]dailyFiatAggRow, len(rows))
	currencySet := make(map[string]struct{})
	for _, row := range rows {
		if row.FiatCurrency == "" {
			continue
		}
		currencySet[row.FiatCurrency] = struct{}{}
		byCcy, ok := index[row.GroupDate]
		if !ok {
			byCcy = make(map[string]dailyFiatAggRow)
			index[row.GroupDate] = byCcy
		}
		byCcy[row.FiatCurrency] = row
	}

	currencies := filterCurrencies
	if len(currencies) == 0 {
		currencies = make([]string, 0, len(currencySet))
		for ccy := range currencySet {
			currencies = append(currencies, ccy)
		}
		sort.Strings(currencies)
	}

	runningActual := make(map[string]string, len(currencies))
	runningEstimated := make(map[string]string, len(currencies))
	runningTotal := make(map[string]string, len(currencies))
	for _, ccy := range currencies {
		runningActual[ccy] = zeroAmount
		runningEstimated[ccy] = zeroAmount
		runningTotal[ccy] = zeroAmount
	}

	list := make([]dto.DailyFiatRow, 0)
	for day := begin; !day.After(endInclusive); day = day.AddDate(0, 0, 1) {
		dateKey := day.Format("2006-01-02")
		row := dto.DailyFiatRow{
			GroupDate:            dateKey,
			CompletedOrderCount:  make(map[string]int64, len(currencies)),
			VolumeCompleted:      make(map[string]string, len(currencies)),
			HedgedOrderCount:     make(map[string]int64, len(currencies)),
			HedgedVolumeUsdt:     make(map[string]string, len(currencies)),
			ActualProfit:         make(map[string]string, len(currencies)),
			EstimatedProfit:      make(map[string]string, len(currencies)),
			ReceivableProfit:     make(map[string]string, len(currencies)),
			TotalActualProfit:    make(map[string]string, len(currencies)),
			TotalEstimatedProfit: make(map[string]string, len(currencies)),
			TotalProfit:          make(map[string]string, len(currencies)),
		}
		byCcy := index[dateKey]
		for _, ccy := range currencies {
			actual := zeroAmount
			estimated := zeroAmount
			if agg, ok := byCcy[ccy]; ok {
				actual = normalizeAmount(agg.ActualProfit)
				estimated = normalizeAmount(agg.EstimatedProfit)
				row.CompletedOrderCount[ccy] = agg.CompletedOrderCount
				row.VolumeCompleted[ccy] = normalizeAmount(agg.VolumeCompleted)
				row.HedgedOrderCount[ccy] = agg.HedgedOrderCount
				row.HedgedVolumeUsdt[ccy] = normalizeAmount(agg.HedgedVolumeUsdt)
			} else {
				row.CompletedOrderCount[ccy] = 0
				row.VolumeCompleted[ccy] = zeroAmount
				row.HedgedOrderCount[ccy] = 0
				row.HedgedVolumeUsdt[ccy] = zeroAmount
			}
			receivable := addDecimal(actual, estimated)
			runningActual[ccy] = addDecimal(runningActual[ccy], actual)
			runningEstimated[ccy] = addDecimal(runningEstimated[ccy], estimated)
			runningTotal[ccy] = addDecimal(runningTotal[ccy], receivable)
			row.ActualProfit[ccy] = actual
			row.EstimatedProfit[ccy] = estimated
			row.ReceivableProfit[ccy] = receivable
			row.TotalActualProfit[ccy] = runningActual[ccy]
			row.TotalEstimatedProfit[ccy] = runningEstimated[ccy]
			row.TotalProfit[ccy] = runningTotal[ccy]
		}
		list = append(list, row)
	}

	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return &dto.DailyFiatGetResp{
		BeginTime:  begin.Format("2006-01-02"),
		EndTime:    endInclusive.Format("2006-01-02"),
		Currencies: currencies,
		List:       list,
	}
}

func parseDailyFiatRange(beginStr, endStr string) (time.Time, time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var begin time.Time
	var err error
	if strings.TrimSpace(beginStr) == "" {
		begin = today.AddDate(0, 0, -defaultDailyFiatDays+1)
	} else {
		begin, err = parseReportTime(beginStr, false)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("beginTime 格式错误，应为 2006-01-02")
		}
	}

	var endExclusive time.Time
	if strings.TrimSpace(endStr) == "" {
		endExclusive = today.AddDate(0, 0, 1)
	} else {
		endExclusive, err = parseReportTime(endStr, true)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("endTime 格式错误，应为 2006-01-02")
		}
	}

	if !begin.Before(endExclusive) {
		return time.Time{}, time.Time{}, errors.New("开始时间必须早于结束时间")
	}
	if endExclusive.Sub(begin) > time.Duration(maxDailyFiatRangeDays)*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("查询跨度不能超过 %d 天", maxDailyFiatRangeDays)
	}
	return begin, endExclusive, nil
}

func parseReportTime(value string, asEnd bool) (time.Time, error) {
	value = strings.TrimSpace(value)
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	var parsed time.Time
	var err error
	for _, layout := range layouts {
		parsed, err = time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			if layout == "2006-01-02" {
				if asEnd {
					return parsed.AddDate(0, 0, 1), nil
				}
				return parsed, nil
			}
			if asEnd {
				return parsed.Add(time.Second), nil
			}
			return parsed, nil
		}
	}
	return time.Time{}, err
}

func splitCurrencies(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		ccy := strings.ToUpper(strings.TrimSpace(part))
		if ccy == "" {
			continue
		}
		if _, ok := seen[ccy]; ok {
			continue
		}
		seen[ccy] = struct{}{}
		out = append(out, ccy)
	}
	return out
}

func parseSide(raw string) (int, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "", "0":
		return 0, nil
	case "1", dto.SideBuy:
		return 1, nil
	case "2", dto.SideSell:
		return 2, nil
	default:
		return 0, errors.New("side 仅支持 1/BUY 或 2/SELL")
	}
}

func normalizeAmount(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return zeroAmount
	}
	r, ok := new(big.Rat).SetString(v)
	if !ok {
		return zeroAmount
	}
	if r.Sign() == 0 {
		return zeroAmount
	}
	return strings.TrimRight(strings.TrimRight(r.FloatString(8), "0"), ".")
}

func addDecimal(a, b string) string {
	x, ok := new(big.Rat).SetString(strings.TrimSpace(a))
	if !ok {
		x = new(big.Rat)
	}
	y, ok := new(big.Rat).SetString(strings.TrimSpace(b))
	if !ok {
		y = new(big.Rat)
	}
	x.Add(x, y)
	if x.Sign() == 0 {
		return zeroAmount
	}
	return strings.TrimRight(strings.TrimRight(x.FloatString(8), "0"), ".")
}
