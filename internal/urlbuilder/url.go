package urlbuilder

import (
	"net/url"
	"time"

	"comparer/internal/config"
)

// BuildURL 是唯一的模式→日期格式分派点，构造完整的查询 URL。
// host 为数据源地址（含协议），mode 为 ModeQuery 或 ModeIndex。
// span 仅对 index 模式生效：为 true 时不携带 startTime/endTime。
func BuildURL(host, mode, value, step string, span bool, start, end time.Time) string {
	q := url.Values{}
	switch mode {
	case config.ModeQuery:
		q.Set("startTime", start.Format(config.FmtQueryType))
		q.Set("endTime", end.Format(config.FmtQueryType))
		q.Set("queryType", value)
		if step != "" {
			q.Set("step", step)
		}
	case config.ModeIndex:
		if !span {
			q.Set("startTime", start.Format(config.FmtIndexType))
			q.Set("endTime", end.Format(config.FmtIndexType))
		}
		q.Set("indexName", value)
	}
	return host + "/service-integration/billing/apis/v1alpha1/si/measure/query?" + q.Encode()
}
