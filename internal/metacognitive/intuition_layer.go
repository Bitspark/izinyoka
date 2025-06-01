package metacognitive

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// IntuitionLayer implements pattern recognition and predictive modeling
type IntuitionLayer struct {
	patternRecognizer *PatternRecognizer
	predictor         *PredictiveModel
	emergentProcessor *EmergentProcessor
	patterns          map[string]Pattern
	mu                sync.RWMutex
	config            IntuitionLayerConfig
}

// IntuitionLayerConfig configures the intuition layer behavior
type IntuitionLayerConfig struct {
	MaxPatterns        int     `yaml:"max_patterns"`
	PatternThreshold   float64 `yaml:"pattern_threshold"`
	PredictionHorizon  int     `yaml:"prediction_horizon"`
	EmergenceThreshold float64 `yaml:"emergence_threshold"`
	LearningRate       float64 `yaml:"learning_rate"`
	NoveltyThreshold   float64 `yaml:"novelty_threshold"`
}

// IntuitionResult represents the output from intuition layer processing
type IntuitionResult struct {
	Predictions        []Prediction
	RecognizedPatterns []Pattern
	EmergentInsights   []EmergentInsight
	Novelty            float64
	Confidence         float64
	Timestamp          time.Time
}

// Pattern represents a recognized pattern
type Pattern struct {
	ID         string
	Name       string
	Features   []Feature
	Frequency  int
	Confidence float64
	LastSeen   time.Time
	Domain     string
	Variations []PatternVariation
}

// Feature represents a feature in a pattern
type Feature struct {
	Name   string
	Value  interface{}
	Weight float64
}

// PatternVariation represents a variation of a base pattern
type PatternVariation struct {
	ID         string
	Features   []Feature
	Similarity float64
	Frequency  int
}

// Prediction represents a prediction made by the intuition layer
type Prediction struct {
	Target     string
	Value      interface{}
	Confidence float64
	Horizon    time.Duration
	Reasoning  string
	Features   []Feature
}

// EmergentInsight represents an emergent insight discovered by the system
type EmergentInsight struct {
	ID          string
	Description string
	Patterns    []string
	Confidence  float64
	Novelty     float64
	Evidence    []string
	CreatedAt   time.Time
}

// PatternRecognizer handles pattern recognition
type PatternRecognizer struct {
	patterns         map[string]Pattern
	featureExtractor *FeatureExtractor
	similarity       *SimilarityCalculator
	mu               sync.RWMutex
}

// PredictiveModel handles predictive modeling
type PredictiveModel struct {
	models       map[string]Model
	ensemble     *EnsembleModel
	trainingData []DataPoint
	mu           sync.RWMutex
}

// EmergentProcessor detects emergent properties and insights
type EmergentProcessor struct {
	insights   map[string]EmergentInsight
	complexity *ComplexityAnalyzer
	emergence  *EmergenceDetector
	mu         sync.RWMutex
}

// Model represents a predictive model
type Model interface {
	Predict(features []Feature) (interface{}, float64, error)
	Train(data []DataPoint) error
	Evaluate() ModelMetrics
}

// DataPoint represents a training data point
type DataPoint struct {
	Features []Feature
	Target   interface{}
	Weight   float64
}

// ModelMetrics represents model performance metrics
type ModelMetrics struct {
	Accuracy  float64
	Precision float64
	Recall    float64
	F1Score   float64
}

// FeatureExtractor extracts features from raw data
type FeatureExtractor struct {
	extractors map[string]func(interface{}) []Feature
}

// SimilarityCalculator calculates similarity between patterns
type SimilarityCalculator struct {
	algorithms map[string]func([]Feature, []Feature) float64
}

// EnsembleModel combines multiple models
type EnsembleModel struct {
	models  []Model
	weights []float64
}

// ComplexityAnalyzer analyzes system complexity
type ComplexityAnalyzer struct {
	metrics map[string]float64
}

// EmergenceDetector detects emergent properties
type EmergenceDetector struct {
	threshold float64
	history   []float64
}

// NewIntuitionLayer creates a new intuition layer
func NewIntuitionLayer(config IntuitionLayerConfig) *IntuitionLayer {
	return &IntuitionLayer{
		patternRecognizer: NewPatternRecognizer(),
		predictor:         NewPredictiveModel(),
		emergentProcessor: NewEmergentProcessor(config.EmergenceThreshold),
		patterns:          make(map[string]Pattern),
		config:            config,
	}
}

