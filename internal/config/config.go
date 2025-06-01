package config

import (
	"fmt"
	"time"
)

// Config represents the complete Izinyoka application configuration
type Config struct {
	Knowledge     KnowledgeConfig     `yaml:"knowledge" json:"knowledge"`
	API           APIConfig           `yaml:"api" json:"api"`
	Metabolic     MetabolicConfig     `yaml:"metabolic" json:"metabolic"`
	Metacognitive MetacognitiveConfig `yaml:"metacognitive" json:"metacognitive"`
	Genomics      GenomicsConfig      `yaml:"genomics" json:"genomics"`
	Logging       LoggingConfig       `yaml:"logging" json:"logging"`
}

// KnowledgeConfig configures the knowledge base component
type KnowledgeConfig struct {
	StoragePath   string `yaml:"storage_path" json:"storage_path"`
	BackupEnabled bool   `yaml:"backup_enabled" json:"backup_enabled"`
	BackupPath    string `yaml:"backup_path" json:"backup_path"`
	MaxItems      int    `yaml:"max_items" json:"max_items"`
}

// APIConfig configures the REST API server
type APIConfig struct {
	Enabled        bool       `yaml:"enabled" json:"enabled"`
	Port           int        `yaml:"port" json:"port"`
	EnableCORS     bool       `yaml:"enable_cors" json:"enable_cors"`
	EnableMetrics  bool       `yaml:"enable_metrics" json:"enable_metrics"`
	EnableAuth     bool       `yaml:"enable_auth" json:"enable_auth"`
	MaxRequestSize int64      `yaml:"max_request_size" json:"max_request_size"`
	RequestTimeout int        `yaml:"request_timeout" json:"request_timeout"`
	RateLimit      int        `yaml:"rate_limit" json:"rate_limit"`
	AuthConfig     AuthConfig `yaml:"auth_config" json:"auth_config"`
}

// AuthConfig configures authentication settings
type AuthConfig struct {
	JWTSecret    string        `yaml:"jwt_secret" json:"jwt_secret"`
	TokenExpiry  time.Duration `yaml:"token_expiry" json:"token_expiry"`
	EnableBearer bool          `yaml:"enable_bearer" json:"enable_bearer"`
	EnableAPIKey bool          `yaml:"enable_api_key" json:"enable_api_key"`
}

// MetabolicConfig configures all metabolic system components
type MetabolicConfig struct {
	Glycolytic GlycolyticConfig `yaml:"glycolytic" json:"glycolytic"`
	Lactate    LactateConfig    `yaml:"lactate" json:"lactate"`
	Dreaming   DreamingConfig   `yaml:"dreaming" json:"dreaming"`
}

// GlycolyticConfig configures the glycolytic cycle component
type GlycolyticConfig struct {
	MaxConcurrentTasks  int  `yaml:"max_concurrent_tasks" json:"max_concurrent_tasks"`
	DefaultTaskPriority int  `yaml:"default_task_priority" json:"default_task_priority"`
	DefaultMaxDuration  int  `yaml:"default_max_duration" json:"default_max_duration"` // seconds
	IdleCheckInterval   int  `yaml:"idle_check_interval" json:"idle_check_interval"`   // seconds
	EnableAutoScale     bool `yaml:"enable_auto_scale" json:"enable_auto_scale"`
	MinWorkers          int  `yaml:"min_workers" json:"min_workers"`
	MaxWorkers          int  `yaml:"max_workers" json:"max_workers"`
}

// LactateConfig configures the lactate cycle component
type LactateConfig struct {
	StoragePath     string `yaml:"storage_path" json:"storage_path"`
	MaxPartials     int    `yaml:"max_partials" json:"max_partials"`
	DefaultTTL      int    `yaml:"default_ttl" json:"default_ttl"`           // hours
	CleanupInterval int    `yaml:"cleanup_interval" json:"cleanup_interval"` // minutes
}

