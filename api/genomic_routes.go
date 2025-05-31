package api

import (
	"github.com/fullscreen-triangle/izinyoka/internal/domain/genomic"
	"github.com/fullscreen-triangle/izinyoka/internal/metacognitive"
	"github.com/gin-gonic/gin"
)

// RegisterGenomicRoutes sets up routes for genomic variant calling
func RegisterGenomicRoutes(router *gin.RouterGroup, orchestrator *metacognitive.Orchestrator) {
	router.POST("/variants", analyzeVariants(orchestrator))
	router.POST("/alignment", alignSequences(orchestrator))
	router.GET("/reference/:id", getReferenceGenome(orchestrator))
}

// analyzeVariants handles genomic variant analysis requests
func analyzeVariants(orchestrator *metacognitive.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			SampleID        string `json:"sampleId" binding:"required"`
			ReferenceID     string `json:"referenceId" binding:"required"`
			ConfidenceLevel int    `json:"confidenceLevel"`
			IncludeFiltered bool   `json:"includeFiltered"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		// Create a genomic variant analysis request
		analysisRequest := &genomic.VariantAnalysisRequest{
			SampleID:        request.SampleID,
			ReferenceID:     request.ReferenceID,
			ConfidenceLevel: request.ConfidenceLevel,
			IncludeFiltered: request.IncludeFiltered,
		}

		// Send to domain-specific handler
		jobID, err := orchestrator.SubmitDomainJob(c.Request.Context(), "genomic", "variant_analysis", analysisRequest)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to submit variant analysis job", "details": err.Error()})
			return
		}

		c.JSON(202, gin.H{
			"message": "Variant analysis job submitted",
			"jobId":   jobID,
		})
	}
}

// alignSequences handles genomic sequence alignment requests
func alignSequences(orchestrator *metacognitive.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			QuerySequence   string `json:"querySequence" binding:"required"`
			ReferenceID     string `json:"referenceId" binding:"required"`
			AlignmentMethod string `json:"alignmentMethod"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		// Create an alignment request
		alignRequest := &genomic.SequenceAlignmentRequest{
			QuerySequence:   request.QuerySequence,
			ReferenceID:     request.ReferenceID,
			AlignmentMethod: request.AlignmentMethod,
		}

		// Send to domain-specific handler
		jobID, err := orchestrator.SubmitDomainJob(c.Request.Context(), "genomic", "sequence_alignment", alignRequest)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to submit alignment job", "details": err.Error()})
			return
		}

		c.JSON(202, gin.H{
			"message": "Sequence alignment job submitted",
			"jobId":   jobID,
		})
	}
}

// getReferenceGenome retrieves information about a reference genome
func getReferenceGenome(orchestrator *metacognitive.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "Reference genome id is required"})
			return
		}

		// Query the domain-specific system for reference genome info
		refGenome, err := orchestrator.QueryDomainResource(c.Request.Context(), "genomic", "reference_genome", id)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve reference genome", "details": err.Error()})
			return
		}

		if refGenome == nil {
			c.JSON(404, gin.H{"error": "Reference genome not found"})
			return
		}

		c.JSON(200, refGenome)
	}
}
