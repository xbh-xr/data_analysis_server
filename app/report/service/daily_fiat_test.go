package service

import (
	"testing"
	"time"
)

func TestNormalizeAndAddDecimal(t *testing.T) {
	if got := normalizeAmount(""); got != "0" {
		t.Fatalf("empty => %s", got)
	}
	if got := normalizeAmount("12.34000000"); got != "12.34" {
		t.Fatalf("trim => %s", got)
	}
	if got := addDecimal("1.1", "2.2"); got != "3.3" {
		t.Fatalf("add => %s", got)
	}
	if got := addDecimal("-1.5", "1.5"); got != "0" {
		t.Fatalf("zero => %s", got)
	}
}

func TestParseSide(t *testing.T) {
	if v, err := parseSide(""); err != nil || v != 0 {
		t.Fatalf("empty side = %d %v", v, err)
	}
	if v, err := parseSide("BUY"); err != nil || v != 1 {
		t.Fatalf("BUY = %d %v", v, err)
	}
	if v, err := parseSide("2"); err != nil || v != 2 {
		t.Fatalf("2 = %d %v", v, err)
	}
	if _, err := parseSide("LONG"); err == nil {
		t.Fatal("expected invalid side")
	}
}

func TestPivotDailyFiatFillsDatesAndCumulative(t *testing.T) {
	begin := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, 8, 3, 23, 59, 59, 0, time.Local)
	rows := []dailyFiatAggRow{
		{GroupDate: "2026-08-01", FiatCurrency: "AED", CompletedOrderCount: 2, VolumeCompleted: "10.5", HedgedOrderCount: 1, HedgedVolumeUsdt: "8", ActualProfit: "1.25", EstimatedProfit: "0.5"},
		{GroupDate: "2026-08-03", FiatCurrency: "AED", CompletedOrderCount: 1, VolumeCompleted: "3", HedgedOrderCount: 1, HedgedVolumeUsdt: "3", ActualProfit: "-0.25", EstimatedProfit: "0"},
		{GroupDate: "2026-08-03", FiatCurrency: "CAD", CompletedOrderCount: 4, VolumeCompleted: "20", HedgedOrderCount: 4, HedgedVolumeUsdt: "20", ActualProfit: "2", EstimatedProfit: "1"},
	}

	resp := pivotDailyFiat(begin, end, nil, rows)
	if len(resp.Currencies) != 2 || resp.Currencies[0] != "AED" || resp.Currencies[1] != "CAD" {
		t.Fatalf("currencies = %#v", resp.Currencies)
	}
	if len(resp.List) != 3 {
		t.Fatalf("days = %d", len(resp.List))
	}
	if resp.List[0].GroupDate != "2026-08-03" || resp.List[2].GroupDate != "2026-08-01" {
		t.Fatalf("order = %#v %#v", resp.List[0].GroupDate, resp.List[2].GroupDate)
	}
	if resp.List[1].CompletedOrderCount["AED"] != 0 || resp.List[1].ActualProfit["AED"] != "0" {
		t.Fatalf("empty day = %#v", resp.List[1])
	}
	if resp.List[2].TotalActualProfit["AED"] != "1.25" || resp.List[2].TotalEstimatedProfit["AED"] != "0.5" || resp.List[2].TotalProfit["AED"] != "1.75" {
		t.Fatalf("day1 totals = %#v", resp.List[2])
	}
	if resp.List[1].TotalProfit["AED"] != "1.75" {
		t.Fatalf("day2 carry = %s", resp.List[1].TotalProfit["AED"])
	}
	if resp.List[0].ActualProfit["AED"] != "-0.25" || resp.List[0].TotalActualProfit["AED"] != "1" || resp.List[0].TotalProfit["AED"] != "1.5" {
		t.Fatalf("day3 aed = %#v", resp.List[0])
	}
	if resp.List[0].ReceivableProfit["CAD"] != "3" || resp.List[0].TotalProfit["CAD"] != "3" {
		t.Fatalf("cad = %#v", resp.List[0])
	}
}

func TestParseDailyFiatRangeDateOnlyInclusive(t *testing.T) {
	begin, endExclusive, err := parseDailyFiatRange("2026-08-01", "2026-08-17")
	if err != nil {
		t.Fatal(err)
	}
	if begin.Format("2006-01-02") != "2026-08-01" {
		t.Fatalf("begin = %s", begin)
	}
	if endExclusive.Format("2006-01-02") != "2026-08-18" {
		t.Fatalf("endExclusive = %s", endExclusive)
	}
}