// DreamingConfig configures the dreaming module
type DreamingConfig struct {
	Enabled   bool    `yaml:"enabled" json:"enabled"`
	Interval  int     `yaml:"interval" json:"interval"`   // minutes
	Duration  int     `yaml:"duration" json:"duration"`   // minutes
	Diversity float64 `yaml:"diversity" json:"diversity"` // 0.0-1.0
	Intensity float64 `yaml:"intensity" json:"intensity"` // 0.0-1.0
}

// MetacognitiveConfig configures the metacognitive architecture
type MetacognitiveConfig struct {
	Context      ContextConfig      `yaml:"context" json:"context"`
	Reasoning    ReasoningConfig    `yaml:"reasoning" json:"reasoning"`
	Intuition    IntuitionConfig    `yaml:"intuition" json:"intuition"`
	Orchestrator OrchestratorConfig `yaml:"orchestrator" json:"orchestrator"`
}

// ContextConfig configures the context layer
type ContextConfig struct {
	MaxMemoryItems      int     `yaml:"max_memory_items" json:"max_memory_items"`
	AttentionThreshold  float64 `yaml:"attention_threshold" json:"attention_threshold"`
	KnowledgeWeight     float64 `yaml:"knowledge_weight" json:"knowledge_weight"`
	TemporalWeight      float64 `yaml:"temporal_weight" json:"temporal_weight"`
	SimilarityThreshold float64 `yaml:"similarity_threshold" json:"similarity_threshold"`
}

// ReasoningConfig configures the reasoning layer
type ReasoningConfig struct {
	MaxRules            int     `yaml:"max_rules" json:"max_rules"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold" json:"confidence_threshold"`
	LogicTimeout        int     `yaml:"logic_timeout" json:"logic_timeout"` // milliseconds
	EnableCausal        bool    `yaml:"enable_causal" json:"enable_causal"`
	MaxInferenceDepth   int     `yaml:"max_inference_depth" json:"max_inference_depth"`
}

// IntuitionConfig configures the intuition layer
type IntuitionConfig struct {
	MaxPatterns        int     `yaml:"max_patterns" json:"max_patterns"`
	PatternThreshold   float64 `yaml:"pattern_threshold" json:"pattern_threshold"`
	PredictionHorizon  int     `yaml:"prediction_horizon" json:"prediction_horizon"` // minutes
	EmergenceThreshold float64 `yaml:"emergence_threshold" json:"emergence_threshold"`
	LearningRate       float64 `yaml:"learning_rate" json:"learning_rate"`
	NoveltyThreshold   float64 `yaml:"novelty_threshold" json:"novelty_threshold"`
}

// OrchestratorConfig configures the metacognitive orchestrator
type OrchestratorConfig struct {
	BalanceThreshold      float64 `yaml:"balance_threshold" json:"balance_threshold"`
	AdaptiveWeights       bool    `yaml:"adaptive_weights" json:"adaptive_weights"`
	PrioritizeConsistency bool    `yaml:"prioritize_consistency" json:"prioritize_consistency"`
}

// GenomicsConfig configures genomic variant calling
type GenomicsConfig struct {
	ReferenceGenome    string   `yaml:"reference_genome" json:"reference_genome"`
	QualityThreshold   int      `yaml:"quality_threshold" json:"quality_threshold"`
	CoverageThreshold  int      `yaml:"coverage_threshold" json:"coverage_threshold"`
	EnableParallel     bool     `yaml:"enable_parallel" json:"enable_parallel"`
	MaxParallelRegions int      `yaml:"max_parallel_regions" json:"max_parallel_regions"`
	VariantFilters     []string `yaml:"variant_filters" json:"variant_filters"`
	OutputFormat       string   `yaml:"output_format" json:"output_format"`
	TempDirectory      string   `yaml:"temp_directory" json:"temp_directory"`
}

// LoggingConfig configures logging behavior
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	OutputFile string `yaml:"output_file" json:"output_file"`
	EnableJSON bool   `yaml:"enable_json" json:"enable_json"`
}

