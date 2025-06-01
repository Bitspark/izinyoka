package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/fullscreen-triangle/izinyoka/internal/metacognitive"
	"github.com/fullscreen-triangle/izinyoka/internal/stream"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ServerConfig configures the API server
type ServerConfig struct {
	Port           int
	EnableCORS     bool
	EnableMetrics  bool
	EnableAuth     bool
	MaxRequestSize int64
	RequestTimeout time.Duration
	RateLimit      int
}

// Server represents the REST API server
type Server struct {
	config        ServerConfig
	router        *mux.Router
	httpServer    *http.Server
	orchestrator  *metacognitive.Orchestrator
	knowledgeBase *knowledge.KnowledgeBase
	logger        *logrus.Logger
	metrics       *Metrics
	shutdown      chan struct{}
	mu            sync.RWMutex
}

// Metrics tracks API performance metrics
type Metrics struct {
	RequestCount   int64         `json:"request_count"`
	ErrorCount     int64         `json:"error_count"`
	AverageLatency time.Duration `json:"average_latency"`
	UpTime         time.Duration `json:"uptime"`
	StartTime      time.Time     `json:"start_time"`
	ActiveRequests int64         `json:"active_requests"`
	mu             sync.RWMutex
}

// Request represents an incoming API request
type ProcessRequest struct {
	Query   string                 `json:"query"`
	Context map[string]interface{} `json:"context,omitempty"`
	Domain  string                 `json:"domain,omitempty"`
	Options ProcessOptions         `json:"options,omitempty"`
}

// ProcessOptions configures request processing
type ProcessOptions struct {
	Sequential     bool    `json:"sequential"`
	ConfidenceMin  float64 `json:"confidence_min"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	EnableDreaming bool    `json:"enable_dreaming"`
	IncludeMetrics bool    `json:"include_metrics"`
}

// Response represents an API response
type ProcessResponse struct {
	Results    []ProcessedResult `json:"results"`
	Confidence float64           `json:"confidence"`
	Metrics    ResponseMetrics   `json:"metrics,omitempty"`
	Error      string            `json:"error,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

// ProcessedResult represents a single processing result
type ProcessedResult struct {
	Source     string                 `json:"source"`
	Content    map[string]interface{} `json:"content"`
	Confidence float64                `json:"confidence"`
	Reasoning  string                 `json:"reasoning,omitempty"`
}

// ResponseMetrics includes processing metrics
type ResponseMetrics struct {
	ProcessingTime time.Duration            `json:"processing_time"`
	LayerTimes     map[string]time.Duration `json:"layer_times"`
	KnowledgeHits  int                      `json:"knowledge_hits"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status     string                     `json:"status"`
	Version    string                     `json:"version"`
	Uptime     string                     `json:"uptime"`
	Components map[string]ComponentHealth `json:"components"`
	Timestamp  time.Time                  `json:"timestamp"`
}

// ComponentHealth represents individual component health
type ComponentHealth struct {
	Status  string                 `json:"status"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewServer creates a new API server instance
func NewServer(config ServerConfig, orchestrator *metacognitive.Orchestrator, knowledgeBase *knowledge.KnowledgeBase) (*Server, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	metrics := &Metrics{
		StartTime: time.Now(),
	}

	server := &Server{
		config:        config,
		orchestrator:  orchestrator,
		knowledgeBase: knowledgeBase,
		logger:        logger,
		metrics:       metrics,
		shutdown:      make(chan struct{}),
	}

	if err := server.setupRouter(); err != nil {
		return nil, fmt.Errorf("failed to setup router: %w", err)
	}

	return server, nil
}

// setupRouter configures the HTTP router with all endpoints
func (s *Server) setupRouter() error {
	s.router = mux.NewRouter()

	// API versioning
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.metricsMiddleware)
	if s.config.EnableAuth {
		api.Use(s.authMiddleware)
	}

	// Health endpoints
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/ready", s.handleReady).Methods("GET")

	// Core processing endpoints
	api.HandleFunc("/process", s.handleProcess).Methods("POST")
	api.HandleFunc("/process/stream", s.handleStreamProcess).Methods("POST")

	// Knowledge base endpoints
	api.HandleFunc("/knowledge", s.handleKnowledgeQuery).Methods("GET")
	api.HandleFunc("/knowledge", s.handleKnowledgeStore).Methods("POST")
	api.HandleFunc("/knowledge/{id}", s.handleKnowledgeGet).Methods("GET")
	api.HandleFunc("/knowledge/{id}", s.handleKnowledgeUpdate).Methods("PUT")
	api.HandleFunc("/knowledge/{id}", s.handleKnowledgeDelete).Methods("DELETE")

	// System information endpoints
	api.HandleFunc("/metrics", s.handleMetrics).Methods("GET")
	api.HandleFunc("/status", s.handleStatus).Methods("GET")
	api.HandleFunc("/domains", s.handleDomains).Methods("GET")

	// Administration endpoints
	api.HandleFunc("/admin/reset", s.handleReset).Methods("POST")
	api.HandleFunc("/admin/backup", s.handleBackup).Methods("POST")

	// Static documentation (if enabled)
	s.router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/static/"))))

	return nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)

	handler := s.router
	if s.config.EnableCORS {
		corsHandler := handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		)(s.router)
		handler = corsHandler
	}

	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    s.config.RequestTimeout,
		WriteTimeout:   s.config.RequestTimeout,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	s.logger.WithField("port", s.config.Port).Info("Starting API server")

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Errorf("Server error: %v", err)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down API server")

	close(s.shutdown)

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}

