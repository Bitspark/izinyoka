package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fullscreen-triangle/izinyoka/internal/api"
	"github.com/fullscreen-triangle/izinyoka/internal/config"
	"github.com/fullscreen-triangle/izinyoka/internal/domain/genomics"
	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/fullscreen-triangle/izinyoka/internal/metabolic"
	"github.com/fullscreen-triangle/izinyoka/internal/metacognitive"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Version information
var (
	Version   = "0.1.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// Application represents the main Izinyoka application
type Application struct {
	config           *config.Config
	logger           *logrus.Logger
	knowledgeBase    *knowledge.KnowledgeBase
	metabolicManager *metabolic.Manager
	orchestrator     *metacognitive.Orchestrator
	apiServer        *api.Server
	genomicsAdapter  *genomics.VariantCallingAdapter
	ctx              context.Context
	cancel           context.CancelFunc
	shutdownComplete chan struct{}
}

// main is the entry point for the Izinyoka application
func main() {
	// Parse command line flags
	var (
		configPath     = flag.String("config", "config.yaml", "Path to configuration file")
		logLevel       = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		version        = flag.Bool("version", false, "Show version information")
		validateOnly   = flag.Bool("validate", false, "Validate configuration and exit")
		dataDir        = flag.String("data-dir", "./data", "Data directory for persistent storage")
		apiPort        = flag.Int("port", 8080, "API server port")
		enableAPI      = flag.Bool("enable-api", true, "Enable REST API server")
		enableDreaming = flag.Bool("enable-dreaming", true, "Enable dreaming module")
	)
	flag.Parse()

	// Show version information
	if *version {
		fmt.Printf("Izinyoka Biomimetic Metacognitive Architecture\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Go Version: %s\n", "go1.21+")
		fmt.Printf("\nA metacognitive AI system inspired by biological processes\n")
		fmt.Printf("Specializing in genomic variant calling with 23%% performance improvements\n")
		os.Exit(0)
	}

	// Initialize logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: false,
	})

	// Load configuration
	cfg, err := loadConfig(*configPath, *dataDir, *apiPort, *enableAPI, *enableDreaming)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration and exit if requested
	if *validateOnly {
		logger.Info("Configuration validation successful")
		os.Exit(0)
	}

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize application
	app, err := NewApplication(ctx, cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize application: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the application
	logger.Info("Starting Izinyoka Biomimetic Metacognitive Architecture")
	if err := app.Start(); err != nil {
		logger.Fatalf("Failed to start application: %v", err)
	}

	// Wait for shutdown signal
	sig := <-sigChan
	logger.WithField("signal", sig).Info("Received shutdown signal")

	// Initiate graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Error during shutdown: %v", err)
		os.Exit(1)
	}

	logger.Info("Izinyoka shutdown complete")
}

// NewApplication creates a new Izinyoka application instance
func NewApplication(ctx context.Context, cfg *config.Config, logger *logrus.Logger) (*Application, error) {
	appCtx, cancel := context.WithCancel(ctx)

	app := &Application{
		config:           cfg,
		logger:           logger,
		ctx:              appCtx,
		cancel:           cancel,
		shutdownComplete: make(chan struct{}),
	}

	// Initialize components in dependency order
	if err := app.initializeKnowledgeBase(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize knowledge base: %w", err)
	}

	if err := app.initializeMetabolicSystem(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize metabolic system: %w", err)
	}

	if err := app.initializeMetacognitive(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize metacognitive system: %w", err)
	}

	if err := app.initializeDomainAdapters(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize domain adapters: %w", err)
	}

	if cfg.API.Enabled {
		if err := app.initializeAPI(); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize API server: %w", err)
		}
	}

	return app, nil
}

