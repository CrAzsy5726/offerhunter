package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"offerhunter/internal/services"
)

var jdParser = &services.JDParser{}

// JobParse handles POST /api/job/parse.
func JobParse(c *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	job, err := jdParser.Parse(c.Request.Context(), req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"job": job})
}