// handleProcess handles cognitive processing requests
func (s *Server) handleProcess(w http.ResponseWriter, r *http.Request) {
	var req ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	startTime := time.Now()

	// Create stream data
	streamData := &stream.StreamData{
		ID:        generateRequestID(),
		Content:   req.Query,
		Metadata:  req.Context,
		Timestamp: time.Now(),
	}

	// Set processing options
	var processingMode metacognitive.ProcessingMode
	if req.Options.Sequential {
		processingMode = metacognitive.ProcessingModeSequential
	} else {
		processingMode = metacognitive.ProcessingModeConcurrent
	}

	// Process through orchestrator
	results, err := s.orchestrator.Process(r.Context(), streamData, processingMode)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Processing failed: %v", err))
		return
	}

	// Convert results to response format
	processedResults := make([]ProcessedResult, 0, len(results))
	totalConfidence := 0.0

	for _, result := range results {
		processedResults = append(processedResults, ProcessedResult{
			Source:     result.Source,
			Content:    result.Data,
			Confidence: result.Confidence,
			Reasoning:  result.Reasoning,
		})
		totalConfidence += result.Confidence
	}

	avgConfidence := totalConfidence / float64(len(results))
	processingTime := time.Since(startTime)

	response := ProcessResponse{
		Results:    processedResults,
		Confidence: avgConfidence,
		Timestamp:  time.Now(),
	}

	if req.Options.IncludeMetrics {
		response.Metrics = ResponseMetrics{
			ProcessingTime: processingTime,
			LayerTimes: map[string]time.Duration{
				"total": processingTime,
			},
		}
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleStreamProcess handles streaming processing requests
func (s *Server) handleStreamProcess(w http.ResponseWriter, r *http.Request) {
	var req ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.writeError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// Create processing channel
	resultChan := make(chan metacognitive.ProcessingResult, 10)
	errorChan := make(chan error, 1)

	// Start processing in background
	go func() {
		defer close(resultChan)
		defer close(errorChan)

		streamData := &stream.StreamData{
			ID:        generateRequestID(),
			Content:   req.Query,
			Metadata:  req.Context,
			Timestamp: time.Now(),
		}

		// Stream processing through orchestrator
		results, err := s.orchestrator.Process(r.Context(), streamData, metacognitive.ProcessingModeConcurrent)
		if err != nil {
			errorChan <- err
			return
		}

		for _, result := range results {
			select {
			case resultChan <- result:
			case <-r.Context().Done():
				return
			}
		}
	}()

	// Stream results to client
	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				// Send completion event
				fmt.Fprintf(w, "event: complete\ndata: {}\n\n")
				flusher.Flush()
				return
			}

			// Send result event
			resultData, _ := json.Marshal(ProcessedResult{
				Source:     result.Source,
				Content:    result.Data,
				Confidence: result.Confidence,
				Reasoning:  result.Reasoning,
			})

			fmt.Fprintf(w, "event: result\ndata: %s\n\n", resultData)
			flusher.Flush()

		case err := <-errorChan:
			// Send error event
			errorData, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorData)
			flusher.Flush()
			return

		case <-r.Context().Done():
			return
		}
	}
}

// handleKnowledgeQuery handles knowledge base queries
func (s *Server) handleKnowledgeQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		s.writeError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	kq := knowledge.KnowledgeQuery{
		Query: query,
		Limit: limit,
	}

	results, err := s.knowledgeBase.Query(r.Context(), kq)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Query failed: %v", err))
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"query":   query,
		"limit":   limit,
		"count":   len(results),
	})
}

