package agent

import "testing"

func TestKeywordIntent(t *testing.T) {
	cases := []struct {
		msg    string
		expect string
	}{
		{"帮我分析一下", "match"},
		{"简历很长...JD很长...帮我分析", "match"},
		{"我想做简历和JD的匹配分析", "match"},
		{"帮我模拟面试", "interview"},
		{"我想转行到AI", "pivot"},
		{"帮我查一下薪资", "salary"},
		{"帮我制定求职计划", "strategy"},
		{"你好", ""},
	}

	for _, c := range cases {
		got, ok := keywordIntent(c.msg)
		if c.expect == "" {
			if ok {
				t.Errorf("keywordIntent(%q) = (%q, true), want no hit", c.msg, got)
			}
			continue
		}
		if !ok || got != c.expect {
			t.Errorf("keywordIntent(%q) = (%q, %v), want (%q, true)", c.msg, got, ok, c.expect)
		}
	}
}
