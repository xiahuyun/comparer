package urlbuilder

import (
	"strings"
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

func TestBuildURLQueryType(t *testing.T) {
	start := mustDate(t, config.FmtIndexType, "2026-08-03")
	end := mustDate(t, config.FmtIndexType, "2026-09-03")
	got := BuildURL("http://h", config.ModeQuery, "resourceusage", "", false, start, end)
	want := "http://h/service-integration/billing/apis/v1alpha1/si/measure/query?endTime=20260903&queryType=resourceusage&startTime=20260803"
	if got != want {
		t.Errorf("U1 期望 %q，实际 %q", want, got)
	}
}

func TestBuildURLQueryTypeWithStep(t *testing.T) {
	start := mustDate(t, config.FmtIndexType, "2026-08-02")
	end := mustDate(t, config.FmtIndexType, "2026-08-03")
	got := BuildURL("http://h", config.ModeQuery, "availability", "day", false, start, end)
	want := "http://h/service-integration/billing/apis/v1alpha1/si/measure/query?endTime=20260803&queryType=availability&startTime=20260802&step=day"
	if got != want {
		t.Errorf("U2 期望 %q，实际 %q", want, got)
	}
}

func TestBuildURLIndexType(t *testing.T) {
	start := mustDate(t, config.FmtIndexType, "2026-08-02")
	end := mustDate(t, config.FmtIndexType, "2026-08-03")
	got := BuildURL("http://h", config.ModeIndex, "servicerequests", "", false, start, end)
	want := "http://h/service-integration/billing/apis/v1alpha1/si/measure/query?endTime=2026-08-03&indexName=servicerequests&startTime=2026-08-02"
	if got != want {
		t.Errorf("U3 期望 %q，实际 %q", want, got)
	}
}

func TestBuildURLIndexTypeSpan(t *testing.T) {
	start := mustDate(t, config.FmtIndexType, "2026-08-02")
	end := mustDate(t, config.FmtIndexType, "2026-08-03")
	got := BuildURL("http://h", config.ModeIndex, "comments", "", true, start, end)
	want := "http://h/service-integration/billing/apis/v1alpha1/si/measure/query?indexName=comments"
	if got != want {
		t.Errorf("U4 期望 %q，实际 %q", want, got)
	}
	if strings.Contains(got, "startTime") || strings.Contains(got, "endTime") {
		t.Errorf("U4 span 模式不应包含时间参数，实际 %q", got)
	}
}

func TestBuildURLIndexTypeZeroPadding(t *testing.T) {
	start := mustDate(t, config.FmtIndexType, "2026-01-05")
	end := mustDate(t, config.FmtIndexType, "2026-02-05")
	got := BuildURL("http://h", config.ModeIndex, "platform", "", false, start, end)
	if !strings.Contains(got, "startTime=2026-01-05") || !strings.Contains(got, "endTime=2026-02-05") {
		t.Errorf("U5 零填充校验失败，实际 %q", got)
	}
}

func TestBuildURLQueryTypePassthroughUnknown(t *testing.T) {
	start := mustDate(t, config.FmtIndexType, "2026-08-03")
	end := mustDate(t, config.FmtIndexType, "2026-09-03")
	got := BuildURL("http://h", config.ModeQuery, "custom-type", "", false, start, end)
	if !strings.Contains(got, "queryType=custom-type") {
		t.Errorf("U6 未知 queryType 应原样透传，实际 %q", got)
	}
}

func TestBuildURLIndexTypePassthroughUnknown(t *testing.T) {
	start := mustDate(t, config.FmtIndexType, "2026-08-03")
	end := mustDate(t, config.FmtIndexType, "2026-09-03")
	got := BuildURL("http://h", config.ModeIndex, "custom-idx", "", false, start, end)
	if !strings.Contains(got, "indexName=custom-idx") {
		t.Errorf("U7 未知 indexType 应原样透传，实际 %q", got)
	}
}
