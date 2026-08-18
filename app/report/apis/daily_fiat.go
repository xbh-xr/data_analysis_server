package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-admin-team/go-admin-core/sdk/api"

	"go-admin/app/report/service"
	"go-admin/app/report/service/dto"
	"go-admin/common/database"
)

type DailyFiat struct {
	api.Api
}

// Get 法币每日完成/对冲/利润报表
// @Summary 法币每日完成/对冲/利润报表
// @Description 按日期行、法币列返回完成单量、USDT 完成额、对冲指标，以及已收/待收/应收利润。side 仅筛选不分组。金额字段为字符串。
// @Tags 报表
// @Accept json
// @Produce json
// @Param beginTime query string false "开始日期（含），2006-01-02 或 2006-01-02 15:04:05，不传默认最近 30 天"
// @Param endTime query string false "结束日期（含），格式同上，不传默认今天"
// @Param fiatCurrency query string false "法币过滤，逗号分隔，如 AED,AUD,BRL,CAD"
// @Param side query string false "方向筛选，1/BUY 或 2/SELL，不传则买卖合计"
// @Success 200 {object} dto.DailyFiatGetReply
// @Failure 400 {object} dto.DailyFiatGetReply "参数错误"
// @Failure 500 {object} dto.DailyFiatGetReply "查询失败 / 业务库未初始化"
// @Router /api/v1/report/daily-fiat [get]
// @Security Bearer
func (e DailyFiat) Get(c *gin.Context) {
	req := dto.DailyFiatGetReq{}
	err := e.MakeContext(c).
		Bind(&req, binding.Query).
		Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(400, err, "参数错误")
		return
	}

	db, err := database.GetBusinessDB()
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, "业务库未初始化")
		return
	}

	s := service.DailyFiat{}
	s.Log = e.Logger
	s.Orm = db.WithContext(c)

	data, err := s.Get(&req)
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	e.OK(data, "查询成功")
}
