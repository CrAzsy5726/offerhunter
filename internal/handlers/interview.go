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
