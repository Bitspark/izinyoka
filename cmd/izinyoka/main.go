package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fullscreen-triangle/izinyoka/api"
	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/fullscreen-triangle/izinyoka/internal/metabolic"
	"github.com/fullscreen-triangle/izinyoka/internal/metacognitive"
	"github.com/fullscreen-triangle/izinyoka/pkg/config"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("Starting Izinyoka - Biomimetic Metacognitive Architecture")

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Info: .env file not found, using environment variables")
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.Info("Initializing Izinyoka")

	// Parse configuration
	cfg, err := config.LoadWithEnv("config/config.yaml", "IZINYOKA")
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("Configuration file not found, using defaults and environment variables")
			cfg, err = config.LoadFromEnv("IZINYOKA")
			if err != nil {
				logger.WithError(err).Fatal("Failed to load configuration from environment")
			}
		} else {
			logger.WithError(err).Fatal("Failed to load configuration")
		}
	}

	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize knowledge base
	kbConfig := knowledge.KnowledgeBaseConfig{
		StoragePath:   cfg.Knowledge.StoragePath,
		BackupEnabled: cfg.Knowledge.BackupEnabled,
		BackupPath:    cfg.Knowledge.BackupPath,
		MaxItems:      cfg.Knowledge.MaxItems,
	}
	knowledgeBase, err := knowledge.NewKnowledgeBase(ctx, kbConfig)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize knowledge base")
	}

	// Initialize glycolytic cycle
	glycolicConfig := metabolic.GlycolyticConfig{
		MaxConcurrentTasks:  cfg.Metabolic.Glycolytic.MaxConcurrentTasks,
		DefaultTaskPriority: cfg.Metabolic.Glycolytic.DefaultTaskPriority,
		DefaultMaxDuration:  cfg.Metabolic.Glycolytic.DefaultMaxDuration,
		IdleCheckInterval:   cfg.Metabolic.Glycolytic.IdleCheckInterval,
		EnableAutoScale:     cfg.Metabolic.Glycolytic.EnableAutoScale,
		MinWorkers:          cfg.Metabolic.Glycolytic.MinWorkers,
		MaxWorkers:          cfg.Metabolic.Glycolytic.MaxWorkers,
	}
	glycolicCycle := metabolic.NewGlycolyticCycle(ctx, glycolicConfig)

	// Initialize dreaming module
	dreamingConfig := metabolic.DreamingConfig{
		EnableAutoDreaming:   cfg.Metabolic.Dreaming.EnableAutoDreaming,
		DreamInterval:        cfg.Metabolic.Dreaming.DreamInterval,
		MaxConcurrentDreams:  cfg.Metabolic.Dreaming.MaxConcurrentDreams,
		MinDreamDuration:     cfg.Metabolic.Dreaming.MinDreamDuration,
		MaxDreamDuration:     cfg.Metabolic.Dreaming.MaxDreamDuration,
		RetainDreamCount:     cfg.Metabolic.Dreaming.RetainDreamCount,
		RareEventProbability: cfg.Metabolic.Dreaming.RareEventProbability,
		AutoApplyInsightProb: cfg.Metabolic.Dreaming.AutoApplyInsightProb,
		RecurrentDreamFactor: cfg.Metabolic.Dreaming.RecurrentDreamFactor,
	}
	dreamingModule := metabolic.NewDreamingModule(ctx, dreamingConfig, knowledgeBase)

	// Initialize lactate cycle
	lactateConfig := metabolic.LactateConfig{
		MaxPartialComputations: cfg.Metabolic.Lactate.MaxPartialComputations,
		DefaultTTL:             cfg.Metabolic.Lactate.DefaultTTL,
		CleanupInterval:        cfg.Metabolic.Lactate.CleanupInterval,
		PrioritizeRecoverable:  cfg.Metabolic.Lactate.PrioritizeRecoverable,
		MaxStorageBytes:        cfg.Metabolic.Lactate.MaxStorageBytes,
	}
	lactateCycle := metabolic.NewLactateCycle(ctx, lactateConfig)

	// Create metabolic manager
	metabolicManager := metabolic.NewManager(glycolicCycle, dreamingModule, lactateCycle)

	// Initialize metacognitive layers
	contextLayer := metacognitive.NewContextLayer()
	reasoningLayer := metacognitive.NewReasoningLayer()
	intuitionLayer := metacognitive.NewIntuitionLayer()

	// Initialize metacognitive orchestrator
	orchestratorConfig := metacognitive.OrchestratorConfig{
		BalanceThreshold:      cfg.Metacognitive.Orchestrator.BalanceThreshold,
		AdaptiveWeights:       cfg.Metacognitive.Orchestrator.AdaptiveWeights,
		PrioritizeConsistency: cfg.Metacognitive.Orchestrator.PrioritizeConsistency,
	}
	orchestrator := metacognitive.NewOrchestrator(
		contextLayer,
		reasoningLayer,
		intuitionLayer,
		metabolicManager,
		orchestratorConfig,
	)

	// Start the orchestrator
	if err := orchestrator.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start orchestrator")
	}

	// Configure and start the API
	router := api.SetupRouter(knowledgeBase, orchestrator, dreamingModule)

	// Configure server
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Start the server in a goroutine
	go func() {
		logger.WithField("addr", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Register sample domain
	registerSampleDomain(orchestrator, logger)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create a deadline for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	// Stop the orchestrator
	if err := orchestrator.Stop(shutdownCtx); err != nil {
		logger.WithError(err).Error("Error stopping orchestrator")
	}

	// Shutdown the knowledge base
	if err := knowledgeBase.Shutdown(); err != nil {
		logger.WithError(err).Error("Error shutting down knowledge base")
	}

	// Shutdown the metabolic manager
	metabolicManager.Shutdown()

	logger.Info("Server exited gracefully")
}

// SampleDomainAdapter is a simple example domain implementation
type SampleDomainAdapter struct{}

func (d *SampleDomainAdapter) ConvertToGeneric(input interface{}) (map[string]interface{}, error) {
	// Simple conversion for example
	switch v := input.(type) {
	case map[string]interface{}:
		return v, nil
	case string:
		return map[string]interface{}{"input": v}, nil
	default:
		return map[string]interface{}{"data": fmt.Sprintf("%v", v)}, nil
	}
}

func (d *SampleDomainAdapter) ConvertFromGeneric(data map[string]interface{}) (interface{}, error) {
	return data, nil
}

func (d *SampleDomainAdapter) QueryResource(resourceType, resourceID string) (interface{}, error) {
	// Simple example implementation
	return map[string]string{
		"type": resourceType,
		"id":   resourceID,
		"info": "This is a sample domain resource",
	}, nil
}

func (d *SampleDomainAdapter) GetSupportedTypes() []string {
	return []string{"process", "analyze", "report"}
}

func (d *SampleDomainAdapter) GetDomainName() string {
	return "sample"
}

func registerSampleDomain(orchestrator *metacognitive.Orchestrator, logger *logrus.Logger) {
	sampleDomain := &SampleDomainAdapter{}
	err := orchestrator.RegisterDomain("sample", sampleDomain)
	if err != nil {
		logger.WithError(err).Error("Failed to register sample domain")
	} else {
		logger.Info("Registered sample domain")
	}
}
