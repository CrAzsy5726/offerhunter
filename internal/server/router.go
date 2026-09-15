package server

import (
	"github.com/gin-gonic/gin"

	"offerhunter/internal/handlers"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "offer-hunter-langchain"})
	})

	r.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File("./static/index.html")
	})

	r.Static("/static", "./static")

	api := r.Group("/api")
	{
		api.POST("/chat", handlers.Chat)
		api.POST("/resume/parse", handlers.ResumeParse)
		api.POST("/resume/upload", handlers.ResumeUpload)
		api.POST("/job/parse", handlers.JobParse)
		api.POST("/match", handlers.Match)
		api.POST("/market", handlers.Market)
		api.POST("/intent/check", handlers.IntentCheck)
		api.POST("/export", handlers.Export)
	}

	return r
}
