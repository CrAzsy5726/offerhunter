package agent

import "testing"

// longResumeJD builds a realistic pasted resume+JD that clears the material
// length threshold.
func longResumeJD() string {
	return "教育背景：2019-2023 XX大学 计算机科学与技术 本科，主修数据结构、操作系统。\n" +
		"项目经历：电商订单系统，负责后端接口开发，使用 Java、Spring Boot、MySQL。\n" +
		"实习经历：2022 年在一家互联网公司实习，参与订单模块开发。\n" +
		"专业技能：Java、Spring Boot、MySQL、Redis。\n" +
		"\n" +
		"招聘 Java 后端工程师\n" +
		"岗位职责：负责后端服务的设计与开发，参与系统架构优化。\n" +
		"任职要求：本科及以上学历，熟悉 Java、Spring Boot、MySQL，有分布式开发经验优先。"
}

func TestKeywordIntent(t *testing.T) {
	material := longResumeJD()

	cases := []struct {
		msg        string
		prevIntent string
		expect     string
	}{
		// Explicit intent phrases win.
		{"帮我分析一下", "", "match"},
		{"简历很长...JD很长...帮我分析", "", "match"},
		{"我想做简历和JD的匹配分析", "", "match"},
		{"帮我模拟面试", "", "interview"},
		{"你好", "", ""},

		// Removed features (salary/strategy/pivot) no longer map to an intent;
		// they fall through so the LLM classifier routes them to "other".
		{"我想转行到AI", "", ""},
		{"帮我查一下薪资", "", ""},
		{"帮我制定求职计划", "", ""},

		// Real pasted material, no explicit phrase.
		// Cold start (no prev intent) → match.
		{material, "", "match"},
		// Continuing an interview flow → interview (not match).
		{material, "interview", "interview"},
		// Explicit "分析" overrides even an in-progress interview flow.
		{material + "\n帮我分析", "interview", "match"},

		// Regression: a bare reply inside an ongoing interview continues it.
		// "前端工程师，重庆，应届生" must not be read as pasted material.
		{"前端工程师，重庆，应届生", "interview", "interview"},
		{"北京，3年经验", "interview", "interview"},
		{"重庆", "interview", "interview"},

		// Same bare reply with no ongoing flow → no deterministic hit, so it
		// falls through to the LLM classifier (never a false "match").
		{"前端工程师，重庆，应届生", "", ""},

		// JD only / resume only → no hit.
		{"招聘Java工程师，岗位职责：负责后端开发，任职要求：熟悉Java", "", ""},
		{"教育背景：XX大学计算机本科，项目经历：电商系统，实习经历：2段", "", ""},
	}

	for _, c := range cases {
		got, ok := keywordIntent(c.msg, c.prevIntent)
		if c.expect == "" {
			if ok {
				t.Errorf("keywordIntent(%.40q, %q) = (%q, true), want no hit", c.msg, c.prevIntent, got)
			}
			continue
		}
		if !ok || got != c.expect {
			t.Errorf("keywordIntent(%.40q, %q) = (%q, %v), want (%q, true)", c.msg, c.prevIntent, got, ok, c.expect)
		}
	}
}
