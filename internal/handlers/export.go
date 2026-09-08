package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"offerhunter/internal/models"
)

type exportRequest struct {
	Report models.MatchReport `json:"report"`
}

// Export handles POST /api/export, generating a Markdown report.
func Export(c *gin.Context) {
	var req exportRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"markdown": generateMarkdown(&req.Report)})
}

func generateMarkdown(report *models.MatchReport) string {
	var sb strings.Builder

	sb.WriteString("# Offer 捕手 - 匹配分析报告\n")
	fmt.Fprintf(&sb, "## 整体匹配度: %d%%\n", report.OverallScore)

	// Threshold
	sb.WriteString("## 一、硬性门槛判定\n")
	if report.ThresholdCheck.Passed {
		sb.WriteString("**已通过所有硬性门槛**\n")
	} else {
		sb.WriteString("**未完全通过硬性门槛**\n")
	}
	for _, item := range report.ThresholdCheck.Items {
		status := map[string]string{"pass": "✅", "fail": "❌", "warn": "⚠️"}[item.Status]
		fmt.Fprintf(&sb, "%s **%s**: %s - %s", status, item.Category, item.Requirement, item.Reason)
		if item.Fixable {
			sb.WriteString("   *可通过简历优化弥补*\n")
		}
		sb.WriteString("\n\n")
	}

	// Competitiveness
	sb.WriteString("## 二、竞争力评分\n")
	sb.WriteString("| 维度 | 分数 | 评语 |\n|------|------|------|\n")
	c := report.Competitiveness
	fmt.Fprintf(&sb, "| 项目经验深度 | %d | %s |\n", c.ProjectDepth.Score, c.ProjectDepth.Comment)
	fmt.Fprintf(&sb, "| 实习含金量 | %d | %s |\n", c.InternQuality.Score, c.InternQuality.Comment)
	fmt.Fprintf(&sb, "| 技术栈匹配 | %d | %s |\n", c.TechStackFit.Score, c.TechStackFit.Comment)
	fmt.Fprintf(&sb, "| **整体竞争力** | **%d** | |\n\n", c.OverallPower)

	// Hidden fit
	sb.WriteString("## 三、隐性契合分析\n")
	fmt.Fprintf(&sb, "**公司风格**: %s\n", report.HiddenFit.CompanyStyle)
	fmt.Fprintf(&sb, "**成长路径**: %s\n", report.HiddenFit.GrowthPath)
	fmt.Fprintf(&sb, "**团队文化**: %s\n", report.HiddenFit.TeamCulture)
	fmt.Fprintf(&sb, "**隐性契合度**: %d%%\n\n", report.HiddenFit.Score)

	// Optimization tips
	sb.WriteString("## 四、简历优化建议\n")
	for _, tip := range report.OptimizationTips {
		fmt.Fprintf(&sb, "### %d. %s (优先级: %d)\n", tip.Priority, tip.Category, tip.Priority)
		fmt.Fprintf(&sb, "**问题**: %s\n", tip.Problem)
		fmt.Fprintf(&sb, "**建议**: %s\n", tip.Suggestion)
		fmt.Fprintf(&sb, "**示例**:\n```\n%s\n```\n\n", tip.Example)
	}

	// Gap analysis
	sb.WriteString("## 五、能力差距分析\n")
	for _, gap := range report.GapAnalysis {
		timeline := "短期"
		if gap.Timeline == "long_term" {
			timeline = "长期"
		}
		fmt.Fprintf(&sb, "- **%s** (重要性: %s, 时间: %s)", gap.Skill, gap.Importance, timeline)
		fmt.Fprintf(&sb, "  学习资源: %s\n\n", gap.Resource)
	}

	// Summary
	sb.WriteString("## 总结\n")
	sb.WriteString(report.Summary)

	return sb.String()
}
