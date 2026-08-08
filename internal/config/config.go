package config

import (
	"fmt"
	"net/http"
	"time"
)

// 比对模式常量。
const (
	ModeQuery = "query" // queryType=xxx
	ModeIndex = "index" // indexName=xxx
)

// 日期格式常量：query-type 用 YYYYMMDD，index-type 用 YYYY-MM-DD。
const (
	FmtQueryType = "20060102"
	FmtIndexType = "2006-01-02"
)

// Config 比对运行配置，由 CLI flag 填充。
type Config struct {
	Mode      string // ModeQuery 或 ModeIndex
	Value     string // queryType 或 indexName 的值（原样透传）
	Step      string
	Span      bool
	Period    int
	ESHost    string
	OSHost    string
	Now       time.Time // 执行参考时间（便于测试注入）
	HTTP      *http.Client
	MaxDiffID int  // 每个 bucket 差异 id 最多展示条数
	Debug     bool // 为 true 时打印每月比对细节；默认仅打印每月汇总
}

// Validate 校验互斥约束并构造运行配置。
// logLevel 支持 "debug"：开启后打印每月比对细节，否则默认仅打印每月汇总。
func Validate(queryType, indexType, step string, span bool, period int, esHost, osHost string, logLevel string) (*Config, error) {
	// 互斥校验
	switch {
	case queryType == "" && indexType == "":
		return nil, fmt.Errorf("必须指定 --query-type 或 --index-type 其中之一")
	case queryType != "" && indexType != "":
		return nil, fmt.Errorf("--query-type 与 --index-type 不能同时指定")
	case step != "" && indexType != "":
		return nil, fmt.Errorf("--step 仅允许在 --query-type 模式下使用")
	case span && queryType != "":
		return nil, fmt.Errorf("--span 仅允许在 --index-type 模式下使用")
	}
	if period < 1 {
		return nil, fmt.Errorf("--period 必须 >= 1，当前为 %d", period)
	}
	if esHost == "" || osHost == "" {
		return nil, fmt.Errorf("必须指定 --es-host 与 --os-host")
	}

	cfg := &Config{
		Period: period,
		Step:   step,
		Span:   span,
		ESHost: esHost,
		OSHost: osHost,
		Debug:  logLevel == "debug",
	}
	if queryType != "" {
		cfg.Mode, cfg.Value = ModeQuery, queryType
	} else {
		cfg.Mode, cfg.Value = ModeIndex, indexType
	}
	return cfg, nil
}
