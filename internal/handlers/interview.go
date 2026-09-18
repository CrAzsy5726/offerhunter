package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"offerhunter/internal/models"
	"offerhunter/internal/services"
)

var interviewGrader = &services.InterviewGrader{}

type interviewGradeRequest struct {
	Question models.InterviewQuestion `json:"question"`
	Answer   string                   `json:"answer"`
}

// InterviewGrade handles POST /api/interview/grade.
func InterviewGrade(c *gin.Context) {
	var req interviewGradeRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	if req.Answer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "答案不能为空"})
		return
	}

	result, err := interviewGrader.Grade(c.Request.Context(), req.Question, req.Answer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

type interviewGenerateRequest struct {
	ThreadID   string `json:"thread_id"`
	ResumeText string `json:"resume_text"`
	JDText     string `json:"jd_text"`
}

// InterviewGenerate handles POST /api/interview/generate.
func InterviewGenerate(c *gin.Context) {
	var req interviewGenerateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	threadID := req.ThreadID
	if threadID == "" {
		threadID = newThreadID()
	}

	questions, err := globalAgent.GenerateInterview(c.Request.Context(), threadID, req.ResumeText, req.JDText)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": questions})
}
