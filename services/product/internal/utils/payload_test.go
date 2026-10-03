package utils

import "testing"

// 局部更新契约:updates_json 为 JSON 对象文本,空串表示无更新。
// 类目域与商品域共用本函数,语义漂移会让两个域的局部更新行为不一致。
func TestDecodeUpdateJSON(t *testing.T) {
	t.Run("空串返回空 map 而非 nil", func(t *testing.T) {
		got, err := DecodeUpdateJSON("")
		if err != nil {
			t.Fatalf("空串不应报错: %v", err)
		}
		if got == nil {
			t.Fatal("空串应返回空 map,返回 nil 会让下游 mapstructure 报错")
		}
		if len(got) != 0 {
			t.Fatalf("空串应解码出空 map, got %v", got)
		}
	})

	t.Run("正常对象按字段名解码", func(t *testing.T) {
		got, err := DecodeUpdateJSON(`{"category_name":"手机","sort_order":3,"is_visible":false}`)
		if err != nil {
			t.Fatalf("正常 JSON 不应报错: %v", err)
		}
		if got["category_name"] != "手机" {
			t.Errorf("category_name = %v", got["category_name"])
		}
		// JSON 数字一律解成 float64:mapstructure 依赖这一点做弱类型转换
		if got["sort_order"] != float64(3) {
			t.Errorf("sort_order = %#v, want float64(3)", got["sort_order"])
		}
		if got["is_visible"] != false {
			t.Errorf("is_visible = %#v", got["is_visible"])
		}
	})

	t.Run("非法 JSON 报错", func(t *testing.T) {
		if _, err := DecodeUpdateJSON(`{"a":`); err == nil {
			t.Fatal("非法 JSON 应报错")
		}
	})

	t.Run("JSON 数组不是对象", func(t *testing.T) {
		if _, err := DecodeUpdateJSON(`[1,2,3]`); err == nil {
			t.Fatal("数组无法解成对象,应报错")
		}
	})
}

func TestJoinWarnings(t *testing.T) {
	if got := JoinWarnings(nil); got != "" {
		t.Errorf("空告警应拼成空串, got %q", got)
	}
	if got := JoinWarnings([]string{"a", "b"}); got != "a; b" {
		t.Errorf("JoinWarnings = %q, want %q", got, "a; b")
	}
}
