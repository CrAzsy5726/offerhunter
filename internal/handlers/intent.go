package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"offerhunter/internal/services"
)

var intentValidator = &services.IntentValidator{}

// IntentCheck handles POST /api/intent/check.
func IntentCheck(c *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	isRelated, reason := intentValidator.IsJobRelated(c.Request.Context(), req.Text)
	c.JSON(http.StatusOK, gin.H{"is_related": isRelated, "reason": reason})
}
