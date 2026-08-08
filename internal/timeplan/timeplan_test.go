package timeplan

import (
	"testing"
	"time"

	"comparer/internal/config"
)

func mustDate(t *testing.T, layout, s string) time.Time {
	t.Helper()
	d, err := time.Parse(layout, s)
	if err != nil {
		t.Fatalf("解析日期 %q 失败: %v", s, err)
	}
	return d
}

func TestMonthlyRangesDefaultOneYear(t *testing.T) {
	today := mustDate(t, config.FmtIndexType, "2026-08-03")
	ranges := MonthlyRanges(today, 1)
	if len(ranges) != 12 {
		t.Fatalf("期望 12 段，实际 %d", len(ranges))
	}
	if got := ranges[0][0].Format(config.FmtIndexType); got != "2025-08-03" {
		t.Errorf("首段 start 期望 2025-08-03，实际 %s", got)
	}
	if got := ranges[11][1].Format(config.FmtIndexType); got != "2026-08-03" {
		t.Errorf("末段 end 期望 2026-08-03，实际 %s", got)
	}
	assertContiguous(t, ranges)
}

func TestMonthlyRangesTwoYears(t *testing.T) {
	today := mustDate(t, config.FmtIndexType, "2026-08-03")
	ranges := MonthlyRanges(today, 2)
	if len(ranges) != 24 {
		t.Fatalf("期望 24 段，实际 %d", len(ranges))
	}
	if got := ranges[0][0].Format(config.FmtIndexType); got != "2024-08-03" {
		t.Errorf("首段 start 期望 2024-08-03，实际 %s", got)
	}
	assertContiguous(t, ranges)
}

func TestMonthlyRangesMonthEndCrossing(t *testing.T) {
	today := mustDate(t, config.FmtIndexType, "2026-03-31")
	ranges := MonthlyRanges(today, 1)
	// 应包含 [2025-03-31, 2025-04-30) 与 [2025-04-30, 2025-05-31)
	if got := ranges[0][0].Format(config.FmtIndexType); got != "2025-03-31" {
		t.Errorf("期望 2025-03-31，实际 %s", got)
	}
	if got := ranges[0][1].Format(config.FmtIndexType); got != "2025-04-30" {
		t.Errorf("期望 2025-04-30，实际 %s", got)
	}
	if got := ranges[1][0].Format(config.FmtIndexType); got != "2025-04-30" {
		t.Errorf("期望 2025-04-30，实际 %s", got)
	}
	if got := ranges[1][1].Format(config.FmtIndexType); got != "2025-05-31" {
		t.Errorf("期望 2025-05-31，实际 %s", got)
	}
	assertContiguous(t, ranges)
}

func TestMonthlyRangesLeapYear(t *testing.T) {
	today := mustDate(t, config.FmtIndexType, "2024-02-29")
	ranges := MonthlyRanges(today, 1)
	if got := ranges[0][0].Format(config.FmtIndexType); got != "2023-02-28" {
		t.Errorf("期望 2023-02-28，实际 %s", got)
	}
	// 回溯一年内所有 start 必须是合法日期，且段与段无缝衔接
	assertContiguous(t, ranges)
}

// assertContiguous 断言所有段左闭右开且首尾无缝、无重叠。
func assertContiguous(t *testing.T, ranges [][2]time.Time) {
	t.Helper()
	for i := 0; i < len(ranges); i++ {
		if ranges[i][0].After(ranges[i][1]) {
			t.Fatalf("段 %d 起点晚于终点: %v > %v", i, ranges[i][0], ranges[i][1])
		}
		if i > 0 && !ranges[i][0].Equal(ranges[i-1][1]) {
			t.Fatalf("段 %d 与 %d 不连续: [%v..%v) vs [%v..%v)",
				i-1, i, ranges[i-1][0], ranges[i-1][1], ranges[i][0], ranges[i][1])
		}
	}
}
