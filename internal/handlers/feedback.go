package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"offerhunter/internal/services"
)

var feedbackStore = services.NewFeedbackStore("data/feedback.json")

type feedbackRequest struct {
	Content string `json:"content"`
}

// FeedbackList handles GET /api/feedback.
func FeedbackList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": feedbackStore.List()})
}

// FeedbackCreate handles POST /api/feedback.
func FeedbackCreate(c *gin.Context) {
	var req feedbackRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "内容不能为空"})
		return
	}
	feedbackStore.Add(content)
	c.JSON(http.StatusOK, gin.H{"items": feedbackStore.List()})
}

// FeedbackUpvote handles POST /api/feedback/:id/upvote.
func FeedbackUpvote(c *gin.Context) {
	if !feedbackStore.Upvote(c.Param("id")) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "意见不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": feedbackStore.List()})
}

// FeedbackReply handles POST /api/feedback/:id/reply.
func FeedbackReply(c *gin.Context) {
	var req feedbackRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "回复不能为空"})
		return
	}
	if !feedbackStore.AddReply(c.Param("id"), content) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "意见不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": feedbackStore.List()})
}

// FeedbackReplyUpvote handles POST /api/feedback/:id/reply/:rid/upvote.
func FeedbackReplyUpvote(c *gin.Context) {
	if !feedbackStore.UpvoteReply(c.Param("id"), c.Param("rid")) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "回复不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": feedbackStore.List()})
}