// handleKnowledgeStore handles storing new knowledge
func (s *Server) handleKnowledgeStore(w http.ResponseWriter, r *http.Request) {
	var item knowledge.KnowledgeItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid knowledge item format")
		return
	}

	if err := s.knowledgeBase.Store(r.Context(), item); err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Storage failed: %v", err))
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Knowledge stored successfully",
		"id":      item.ID,
	})
}

// handleHealth handles health checks
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	components := make(map[string]ComponentHealth)

	// Check orchestrator health
	if s.orchestrator != nil && s.orchestrator.IsRunning() {
		components["orchestrator"] = ComponentHealth{Status: "healthy"}
	} else {
		components["orchestrator"] = ComponentHealth{Status: "unhealthy"}
	}

	// Check knowledge base health
	if s.knowledgeBase != nil {
		components["knowledge_base"] = ComponentHealth{Status: "healthy"}
	} else {
		components["knowledge_base"] = ComponentHealth{Status: "unhealthy"}
	}

	// Overall status
	status := "healthy"
	for _, comp := range components {
		if comp.Status != "healthy" {
			status = "unhealthy"
			break
		}
	}

	response := HealthResponse{
		Status:     status,
		Version:    "0.1.0",
		Uptime:     time.Since(s.metrics.StartTime).String(),
		Components: components,
		Timestamp:  time.Now(),
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleReady handles readiness checks
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if s.orchestrator == nil || !s.orchestrator.IsRunning() {
		s.writeError(w, http.StatusServiceUnavailable, "System not ready")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ready":     true,
		"timestamp": time.Now(),
	})
}

// handleMetrics handles metrics requests
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	s.metrics.mu.RLock()
	metrics := *s.metrics
	s.metrics.mu.RUnlock()

	metrics.UpTime = time.Since(metrics.StartTime)

	s.writeJSON(w, http.StatusOK, metrics)
}

// handleStatus handles system status requests
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"version":   "0.1.0",
		"uptime":    time.Since(s.metrics.StartTime).String(),
		"timestamp": time.Now(),
	}

	if s.orchestrator != nil && s.orchestrator.IsRunning() {
		state := s.orchestrator.GetState()
		status["orchestrator"] = map[string]interface{}{
			"running":          true,
			"processing_count": state.ProcessingCount,
			"error_count":      state.ErrorCount,
		}
	} else {
		status["orchestrator"] = map[string]interface{}{
			"running": false,
		}
	}

	s.writeJSON(w, http.StatusOK, status)
}

// handleDomains handles domain information requests
func (s *Server) handleDomains(w http.ResponseWriter, r *http.Request) {
	domains := []map[string]interface{}{
		{
			"name":        "genomics",
			"description": "Genomic variant calling and analysis",
			"available":   true,
		},
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"domains": domains,
		"count":   len(domains),
	})
}

// Middleware functions

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		s.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": time.Since(start),
			"remote":   r.RemoteAddr,
		}).Info("API request")
	})
}

func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.metrics.mu.Lock()
		s.metrics.RequestCount++
		s.metrics.ActiveRequests++
		s.metrics.mu.Unlock()

		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		s.metrics.mu.Lock()
		s.metrics.ActiveRequests--
		// Update average latency
		if s.metrics.RequestCount > 0 {
			s.metrics.AverageLatency = (s.metrics.AverageLatency*time.Duration(s.metrics.RequestCount-1) + duration) / time.Duration(s.metrics.RequestCount)
		}
		s.metrics.mu.Unlock()
	})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic authentication check
		token := r.Header.Get("Authorization")
		if token == "" {
			s.writeError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		// TODO: Implement proper token validation
		next.ServeHTTP(w, r)
	})
}

// Utility functions

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.metrics.mu.Lock()
	s.metrics.ErrorCount++
	s.metrics.mu.Unlock()

	s.writeJSON(w, status, map[string]string{
		"error":     message,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Additional handlers for completeness

func (s *Server) handleKnowledgeGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Implement get by ID
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Get by ID not yet implemented",
	})
}

func (s *Server) handleKnowledgeUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Implement update by ID
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Update by ID not yet implemented",
	})
}

func (s *Server) handleKnowledgeDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Implement delete by ID
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Delete by ID not yet implemented",
	})
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement system reset
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "System reset not yet implemented",
	})
}

func (s *Server) handleBackup(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement backup functionality
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Backup not yet implemented",
	})
}
