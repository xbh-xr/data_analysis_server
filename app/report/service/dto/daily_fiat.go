package dto

const (
	SideBuy  = "BUY"
	SideSell = "SELL"
)

// DailyFiatGetReq 按法币拆分的每日完成/对冲/利润报表查询
type DailyFiatGetReq struct {
	BeginTime    string `form:"beginTime" json:"beginTime" example:"2026-07-30"`            // 开始日期，含当天；不传默认最近 30 天
	EndTime      string `form:"endTime" json:"endTime" example:"2026-08-17"`                // 结束日期，含当天；不传默认今天
	FiatCurrency string `form:"fiatCurrency" json:"fiatCurrency" example:"AED,AUD,BRL,CAD"` // 法币过滤，逗号分隔；不传则返回区间内出现过的全部法币
	Side         string `form:"side" json:"side" example:"BUY"`                             // 可选筛选，1/BUY 或 2/SELL；不传则买卖合计
}

// DailyFiatGetReply 接口外层包装，与全局 {code, data, msg, requestId} 一致
type DailyFiatGetReply struct {
	Code      int              `json:"code" example:"200"`
	Data      DailyFiatGetResp `json:"data"`
	Msg       string           `json:"msg" example:"查询成功"`
	RequestId string           `json:"requestId"`
}

// DailyFiatGetResp 宽表结构：每一行是一天，各指标按法币展开
type DailyFiatGetResp struct {
	BeginTime  string         `json:"beginTime" example:"2026-07-30"`
	EndTime    string         `json:"endTime" example:"2026-08-17"`
	Currencies []string       `json:"currencies"`
	List       []DailyFiatRow `json:"list"`
}

// DailyFiatRow 报表一行，对应一个日期
type DailyFiatRow struct {
	GroupDate            string            `json:"groupDate" example:"2026-08-01"`
	CompletedOrderCount  map[string]int64  `json:"completedOrderCount"`
	VolumeCompleted      map[string]string `json:"volumeCompleted"`
	HedgedOrderCount     map[string]int64  `json:"hedgedOrderCount"`
	HedgedVolumeUsdt     map[string]string `json:"hedgedVolumeUsdt"`
	ActualProfit         map[string]string `json:"actualProfit"`
	EstimatedProfit      map[string]string `json:"estimatedProfit"`
	ReceivableProfit     map[string]string `json:"receivableProfit"`
	TotalActualProfit    map[string]string `json:"totalActualProfit"`
	TotalEstimatedProfit map[string]string `json:"totalEstimatedProfit"`
	TotalProfit          map[string]string `json:"totalProfit"`
}
