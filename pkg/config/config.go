package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the main system configuration
type Config struct {
	Server        ServerConfig           `yaml:"server" json:"server"`
	Knowledge     KnowledgeConfig        `yaml:"knowledge" json:"knowledge"`
	Metabolic     MetabolicConfig        `yaml:"metabolic" json:"metabolic"`
	Metacognitive MetacognitiveConfig    `yaml:"metacognitive" json:"metacognitive"`
	Logging       LoggingConfig          `yaml:"logging" json:"logging"`
	Domains       map[string]interface{} `yaml:"domains" json:"domains"`
}

// ServerConfig contains the HTTP server configuration
type ServerConfig struct {
	Host           string   `yaml:"host" json:"host"`
	Port           int      `yaml:"port" json:"port"`
	ReadTimeout    int      `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout   int      `yaml:"writeTimeout" json:"writeTimeout"`
	MaxHeaderBytes int      `yaml:"maxHeaderBytes" json:"maxHeaderBytes"`
	TLS            bool     `yaml:"tls" json:"tls"`
	CertFile       string   `yaml:"certFile" json:"certFile"`
	KeyFile        string   `yaml:"keyFile" json:"keyFile"`
	EnableCORS     bool     `yaml:"enableCORS" json:"enableCORS"`
	TrustedProxies []string `yaml:"trustedProxies" json:"trustedProxies"`
}

// KnowledgeConfig contains the knowledge base configuration
type KnowledgeConfig struct {
	StoragePath   string `yaml:"storagePath" json:"storagePath"`
	BackupEnabled bool   `yaml:"backupEnabled" json:"backupEnabled"`
	BackupPath    string `yaml:"backupPath" json:"backupPath"`
	MaxItems      int    `yaml:"maxItems" json:"maxItems"`
}

// MetabolicConfig contains the metabolic components configuration
type MetabolicConfig struct {
	Glycolytic GlycolyticConfig `yaml:"glycolytic" json:"glycolytic"`
	Dreaming   DreamingConfig   `yaml:"dreaming" json:"dreaming"`
	Lactate    LactateConfig    `yaml:"lactate" json:"lactate"`
}

// GlycolyticConfig contains configuration for the glycolytic cycle
type GlycolyticConfig struct {
	MaxConcurrentTasks  int           `yaml:"maxConcurrentTasks" json:"maxConcurrentTasks"`
	DefaultTaskPriority int           `yaml:"defaultTaskPriority" json:"defaultTaskPriority"`
	DefaultMaxDuration  time.Duration `yaml:"defaultMaxDuration" json:"defaultMaxDuration"`
	IdleCheckInterval   time.Duration `yaml:"idleCheckInterval" json:"idleCheckInterval"`
	EnableAutoScale     bool          `yaml:"enableAutoScale" json:"enableAutoScale"`
	MinWorkers          int           `yaml:"minWorkers" json:"minWorkers"`
	MaxWorkers          int           `yaml:"maxWorkers" json:"maxWorkers"`
}

// DreamingConfig contains configuration for the dreaming module
type DreamingConfig struct {
	EnableAutoDreaming   bool          `yaml:"enableAutoDreaming" json:"enableAutoDreaming"`
	DreamInterval        time.Duration `yaml:"dreamInterval" json:"dreamInterval"`
	MaxConcurrentDreams  int           `yaml:"maxConcurrentDreams" json:"maxConcurrentDreams"`
	MinDreamDuration     time.Duration `yaml:"minDreamDuration" json:"minDreamDuration"`
	MaxDreamDuration     time.Duration `yaml:"maxDreamDuration" json:"maxDreamDuration"`
	RetainDreamCount     int           `yaml:"retainDreamCount" json:"retainDreamCount"`
	RareEventProbability float64       `yaml:"rareEventProbability" json:"rareEventProbability"`
	AutoApplyInsightProb float64       `yaml:"autoApplyInsightProbability" json:"autoApplyInsightProbability"`
	RecurrentDreamFactor float64       `yaml:"recurrentDreamFactor" json:"recurrentDreamFactor"`
}

// LactateConfig contains configuration for the lactate cycle
type LactateConfig struct {
	MaxPartialComputations int           `yaml:"maxPartialComputations" json:"maxPartialComputations"`
	DefaultTTL             time.Duration `yaml:"defaultTTL" json:"defaultTTL"`
	CleanupInterval        time.Duration `yaml:"cleanupInterval" json:"cleanupInterval"`
	PrioritizeRecoverable  bool          `yaml:"prioritizeRecoverable" json:"prioritizeRecoverable"`
	MaxStorageBytes        int64         `yaml:"maxStorageBytes" json:"maxStorageBytes"`
}

// MetacognitiveConfig contains the metacognitive components configuration
type MetacognitiveConfig struct {
	Context      ContextConfig      `yaml:"context" json:"context"`
	Reasoning    ReasoningConfig    `yaml:"reasoning" json:"reasoning"`
	Intuition    IntuitionConfig    `yaml:"intuition" json:"intuition"`
	Orchestrator OrchestratorConfig `yaml:"orchestrator" json:"orchestrator"`
}

// ContextConfig contains configuration for the context layer
type ContextConfig struct {
	ContextRetentionSize int           `yaml:"contextRetentionSize" json:"contextRetentionSize"`
	ContextExpiration    time.Duration `yaml:"contextExpiration" json:"contextExpiration"`
	KnowledgeThreshold   float64       `yaml:"knowledgeThreshold" json:"knowledgeThreshold"`
	MaxContexts          int           `yaml:"maxContexts" json:"maxContexts"`
}

// ReasoningConfig contains configuration for the reasoning layer
type ReasoningConfig struct {
	MaxReasoningDepth   int     `yaml:"maxReasoningDepth" json:"maxReasoningDepth"`
	ConfidenceThreshold float64 `yaml:"confidenceThreshold" json:"confidenceThreshold"`
	TimeoutMultiplier   float64 `yaml:"timeoutMultiplier" json:"timeoutMultiplier"`
}

// IntuitionConfig contains configuration for the intuition layer
type IntuitionConfig struct {
	PatternMatchThreshold float64 `yaml:"patternMatchThreshold" json:"patternMatchThreshold"`
	MaxPatternComplexity  int     `yaml:"maxPatternComplexity" json:"maxPatternComplexity"`
	EnableFastPath        bool    `yaml:"enableFastPath" json:"enableFastPath"`
}

// OrchestratorConfig contains configuration for the metacognitive orchestrator
type OrchestratorConfig struct {
	BalanceThreshold      float64 `yaml:"balanceThreshold" json:"balanceThreshold"`
	AdaptiveWeights       bool    `yaml:"adaptiveWeights" json:"adaptiveWeights"`
	PrioritizeConsistency bool    `yaml:"prioritizeConsistency" json:"prioritizeConsistency"`
}

// LoggingConfig contains the logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	Output     string `yaml:"output" json:"output"`
	EnableFile bool   `yaml:"enableFile" json:"enableFile"`
	FilePath   string `yaml:"filePath" json:"filePath"`
	MaxSize    int    `yaml:"maxSize" json:"maxSize"`
	MaxBackups int    `yaml:"maxBackups" json:"maxBackups"`
	MaxAge     int    `yaml:"maxAge" json:"maxAge"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	// Read file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Set defaults if needed
	applyDefaults(&config)

	return &config, nil
}

