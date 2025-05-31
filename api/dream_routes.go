package api

import (
	"fmt"

	"github.com/fullscreen-triangle/izinyoka/internal/metabolic"
	"github.com/gin-gonic/gin"
)

// RegisterDreamRoutes sets up routes for the dreaming module
func RegisterDreamRoutes(router *gin.RouterGroup, dm *metabolic.DreamingModule) {
	router.POST("/generate", generateDream(dm))
	router.GET("/retrieve/:id", retrieveDream(dm))
	router.GET("/list", listDreams(dm))
	router.POST("/feedback/:id", provideDreamFeedback(dm))
}

// Handler to generate a new dream
func generateDream(dm *metabolic.DreamingModule) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Context  map[string]interface{} `json:"context"`
			Duration int                    `json:"duration"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		dream, err := dm.GenerateDream(c.Request.Context(), request.Context, request.Duration)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate dream", "details": err.Error()})
			return
		}

		c.JSON(200, dream)
	}
}

// Handler to retrieve a dream by ID
func retrieveDream(dm *metabolic.DreamingModule) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "id parameter is required"})
			return
		}

		dream, err := dm.GetDream(c.Request.Context(), id)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve dream", "details": err.Error()})
			return
		}

		if dream == nil {
			c.JSON(404, gin.H{"error": "Dream not found"})
			return
		}

		c.JSON(200, dream)
	}
}

// Handler to list dreams
func listDreams(dm *metabolic.DreamingModule) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 10 // Default limit

		if limitParam := c.Query("limit"); limitParam != "" {
			if _, err := fmt.Sscanf(limitParam, "%d", &limit); err != nil {
				c.JSON(400, gin.H{"error": "Invalid limit parameter"})
				return
			}
		}

		dreams, err := dm.ListDreams(c.Request.Context(), limit)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to list dreams", "details": err.Error()})
			return
		}

		c.JSON(200, dreams)
	}
}

// Handler to provide feedback on a dream
func provideDreamFeedback(dm *metabolic.DreamingModule) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "id parameter is required"})
			return
		}

		var feedback struct {
			Useful    bool   `json:"useful"`
			Comments  string `json:"comments"`
			ApplyToKB bool   `json:"applyToKnowledgeBase"`
		}

		if err := c.ShouldBindJSON(&feedback); err != nil {
			c.JSON(400, gin.H{"error": "Invalid feedback format", "details": err.Error()})
			return
		}

		err := dm.ProvideFeedback(c.Request.Context(), id, feedback.Useful, feedback.Comments, feedback.ApplyToKB)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to process feedback", "details": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Feedback processed successfully"})
	}
}