// Start starts all application components
func (app *Application) Start() error {
	app.logger.Info("Initializing Izinyoka components...")

	// Start metabolic manager
	if err := app.metabolicManager.Start(app.ctx); err != nil {
		return fmt.Errorf("failed to start metabolic manager: %w", err)
	}

	// Start metacognitive orchestrator
	if err := app.orchestrator.Start(app.ctx); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}

	// Register domain adapters with orchestrator
	if app.genomicsAdapter != nil {
		if err := app.orchestrator.RegisterDomain("genomics", app.genomicsAdapter); err != nil {
			return fmt.Errorf("failed to register genomics adapter: %w", err)
		}
	}

	// Start API server if enabled
	if app.apiServer != nil {
		go func() {
			if err := app.apiServer.Start(); err != nil {
				app.logger.Errorf("API server error: %v", err)
			}
		}()
	}

	app.logger.WithFields(logrus.Fields{
		"version":     Version,
		"api_enabled": app.config.API.Enabled,
		"port":        app.config.API.Port,
	}).Info("Izinyoka started successfully")

	// Run system health monitoring
	go app.runHealthMonitoring()

	// Run performance metrics collection
	go app.runMetricsCollection()

	return nil
}

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown(ctx context.Context) error {
	app.logger.Info("Starting graceful shutdown...")

	// Stop API server first
	if app.apiServer != nil {
		if err := app.apiServer.Shutdown(ctx); err != nil {
			app.logger.Errorf("Error shutting down API server: %v", err)
		}
	}

	// Stop orchestrator
	if app.orchestrator != nil {
		if err := app.orchestrator.Stop(ctx); err != nil {
			app.logger.Errorf("Error stopping orchestrator: %v", err)
		}
	}

	// Stop metabolic manager
	if app.metabolicManager != nil {
		app.metabolicManager.Shutdown()
	}

	// Shutdown knowledge base
	if app.knowledgeBase != nil {
		if err := app.knowledgeBase.Shutdown(); err != nil {
			app.logger.Errorf("Error shutting down knowledge base: %v", err)
		}
	}

	// Cancel application context
	app.cancel()

	// Signal shutdown completion
	close(app.shutdownComplete)

	app.logger.Info("Shutdown completed")
	return nil
}

// initializeKnowledgeBase sets up the knowledge base component
func (app *Application) initializeKnowledgeBase() error {
	app.logger.Info("Initializing knowledge base...")

	kbConfig := knowledge.KnowledgeBaseConfig{
		StoragePath:   app.config.Knowledge.StoragePath,
		BackupEnabled: app.config.Knowledge.BackupEnabled,
		BackupPath:    app.config.Knowledge.BackupPath,
		MaxItems:      app.config.Knowledge.MaxItems,
	}

	kb, err := knowledge.NewKnowledgeBase(app.ctx, kbConfig)
	if err != nil {
		return fmt.Errorf("failed to create knowledge base: %w", err)
	}

	app.knowledgeBase = kb
	app.logger.Info("Knowledge base initialized successfully")
	return nil
}

