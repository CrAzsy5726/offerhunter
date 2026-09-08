package models

// ── Resume models ──

type Education struct {
	School    string `json:"school"`
	Degree    string `json:"degree"`
	Major     string `json:"major"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	GPA       string `json:"gpa,omitempty"`
}

type Project struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Role         string   `json:"role,omitempty"`
	TechStack    []string `json:"tech_stack"`
	Duration     string   `json:"duration,omitempty"`
	Achievements []string `json:"achievements"`
}

type Experience struct {
	Company      string   `json:"company"`
	Position     string   `json:"position"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	Description  string   `json:"description"`
	Achievements []string `json:"achievements"`
	TechStack    []string `json:"tech_stack"`
}

type ResumeMetadata struct {
	TotalProjects   int      `json:"total_projects"`
	TotalExperience int      `json:"total_experience"`
	TopSkills       []string `json:"top_skills"`
	ExperienceLevel string   `json:"experience_level"`
}

type Resume struct {
	ID          string          `json:"id"`
	RawText     string          `json:"raw_text"`
	Education   []Education     `json:"education"`
	Skills      []string        `json:"skills"`
	Projects    []Project       `json:"projects"`
	Experience  []Experience    `json:"experience"`
	Personality []string        `json:"personality"`
	Metadata    ResumeMetadata  `json:"metadata"`
}

// ── Job models ──

type Requirement struct {
	Category string `json:"category"`
	Value    string `json:"value"`
	IsHard   bool   `json:"is_hard"`
	Priority string `json:"priority"`
}

type Job struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	Company      string        `json:"company,omitempty"`
	Location     string        `json:"location,omitempty"`
	Salary       string        `json:"salary,omitempty"`
	Requirements []Requirement `json:"requirements"`
	BonusItems   []string      `json:"bonus_items"`
	HiddenPrefs  []string      `json:"hidden_prefs"`
	Description  string        `json:"description"`
	TeamInfo     string        `json:"team_info,omitempty"`
	GrowthPath   string        `json:"growth_path,omitempty"`
}

// ── Match report models ──

type ThresholdItem struct {
	Category    string `json:"category"`
	Requirement string `json:"requirement"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
	Fixable     bool   `json:"fixable"`
}

type ThresholdResult struct {
	Passed bool            `json:"passed"`
	Items  []ThresholdItem `json:"items"`
}

type ScoreItem struct {
	Score    int    `json:"score"`
	Comment  string `json:"comment"`
	Evidence string `json:"evidence"`
}

type Competitiveness struct {
	ProjectDepth  ScoreItem `json:"project_depth"`
	InternQuality ScoreItem `json:"intern_quality"`
	TechStackFit  ScoreItem `json:"tech_stack_fit"`
	OverallPower  int       `json:"overall_power"`
}

type HiddenFit struct {
	CompanyStyle string `json:"company_style"`
	GrowthPath   string `json:"growth_path"`
	TeamCulture  string `json:"team_culture"`
	Score        int    `json:"score"`
}

type OptimizationTip struct {
	Priority    int    `json:"priority"`
	Category    string `json:"category"`
	Problem     string `json:"problem"`
	Suggestion  string `json:"suggestion"`
	Example     string `json:"example"`
	HasEvidence bool   `json:"has_evidence"`
}

type Benchmark struct {
	PassLine  int    `json:"pass_line"`
	Percentile string `json:"percentile"`
	Comment   string `json:"comment"`
}

type GapItem struct {
	Skill      string `json:"skill"`
	Importance string `json:"importance"`
	Timeline   string `json:"timeline"`
	Resource   string `json:"resource"`
}

type MatchReport struct {
	ID               string            `json:"id"`
	ResumeID         string            `json:"resume_id"`
	JobID            string            `json:"job_id"`
	OverallScore     int               `json:"overall_score"`
	ThresholdCheck   ThresholdResult   `json:"threshold_check"`
	Competitiveness  Competitiveness   `json:"competitiveness"`
	HiddenFit        HiddenFit         `json:"hidden_fit"`
	OptimizationTips []OptimizationTip `json:"optimization_tips"`
	GapAnalysis      []GapItem         `json:"gap_analysis"`
	Benchmark        Benchmark         `json:"benchmark"`
	Summary          string            `json:"summary"`
}

// ── Market info ──

type MarketInfo struct {
	SalaryRange string   `json:"salary_range"`
	HotSkills   []string `json:"hot_skills"`
	Trend       string   `json:"trend"`
	Competition string   `json:"competition"`
	Summary     string   `json:"summary"`
}
