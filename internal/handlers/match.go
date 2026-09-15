package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"offerhunter/internal/services"
)

type matchRequest struct {
	ResumeText string `json:"resume_text"`
	JobText    string `json:"job_text"`
}

var (
	matcherSvc   = &services.Matcher{}
	optimizerSvc = &services.Optimizer{}
	gapSvc       = &services.GapAnalyzer{}

	matchCacheMu sync.Mutex
	matchCache   = map[string]any{}
)

// Match handles POST /api/match with SHA-256 caching.
func Match(c *gin.Context) {
	var req matchRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	key := cacheKey(req.ResumeText, req.JobText)

	matchCacheMu.Lock()
	if cached, ok := matchCache[key]; ok {
		matchCacheMu.Unlock()
		c.JSON(http.StatusOK, gin.H{"report": cached, "cached": true})
		return
	}
	matchCacheMu.Unlock()

	ctx := c.Request.Context()

	resume, err := resumeParser.Parse(ctx, req.ResumeText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	job, err := jdParser.Parse(ctx, req.JobText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	report, err := matcherSvc.Analyze(ctx, resume, job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	tips, err := optimizerSvc.GenerateTips(ctx, resume, job, report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	gaps, err := gapSvc.Analyze(ctx, resume, job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	report.OptimizationTips = tips
	report.GapAnalysis = gaps

	matchCacheMu.Lock()
	matchCache[key] = report
	matchCacheMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"report": report})
}

func cacheKey(resumeText, jobText string) string {
	h := sha256.Sum256([]byte(resumeText + "||" + jobText))
	return hex.EncodeToString(h[:])
}