// Validate validates the entire configuration
func (c *Config) Validate() error {
	if err := c.Knowledge.Validate(); err != nil {
		return fmt.Errorf("knowledge config: %w", err)
	}

	if err := c.API.Validate(); err != nil {
		return fmt.Errorf("API config: %w", err)
	}

	if err := c.Metabolic.Validate(); err != nil {
		return fmt.Errorf("metabolic config: %w", err)
	}

	if err := c.Metacognitive.Validate(); err != nil {
		return fmt.Errorf("metacognitive config: %w", err)
	}

	if err := c.Genomics.Validate(); err != nil {
		return fmt.Errorf("genomics config: %w", err)
	}

	return nil
}

// Validate validates knowledge configuration
func (k *KnowledgeConfig) Validate() error {
	if k.StoragePath == "" {
		return fmt.Errorf("storage_path is required")
	}

	if k.MaxItems <= 0 {
		return fmt.Errorf("max_items must be positive")
	}

	if k.BackupEnabled && k.BackupPath == "" {
		return fmt.Errorf("backup_path is required when backup is enabled")
	}

	return nil
}

// Validate validates API configuration
func (a *APIConfig) Validate() error {
	if a.Enabled {
		if a.Port <= 0 || a.Port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}

		if a.MaxRequestSize <= 0 {
			return fmt.Errorf("max_request_size must be positive")
		}

		if a.RequestTimeout <= 0 {
			return fmt.Errorf("request_timeout must be positive")
		}

		if a.RateLimit <= 0 {
			return fmt.Errorf("rate_limit must be positive")
		}
	}

	return nil
}

// Validate validates metabolic configuration
func (m *MetabolicConfig) Validate() error {
	if err := m.Glycolytic.Validate(); err != nil {
		return fmt.Errorf("glycolytic: %w", err)
	}

	if err := m.Lactate.Validate(); err != nil {
		return fmt.Errorf("lactate: %w", err)
	}

	if err := m.Dreaming.Validate(); err != nil {
		return fmt.Errorf("dreaming: %w", err)
	}

	return nil
}

// Validate validates glycolytic configuration
func (g *GlycolyticConfig) Validate() error {
	if g.MaxConcurrentTasks <= 0 {
		return fmt.Errorf("max_concurrent_tasks must be positive")
	}

	if g.DefaultTaskPriority < 1 || g.DefaultTaskPriority > 10 {
		return fmt.Errorf("default_task_priority must be between 1 and 10")
	}

	if g.DefaultMaxDuration <= 0 {
		return fmt.Errorf("default_max_duration must be positive")
	}

	if g.IdleCheckInterval <= 0 {
		return fmt.Errorf("idle_check_interval must be positive")
	}

	if g.EnableAutoScale {
		if g.MinWorkers <= 0 {
			return fmt.Errorf("min_workers must be positive when auto-scaling is enabled")
		}

		if g.MaxWorkers <= g.MinWorkers {
			return fmt.Errorf("max_workers must be greater than min_workers")
		}
	}

	return nil
}

// Validate validates lactate configuration
func (l *LactateConfig) Validate() error {
	if l.StoragePath == "" {
		return fmt.Errorf("storage_path is required")
	}

	if l.MaxPartials <= 0 {
		return fmt.Errorf("max_partials must be positive")
	}

	if l.DefaultTTL <= 0 {
		return fmt.Errorf("default_ttl must be positive")
	}

	if l.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup_interval must be positive")
	}

	return nil
}

// Validate validates dreaming configuration
func (d *DreamingConfig) Validate() error {
	if d.Enabled {
		if d.Interval <= 0 {
			return fmt.Errorf("interval must be positive when dreaming is enabled")
		}

		if d.Duration <= 0 {
			return fmt.Errorf("duration must be positive when dreaming is enabled")
		}

		if d.Diversity < 0.0 || d.Diversity > 1.0 {
			return fmt.Errorf("diversity must be between 0.0 and 1.0")
		}

		if d.Intensity < 0.0 || d.Intensity > 1.0 {
			return fmt.Errorf("intensity must be between 0.0 and 1.0")
		}
	}

	return nil
}

