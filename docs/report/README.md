# 报表接口文档（前端）

鉴权：请求头 `Authorization: Bearer <token>`，与现有后台接口相同。

统一响应：

```json
{
  "code": 200,
  "data": {},
  "msg": "查询成功",
  "requestId": ""
}
```

`code !== 200` 时按现有全局错误处理即可。金额类字段为 **string**，请用 `Number(value)` 或 decimal 库展示，不要直接当 number 做累加。

---

## 法币日报表

对应前端表格：行 = 日期（`groupDate`），列 = 各法币 × 完成/对冲/利润指标。

- **URL**：`GET /api/v1/report/daily-fiat`
- **说明**：本文档
- **OpenAPI**：[report_swagger.json](./report_swagger.json) / [report_swagger.yaml](./report_swagger.yaml)
- **Swagger UI**（非 prod）：`/swagger/report/index.html`

### Query

| 参数 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- |
| beginTime | 否 | string | 开始日期，含当天。`2006-01-02` 或 `2006-01-02 15:04:05`。不传默认最近 30 天 |
| endTime | 否 | string | 结束日期，含当天。格式同上。不传默认今天 |
| fiatCurrency | 否 | string | 法币过滤，逗号分隔，如 `AED,AUD,BRL,CAD`。不传则返回区间内出现过的全部法币 |
| side | 否 | string | 方向筛选：`1`/`BUY` 或 `2`/`SELL`。不传则买卖合计，**不做分组** |

查询跨度最大 **366** 天。`beginTime` 必须早于 `endTime`。

### 示例

```http
GET /api/v1/report/daily-fiat?beginTime=2026-07-30&endTime=2026-08-17&fiatCurrency=AED,AUD,BRL,CAD&side=SELL
```

### data 字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| beginTime | string | 实际查询起始日期 `YYYY-MM-DD` |
| endTime | string | 实际查询结束日期（含） |
| currencies | string[] | 表头法币顺序 |
| list | object[] | 按日期倒序；区间内无成交的日期也会返回，指标为 0 |

### list[] 一行

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| groupDate | string | 日期 `YYYY-MM-DD` |
| completedOrderCount | `{ [fiat]: number }` | Completed Order Count |
| volumeCompleted | `{ [fiat]: string }` | Volume completed (USDT)，`fiat_amount / rate` |
| hedgedOrderCount | `{ [fiat]: number }` | Hedged Order Count |
| hedgedVolumeUsdt | `{ [fiat]: string }` | Hedged Volume_USDT，`hedged_fiat_amount / hedged_rate` |
| actualProfit | `{ [fiat]: string }` | 已收：订单已完成且对冲已完成（hedges.status=3），取 `actual_profit.fiat` |
| estimatedProfit | `{ [fiat]: string }` | 待收：订单已完成但对冲未完成，BUY 取 `buy_check_point.estimated_profit.fiat`，SELL 取 `sell_check_point.estimated_profit.fiat` |
| receivableProfit | `{ [fiat]: string }` | 应收：当日已收 + 待收 |
| totalActualProfit | `{ [fiat]: string }` | 已收累计 |
| totalEstimatedProfit | `{ [fiat]: string }` | 待收累计 |
| totalProfit | `{ [fiat]: string }` | 应收累计 |

### 响应示例

```json
{
  "code": 200,
  "msg": "查询成功",
  "requestId": "",
  "data": {
    "beginTime": "2026-08-16",
    "endTime": "2026-08-17",
    "currencies": ["CAD"],
    "list": [
      {
        "groupDate": "2026-08-17",
        "completedOrderCount": { "CAD": 15 },
        "volumeCompleted": { "CAD": "12082.06" },
        "hedgedOrderCount": { "CAD": 15 },
        "hedgedVolumeUsdt": { "CAD": "12082.06" },
        "actualProfit": { "CAD": "-35.56431075" },
        "estimatedProfit": { "CAD": "2.32" },
        "receivableProfit": { "CAD": "-33.24431075" },
        "totalActualProfit": { "CAD": "-18.53279017" },
        "totalEstimatedProfit": { "CAD": "2.32" },
        "totalProfit": { "CAD": "-16.21279017" }
      }
    ]
  }
}
```

### 前端渲染建议

1. 表头：指标 × `data.currencies`。
2. 取数：`row.actualProfit[ccy]`、`row.estimatedProfit[ccy]`。
3. 需要只看 BUY 或 SELL 时传 `side`，不要在前端再拆方向。
4. 累计字段后端已算好。
5. 利润可能为负数。

### 统计口径

- 完成单：`source_status = 6`，且 `flow_status` 为 `19`（已结算）或 `12`（对冲中）。同一笔订单只进利润的其中一个桶：
  - 对冲已完成（`hedges.status = 3` SUCCESSFUL）→ `actualProfit`
  - 对冲未完成（无对冲 / PENDING / FAILED 等）→ `estimatedProfit`
- `side` 只作为 WHERE 筛选：`1` BUY / `2` SELL，不传则买卖合计进同一单元格
- 完成额 USDT：`fiat_amount / rate`（`rate` = 1 USD 对应的法币）
- 对冲额 USDT：`hedged_fiat_amount / hedged_rate`；对冲汇率为 0 时回退到订单汇率
- 对冲状态：0 UNKNOWN / 1 PENDING / 2 PLACING_AN_ORDER / 3 SUCCESSFUL / 4 FAILED
- 已收：对冲 SUCCESSFUL，取订单 `actual_profit.fiat`（此时不用预估利润）
- 待收：对冲非 SUCCESSFUL 时，BUY 取 `buy_check_point.estimated_profit.fiat`，SELL 取 `sell_check_point.estimated_profit.fiat`
- 应收 = 已收 + 待收
- 日期：按 `source_created_at` 的自然日
