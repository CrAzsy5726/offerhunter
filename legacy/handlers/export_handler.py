"""Markdown export endpoint."""
from fastapi import APIRouter, HTTPException
from models import ExportRequest

router = APIRouter(prefix="/api", tags=["export"])


@router.post("/export")
async def export_report(req: ExportRequest):
    try:
        markdown = generate_markdown(req.report)
        return {"markdown": markdown}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


def generate_markdown(report) -> str:
    sb = []

    sb.append("# Offer 捕手 - 匹配分析报告\n")
    sb.append(f"## 整体匹配度: {report.overall_score}%\n")

    # Threshold
    sb.append("## 一、硬性门槛判定\n")
    if report.threshold_check.passed:
        sb.append("**已通过所有硬性门槛**\n")
    else:
        sb.append("**未完全通过硬性门槛**\n")

    for item in report.threshold_check.items:
        status = {"pass": "✅", "fail": "❌", "warn": "⚠️"}.get(item.status, "")
        sb.append(f"{status} **{item.category}**: {item.requirement} - {item.reason}")
        if item.fixable:
            sb.append("   *可通过简历优化弥补*\n")
        sb.append("")

    # Competitiveness
    sb.append("## 二、竞争力评分\n")
    sb.append("| 维度 | 分数 | 评语 |")
    sb.append("|------|------|------|")
    c = report.competitiveness
    sb.append(f"| 项目经验深度 | {c.project_depth.score} | {c.project_depth.comment} |")
    sb.append(f"| 实习含金量 | {c.intern_quality.score} | {c.intern_quality.comment} |")
    sb.append(f"| 技术栈匹配 | {c.tech_stack_fit.score} | {c.tech_stack_fit.comment} |")
    sb.append(f"| **整体竞争力** | **{c.overall_power}** | |\n")

    # Hidden fit
    sb.append("## 三、隐性契合分析\n")
    sb.append(f"**公司风格**: {report.hidden_fit.company_style}\n")
    sb.append(f"**成长路径**: {report.hidden_fit.growth_path}\n")
    sb.append(f"**团队文化**: {report.hidden_fit.team_culture}\n")
    sb.append(f"**隐性契合度**: {report.hidden_fit.score}%\n")

    # Optimization tips
    sb.append("## 四、简历优化建议\n")
    for tip in report.optimization_tips:
        sb.append(f"### {tip.priority}. {tip.category} (优先级: {tip.priority})\n")
        sb.append(f"**问题**: {tip.problem}\n")
        sb.append(f"**建议**: {tip.suggestion}\n")
        sb.append(f"**示例**:\n```\n{tip.example}\n```\n")

    # Gap analysis
    sb.append("## 五、能力差距分析\n")
    for gap in report.gap_analysis:
        timeline = "短期" if gap.timeline == "long_term" else "长期"
        sb.append(f"- **{gap.skill}** (重要性: {gap.importance}, 时间: {timeline})")
        sb.append(f"  学习资源: {gap.resource}\n")

    # Summary
    sb.append("## 总结\n")
    sb.append(report.summary)

    return "\n".join(sb)