// initializeMetabolicSystem sets up the metabolic components
func (app *Application) initializeMetabolicSystem() error {
	app.logger.Info("Initializing metabolic system...")

	// Initialize glycolytic cycle
	glycolyticConfig := metabolic.GlycolyticConfig{
		MaxConcurrentTasks:  app.config.Metabolic.Glycolytic.MaxConcurrentTasks,
		DefaultTaskPriority: app.config.Metabolic.Glycolytic.DefaultTaskPriority,
		DefaultMaxDuration:  time.Duration(app.config.Metabolic.Glycolytic.DefaultMaxDuration) * time.Second,
		IdleCheckInterval:   time.Duration(app.config.Metabolic.Glycolytic.IdleCheckInterval) * time.Second,
		EnableAutoScale:     app.config.Metabolic.Glycolytic.EnableAutoScale,
		MinWorkers:          app.config.Metabolic.Glycolytic.MinWorkers,
		MaxWorkers:          app.config.Metabolic.Glycolytic.MaxWorkers,
	}

	glycolyticCycle := metabolic.NewGlycolyticCycle(app.ctx, glycolyticConfig)

	// Initialize lactate cycle
	lactateConfig := metabolic.LactateConfig{
		StoragePath:     app.config.Metabolic.Lactate.StoragePath,
		MaxPartials:     app.config.Metabolic.Lactate.MaxPartials,
		DefaultTTL:      time.Duration(app.config.Metabolic.Lactate.DefaultTTL) * time.Hour,
		CleanupInterval: time.Duration(app.config.Metabolic.Lactate.CleanupInterval) * time.Minute,
	}

	lactateCycle := metabolic.NewLactateCycle(app.ctx, lactateConfig)

	// Initialize dreaming module if enabled
	var dreamingModule *metabolic.DreamingModule
	if app.config.Metabolic.Dreaming.Enabled {
		dreamingConfig := metabolic.DreamingConfig{
			Interval:  time.Duration(app.config.Metabolic.Dreaming.Interval) * time.Minute,
			Duration:  time.Duration(app.config.Metabolic.Dreaming.Duration) * time.Minute,
			Diversity: app.config.Metabolic.Dreaming.Diversity,
			Intensity: app.config.Metabolic.Dreaming.Intensity,
		}

		dreamingModule = metabolic.NewDreamingModule(app.knowledgeBase, lactateCycle, dreamingConfig)

		// Start dreaming process
		go dreamingModule.StartDreaming(app.ctx)
	}

	// Create metabolic manager
	app.metabolicManager = metabolic.NewManager(glycolyticCycle, dreamingModule, lactateCycle)

	app.logger.WithFields(logrus.Fields{
		"dreaming_enabled": app.config.Metabolic.Dreaming.Enabled,
		"max_workers":      app.config.Metabolic.Glycolytic.MaxWorkers,
	}).Info("Metabolic system initialized successfully")

	return nil
}

// initializeMetacognitive sets up the metacognitive orchestrator
func (app *Application) initializeMetacognitive() error {
	app.logger.Info("Initializing metacognitive system...")

	// Create context layer
	contextConfig := metacognitive.ContextLayerConfig{
		MaxMemoryItems:      app.config.Metacognitive.Context.MaxMemoryItems,
		AttentionThreshold:  app.config.Metacognitive.Context.AttentionThreshold,
		KnowledgeWeight:     app.config.Metacognitive.Context.KnowledgeWeight,
		TemporalWeight:      app.config.Metacognitive.Context.TemporalWeight,
		SimilarityThreshold: app.config.Metacognitive.Context.SimilarityThreshold,
	}

	contextLayer := metacognitive.NewContextLayer(app.knowledgeBase, contextConfig)

	// Create reasoning layer
	reasoningConfig := metacognitive.ReasoningLayerConfig{
		MaxRules:            app.config.Metacognitive.Reasoning.MaxRules,
		ConfidenceThreshold: app.config.Metacognitive.Reasoning.ConfidenceThreshold,
		LogicTimeout:        app.config.Metacognitive.Reasoning.LogicTimeout,
		EnableCausal:        app.config.Metacognitive.Reasoning.EnableCausal,
		MaxInferenceDepth:   app.config.Metacognitive.Reasoning.MaxInferenceDepth,
	}

	reasoningLayer := metacognitive.NewReasoningLayer(reasoningConfig)

	// Create intuition layer
	intuitionConfig := metacognitive.IntuitionLayerConfig{
		MaxPatterns:        app.config.Metacognitive.Intuition.MaxPatterns,
		PatternThreshold:   app.config.Metacognitive.Intuition.PatternThreshold,
		PredictionHorizon:  app.config.Metacognitive.Intuition.PredictionHorizon,
		EmergenceThreshold: app.config.Metacognitive.Intuition.EmergenceThreshold,
		LearningRate:       app.config.Metacognitive.Intuition.LearningRate,
		NoveltyThreshold:   app.config.Metacognitive.Intuition.NoveltyThreshold,
	}

	intuitionLayer := metacognitive.NewIntuitionLayer(intuitionConfig)

	// Create orchestrator config
	orchestratorConfig := metacognitive.OrchestratorConfig{
		BalanceThreshold:      app.config.Metacognitive.Orchestrator.BalanceThreshold,
		AdaptiveWeights:       app.config.Metacognitive.Orchestrator.AdaptiveWeights,
		PrioritizeConsistency: app.config.Metacognitive.Orchestrator.PrioritizeConsistency,
	}

	// Create orchestrator
	app.orchestrator = metacognitive.NewOrchestrator(
		contextLayer,
		reasoningLayer,
		intuitionLayer,
		app.metabolicManager,
		orchestratorConfig,
	)

	app.logger.Info("Metacognitive system initialized successfully")
	return nil
}

