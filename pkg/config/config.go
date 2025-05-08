package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config holds all system configuration
type Config struct {
	// General settings
	LogLevel      string `yaml:"log_level"`
	KnowledgePath string `yaml:"knowledge_path"`
	ServerPort    int    `yaml:"server_port"`

	// Metacognitive settings
	Metacognitive MetacognitiveConfig `yaml:"metacognitive"`

	// Metabolic settings
	Metabolic MetabolicConfig `yaml:"metabolic"`

	// Domain-specific configuration
	DomainConfig map[string]interface{} `yaml:"domain"`
}

// MetacognitiveConfig holds settings for the metacognitive orchestrator
type MetacognitiveConfig struct {
	ContextLayer struct {
		Threshold  float64 `yaml:"threshold"`
		BufferSize int     `yaml:"buffer_size"`
	} `yaml:"context_layer"`

	ReasoningLayer struct {
		MaxDepth int           `yaml:"max_depth"`
		Timeout  time.Duration `yaml:"timeout"`
	} `yaml:"reasoning_layer"`

	IntuitionLayer struct {
		HeuristicWeight  float64 `yaml:"heuristic_weight"`
		PatternThreshold float64 `yaml:"pattern_threshold"`
	} `yaml:"intuition_layer"`
}

// MetabolicConfig holds settings for metabolic components
type MetabolicConfig struct {
	Glycolytic GlycolicConfig `yaml:"glycolytic"`
	Dreaming   DreamingConfig `yaml:"dreaming"`
	Lactate    LactateConfig  `yaml:"lactate"`
}

// GlycolicConfig defines glycolytic cycle parameters
type GlycolicConfig struct {
	TaskBatchSize  int `yaml:"task_batch_size"`
	ResourceLimit  int `yaml:"resource_limit"`
	TimeoutMs      int `yaml:"timeout_ms"`
	PriorityLevels int `yaml:"priority_levels"`
}

// DreamingConfig defines dreaming module parameters
type DreamingConfig struct {
	Interval  time.Duration `yaml:"interval"`
	Duration  time.Duration `yaml:"duration"`
	Diversity float64       `yaml:"diversity"`
	Intensity float64       `yaml:"intensity"`
}

// LactateConfig defines lactate cycle parameters
type LactateConfig struct {
	MaxIncompleteTasks int     `yaml:"max_incomplete_tasks"`
	RecycleThreshold   float64 `yaml:"recycle_threshold"`
	PersistPath        string  `yaml:"persist_path"`
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	// Load .env file if present
	godotenv.Load()

	// Load configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
