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
		// resume + JD pasted together, no explicit "分析"
		{"教育背景：XX大学计算机本科\n项目经历：电商系统\n\n招聘Java工程师，岗位职责：负责后端开发，任职要求：熟悉Java", "match"},
		{"学校：XX大学 学历：本科 技能：Java\n\n岗位：Java后端 薪资范围：15-25K", "match"},
		// JD only, no resume
		{"招聘Java工程师，岗位职责：负责后端开发，任职要求：熟悉Java", ""},
		// resume only, no JD
		{"教育背景：XX大学计算机本科，项目经历：电商系统，实习经历：2段", ""},
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