// NewPatternRecognizer creates a new pattern recognizer
func NewPatternRecognizer() *PatternRecognizer {
	return &PatternRecognizer{
		patterns:         make(map[string]Pattern),
		featureExtractor: NewFeatureExtractor(),
		similarity:       NewSimilarityCalculator(),
	}
}

// NewPredictiveModel creates a new predictive model
func NewPredictiveModel() *PredictiveModel {
	return &PredictiveModel{
		models:       make(map[string]Model),
		ensemble:     NewEnsembleModel(),
		trainingData: make([]DataPoint, 0),
	}
}

// NewEmergentProcessor creates a new emergent processor
func NewEmergentProcessor(threshold float64) *EmergentProcessor {
	return &EmergentProcessor{
		insights:   make(map[string]EmergentInsight),
		complexity: &ComplexityAnalyzer{metrics: make(map[string]float64)},
		emergence:  &EmergenceDetector{threshold: threshold, history: make([]float64, 0)},
	}
}

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor() *FeatureExtractor {
	extractors := make(map[string]func(interface{}) []Feature)

	// Add basic feature extractors
	extractors["numeric"] = extractNumericFeatures
	extractors["text"] = extractTextFeatures
	extractors["temporal"] = extractTemporalFeatures

	return &FeatureExtractor{extractors: extractors}
}

// NewSimilarityCalculator creates a new similarity calculator
func NewSimilarityCalculator() *SimilarityCalculator {
	algorithms := make(map[string]func([]Feature, []Feature) float64)

	// Add similarity algorithms
	algorithms["cosine"] = cosineSimilarity
	algorithms["euclidean"] = euclideanSimilarity
	algorithms["jaccard"] = jaccardSimilarity

	return &SimilarityCalculator{algorithms: algorithms}
}

// NewEnsembleModel creates a new ensemble model
func NewEnsembleModel() *EnsembleModel {
	return &EnsembleModel{
		models:  make([]Model, 0),
		weights: make([]float64, 0),
	}
}

// Process processes data through the intuition layer
func (il *IntuitionLayer) Process(ctx context.Context, data map[string]interface{}, reasoningResult *ReasoningResult) (*IntuitionResult, error) {
	start := time.Now()

	// Extract features from input data
	features := il.patternRecognizer.featureExtractor.ExtractFeatures(data)

	// Recognize patterns
	recognizedPatterns := il.patternRecognizer.RecognizePatterns(features)

	// Make predictions
	predictions := il.predictor.MakePredictions(features, reasoningResult)

	// Detect emergent insights
	emergentInsights := il.emergentProcessor.DetectEmergence(recognizedPatterns, predictions)

	// Calculate novelty
	novelty := il.calculateNovelty(recognizedPatterns, features)

	// Calculate confidence
	confidence := il.calculateConfidence(predictions, recognizedPatterns)

	result := &IntuitionResult{
		Predictions:        predictions,
		RecognizedPatterns: recognizedPatterns,
		EmergentInsights:   emergentInsights,
		Novelty:            novelty,
		Confidence:         confidence,
		Timestamp:          time.Now(),
	}

	// Update patterns and models
	il.updatePatterns(recognizedPatterns)
	il.updateModels(features, reasoningResult)

	return result, nil
}

// ExtractFeatures extracts features from data
func (fe *FeatureExtractor) ExtractFeatures(data map[string]interface{}) []Feature {
	var allFeatures []Feature

	for key, value := range data {
		// Determine feature type and extract accordingly
		switch v := value.(type) {
		case float64, int, int64:
			features := fe.extractors["numeric"](v)
			for i := range features {
				features[i].Name = key + "_" + features[i].Name
			}
			allFeatures = append(allFeatures, features...)
		case string:
			features := fe.extractors["text"](v)
			for i := range features {
				features[i].Name = key + "_" + features[i].Name
			}
			allFeatures = append(allFeatures, features...)
		case time.Time:
			features := fe.extractors["temporal"](v)
			for i := range features {
				features[i].Name = key + "_" + features[i].Name
			}
			allFeatures = append(allFeatures, features...)
		}
	}

	return allFeatures
}

