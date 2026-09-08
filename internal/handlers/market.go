package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"offerhunter/internal/services"
)

type marketRequest struct {
	JobInfo       string `json:"job_info"`
	JobPreference string `json:"job_preference"`
}

var marketSvc = &services.MarketResearchService{}

// Market handles POST /api/market.
func Market(c *gin.Context) {
	var req marketRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	info, err := marketSvc.Research(c.Request.Context(), req.JobInfo, req.JobPreference)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"market": info})
}
