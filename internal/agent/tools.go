package agent

import (
	"context"
	"fmt"
	"strings"

	"offerhunter/internal/models"
	"offerhunter/internal/services"
)

// Tool describes a callable function exposed to the LLM.
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any // JSON Schema object
}

// ToolFunc executes a tool given its JSON arguments, returning a
// natural-language summary for the agent to read.
type ToolFunc func(ctx context.Context, args map[string]string, st *ThreadState) (string, error)

var (
	resumeParser = &services.ResumeParser{}
	jdParser     = &services.JDParser{}
	matcher      = &services.Matcher{}
	optimizer    = &services.Optimizer{}
	gapAnalyzer  = &services.GapAnalyzer{}
)

func strParam(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

// allTools is the flat tool registry for the match intent (phase 1).
var allTools = []Tool{
	{
		Name: "parse_resume",
		Description: "解析用户的简历文本，提取结构化信息（教育背景、技能、项目经历、实习经历等）。" +
			"调用时机：用户提供了一份简历（文本内容较长，包含教育/技能/项目/实习等信息）。" +
			"text: 完整的简历原文，必须传入原始文本，不要做任何删减或摘要。",
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{"text": strParam("完整的简历原文")},
		},
	},
	{
		Name: "parse_jd",
		Description: "解析岗位描述(JD)文本，提取岗位名称、公司、硬性要求、软性要求、隐性偏好等信息。" +
			"调用时机：用户提供了一份招聘JD（包含岗位职责、任职要求、加分项等）。" +
			"text: 完整的JD原文，必须传入原始文本，不要做任何删减或摘要。",
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{"text": strParam("完整的JD原文")},
		},
	},
	{
		Name: "analyze_match",
		Description: "执行完整的简历-岗位匹配分析。包括三层分析：硬性门槛判定、竞争力评分（项目深度/实习含金量/技术栈匹配）、隐性契合分析，" +
			"以及简历优化建议和能力差距分析。调用时机：当简历和JD都已就绪，用户表达了匹配意愿，或双方信息已经完整时调用。" +
			"resume_text: 简历原文（如果之前已通过parse_resume解析过，可留空让工具自动从上下文获取）。" +
			"job_text: JD原文（如果之前已通过parse_jd解析过，可留空让工具自动从上下文获取）。",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"resume_text": strParam("简历原文（可留空）"),
				"job_text":    strParam("JD原文（可留空）"),
			},
		},
	},
}

// toolExec maps a tool name to its implementation.
var toolExec = map[string]ToolFunc{
	"parse_resume":  execParseResume,
	"parse_jd":      execParseJD,
	"analyze_match": execAnalyzeMatch,
}

func execParseResume(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	text := args["text"]
	resume, err := resumeParser.Parse(ctx, text)
	if err != nil {
		return "", err
	}
	st.resumeText = text
	st.resume = resume

	skills := resume.Metadata.TopSkills
	if len(skills) == 0 {
		skills = resume.Skills
		if len(skills) > 5 {
			skills = skills[:5]
		}
	}
	projects := resume.Metadata.TotalProjects
	exps := resume.Metadata.TotalExperience
	level := resume.Metadata.ExperienceLevel
	if level == "" {
		level = "未知"
	}
	levelCN := map[string]string{"entry": "应届", "junior": "1-3年", "mid": "3-5年", "senior": "5年+"}[level]
	if levelCN == "" {
		levelCN = level
	}

	return fmt.Sprintf("简历解析成功。经验水平: %s，核心技能: %s，项目经历: %d个，实习/工作经历: %d个。",
		levelCN, strings.Join(skills[:min(6, len(skills))], ", "), projects, exps), nil
}

func execParseJD(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	text := args["text"]
	job, err := jdParser.Parse(ctx, text)
	if err != nil {
		return "", err
	}
	st.jobText = text
	st.job = job

	hardReqs := []models.Requirement{}
	for _, r := range job.Requirements {
		if r.IsHard {
			hardReqs = append(hardReqs, r)
		}
	}
	title := job.Title
	if title == "" {
		title = "未知岗位"
	}
	company := ""
	if job.Company != "" {
		company = "@" + job.Company
	}
	hardPreview := []string{}
	for _, r := range hardReqs {
		if len(hardPreview) >= 4 {
			break
		}
		v := r.Value
		if len(v) > 20 {
			v = v[:20]
		}
		hardPreview = append(hardPreview, v)
	}

	return fmt.Sprintf("JD解析成功。岗位: %s %s，硬性要求: %d项（%s），软性要求: %d项，加分项: %d项。",
		title, company, len(hardReqs), strings.Join(hardPreview, ", "), len(job.Requirements)-len(hardReqs), len(job.BonusItems)), nil
}

func execAnalyzeMatch(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	resumeText := args["resume_text"]
	if resumeText == "" {
		resumeText = st.resumeText
	}
	if resumeText == "" {
		return "错误：没有找到简历信息。请先要求用户提供简历。", nil
	}
	jobText := args["job_text"]
	if jobText == "" {
		jobText = st.jobText
	}
	if jobText == "" {
		return "错误：没有找到岗位信息。请先要求用户提供JD。", nil
	}

	// Reuse already-parsed objects when the text is unchanged.
	var resume *models.Resume
	if st.resume != nil && st.resumeText == resumeText {
		resume = st.resume
	} else {
		var err error
		resume, err = resumeParser.Parse(ctx, resumeText)
		if err != nil {
			return "", err
		}
	}
	var job *models.Job
	if st.job != nil && st.jobText == jobText {
		job = st.job
	} else {
		var err error
		job, err = jdParser.Parse(ctx, jobText)
		if err != nil {
			return "", err
		}
	}

	report, err := matcher.Analyze(ctx, resume, job)
	if err != nil {
		return "", err
	}
	tips, err := optimizer.GenerateTips(ctx, resume, job, report)
	if err != nil {
		return "", err
	}
	gaps, err := gapAnalyzer.Analyze(ctx, resume, job)
	if err != nil {
		return "", err
	}
	report.OptimizationTips = tips
	report.GapAnalysis = gaps

	st.report = report
	st.resumeText = resumeText
	st.jobText = jobText

	evTips := 0
	learnTips := 0
	for _, t := range tips {
		if t.HasEvidence {
			evTips++
		} else {
			learnTips++
		}
	}
	c := report.Competitiveness

	return fmt.Sprintf("匹配分析完成！\n整体匹配度: %d%% (%s)。\n硬性门槛: %s。\n竞争力: 项目深度 %d / 实习含金量 %d / 技术栈匹配 %d / 整体 %d。\n隐性契合度: %d%%。\n优化建议: %d条有据可依 + %d条技能提升。\n能力差距: %d项需弥补。\n参考线: %d分，排名: %s。",
		report.OverallScore, grade(report.OverallScore),
		thresholdSummary(report), c.ProjectDepth.Score, c.InternQuality.Score, c.TechStackFit.Score, c.OverallPower,
		report.HiddenFit.Score, evTips, learnTips, len(gaps),
		report.Benchmark.PassLine, emptyOr(report.Benchmark.Percentile, "未评估")), nil
}

func grade(s int) string {
	switch {
	case s >= 80:
		return "优秀"
	case s >= 60:
		return "良好"
	case s >= 40:
		return "一般"
	default:
		return "需改进"
	}
}

func thresholdSummary(r *models.MatchReport) string {
	passed := 0
	for _, it := range r.ThresholdCheck.Items {
		if it.Status == "pass" {
			passed++
		}
	}
	return fmt.Sprintf("%d/%d 通过", passed, len(r.ThresholdCheck.Items))
}

func emptyOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