// LoadWithEnv loads configuration from a YAML file and overrides with environment variables
func LoadWithEnv(path string, envPrefix string) (*Config, error) {
	// First load from file
	config, err := Load(path)
	if err != nil {
		return nil, err
	}

	// Then override with environment variables
	applyEnvironmentOverrides(config, envPrefix)

	return config, nil
}

// LoadFromEnv loads configuration from environment variables only
func LoadFromEnv(envPrefix string) (*Config, error) {
	config := &Config{}
	applyDefaults(config)
	applyEnvironmentOverrides(config, envPrefix)
	return config, nil
}

// LoadEnvFile loads environment variables from a .env file
func LoadEnvFile(path string) error {
	return godotenv.Load(path)
}

// Save saves the configuration to a YAML file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// applyDefaults sets default values for configuration
func applyDefaults(config *Config) {
	// Server defaults
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30
	}
	if config.Server.MaxHeaderBytes == 0 {
		config.Server.MaxHeaderBytes = 1 << 20 // 1 MB
	}

	// Knowledge defaults
	if config.Knowledge.StoragePath == "" {
		config.Knowledge.StoragePath = "./data/knowledge"
	}
	if config.Knowledge.MaxItems == 0 {
		config.Knowledge.MaxItems = 10000
	}
	if config.Knowledge.BackupEnabled && config.Knowledge.BackupPath == "" {
		config.Knowledge.BackupPath = "./data/backup"
	}

	// Metabolic defaults - glycolytic cycle
	if config.Metabolic.Glycolytic.MaxConcurrentTasks == 0 {
		config.Metabolic.Glycolytic.MaxConcurrentTasks = 10
	}
	if config.Metabolic.Glycolytic.DefaultTaskPriority == 0 {
		config.Metabolic.Glycolytic.DefaultTaskPriority = 5
	}
	if config.Metabolic.Glycolytic.DefaultMaxDuration == 0 {
		config.Metabolic.Glycolytic.DefaultMaxDuration = 5 * time.Minute
	}
	if config.Metabolic.Glycolytic.IdleCheckInterval == 0 {
		config.Metabolic.Glycolytic.IdleCheckInterval = 30 * time.Second
	}
	if config.Metabolic.Glycolytic.MinWorkers == 0 {
		config.Metabolic.Glycolytic.MinWorkers = 2
	}
	if config.Metabolic.Glycolytic.MaxWorkers == 0 {
		config.Metabolic.Glycolytic.MaxWorkers = 20
	}

	// Metabolic defaults - dreaming module
	if config.Metabolic.Dreaming.DreamInterval == 0 {
		config.Metabolic.Dreaming.DreamInterval = 30 * time.Minute
	}
	if config.Metabolic.Dreaming.MaxConcurrentDreams == 0 {
		config.Metabolic.Dreaming.MaxConcurrentDreams = 3
	}
	if config.Metabolic.Dreaming.MinDreamDuration == 0 {
		config.Metabolic.Dreaming.MinDreamDuration = 10 * time.Second
	}
	if config.Metabolic.Dreaming.MaxDreamDuration == 0 {
		config.Metabolic.Dreaming.MaxDreamDuration = 2 * time.Minute
	}
	if config.Metabolic.Dreaming.RetainDreamCount == 0 {
		config.Metabolic.Dreaming.RetainDreamCount = 100
	}
	if config.Metabolic.Dreaming.RareEventProbability == 0 {
		config.Metabolic.Dreaming.RareEventProbability = 0.1
	}
	if config.Metabolic.Dreaming.AutoApplyInsightProb == 0 {
		config.Metabolic.Dreaming.AutoApplyInsightProb = 0.3
	}
	if config.Metabolic.Dreaming.RecurrentDreamFactor == 0 {
		config.Metabolic.Dreaming.RecurrentDreamFactor = 0.2
	}

	// Metabolic defaults - lactate cycle
	if config.Metabolic.Lactate.MaxPartialComputations == 0 {
		config.Metabolic.Lactate.MaxPartialComputations = 1000
	}
	if config.Metabolic.Lactate.DefaultTTL == 0 {
		config.Metabolic.Lactate.DefaultTTL = 24 * time.Hour
	}
	if config.Metabolic.Lactate.CleanupInterval == 0 {
		config.Metabolic.Lactate.CleanupInterval = 10 * time.Minute
	}
	if config.Metabolic.Lactate.MaxStorageBytes == 0 {
		config.Metabolic.Lactate.MaxStorageBytes = 100 * 1024 * 1024 // 100 MB
	}

	// Metacognitive defaults - context layer
	if config.Metacognitive.Context.ContextRetentionSize == 0 {
		config.Metacognitive.Context.ContextRetentionSize = 100
	}
	if config.Metacognitive.Context.ContextExpiration == 0 {
		config.Metacognitive.Context.ContextExpiration = 24 * time.Hour
	}
	if config.Metacognitive.Context.KnowledgeThreshold == 0 {
		config.Metacognitive.Context.KnowledgeThreshold = 0.7
	}
	if config.Metacognitive.Context.MaxContexts == 0 {
		config.Metacognitive.Context.MaxContexts = 10
	}

	// Metacognitive defaults - reasoning layer
	if config.Metacognitive.Reasoning.MaxReasoningDepth == 0 {
		config.Metacognitive.Reasoning.MaxReasoningDepth = 5
	}
	if config.Metacognitive.Reasoning.ConfidenceThreshold == 0 {
		config.Metacognitive.Reasoning.ConfidenceThreshold = 0.8
	}
	if config.Metacognitive.Reasoning.TimeoutMultiplier == 0 {
		config.Metacognitive.Reasoning.TimeoutMultiplier = 1.5
	}

	// Metacognitive defaults - intuition layer
	if config.Metacognitive.Intuition.PatternMatchThreshold == 0 {
		config.Metacognitive.Intuition.PatternMatchThreshold = 0.6
	}
	if config.Metacognitive.Intuition.MaxPatternComplexity == 0 {
		config.Metacognitive.Intuition.MaxPatternComplexity = 10
	}

	// Metacognitive defaults - orchestrator
	if config.Metacognitive.Orchestrator.BalanceThreshold == 0 {
		config.Metacognitive.Orchestrator.BalanceThreshold = 0.1
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}
	if config.Logging.MaxSize == 0 {
		config.Logging.MaxSize = 10 // 10 MB
	}
	if config.Logging.MaxBackups == 0 {
		config.Logging.MaxBackups = 3
	}
	if config.Logging.MaxAge == 0 {
		config.Logging.MaxAge = 28 // 28 days
	}

	// Initialize empty maps
	if config.Domains == nil {
		config.Domains = make(map[string]interface{})
	}
}

