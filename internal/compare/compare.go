package compare

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"comparer/internal/config"
	"comparer/internal/parse"
	"comparer/internal/timeplan"
	"comparer/internal/urlbuilder"
)

// BucketDiff 单个 bucket 的比对差异。
type BucketDiff struct {
	Key          string
	LenES, LenOS int
	OnlyES       []string
	OnlyOS       []string
	MissingSide  string // 缺失的一侧："es" / "os" / ""
}

// Pass 判断该 bucket 是否一致。
func (d BucketDiff) Pass() bool {
	return d.MissingSide == "" && len(d.OnlyES) == 0 && len(d.OnlyOS) == 0 && d.LenES == d.LenOS
}

// MonthResult 一个月度区间的比对结果。
type MonthResult struct {
	Start, End time.Time
	Diffs      []BucketDiff
}

// Pass 判断该月是否全量一致。
func (m MonthResult) Pass() bool {
	for _, d := range m.Diffs {
		if !d.Pass() {
			return false
		}
	}
	return true
}

// compareBucket 比较两端同一 bucket 的 id 列表，返回差异。
func compareBucket(key string, esIds, osIds []string) BucketDiff {
	diff := BucketDiff{Key: key, LenES: len(esIds), LenOS: len(osIds)}

	esSet := make(map[string]bool, len(esIds))
	for _, id := range esIds {
		esSet[id] = true
	}
	osSet := make(map[string]bool, len(osIds))
	for _, id := range osIds {
		osSet[id] = true
	}
	for id := range esSet {
		if !osSet[id] {
			diff.OnlyES = append(diff.OnlyES, id)
		}
	}
	for id := range osSet {
		if !esSet[id] {
			diff.OnlyOS = append(diff.OnlyOS, id)
		}
	}
	sort.Strings(diff.OnlyES)
	sort.Strings(diff.OnlyOS)
	return diff
}

// compareMonth 比较两端同一月份的全部 bucket，返回结果。
// debug 为 true 时打印每个 bucket 的比对过程日志。
func compareMonth(es, os parse.BucketedResponse, debug bool) MonthResult {
	keys := make(map[string]struct{}, len(es)+len(os))
	for k := range es {
		keys[k] = struct{}{}
	}
	for k := range os {
		keys[k] = struct{}{}
	}

	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	res := MonthResult{}
	for _, key := range sorted {
		esIds, esOK := es[key]
		osIds, osOK := os[key]
		switch {
		case esOK && !osOK:
			if debug {
				fmt.Println("ES 缺失 bucket:", key)
			}
			res.Diffs = append(res.Diffs, BucketDiff{Key: key, LenES: len(esIds), LenOS: 0, MissingSide: "os"})
		case !esOK && osOK:
			if debug {
				fmt.Println("OS 缺失 bucket:", key)
			}
			res.Diffs = append(res.Diffs, BucketDiff{Key: key, LenES: 0, LenOS: len(osIds), MissingSide: "es"})
		default:
			if debug {
				fmt.Println("比对 bucket:", key)
			}
			res.Diffs = append(res.Diffs, compareBucket(key, esIds, osIds))
		}
	}
	return res
}

