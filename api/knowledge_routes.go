package api

import (
	"fmt"
	"strings"

	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/gin-gonic/gin"
)

// RegisterKnowledgeRoutes sets up routes for knowledge management
func RegisterKnowledgeRoutes(router *gin.RouterGroup, kb *knowledge.KnowledgeBase) {
	router.GET("/item/:id", getKnowledgeItem(kb))
	router.POST("/item", createKnowledgeItem(kb))
	router.PUT("/item/:id", updateKnowledgeItem(kb))
	router.DELETE("/item/:id", deleteKnowledgeItem(kb))
	router.GET("/query", queryKnowledgeItems(kb))
}

// Handler to get a knowledge item by ID
func getKnowledgeItem(kb *knowledge.KnowledgeBase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "id parameter is required"})
			return
		}

		query := &knowledge.KnowledgeQuery{
			ID: id,
		}

		items, err := kb.Query(query)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve knowledge item", "details": err.Error()})
			return
		}

		if len(items) == 0 {
			c.JSON(404, gin.H{"error": "Knowledge item not found"})
			return
		}

		c.JSON(200, items[0])
	}
}

// Handler to create a new knowledge item
func createKnowledgeItem(kb *knowledge.KnowledgeBase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var item knowledge.KnowledgeItem
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(400, gin.H{"error": "Invalid knowledge item format", "details": err.Error()})
			return
		}

		err := kb.Store(item)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to store knowledge item", "details": err.Error()})
			return
		}

		c.JSON(201, gin.H{"message": "Knowledge item created", "id": item.ID})
	}
}

// Handler to update an existing knowledge item
func updateKnowledgeItem(kb *knowledge.KnowledgeBase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "id parameter is required"})
			return
		}

		var item knowledge.KnowledgeItem
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(400, gin.H{"error": "Invalid knowledge item format", "details": err.Error()})
			return
		}

		err := kb.Update(id, item)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to update knowledge item", "details": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Knowledge item updated", "id": id})
	}
}

// Handler to delete a knowledge item
func deleteKnowledgeItem(kb *knowledge.KnowledgeBase) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "id parameter is required"})
			return
		}

		err := kb.Delete(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete knowledge item", "details": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Knowledge item deleted", "id": id})
	}
}

// Handler to query multiple knowledge items
func queryKnowledgeItems(kb *knowledge.KnowledgeBase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var query knowledge.KnowledgeQuery

		// Extract query parameters
		if itemType := c.Query("type"); itemType != "" {
			query.Type = itemType
		}

		// Handle tags as a comma-separated list
		if tagList := c.Query("tags"); tagList != "" {
			query.Tags = splitAndTrim(tagList)
		}

		// Handle pagination
		if limit := c.Query("limit"); limit != "" {
			// Parse limit, default to 100 on error
			limitVal := 100
			if _, err := fmt.Sscanf(limit, "%d", &limitVal); err == nil {
				query.Limit = limitVal
			}
		}

		items, err := kb.Query(&query)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to query knowledge items", "details": err.Error()})
			return
		}

		c.JSON(200, items)
	}
}

// splitAndTrim splits a comma-separated string and trims whitespace from each element
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