// applyEnvironmentOverrides overrides configuration with environment variables
func applyEnvironmentOverrides(config *Config, prefix string) {
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix = prefix + "_"
	}

	// Server overrides
	if host := os.Getenv(prefix + "SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if portStr := os.Getenv(prefix + "SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Server.Port = port
		}
	}

	// Knowledge overrides
	if path := os.Getenv(prefix + "KNOWLEDGE_PATH"); path != "" {
		config.Knowledge.StoragePath = path
	}
	if backupStr := os.Getenv(prefix + "KNOWLEDGE_BACKUP_ENABLED"); backupStr != "" {
		config.Knowledge.BackupEnabled = backupStr == "true" || backupStr == "1" || backupStr == "yes"
	}
	if path := os.Getenv(prefix + "KNOWLEDGE_BACKUP_PATH"); path != "" {
		config.Knowledge.BackupPath = path
	}

	// Metabolic - glycolytic overrides
	if tasksStr := os.Getenv(prefix + "GLYCOLYTIC_MAX_TASKS"); tasksStr != "" {
		if tasks, err := strconv.Atoi(tasksStr); err == nil {
			config.Metabolic.Glycolytic.MaxConcurrentTasks = tasks
		}
	}
	if priorityStr := os.Getenv(prefix + "GLYCOLYTIC_DEFAULT_PRIORITY"); priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil {
			config.Metabolic.Glycolytic.DefaultTaskPriority = priority
		}
	}
	if durationStr := os.Getenv(prefix + "GLYCOLYTIC_DEFAULT_DURATION"); durationStr != "" {
		if duration, err := time.ParseDuration(durationStr); err == nil {
			config.Metabolic.Glycolytic.DefaultMaxDuration = duration
		}
	}

	// Metabolic - dreaming overrides
	if autoStr := os.Getenv(prefix + "DREAMING_AUTO_ENABLED"); autoStr != "" {
		config.Metabolic.Dreaming.EnableAutoDreaming = autoStr == "true" || autoStr == "1" || autoStr == "yes"
	}
	if intervalStr := os.Getenv(prefix + "DREAMING_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			config.Metabolic.Dreaming.DreamInterval = interval
		}
	}
	if concurrentStr := os.Getenv(prefix + "DREAMING_MAX_CONCURRENT"); concurrentStr != "" {
		if concurrent, err := strconv.Atoi(concurrentStr); err == nil {
			config.Metabolic.Dreaming.MaxConcurrentDreams = concurrent
		}
	}

	// Metabolic - lactate overrides
	if maxStr := os.Getenv(prefix + "LACTATE_MAX_PARTIALS"); maxStr != "" {
		if max, err := strconv.Atoi(maxStr); err == nil {
			config.Metabolic.Lactate.MaxPartialComputations = max
		}
	}
	if ttlStr := os.Getenv(prefix + "LACTATE_DEFAULT_TTL"); ttlStr != "" {
		if ttl, err := time.ParseDuration(ttlStr); err == nil {
			config.Metabolic.Lactate.DefaultTTL = ttl
		}
	}
	if sizeStr := os.Getenv(prefix + "LACTATE_MAX_STORAGE"); sizeStr != "" {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			config.Metabolic.Lactate.MaxStorageBytes = size
		}
	}

	// Logging overrides
	if level := os.Getenv(prefix + "LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
	if format := os.Getenv(prefix + "LOG_FORMAT"); format != "" {
		config.Logging.Format = format
	}
	if output := os.Getenv(prefix + "LOG_OUTPUT"); output != "" {
		config.Logging.Output = output
	}
	if path := os.Getenv(prefix + "LOG_FILE_PATH"); path != "" {
		config.Logging.FilePath = path
		config.Logging.EnableFile = true
	}
}