// fetch 拉取单个 host 的数据并解析。
// debug 为 true 时打印请求 URL 与原始/归一化响应。
func fetch(client *http.Client, url string, debug bool) (parse.BucketedResponse, error) {
	if debug {
		fmt.Println("拉取 url:", url)
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if debug {
		fmt.Println("解析响应:", string(body))
	}
	data, _, err := parse.ParseResponse(body)
	if debug {
		fmt.Println("解析结果:", data)
	}
	return data, err
}

// CompareRunner 执行完整比对流程（月度循环 + 月内并发拉取 + 比对 + 报告输出）。
// 返回所有月份的比对结果（span 模式返回单月结果），便于上层计算退出码。
func CompareRunner(cfg *config.Config) ([]MonthResult, error) {
	client := cfg.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	maxDiff := cfg.MaxDiffID
	if maxDiff <= 0 {
		maxDiff = 5
	}

	modeLine := fmt.Sprintf("模式: %s=%s", cfg.Mode, cfg.Value)
	period := cfg.Period
	if period < 1 {
		period = 1
	}

	fmt.Println("== comparer 比对结果 ==")
	if cfg.Span {
		fmt.Printf("%s  ·  span（单次比对，无时间参数）\n", modeLine)
	} else {
		fmt.Printf("%s  ·  周期: %d 年（%d 个月度区间）\n", modeLine, period, period*12)
	}
	fmt.Printf("ES: %s\n", cfg.ESHost)
	fmt.Printf("OS: %s\n", cfg.OSHost)
	fmt.Println("---")

	now := cfg.Now
	if now.IsZero() {
		now = time.Now()
	}

	results := make([]MonthResult, 0, period*12)

	// -span 走单次比对
	if cfg.Span {
		es, err := fetch(client, urlbuilder.BuildURL(cfg.ESHost, cfg.Mode, cfg.Value, cfg.Step, cfg.Span, now, now), cfg.Debug)
		if err != nil {
			return nil, fmt.Errorf("拉取 es 失败: %w", err)
		}
		os, err := fetch(client, urlbuilder.BuildURL(cfg.OSHost, cfg.Mode, cfg.Value, cfg.Step, cfg.Span, now, now), cfg.Debug)
		if err != nil {
			return nil, fmt.Errorf("拉取 os 失败: %w", err)
		}
		month := compareMonth(es, os, cfg.Debug)
		month.Start, month.End = now, now
		results = append(results, month)
		printMonth(cfg.Debug, month, now, now, maxDiff)
		printSummary(results)
		return results, nil
	}

	totalPass, totalFail := 0, 0
	for _, r := range timeplan.MonthlyRanges(now, period) {
		start, end := r[0], r[1]

		var wg sync.WaitGroup
		var es, os parse.BucketedResponse
		var esErr, osErr error
		wg.Add(2)
		go func() {
			defer wg.Done()
			es, esErr = fetch(client, urlbuilder.BuildURL(cfg.ESHost, cfg.Mode, cfg.Value, cfg.Step, cfg.Span, start, end), cfg.Debug)
		}()
		go func() {
			defer wg.Done()
			os, osErr = fetch(client, urlbuilder.BuildURL(cfg.OSHost, cfg.Mode, cfg.Value, cfg.Step, cfg.Span, start, end), cfg.Debug)
		}()
		wg.Wait()

		if esErr != nil {
			log.Printf("月度 [%s ~ %s] es 拉取失败: %v（跳过该月）", start.Format(config.FmtIndexType), end.Format(config.FmtIndexType), esErr)
			totalFail++
			continue
		}
		if osErr != nil {
			log.Printf("月度 [%s ~ %s] os 拉取失败: %v（跳过该月）", start.Format(config.FmtIndexType), end.Format(config.FmtIndexType), osErr)
			totalFail++
			continue
		}

		month := compareMonth(es, os, cfg.Debug)
		month.Start, month.End = start, end
		results = append(results, month)
		if month.Pass() {
			totalPass++
		} else {
			totalFail++
		}
		printMonth(cfg.Debug, month, start, end, maxDiff)
	}

	fmt.Println("---")
	fmt.Printf("月度: %d PASS / %d FAIL\n", totalPass, totalFail)
	if totalFail > 0 {
		fmt.Println("整体结论: ✗ FAIL（以最差月份为准）")
		return results, nil
	}
	fmt.Println("整体结论: ✓ PASS")
	return results, nil
}

// printMonth 打印单个月的比对结果。
// debug 为 true 时打印每月完整细节（含各 bucket 差异）；为 false 时仅打印每月汇总一行。
func printMonth(debug bool, m MonthResult, start, end time.Time, maxDiff int) {
	if !debug {
		fmt.Printf("[%s ~ %s]  %s\n", start.Format(config.FmtIndexType), end.Format(config.FmtIndexType), statusOf(m))
		return
	}

	status := statusOf(m)
	fmt.Printf("[%s ~ %s]  %s\n", start.Format(config.FmtIndexType), end.Format(config.FmtIndexType), status)
	if m.Pass() {
		fmt.Printf("  (%d 个 bucket / 全量一致)\n", len(m.Diffs))
		return
	}
	for _, d := range m.Diffs {
		switch {
		case d.MissingSide == "os":
			fmt.Printf("  bucket %s: 缺失 os bucket\n", d.Key)
		case d.MissingSide == "es":
			fmt.Printf("  bucket %s: 缺失 es bucket\n", d.Key)
		case d.LenES != d.LenOS:
			fmt.Printf("  bucket %s: 总量不一致 es=%d  os=%d\n", d.Key, d.LenES, d.LenOS)
			printIDs("    仅 es", d.OnlyES, maxDiff)
			printIDs("    仅 os", d.OnlyOS, maxDiff)
		case len(d.OnlyES) > 0 || len(d.OnlyOS) > 0:
			fmt.Printf("  bucket %s: id 集合不一致\n", d.Key)
			printIDs("    仅 es", d.OnlyES, maxDiff)
			printIDs("    仅 os", d.OnlyOS, maxDiff)
		default:
			fmt.Printf("  bucket %s:  ✓ PASS\n", d.Key)
		}
	}
}

// statusOf 返回月度的 ✓ PASS / ✗ FAIL 状态标记。
func statusOf(m MonthResult) string {
	if m.Pass() {
		return "✓ PASS"
	}
	return "✗ FAIL"
}

// printIDs 打印 id 列表，超出 maxDiff 条时提示数量。
func printIDs(label string, ids []string, maxDiff int) {
	if len(ids) == 0 {
		return
	}
	shown := ids
	suffix := ""
	if len(ids) > maxDiff {
		shown = ids[:maxDiff]
		suffix = fmt.Sprintf("  (共 %d 条，仅显示前 %d 条)", len(ids), maxDiff)
	}
	fmt.Printf("%s: [%s]%s\n", label, joinIDs(shown), suffix)
}

// joinIDs 以逗号拼接 id 列表。
func joinIDs(ids []string) string {
	out := ""
	for i, id := range ids {
		if i > 0 {
			out += ", "
		}
		out += id
	}
	return out
}

// printSummary 打印整体汇总（span 单次模式）。
func printSummary(results []MonthResult) {
	pass, fail := 0, 0
	for _, r := range results {
		if r.Pass() {
			pass++
		} else {
			fail++
		}
	}
	fmt.Println("---")
	fmt.Printf("月度: %d PASS / %d FAIL\n", pass, fail)
	if fail > 0 {
		fmt.Println("整体结论: ✗ FAIL（以最差月份为准）")
		return
	}
	fmt.Println("整体结论: ✓ PASS")
}