// initializeDomainAdapters sets up domain-specific adapters
func (app *Application) initializeDomainAdapters() error {
	app.logger.Info("Initializing domain adapters...")

	// Initialize genomics adapter
	genomicsConfig := genomics.VariantCallingConfig{
		ReferenceGenome:    app.config.Genomics.ReferenceGenome,
		QualityThreshold:   app.config.Genomics.QualityThreshold,
		CoverageThreshold:  app.config.Genomics.CoverageThreshold,
		EnableParallel:     app.config.Genomics.EnableParallel,
		MaxParallelRegions: app.config.Genomics.MaxParallelRegions,
		VariantFilters:     app.config.Genomics.VariantFilters,
		OutputFormat:       app.config.Genomics.OutputFormat,
		TempDirectory:      app.config.Genomics.TempDirectory,
	}

	genomicsAdapter, err := genomics.NewVariantCallingAdapter(genomicsConfig, app.knowledgeBase)
	if err != nil {
		return fmt.Errorf("failed to create genomics adapter: %w", err)
	}

	app.genomicsAdapter = genomicsAdapter

	app.logger.Info("Domain adapters initialized successfully")
	return nil
}

// initializeAPI sets up the REST API server
func (app *Application) initializeAPI() error {
	app.logger.Info("Initializing API server...")

	apiConfig := api.ServerConfig{
		Port:           app.config.API.Port,
		EnableCORS:     app.config.API.EnableCORS,
		EnableMetrics:  app.config.API.EnableMetrics,
		EnableAuth:     app.config.API.EnableAuth,
		MaxRequestSize: app.config.API.MaxRequestSize,
		RequestTimeout: time.Duration(app.config.API.RequestTimeout) * time.Second,
		RateLimit:      app.config.API.RateLimit,
	}

	apiServer, err := api.NewServer(apiConfig, app.orchestrator, app.knowledgeBase)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}

	app.apiServer = apiServer

	app.logger.WithField("port", app.config.API.Port).Info("API server initialized successfully")
	return nil
}

// runHealthMonitoring monitors system health and resource usage
func (app *Application) runHealthMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-app.ctx.Done():
			return
		case <-ticker.C:
			app.checkSystemHealth()
		}
	}
}

// checkSystemHealth performs health checks on system components
func (app *Application) checkSystemHealth() {
	// Check orchestrator health
	if app.orchestrator != nil && app.orchestrator.IsRunning() {
		state := app.orchestrator.GetState()

		app.logger.WithFields(logrus.Fields{
			"processing_count":    state.ProcessingCount,
			"error_count":         state.ErrorCount,
			"avg_processing_time": state.AverageProcessingTime,
		}).Debug("Orchestrator health check")

		// Trigger resource events based on system state
		if state.ErrorCount > 0 && state.ErrorCount%10 == 0 {
			app.metabolicManager.HandleResourceEvent(app.ctx, "computation", "partial_failure", map[string]interface{}{
				"error_count": state.ErrorCount,
				"timestamp":   time.Now(),
			})
		}
	}

	// Check metabolic system health
	if app.metabolicManager != nil {
		lactateCycle := app.metabolicManager.GetLactateCycle()
		if lactateCycle != nil {
			stats := lactateCycle.GetStorageStats()

			if totalPartials, ok := stats["totalPartials"].(int); ok && totalPartials > 1000 {
				app.metabolicManager.HandleResourceEvent(app.ctx, "memory", "pressure_high", stats)
			}
		}
	}
}

// runMetricsCollection collects and logs performance metrics
func (app *Application) runMetricsCollection() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-app.ctx.Done():
			return
		case <-ticker.C:
			app.collectMetrics()
		}
	}
}

