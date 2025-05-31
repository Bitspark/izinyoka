package api

import (
	"github.com/fullscreen-triangle/izinyoka/internal/metacognitive"
	"github.com/fullscreen-triangle/izinyoka/internal/stream"
	"github.com/gin-gonic/gin"
)

// RegisterProcessRoutes sets up routes for general processing
func RegisterProcessRoutes(router *gin.RouterGroup, orchestrator *metacognitive.Orchestrator) {
	router.POST("/job", submitJob(orchestrator))
	router.GET("/job/:id", getJobStatus(orchestrator))
	router.GET("/job/:id/results", getJobResults(orchestrator))
	router.DELETE("/job/:id", cancelJob(orchestrator))
}

// submitJob submits a new processing job to the orchestrator
func submitJob(orchestrator *metacognitive.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			InputData   map[string]interface{} `json:"inputData"`
			ProcessType string                 `json:"processType"`
			Priority    int                    `json:"priority"`
			CallbackURL string                 `json:"callbackUrl,omitempty"`
			MaxDuration int                    `json:"maxDuration,omitempty"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		// Create a job from the request
		job := &stream.Job{
			InputData:   request.InputData,
			ProcessType: request.ProcessType,
			Priority:    request.Priority,
			MaxDuration: request.MaxDuration,
		}

		// Submit the job
		jobID, err := orchestrator.SubmitJob(c.Request.Context(), job)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to submit job", "details": err.Error()})
			return
		}

		c.JSON(202, gin.H{
			"message": "Job submitted successfully",
			"jobId":   jobID,
		})
	}
}

// getJobStatus retrieves the status of a job
func getJobStatus(orchestrator *metacognitive.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")
		if jobID == "" {
			c.JSON(400, gin.H{"error": "job id parameter is required"})
			return
		}

		status, err := orchestrator.GetJobStatus(c.Request.Context(), jobID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve job status", "details": err.Error()})
			return
		}

		if status == nil {
			c.JSON(404, gin.H{"error": "Job not found"})
			return
		}

		c.JSON(200, status)
	}
}

// getJobResults retrieves the results of a completed job
func getJobResults(orchestrator *metacognitive.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")
		if jobID == "" {
			c.JSON(400, gin.H{"error": "job id parameter is required"})
			return
		}

		results, err := orchestrator.GetJobResults(c.Request.Context(), jobID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve job results", "details": err.Error()})
			return
		}

		if results == nil {
			c.JSON(404, gin.H{"error": "Job results not found"})
			return
		}

		c.JSON(200, results)
	}
}

// cancelJob cancels a running job
func cancelJob(orchestrator *metacognitive.Orchestrator) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")
		if jobID == "" {
			c.JSON(400, gin.H{"error": "job id parameter is required"})
			return
		}

		err := orchestrator.CancelJob(c.Request.Context(), jobID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to cancel job", "details": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Job canceled successfully"})
	}
}
