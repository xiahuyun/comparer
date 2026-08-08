package parse

import "encoding/json"

// BucketedResponse 归一化响应：bucketKey → 记录 id 列表。
// 顶层 key 形态因业务而异（按日/按周/按年），不做识别与匹配，直接作为字符串参与集合运算。
type BucketedResponse map[string][]string

// ParseResponse 将响应 JSON 解析为 BucketedResponse。
// 返回 (结果, 每条无 id 记录对应的 bucket 计数, 错误)。
// 无 id 或 id 为空的记录被跳过，不计入 id 列表，仅累计告警计数，不视为错误。
func ParseResponse(body []byte) (BucketedResponse, map[string]int, error) {
	var raw map[string][]map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, nil, err
	}

	res := make(BucketedResponse, len(raw))
	missing := make(map[string]int)
	for key, records := range raw {
		ids := make([]string, 0, len(records))
		for _, rec := range records {
			if idVal, ok := rec["id"]; ok {
				if id, ok := idVal.(string); ok && id != "" {
					ids = append(ids, id)
					continue
				}
			}
			missing[key]++
		}
		// 保留空 bucket（即使数组为空也不丢失 key）
		res[key] = ids
	}
	return res, missing, nil
}