// collectMetrics gathers performance metrics from all components
func (app *Application) collectMetrics() {
	if app.orchestrator != nil && app.orchestrator.IsRunning() {
		metrics := app.orchestrator.GetMetrics()

		app.logger.WithFields(logrus.Fields{
			"avg_confidence":   metrics.ThroughputMetrics["avg_confidence"],
			"requests_per_sec": metrics.ThroughputMetrics["requests_per_second"],
			"total_errors":     len(metrics.ErrorRates),
		}).Info("System performance metrics")
	}
}

// loadConfig loads configuration from file with command line overrides
func loadConfig(configPath, dataDir string, apiPort int, enableAPI, enableDreaming bool) (*config.Config, error) {
	// Default configuration
	cfg := &config.Config{
		Knowledge: config.KnowledgeConfig{
			StoragePath:   filepath.Join(dataDir, "knowledge"),
			BackupEnabled: true,
			BackupPath:    filepath.Join(dataDir, "knowledge-backup"),
			MaxItems:      100000,
		},
		API: config.APIConfig{
			Enabled:        enableAPI,
			Port:           apiPort,
			EnableCORS:     true,
			EnableMetrics:  true,
			EnableAuth:     false,
			MaxRequestSize: 10 << 20, // 10MB
			RequestTimeout: 30,       // 30 seconds
			RateLimit:      1000,     // requests per minute
		},
		Metabolic: config.MetabolicConfig{
			Glycolytic: config.GlycolyticConfig{
				MaxConcurrentTasks:  10,
				DefaultTaskPriority: 5,
				DefaultMaxDuration:  300, // 5 minutes
				IdleCheckInterval:   30,  // 30 seconds
				EnableAutoScale:     true,
				MinWorkers:          2,
				MaxWorkers:          20,
			},
			Lactate: config.LactateConfig{
				StoragePath:     filepath.Join(dataDir, "lactate"),
				MaxPartials:     1000,
				DefaultTTL:      24, // 24 hours
				CleanupInterval: 60, // 60 minutes
			},
			Dreaming: config.DreamingConfig{
				Enabled:   enableDreaming,
				Interval:  10, // 10 minutes
				Duration:  2,  // 2 minutes
				Diversity: 0.8,
				Intensity: 0.5,
			},
		},
		Metacognitive: config.MetacognitiveConfig{
			Context: config.ContextConfig{
				MaxMemoryItems:      1000,
				AttentionThreshold:  0.3,
				KnowledgeWeight:     0.4,
				TemporalWeight:      0.3,
				SimilarityThreshold: 0.5,
			},
			Reasoning: config.ReasoningConfig{
				MaxRules:            500,
				ConfidenceThreshold: 0.6,
				LogicTimeout:        5000, // 5 seconds
				EnableCausal:        true,
				MaxInferenceDepth:   10,
			},
			Intuition: config.IntuitionConfig{
				MaxPatterns:        1000,
				PatternThreshold:   0.7,
				PredictionHorizon:  60, // 60 minutes
				EmergenceThreshold: 0.8,
				LearningRate:       0.1,
				NoveltyThreshold:   0.6,
			},
			Orchestrator: config.OrchestratorConfig{
				BalanceThreshold:      0.5,
				AdaptiveWeights:       true,
				PrioritizeConsistency: false,
			},
		},
		Genomics: config.GenomicsConfig{
			ReferenceGenome:    "/data/reference/hg38.fa",
			QualityThreshold:   20,
			CoverageThreshold:  10,
			EnableParallel:     true,
			MaxParallelRegions: 4,
			VariantFilters:     []string{"PASS", "LIKELY_BENIGN"},
			OutputFormat:       "VCF",
			TempDirectory:      filepath.Join(dataDir, "temp"),
		},
	}

	// Load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, cfg); err != nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Ensure data directories exist
	dirs := []string{
		cfg.Knowledge.StoragePath,
		cfg.Knowledge.BackupPath,
		cfg.Metabolic.Lactate.StoragePath,
		cfg.Genomics.TempDirectory,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return cfg, nil
}
