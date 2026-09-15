package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"offerhunter/internal/models"
	"offerhunter/internal/services"
)

type resumeUploadRequest struct {
	Text   string   `json:"text"`
	Images []string `json:"images"`
}

var resumeParser = &services.ResumeParser{}

// ResumeParse handles POST /api/resume/parse (plain text).
func ResumeParse(c *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	resume, err := resumeParser.Parse(c.Request.Context(), req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"resume": resume})
}

// ResumeUpload handles POST /api/resume/upload (text or image OCR).
func ResumeUpload(c *gin.Context) {
	var req resumeUploadRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if len(req.Images) > 0 {
		text, err := resumeParser.ParseFromImages(c.Request.Context(), req.Images)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		resume := &models.Resume{RawText: text}
		c.JSON(http.StatusOK, gin.H{"resume": resume})
		return
	}

	if req.Text != "" {
		resume, err := resumeParser.Parse(c.Request.Context(), req.Text)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"resume": resume})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"detail": "No text or images provided"})
}
