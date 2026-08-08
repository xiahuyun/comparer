package compare

import (
	"reflect"
	"testing"

	"comparer/internal/parse"
)

func TestComparePass(t *testing.T) {
	es := parse.BucketedResponse{"bk1": {"a", "b", "c"}}
	os := parse.BucketedResponse{"bk1": {"a", "b", "c"}}
	res := compareMonth(es, os, false)
	if !res.Pass() {
		t.Errorf("C1 期望 PASS，实际 %v", res.Diffs)
	}
}

func TestCompareIDSetDiff(t *testing.T) {
	es := parse.BucketedResponse{"bk1": {"a", "b", "c"}}
	os := parse.BucketedResponse{"bk1": {"a", "b", "d"}}
	res := compareMonth(es, os, false)
	if res.Pass() {
		t.Fatal("C2 期望 FAIL")
	}
	d := res.Diffs[0]
	if d.LenES != d.LenOS {
		t.Errorf("C2 长度应相同，实际 es=%d os=%d", d.LenES, d.LenOS)
	}
	if !reflect.DeepEqual(d.OnlyES, []string{"c"}) {
		t.Errorf("C2 仅es 期望 [c]，实际 %v", d.OnlyES)
	}
	if !reflect.DeepEqual(d.OnlyOS, []string{"d"}) {
		t.Errorf("C2 仅os 期望 [d]，实际 %v", d.OnlyOS)
	}
}

func TestCompareTotalMismatch(t *testing.T) {
	es := parse.BucketedResponse{"bk1": {"a", "b", "c"}}
	os := parse.BucketedResponse{"bk1": {"a", "b"}}
	res := compareMonth(es, os, false)
	d := res.Diffs[0]
	if d.LenES != 3 || d.LenOS != 2 {
		t.Errorf("C3 长度期望 es=3 os=2，实际 es=%d os=%d", d.LenES, d.LenOS)
	}
	if !reflect.DeepEqual(d.OnlyES, []string{"c"}) {
		t.Errorf("C3 仅es 期望 [c]，实际 %v", d.OnlyES)
	}
}

func TestCompareMissingOSBucket(t *testing.T) {
	es := parse.BucketedResponse{"bk1": {"a"}, "bk2": {"b"}}
	os := parse.BucketedResponse{"bk1": {"a"}}
	res := compareMonth(es, os, false)
	if res.Pass() {
		t.Fatal("C4 期望 FAIL")
	}
	found := false
	for _, d := range res.Diffs {
		if d.Key == "bk2" && d.MissingSide == "os" {
			found = true
		}
	}
	if !found {
		t.Errorf("C4 期望 bk2 标记缺失 os，实际 %v", res.Diffs)
	}
}

func TestCompareMissingESBucket(t *testing.T) {
	es := parse.BucketedResponse{"bk1": {"a"}}
	os := parse.BucketedResponse{"bk1": {"a"}, "bk3": {"c"}}
	res := compareMonth(es, os, false)
	if res.Pass() {
		t.Fatal("C5 期望 FAIL")
	}
	found := false
	for _, d := range res.Diffs {
		if d.Key == "bk3" && d.MissingSide == "es" {
			found = true
		}
	}
	if !found {
		t.Errorf("C5 期望 bk3 标记缺失 es，实际 %v", res.Diffs)
	}
}

func TestCompareEmptyBucketsPass(t *testing.T) {
	es := parse.BucketedResponse{"bk1": {}}
	os := parse.BucketedResponse{"bk1": {}}
	res := compareMonth(es, os, false)
	if !res.Pass() {
		t.Errorf("C6 两个空 bucket 应 PASS，实际 %v", res.Diffs)
	}
}

func TestCompareMixedBuckets(t *testing.T) {
	es := parse.BucketedResponse{
		"bk1": {"a", "b"}, // 一致
		"bk2": {"x", "y"}, // id 差集
		"bk3": {"p"},      // 缺失 os
	}
	os := parse.BucketedResponse{
		"bk1": {"a", "b"},
		"bk2": {"x", "z"},
	}
	res := compareMonth(es, os, false)
	if res.Pass() {
		t.Fatal("C7 期望 FAIL")
	}
	byKey := map[string]BucketDiff{}
	for _, d := range res.Diffs {
		byKey[d.Key] = d
	}
	if !byKey["bk1"].Pass() {
		t.Errorf("C7 bk1 应一致，实际 %v", byKey["bk1"])
	}
	if byKey["bk2"].MissingSide != "" || !reflect.DeepEqual(byKey["bk2"].OnlyES, []string{"y"}) || !reflect.DeepEqual(byKey["bk2"].OnlyOS, []string{"z"}) {
		t.Errorf("C7 bk2 差异错误: %v", byKey["bk2"])
	}
	if byKey["bk3"].MissingSide != "os" {
		t.Errorf("C7 bk3 应缺失 os，实际 %v", byKey["bk3"])
	}
}
