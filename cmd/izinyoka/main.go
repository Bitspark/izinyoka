package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fullscreen-triangle/izinyoka/internal/domain"
	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/fullscreen-triangle/izinyoka/internal/metabolic"
	"github.com/fullscreen-triangle/izinyoka/internal/metacognitive"
	"github.com/fullscreen-triangle/izinyoka/pkg/config"
)

func main() {
	configPath := flag.String("config", "config/system.yaml", "Path to configuration file")
	domainName := flag.String("domain", "genomic", "Domain to use")
	initKnowledge := flag.Bool("init-knowledge", false, "Initialize knowledge base")
	debugMode := flag.Bool("dev", false, "Run in development mode")
	port := flag.Int("port", 8080, "Port for HTTP server")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadFromFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up knowledge base
	kb, err := knowledge.NewKnowledgeBase(cfg.KnowledgePath)
	if err != nil {
		log.Fatalf("Failed to initialize knowledge base: %v", err)
	}

	if *initKnowledge {
		log.Println("Initializing knowledge base...")
		if err := kb.Initialize(); err != nil {
			log.Fatalf("Failed to initialize knowledge base: %v", err)
		}
		log.Println("Knowledge base initialized successfully")
		return
	}

	// Set up domain adapter
	domainAdapter, err := domain.GetAdapter(*domainName, cfg.DomainConfig)
	if err != nil {
		log.Fatalf("Failed to initialize domain adapter: %v", err)
	}

	// Set up metabolic components
	glycolicCycle := metabolic.NewGlycolicCycle(kb, cfg.Metabolic.Glycolytic)
	lactateCycle := metabolic.NewLactateCycle(kb, cfg.Metabolic.Lactate)
	dreamingModule := metabolic.NewDreamingModule(kb, lactateCycle, cfg.Metabolic.Dreaming)

	// Set up metacognitive orchestrator
	orchestrator := metacognitive.NewOrchestrator(
		kb,
		domainAdapter.GetContextProcessor(),
		domainAdapter.GetReasoningProcessor(),
		domainAdapter.GetIntuitionProcessor(),
		glycolicCycle,
		dreamingModule,
		lactateCycle,
	)

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start dreaming in background
	go dreamingModule.StartDreaming(ctx)

	// Handle graceful shutdown
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	// Start processing
	log.Printf("Izinyoka started with %s domain on port %d\n", *domainName, *port)

	<-shutdownCh
	log.Println("Shutting down gracefully...")

	// Give some time for ongoing processes to complete
	cancelCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	orchestrator.Shutdown(cancelCtx)
	log.Println("Shutdown complete")
}
