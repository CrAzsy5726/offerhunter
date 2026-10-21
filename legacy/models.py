from pydantic import BaseModel, Field
from typing import Optional


# ── Resume models ──

class Education(BaseModel):
    school: str = ""
    degree: str = ""
    major: str = ""
    start_date: str = ""
    end_date: str = ""
    gpa: Optional[str] = None


class Project(BaseModel):
    name: str = ""
    description: str = ""
    role: Optional[str] = None
    tech_stack: list[str] = Field(default_factory=list)
    duration: Optional[str] = None
    achievements: list[str] = Field(default_factory=list)


class Experience(BaseModel):
    company: str = ""
    position: str = ""
    start_date: str = ""
    end_date: str = ""
    description: str = ""
    achievements: list[str] = Field(default_factory=list)
    tech_stack: list[str] = Field(default_factory=list)


class ResumeMetadata(BaseModel):
    total_projects: int = 0
    total_experience: int = 0
    top_skills: list[str] = Field(default_factory=list)
    experience_level: str = ""


class Resume(BaseModel):
    id: str = ""
    raw_text: str = ""
    education: list[Education] = Field(default_factory=list)
    skills: list[str] = Field(default_factory=list)
    projects: list[Project] = Field(default_factory=list)
    experience: list[Experience] = Field(default_factory=list)
    personality: list[str] = Field(default_factory=list)
    metadata: ResumeMetadata = Field(default_factory=ResumeMetadata)


# ── Job models ──

class Requirement(BaseModel):
    category: str = ""
    value: str = ""
    is_hard: bool = False
    priority: str = "medium"


class Job(BaseModel):
    id: str = ""
    title: str = ""
    company: Optional[str] = None
    location: Optional[str] = None
    salary: Optional[str] = None
    requirements: list[Requirement] = Field(default_factory=list)
    bonus_items: list[str] = Field(default_factory=list)
    hidden_prefs: list[str] = Field(default_factory=list)
    description: str = ""
    team_info: Optional[str] = None
    growth_path: Optional[str] = None


# ── Match report models ──

class ThresholdItem(BaseModel):
    category: str = ""
    requirement: str = ""
    status: str = ""  # "pass", "fail", "warn"
    reason: str = ""
    fixable: bool = False


class ThresholdResult(BaseModel):
    passed: bool = True
    items: list[ThresholdItem] = Field(default_factory=list)


class ScoreItem(BaseModel):
    score: int = 0
    comment: str = ""
    evidence: str = ""


class Competitiveness(BaseModel):
    project_depth: ScoreItem = Field(default_factory=lambda: ScoreItem(score=0, comment="", evidence=""))
    intern_quality: ScoreItem = Field(default_factory=lambda: ScoreItem(score=0, comment="", evidence=""))
    tech_stack_fit: ScoreItem = Field(default_factory=lambda: ScoreItem(score=0, comment="", evidence=""))
    overall_power: int = 0


class HiddenFit(BaseModel):
    company_style: str = ""
    growth_path: str = ""
    team_culture: str = ""
    score: int = 0


class OptimizationTip(BaseModel):
    priority: int = 0
    category: str = ""
    problem: str = ""
    suggestion: str = ""
    example: str = ""
    has_evidence: bool = False


class Benchmark(BaseModel):
    pass_line: int = 65
    percentile: str = ""
    comment: str = ""


class GapItem(BaseModel):
    skill: str = ""
    importance: str = "medium"
    timeline: str = "short_term"
    resource: str = ""


class MatchReport(BaseModel):
    id: str = ""
    resume_id: str = ""
    job_id: str = ""
    overall_score: int = 0
    threshold_check: ThresholdResult = Field(default_factory=ThresholdResult)
    competitiveness: Competitiveness = Field(default_factory=Competitiveness)
    hidden_fit: HiddenFit = Field(default_factory=HiddenFit)
    optimization_tips: list[OptimizationTip] = Field(default_factory=list)
    gap_analysis: list[GapItem] = Field(default_factory=list)
    benchmark: Benchmark = Field(default_factory=Benchmark)
    summary: str = ""


# ── API request/response models ──

class ResumeParseRequest(BaseModel):
    text: str


class ResumeUploadRequest(BaseModel):
    text: Optional[str] = None
    images: Optional[list[str]] = None


class JobParseRequest(BaseModel):
    text: str


class MatchRequest(BaseModel):
    resume_text: str
    job_text: str


class ExportRequest(BaseModel):
    report: MatchReport


class MarketResearchRequest(BaseModel):
    job_info: str
    job_preference: str = ""


class IntentCheckRequest(BaseModel):
    text: str


class MarketInfo(BaseModel):
    salary_range: str = ""
    hot_skills: list[str] = Field(default_factory=list)
    trend: str = ""
    competition: str = ""
    summary: str = ""


class ChatRequest(BaseModel):
    message: str
    thread_id: str = ""
    mode: str = "default"  # default, interview, strategist, pivot, negotiation
