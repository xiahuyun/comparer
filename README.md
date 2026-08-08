# comparer

校验 elasticsearch 迁移到 opensearch 之后数据是否一致的命令行工具。

仅比较每个 bucket 的记录 **id 字段集合**与**数据总量**，不比较其它字段。

## 功能特性

- 两种比对模式：`--query-type`（按查询类型）与 `--index-type`（按索引名），二选一。
- 自动按日历月切片：以当前日期为终点，向前回推 `--period` 年，分成 `period*12` 个左闭右开、无间隙无重叠的月度区间，无需手动指定 `startTime` / `endTime`。
- es 与 os 两端并发拉取，逐月比对。
- 默认仅打印每月一行汇总；`--log=debug` 时打印每月逐 bucket 的完整比对细节。
- 退出码三态：`0` 比对全部一致（PASS）、`1` 存在不一致（FAIL）、`2` 参数或网络等非比对错误。

## 构建

```bash
go build -o comparer .
```

## 用法

```bash
# 按查询类型比对（query-type 时间戳为 YYYYMMDD）
comparer --es-host http://<es-host> --os-host http://<os-host> \
    --query-type <query-type> [--step <step>] [--period <n>]

# 按索引名比对（index-type 时间戳为 YYYY-MM-DD）
comparer --es-host http://<es-host> --os-host http://<os-host> \
    --index-type <index-name> [--span] [--period <n>]

# 打印每月比对细节
comparer --es-host http://<es-host> --os-host http://<os-host> \
    --query-type <query-type> --log=debug
```

> 注意：长 flag 一律使用**双横线**（`--`），如 `--query-type`。单横线 `-` 会被当作 shorthand 解析报错。

### 参数说明

| 参数 | 说明 | 默认值 |
|---|---|---|
| `--query-type` | 比对模式 A，透传给服务端的 `queryType` 值 | - |
| `--index-type` | 比对模式 B，透传给服务端的 `indexName` 值 | - |
| `--step` | 仅 `--query-type` 模式可用，追加 `step` 参数（如 availability 传 `day`） | - |
| `--span` | 仅 `--index-type` 模式可用，表示不传 `startTime`/`endTime` | `false` |
| `--period` | 比对周期（年），即 `period*12` 个月度区间，必须 `>=1` | `1` |
| `--es-host` | 迁移前 elasticsearch 地址（含协议，如 `http://10.250.140.11`） | - |
| `--os-host` | 迁移后 opensearch 地址（含协议，如 `http://10.250.150.11`） | - |
| `--log` | 日志级别，`debug` 时打印每月比对细节；默认仅打印每月汇总 | `""` |

## 输出示例

`--log=debug` 下的月度比对输出：

```
== comparer 比对结果 ==
模式: query=resourceusage  ·  周期: 1 年（12 个月度区间）
ES: http://10.250.140.11
OS: http://10.250.150.11
---
[2025-08-08 ~ 2025-09-08]  ✓ PASS
[2025-09-08 ~ 2025-10-08]  ✓ PASS
...
[2026-07-08 ~ 2026-08-08]  ✗ FAIL
  bucket billing-si-resourceusage-20260804: 总量不一致 es=3  os=2
    仅 es: [c637ef...]
---
月度: 11 PASS / 1 FAIL
整体结论: ✗ FAIL（以最差月份为准）
```

## 运行测试

```bash
go test ./...
```

## 目录结构

```
comparer/
├── main.go                 # 程序入口
├── cmd/
│   └── root.go            # cobra 根命令与参数解析
└── internal/
    ├── config/            # 配置定义与参数校验
    ├── timeplan/          # 月度切片算法（锚定日语义）
    ├── urlbuilder/        # URL 构造（模式 → 日期格式分派）
    ├── parse/             # JSON 响应解析与归一化
    └── compare/           # 比对逻辑与报告输出
```