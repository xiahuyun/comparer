package parse

import (
	"reflect"
	"testing"
)

func TestParseResponseDailyBuckets(t *testing.T) {
	body := `{
	  "billing-si-resourceusage-20260804": [
		{"id": "a", "x": 1},
		{"id": "b", "x": 2}
	  ]
	}`
	got, missing, err := ParseResponse([]byte(body))
	if err != nil {
		t.Fatalf("P1 解析失败: %v", err)
	}
	want := BucketedResponse{"billing-si-resourceusage-20260804": {"a", "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("P1 期望 %v，实际 %v", want, got)
	}
	if len(missing) != 0 {
		t.Errorf("P1 不应有缺失 id 告警，实际 %v", missing)
	}
}

func TestParseResponseWeeklyBuckets(t *testing.T) {
	body := `{
	  "20260727-20260803": [
		{"id": "x", "v": 1}
	  ]
	}`
	got, _, err := ParseResponse([]byte(body))
	if err != nil {
		t.Fatalf("P2 解析失败: %v", err)
	}
	want := BucketedResponse{"20260727-20260803": {"x"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("P2 期望 %v，实际 %v", want, got)
	}
}

func TestParseResponseYearlyBuckets(t *testing.T) {
	body := `{
	  "2026": [
		{"id": "y", "m": {}}
	  ]
	}`
	got, _, err := ParseResponse([]byte(body))
	if err != nil {
		t.Fatalf("P3 解析失败: %v", err)
	}
	want := BucketedResponse{"2026": {"y"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("P3 期望 %v，实际 %v", want, got)
	}
}

func TestParseResponseMultiBucketWithEmpty(t *testing.T) {
	body := `{
	  "bk1": [{"id": "a"}],
	  "bk2": [{"id": "b"}, {"id": "c"}],
	  "bk3": []
	}`
	got, _, err := ParseResponse([]byte(body))
	if err != nil {
		t.Fatalf("P4 解析失败: %v", err)
	}
	want := BucketedResponse{"bk1": {"a"}, "bk2": {"b", "c"}, "bk3": {}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("P4 期望 %v，实际 %v", want, got)
	}
	if _, ok := got["bk3"]; !ok {
		t.Error("P4 空数组 bucket 不应丢失")
	}
}

func TestParseResponseMissingID(t *testing.T) {
	body := `{
	  "bk1": [
		{"id": "a"},
		{"name": "no-id"},
		{"id": ""},
		{"id": "b"}
	  ]
	}`
	got, missing, err := ParseResponse([]byte(body))
	if err != nil {
		t.Fatalf("P5 解析失败: %v", err)
	}
	want := BucketedResponse{"bk1": {"a", "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("P5 期望 %v，实际 %v", want, got)
	}
	if missing["bk1"] != 2 {
		t.Errorf("P5 期望 bk1 缺失 2 条，实际 %d", missing["bk1"])
	}
}
