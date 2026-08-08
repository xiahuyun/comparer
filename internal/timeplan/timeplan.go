package timeplan

import "time"

// MonthlyRanges 以 today 为终点，向回推 period 年，按日历月切片为 period*12 个左闭右开区间。
// 段与段之间无间隙、无重叠。
//
// 采用"锚定日"语义：锚定 today 的日序号 anchorDay，起始日与逐月推进均以 anchorDay 对齐；
// 若目标月/年没有该日（如 31 号在 4 月、29 号在平年 2 月），则回退到该月最后一天。
func MonthlyRanges(today time.Time, period int) [][2]time.Time {
	if period < 1 {
		period = 1
	}
	anchorDay := today.Day()
	start := addYearsAnchored(today, period, anchorDay)

	ranges := make([][2]time.Time, 0, period*12)
	cur := start
	for i := 0; i < period*12; i++ {
		next := addMonthAnchored(cur, anchorDay)
		ranges = append(ranges, [2]time.Time{cur, next})
		cur = next
	}
	return ranges
}

// addYearsAnchored 返回 t 回退 n 年的起始日，日序号取 min(anchorDay, 目标年同月最后一天)。
func addYearsAnchored(t time.Time, n, anchorDay int) time.Time {
	y := t.Year() - n
	lastDay := time.Date(y, t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
	day := anchorDay
	if day > lastDay {
		day = lastDay
	}
	return time.Date(y, t.Month(), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// addMonthAnchored 返回 t 的下一个月，日序号取 min(anchorDay, 目标月最后一天)。
// 保留 t 的时分秒与时区。
func addMonthAnchored(t time.Time, anchorDay int) time.Time {
	y, m, _ := t.Date()
	firstOfNext := time.Date(y, m+1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	lastDay := firstOfNext.AddDate(0, 1, -1).Day()
	day := anchorDay
	if day > lastDay {
		day = lastDay
	}
	return time.Date(y, m+1, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}
