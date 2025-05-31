package api

import (
	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/fullscreen-triangle/izinyoka/internal/metabolic"
	"github.com/fullscreen-triangle/izinyoka/internal/metacognitive"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the API router with all routes
func SetupRouter(
	knowledgeBase *knowledge.KnowledgeBase,
	orchestrator *metacognitive.Orchestrator,
	dreamingModule *metabolic.DreamingModule,
) *gin.Engine {
	// Set to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Register API routes
	v1 := router.Group("/api/v1")
	{
		// Knowledge endpoints
		knowledgeRoutes := v1.Group("/knowledge")
		RegisterKnowledgeRoutes(knowledgeRoutes, knowledgeBase)

		// Process/job endpoints
		processRoutes := v1.Group("/process")
		RegisterProcessRoutes(processRoutes, orchestrator)

		// Dreaming endpoints
		dreamRoutes := v1.Group("/dreams")
		RegisterDreamRoutes(dreamRoutes, dreamingModule)

		// Domain-specific endpoints
		// Here we include a genomic domain as an example
		genomicRoutes := v1.Group("/genomic")
		RegisterGenomicRoutes(genomicRoutes, orchestrator)
	}

	// Serve static files if needed
	router.Static("/docs", "./web/docs")
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	// Serve the API docs
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/docs")
	})

	return router
}
