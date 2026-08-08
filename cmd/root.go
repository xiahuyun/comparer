package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"comparer/internal/compare"
	"comparer/internal/config"
)

// exitCode 保存最终进程退出码：0=PASS，1=FAIL，2=参数/网络等非比对错误。
var exitCode int

// Execute 构建根命令并执行，返回进程退出码。
func Execute() int {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		return 2
	}
	return exitCode
}

// newRootCmd 构建 comparer 根命令。
func newRootCmd() *cobra.Command {
	var (
		queryType string
		indexType string
		step      string
		span      bool
		period    int
		esHost    string
		osHost    string
		logLevel  string
	)

	cmd := &cobra.Command{
		Use:   "comparer",
		Short: "比较 elasticsearch 与 opensearch 迁移前后数据是否一致",
		Long: `comparer 用于验证 elasticsearch 迁移到 opensearch 后数据是否一致。

仅比较每个 bucket 的记录 id 字段集合与数据总量，不比较其它字段。
startTime / endTime 由 comparer 自动按日历月切片计算，用户无需指定。

用法：
  comparer --query-type <query-type> [flags]
  comparer --index-type <index-type> [flags]`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Validate(queryType, indexType, step, span, period, esHost, osHost, logLevel)
			if err != nil {
				return err
			}
			results, err := compare.CompareRunner(cfg)
			if err != nil {
				return err
			}
			exitCode = 0
			for _, r := range results {
				if !r.Pass() {
					exitCode = 1
					break
				}
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&queryType, "query-type", "", "比对模式 A：透传给服务端的 queryType 值（与 --index-type 二选一）")
	f.StringVar(&indexType, "index-type", "", "比对模式 B：透传给服务端的 indexName 值（与 --query-type 二选一）")
	f.StringVar(&step, "step", "", "仅在 query-type 模式下可用，追加 step 参数（如 availability 传 day）")
	f.BoolVar(&span, "span", false, "仅在 index-type 模式下可用，表示不传 startTime/endTime")
	f.IntVar(&period, "period", 1, "比对周期（年），默认 1（=12 个月度区间），必须 >= 1")
	f.StringVar(&esHost, "es-host", "", "迁移前 elasticsearch 地址（含协议，如 http://10.250.140.11）")
	f.StringVar(&osHost, "os-host", "", "迁移后 opensearch 地址（含协议，如 http://10.250.150.11）")
	f.StringVar(&logLevel, "log", "", "日志级别：debug 时打印每月比对细节；默认仅打印每月汇总")

	return cmd
}