// Validate validates metacognitive configuration
func (m *MetacognitiveConfig) Validate() error {
	if err := m.Context.Validate(); err != nil {
		return fmt.Errorf("context: %w", err)
	}

	if err := m.Reasoning.Validate(); err != nil {
		return fmt.Errorf("reasoning: %w", err)
	}

	if err := m.Intuition.Validate(); err != nil {
		return fmt.Errorf("intuition: %w", err)
	}

	if err := m.Orchestrator.Validate(); err != nil {
		return fmt.Errorf("orchestrator: %w", err)
	}

	return nil
}

// Validate validates context configuration
func (c *ContextConfig) Validate() error {
	if c.MaxMemoryItems <= 0 {
		return fmt.Errorf("max_memory_items must be positive")
	}

	if c.AttentionThreshold < 0.0 || c.AttentionThreshold > 1.0 {
		return fmt.Errorf("attention_threshold must be between 0.0 and 1.0")
	}

	if c.KnowledgeWeight < 0.0 || c.KnowledgeWeight > 1.0 {
		return fmt.Errorf("knowledge_weight must be between 0.0 and 1.0")
	}

	if c.TemporalWeight < 0.0 || c.TemporalWeight > 1.0 {
		return fmt.Errorf("temporal_weight must be between 0.0 and 1.0")
	}

	if c.SimilarityThreshold < 0.0 || c.SimilarityThreshold > 1.0 {
		return fmt.Errorf("similarity_threshold must be between 0.0 and 1.0")
	}

	return nil
}

// Validate validates reasoning configuration
func (r *ReasoningConfig) Validate() error {
	if r.MaxRules <= 0 {
		return fmt.Errorf("max_rules must be positive")
	}

	if r.ConfidenceThreshold < 0.0 || r.ConfidenceThreshold > 1.0 {
		return fmt.Errorf("confidence_threshold must be between 0.0 and 1.0")
	}

	if r.LogicTimeout <= 0 {
		return fmt.Errorf("logic_timeout must be positive")
	}

	if r.MaxInferenceDepth <= 0 {
		return fmt.Errorf("max_inference_depth must be positive")
	}

	return nil
}

// Validate validates intuition configuration
func (i *IntuitionConfig) Validate() error {
	if i.MaxPatterns <= 0 {
		return fmt.Errorf("max_patterns must be positive")
	}

	if i.PatternThreshold < 0.0 || i.PatternThreshold > 1.0 {
		return fmt.Errorf("pattern_threshold must be between 0.0 and 1.0")
	}

	if i.PredictionHorizon <= 0 {
		return fmt.Errorf("prediction_horizon must be positive")
	}

	if i.EmergenceThreshold < 0.0 || i.EmergenceThreshold > 1.0 {
		return fmt.Errorf("emergence_threshold must be between 0.0 and 1.0")
	}

	if i.LearningRate <= 0.0 || i.LearningRate > 1.0 {
		return fmt.Errorf("learning_rate must be between 0.0 and 1.0")
	}

	if i.NoveltyThreshold < 0.0 || i.NoveltyThreshold > 1.0 {
		return fmt.Errorf("novelty_threshold must be between 0.0 and 1.0")
	}

	return nil
}

// Validate validates orchestrator configuration
func (o *OrchestratorConfig) Validate() error {
	if o.BalanceThreshold < 0.0 || o.BalanceThreshold > 1.0 {
		return fmt.Errorf("balance_threshold must be between 0.0 and 1.0")
	}

	return nil
}

// Validate validates genomics configuration
func (g *GenomicsConfig) Validate() error {
	if g.ReferenceGenome == "" {
		return fmt.Errorf("reference_genome is required")
	}

	if g.QualityThreshold < 0 {
		return fmt.Errorf("quality_threshold must be non-negative")
	}

	if g.CoverageThreshold < 0 {
		return fmt.Errorf("coverage_threshold must be non-negative")
	}

	if g.EnableParallel && g.MaxParallelRegions <= 0 {
		return fmt.Errorf("max_parallel_regions must be positive when parallel processing is enabled")
	}

	if g.OutputFormat == "" {
		return fmt.Errorf("output_format is required")
	}

	if g.TempDirectory == "" {
		return fmt.Errorf("temp_directory is required")
	}

	return nil
}