// RecognizePatterns recognizes patterns in the given features
func (pr *PatternRecognizer) RecognizePatterns(features []Feature) []Pattern {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var recognized []Pattern

	for _, pattern := range pr.patterns {
		similarity := pr.similarity.algorithms["cosine"](features, pattern.Features)
		if similarity >= 0.7 { // Threshold for pattern recognition
			pattern.LastSeen = time.Now()
			pattern.Frequency++
			recognized = append(recognized, pattern)
		}
	}

	// Sort by confidence
	sort.Slice(recognized, func(i, j int) bool {
		return recognized[i].Confidence > recognized[j].Confidence
	})

	return recognized
}

// MakePredictions makes predictions based on features and reasoning
func (pm *PredictiveModel) MakePredictions(features []Feature, reasoningResult *ReasoningResult) []Prediction {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var predictions []Prediction

	// Use ensemble model to make predictions
	if len(pm.ensemble.models) > 0 {
		value, confidence, err := pm.ensemble.Predict(features)
		if err == nil {
			prediction := Prediction{
				Target:     "primary_prediction",
				Value:      value,
				Confidence: confidence,
				Horizon:    time.Minute * 10, // Default horizon
				Reasoning:  "Ensemble model prediction",
				Features:   features,
			}
			predictions = append(predictions, prediction)
		}
	}

	return predictions
}

// DetectEmergence detects emergent insights
func (ep *EmergentProcessor) DetectEmergence(patterns []Pattern, predictions []Prediction) []EmergentInsight {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	var insights []EmergentInsight

	// Analyze complexity increase
	currentComplexity := ep.complexity.CalculateComplexity(patterns, predictions)
	ep.emergence.history = append(ep.emergence.history, currentComplexity)

	// Detect emergent properties
	if len(ep.emergence.history) > 5 {
		recentHistory := ep.emergence.history[len(ep.emergence.history)-5:]
		if ep.detectEmergentProperty(recentHistory) {
			insight := EmergentInsight{
				ID:          fmt.Sprintf("insight_%d", time.Now().Unix()),
				Description: "Emergent complexity detected",
				Patterns:    extractPatternIDs(patterns),
				Confidence:  0.8,
				Novelty:     0.7,
				Evidence:    []string{"complexity_increase", "pattern_interaction"},
				CreatedAt:   time.Now(),
			}
			insights = append(insights, insight)
			ep.insights[insight.ID] = insight
		}
	}

	return insights
}

// calculateNovelty calculates novelty score
func (il *IntuitionLayer) calculateNovelty(patterns []Pattern, features []Feature) float64 {
	if len(patterns) == 0 {
		return 1.0 // Completely novel
	}

	// Calculate average pattern frequency (lower frequency = higher novelty)
	totalFreq := 0
	for _, pattern := range patterns {
		totalFreq += pattern.Frequency
	}

	avgFreq := float64(totalFreq) / float64(len(patterns))

	// Normalize to 0-1 scale (assuming max frequency of 1000)
	novelty := 1.0 - math.Min(avgFreq/1000.0, 1.0)

	return novelty
}

// calculateConfidence calculates overall confidence
func (il *IntuitionLayer) calculateConfidence(predictions []Prediction, patterns []Pattern) float64 {
	if len(predictions) == 0 && len(patterns) == 0 {
		return 0.0
	}

	totalConf := 0.0
	count := 0

	for _, pred := range predictions {
		totalConf += pred.Confidence
		count++
	}

	for _, pattern := range patterns {
		totalConf += pattern.Confidence
		count++
	}

	if count == 0 {
		return 0.0
	}

	return totalConf / float64(count)
}

// Helper functions

// extractNumericFeatures extracts features from numeric values
func extractNumericFeatures(value interface{}) []Feature {
	var features []Feature

	switch v := value.(type) {
	case float64:
		features = append(features, Feature{Name: "value", Value: v, Weight: 1.0})
		features = append(features, Feature{Name: "magnitude", Value: math.Abs(v), Weight: 0.8})
	case int:
		f := float64(v)
		features = append(features, Feature{Name: "value", Value: f, Weight: 1.0})
		features = append(features, Feature{Name: "magnitude", Value: math.Abs(f), Weight: 0.8})
	}

	return features
}

// extractTextFeatures extracts features from text values
func extractTextFeatures(value interface{}) []Feature {
	var features []Feature

	if str, ok := value.(string); ok {
		features = append(features, Feature{Name: "length", Value: float64(len(str)), Weight: 0.6})
		features = append(features, Feature{Name: "has_content", Value: 1.0, Weight: 0.8})
	}

	return features
}

