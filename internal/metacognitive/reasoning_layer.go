package metacognitive

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// ReasoningLayer implements logical deduction and rule-based processing
type ReasoningLayer struct {
	ruleEngine      *RuleEngine
	causalInference *CausalInferenceEngine
	logicSolver     *LogicSolver
	rules           map[string]Rule
	mu              sync.RWMutex
	config          ReasoningLayerConfig
}

// ReasoningLayerConfig configures the reasoning layer behavior
type ReasoningLayerConfig struct {
	MaxRules            int     `yaml:"max_rules"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	LogicTimeout        int     `yaml:"logic_timeout_ms"`
	EnableCausal        bool    `yaml:"enable_causal"`
	MaxInferenceDepth   int     `yaml:"max_inference_depth"`
}

// ReasoningResult represents the output from reasoning layer processing
type ReasoningResult struct {
	Conclusions    []Conclusion
	CausalChain    []CausalLink
	AppliedRules   []Rule
	Confidence     float64
	DerivationPath []string
	Timestamp      time.Time
}

// Rule represents a logical rule in the reasoning system
type Rule struct {
	ID          string
	Name        string
	Conditions  []Condition
	Conclusions []Conclusion
	Confidence  float64
	Priority    int
	Domain      string
	CreatedAt   time.Time
}

// Condition represents a condition in a logical rule
type Condition struct {
	Field    string
	Operator string
	Value    interface{}
	Weight   float64
}

// Conclusion represents a conclusion from applying a rule
type Conclusion struct {
	Field      string
	Value      interface{}
	Confidence float64
	Reasoning  string
}

// CausalLink represents a causal relationship
type CausalLink struct {
	Cause      string
	Effect     string
	Strength   float64
	Evidence   []string
	Confidence float64
}

// RuleEngine manages and applies logical rules
type RuleEngine struct {
	rules   map[string]Rule
	mu      sync.RWMutex
	metrics map[string]float64
}

// CausalInferenceEngine handles causal reasoning
type CausalInferenceEngine struct {
	causalGraph map[string][]CausalLink
	mu          sync.RWMutex
}

// LogicSolver validates logical conclusions
type LogicSolver struct {
	timeout time.Duration
}

// NewReasoningLayer creates a new reasoning layer
func NewReasoningLayer(config ReasoningLayerConfig) *ReasoningLayer {
	return &ReasoningLayer{
		ruleEngine:      NewRuleEngine(),
		causalInference: NewCausalInferenceEngine(),
		logicSolver:     NewLogicSolver(time.Duration(config.LogicTimeout) * time.Millisecond),
		rules:           make(map[string]Rule),
		config:          config,
	}
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		rules:   make(map[string]Rule),
		metrics: make(map[string]float64),
	}
}

// NewCausalInferenceEngine creates a new causal inference engine
func NewCausalInferenceEngine() *CausalInferenceEngine {
	return &CausalInferenceEngine{
		causalGraph: make(map[string][]CausalLink),
	}
}

// NewLogicSolver creates a new logic solver
func NewLogicSolver(timeout time.Duration) *LogicSolver {
	return &LogicSolver{
		timeout: timeout,
	}
}

// Process processes data through the reasoning layer
func (rl *ReasoningLayer) Process(ctx context.Context, contextData map[string]interface{}) (*ReasoningResult, error) {
	start := time.Now()

	// Find applicable rules
	applicableRules := rl.ruleEngine.FindApplicableRules(contextData)

	// Apply rules to generate conclusions
	conclusions := rl.applyRules(applicableRules, contextData)

	// Perform causal inference if enabled
	var causalChain []CausalLink
	if rl.config.EnableCausal {
		causalChain = rl.causalInference.Infer(contextData, applicableRules)
	}

	// Validate conclusions with logic solver
	validatedConclusions := rl.logicSolver.Validate(conclusions)

	// Calculate confidence
	confidence := rl.calculateConfidence(validatedConclusions, applicableRules)

	// Build derivation path
	derivationPath := rl.buildDerivationPath(applicableRules, conclusions)

	result := &ReasoningResult{
		Conclusions:    validatedConclusions,
		CausalChain:    causalChain,
		AppliedRules:   applicableRules,
		Confidence:     confidence,
		DerivationPath: derivationPath,
		Timestamp:      time.Now(),
	}

	// Update rule metrics
	rl.updateRuleMetrics(applicableRules, confidence, time.Since(start))

	return result, nil
}

// FindApplicableRules finds rules that can be applied to the given data
func (re *RuleEngine) FindApplicableRules(data map[string]interface{}) []Rule {
	re.mu.RLock()
	defer re.mu.RUnlock()

	var applicable []Rule

	for _, rule := range re.rules {
		if re.ruleMatches(rule, data) {
			applicable = append(applicable, rule)
		}
	}

	return applicable
}

// ruleMatches checks if a rule's conditions are satisfied by the data
func (re *RuleEngine) ruleMatches(rule Rule, data map[string]interface{}) bool {
	for _, condition := range rule.Conditions {
		if !re.conditionMatches(condition, data) {
			return false
		}
	}
	return true
}

// conditionMatches checks if a single condition is satisfied
func (re *RuleEngine) conditionMatches(condition Condition, data map[string]interface{}) bool {
	value, exists := data[condition.Field]
	if !exists {
		return false
	}

	switch condition.Operator {
	case "equals":
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", condition.Value)
	case "contains":
		str, ok := value.(string)
		if !ok {
			return false
		}
		substr, ok := condition.Value.(string)
		if !ok {
			return false
		}
		return fmt.Sprintf("%v", str) == fmt.Sprintf("%v", substr)
	case "greater_than":
		return re.compareNumerically(value, condition.Value, ">")
	case "less_than":
		return re.compareNumerically(value, condition.Value, "<")
	case "exists":
		return exists
	default:
		return false
	}
}

// compareNumerically compares two values numerically
func (re *RuleEngine) compareNumerically(v1, v2 interface{}, op string) bool {
	f1, ok1 := re.toFloat64(v1)
	f2, ok2 := re.toFloat64(v2)

	if !ok1 || !ok2 {
		return false
	}

	switch op {
	case ">":
		return f1 > f2
	case "<":
		return f1 < f2
	case ">=":
		return f1 >= f2
	case "<=":
		return f1 <= f2
	case "==":
		return math.Abs(f1-f2) < 1e-9
	default:
		return false
	}
}

// toFloat64 converts interface{} to float64
func (re *RuleEngine) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		return 0, false
	}
}

// applyRules applies the given rules to generate conclusions
func (rl *ReasoningLayer) applyRules(rules []Rule, contextData map[string]interface{}) []Conclusion {
	var conclusions []Conclusion

	for _, rule := range rules {
		for _, conclusion := range rule.Conclusions {
			// Apply rule-specific reasoning
			adjustedConclusion := conclusion
			adjustedConclusion.Confidence *= rule.Confidence
			adjustedConclusion.Reasoning = fmt.Sprintf("Applied rule: %s", rule.Name)

			conclusions = append(conclusions, adjustedConclusion)
		}
	}

	return conclusions
}

// Infer performs causal inference
func (cie *CausalInferenceEngine) Infer(contextData map[string]interface{}, rules []Rule) []CausalLink {
	cie.mu.RLock()
	defer cie.mu.RUnlock()

	var causalChain []CausalLink

	// Simple causal inference based on rules
	for _, rule := range rules {
		for _, condition := range rule.Conditions {
			for _, conclusion := range rule.Conclusions {
				link := CausalLink{
					Cause:      condition.Field,
					Effect:     conclusion.Field,
					Strength:   rule.Confidence * condition.Weight,
					Evidence:   []string{rule.ID},
					Confidence: rule.Confidence,
				}
				causalChain = append(causalChain, link)
			}
		}
	}

	return causalChain
}

// Validate validates conclusions using logical constraints
func (ls *LogicSolver) Validate(conclusions []Conclusion) []Conclusion {
	var validated []Conclusion

	// Simple validation - in a real implementation, this would be more sophisticated
	for _, conclusion := range conclusions {
		if conclusion.Confidence >= 0.1 { // Minimum confidence threshold
			validated = append(validated, conclusion)
		}
	}

	return validated
}

// calculateConfidence calculates overall confidence for the reasoning result
func (rl *ReasoningLayer) calculateConfidence(conclusions []Conclusion, rules []Rule) float64 {
	if len(conclusions) == 0 {
		return 0.0
	}

	// Weighted average of conclusion confidences
	totalConf := 0.0
	totalWeight := 0.0

	for _, conclusion := range conclusions {
		weight := 1.0 // Could be based on rule priority
		totalConf += conclusion.Confidence * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalConf / totalWeight
}

// buildDerivationPath creates a path showing how conclusions were derived
func (rl *ReasoningLayer) buildDerivationPath(rules []Rule, conclusions []Conclusion) []string {
	var path []string

	for _, rule := range rules {
		path = append(path, fmt.Sprintf("Applied rule: %s", rule.Name))
	}

	for _, conclusion := range conclusions {
		path = append(path, fmt.Sprintf("Concluded: %s = %v (confidence: %.2f)",
			conclusion.Field, conclusion.Value, conclusion.Confidence))
	}

	return path
}

// updateRuleMetrics updates performance metrics for applied rules
func (rl *ReasoningLayer) updateRuleMetrics(rules []Rule, confidence float64, duration time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for _, rule := range rules {
		metricKey := fmt.Sprintf("rule_%s_performance", rule.ID)
		rl.ruleEngine.metrics[metricKey] = confidence

		durationKey := fmt.Sprintf("rule_%s_duration", rule.ID)
		rl.ruleEngine.metrics[durationKey] = duration.Seconds()
	}
}

// AddRule adds a new rule to the reasoning system
func (rl *ReasoningLayer) AddRule(rule Rule) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if len(rl.rules) >= rl.config.MaxRules {
		return fmt.Errorf("maximum number of rules (%d) reached", rl.config.MaxRules)
	}

	rule.CreatedAt = time.Now()
	rl.rules[rule.ID] = rule
	rl.ruleEngine.rules[rule.ID] = rule

	return nil
}

// GetRules returns all rules in the system
func (rl *ReasoningLayer) GetRules() []Rule {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	rules := make([]Rule, 0, len(rl.rules))
	for _, rule := range rl.rules {
		rules = append(rules, rule)
	}

	return rules
}
