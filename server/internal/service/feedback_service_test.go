package service

import "testing"

func TestSanitize(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "去除首尾空白", in: "  hello  ", want: "hello"},
		{name: "保留换行", in: "第一行\n第二行", want: "第一行\n第二行"},
		{name: "剔除控制字符", in: "正常\x00文本\x07", want: "正常文本"},
		{name: "空字符串", in: "   ", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sanitize(tc.in); got != tc.want {
				t.Fatalf("sanitize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsHTTPURL(t *testing.T) {
	cases := map[string]bool{
		"https://example.com/report?tab=daily": true,
		"http://localhost:9090":                true,
		"ftp://example.com/file":               false,
		"example.com/report":                   false,
		"":                                     false,
	}

	for input, want := range cases {
		if got := isHTTPURL(input); got != want {
			t.Fatalf("isHTTPURL(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestTruncateCountsRunes(t *testing.T) {
	if got := truncate("反馈反馈反馈", 3); got != "反馈反" {
		t.Fatalf("truncate 中文失败: %q", got)
	}
	if got := truncate("abc", 5); got != "abc" {
		t.Fatalf("truncate 短字符串不应改变: %q", got)
	}
}