// extractTemporalFeatures extracts features from temporal values
func extractTemporalFeatures(value interface{}) []Feature {
	var features []Feature

	if t, ok := value.(time.Time); ok {
		now := time.Now()
		features = append(features, Feature{Name: "hour", Value: float64(t.Hour()), Weight: 0.7})
		features = append(features, Feature{Name: "age_hours", Value: now.Sub(t).Hours(), Weight: 0.9})
	}

	return features
}

// cosineSimilarity calculates cosine similarity between two feature vectors
func cosineSimilarity(f1, f2 []Feature) float64 {
	if len(f1) == 0 || len(f2) == 0 {
		return 0.0
	}

	// Simple implementation - in practice would be more sophisticated
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	// Convert to maps for easier lookup
	map1 := make(map[string]float64)
	map2 := make(map[string]float64)

	for _, f := range f1 {
		if val, ok := f.Value.(float64); ok {
			map1[f.Name] = val
			norm1 += val * val
		}
	}

	for _, f := range f2 {
		if val, ok := f.Value.(float64); ok {
			map2[f.Name] = val
			norm2 += val * val
		}
	}

	// Calculate dot product
	for name, val1 := range map1 {
		if val2, exists := map2[name]; exists {
			dotProduct += val1 * val2
		}
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// euclideanSimilarity calculates euclidean similarity
func euclideanSimilarity(f1, f2 []Feature) float64 {
	// Placeholder implementation
	return cosineSimilarity(f1, f2) * 0.8
}

// jaccardSimilarity calculates jaccard similarity
func jaccardSimilarity(f1, f2 []Feature) float64 {
	// Placeholder implementation
	return cosineSimilarity(f1, f2) * 0.9
}

// Predict makes a prediction using the ensemble
func (em *EnsembleModel) Predict(features []Feature) (interface{}, float64, error) {
	if len(em.models) == 0 {
		return nil, 0.0, fmt.Errorf("no models in ensemble")
	}

	// Simple averaging for now
	totalConf := 0.0
	validPredictions := 0

	for i, model := range em.models {
		_, conf, err := model.Predict(features)
		if err == nil {
			weight := 1.0
			if i < len(em.weights) {
				weight = em.weights[i]
			}
			totalConf += conf * weight
			validPredictions++
		}
	}

	if validPredictions == 0 {
		return nil, 0.0, fmt.Errorf("no valid predictions")
	}

	avgConf := totalConf / float64(validPredictions)
	return "ensemble_prediction", avgConf, nil
}

// CalculateComplexity calculates system complexity
func (ca *ComplexityAnalyzer) CalculateComplexity(patterns []Pattern, predictions []Prediction) float64 {
	// Simple complexity metric based on number of patterns and predictions
	complexity := float64(len(patterns)*2 + len(predictions))
	ca.metrics["current_complexity"] = complexity
	return complexity
}

// detectEmergentProperty detects if emergent properties are arising
func (ed *EmergenceDetector) detectEmergentProperty(history []float64) bool {
	if len(history) < 3 {
		return false
	}

	// Check for increasing complexity trend
	increasing := 0
	for i := 1; i < len(history); i++ {
		if history[i] > history[i-1] {
			increasing++
		}
	}

	// If most recent values show increasing complexity
	return float64(increasing)/float64(len(history)-1) > ed.threshold
}

// extractPatternIDs extracts pattern IDs from a slice of patterns
func extractPatternIDs(patterns []Pattern) []string {
	ids := make([]string, len(patterns))
	for i, pattern := range patterns {
		ids[i] = pattern.ID
	}
	return ids
}

// updatePatterns updates the pattern database
func (il *IntuitionLayer) updatePatterns(patterns []Pattern) {
	il.mu.Lock()
	defer il.mu.Unlock()

	for _, pattern := range patterns {
		il.patterns[pattern.ID] = pattern
		il.patternRecognizer.patterns[pattern.ID] = pattern
	}
}

// updateModels updates the predictive models
func (il *IntuitionLayer) updateModels(features []Feature, reasoningResult *ReasoningResult) {
	// Create training data point
	if reasoningResult != nil && len(reasoningResult.Conclusions) > 0 {
		dataPoint := DataPoint{
			Features: features,
			Target:   reasoningResult.Conclusions[0].Value,
			Weight:   reasoningResult.Confidence,
		}

		il.predictor.trainingData = append(il.predictor.trainingData, dataPoint)

		// Limit training data size
		if len(il.predictor.trainingData) > 1000 {
			il.predictor.trainingData = il.predictor.trainingData[100:]
		}
	}
}
