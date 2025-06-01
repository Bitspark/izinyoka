package metacognitive

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/fullscreen-triangle/izinyoka/internal/stream"
)

// ContextLayer implements the knowledge integration and situational awareness layer
type ContextLayer struct {
	knowledgeBase    *knowledge.KnowledgeBase
	attentionWeights map[string]float64
	workingMemory    map[string]interface{}
	confidenceCalc   *ConfidenceCalculator
	mu               sync.RWMutex
	config           ContextLayerConfig
}

// ContextLayerConfig configures the context layer behavior
type ContextLayerConfig struct {
	MaxMemoryItems      int     `yaml:"max_memory_items"`
	AttentionThreshold  float64 `yaml:"attention_threshold"`
	KnowledgeWeight     float64 `yaml:"knowledge_weight"`
	TemporalWeight      float64 `yaml:"temporal_weight"`
	SimilarityThreshold float64 `yaml:"similarity_threshold"`
}

// ContextResult represents the output from context layer processing
type ContextResult struct {
	Interpretation map[string]interface{}
	Confidence     float64
	Knowledge      []knowledge.KnowledgeItem
	AttentionMask  map[string]float64
	Timestamp      time.Time
}

// AttentionMechanism computes attention weights for different aspects of data
type AttentionMechanism struct {
	weights map[string]float64
	decay   float64
}

// ConfidenceCalculator computes confidence scores for interpretations
type ConfidenceCalculator struct {
	knowledgeWeight   float64
	noveltyWeight     float64
	consistencyWeight float64
}

// NewContextLayer creates a new context layer with knowledge base integration
func NewContextLayer(kb *knowledge.KnowledgeBase, config ContextLayerConfig) *ContextLayer {
	return &ContextLayer{
		knowledgeBase:    kb,
		attentionWeights: make(map[string]float64),
		workingMemory:    make(map[string]interface{}),
		confidenceCalc:   NewConfidenceCalculator(0.4, 0.3, 0.3),
		config:           config,
	}
}

// NewAttentionMechanism creates a new attention mechanism
func NewAttentionMechanism(decay float64) *AttentionMechanism {
	return &AttentionMechanism{
		weights: make(map[string]float64),
		decay:   decay,
	}
}

// NewConfidenceCalculator creates a new confidence calculator
func NewConfidenceCalculator(knowledgeWeight, noveltyWeight, consistencyWeight float64) *ConfidenceCalculator {
	return &ConfidenceCalculator{
		knowledgeWeight:   knowledgeWeight,
		noveltyWeight:     noveltyWeight,
		consistencyWeight: consistencyWeight,
	}
}

// Process processes data through the context layer
func (cl *ContextLayer) Process(ctx context.Context, data *stream.StreamData) (*ContextResult, error) {
	start := time.Now()
	defer func() {
		data.LogProcessing("context_layer", "process", time.Since(start))
	}()

	// Extract relevant knowledge from knowledge base
	relevantKnowledge, err := cl.queryRelevantKnowledge(data)
	if err != nil {
		data.LogError("context_layer", "knowledge_query", err.Error())
		relevantKnowledge = []knowledge.KnowledgeItem{} // Continue with empty knowledge
	}

	// Compute attention weights
	attentionWeights := cl.computeAttentionWeights(data, relevantKnowledge)

	// Generate contextualized interpretation
	interpretation := cl.generateInterpretation(data, relevantKnowledge, attentionWeights)

	// Calculate confidence
	confidence := cl.confidenceCalc.Calculate(interpretation, relevantKnowledge, data)

	// Update working memory
	cl.updateWorkingMemory(data, interpretation)

	result := &ContextResult{
		Interpretation: interpretation,
		Confidence:     confidence,
		Knowledge:      relevantKnowledge,
		AttentionMask:  attentionWeights,
		Timestamp:      time.Now(),
	}

	// Add processing metadata
	data.AddMetadata("context_confidence", confidence)
	data.AddMetadata("knowledge_items_used", len(relevantKnowledge))

	return result, nil
}

// queryRelevantKnowledge retrieves relevant knowledge items from the knowledge base
func (cl *ContextLayer) queryRelevantKnowledge(data *stream.StreamData) ([]knowledge.KnowledgeItem, error) {
	// Build query based on data type and content
	query := &knowledge.KnowledgeQuery{
		Type:  data.Type,
		Limit: 10,
	}

	// Add tag-based search if context hint exists
	if contextHint, exists := data.Metadata["context_hint"]; exists {
		if hint, ok := contextHint.(string); ok {
			query.Tags = []string{hint}
		}
	}

	// Query the knowledge base
	items, err := cl.knowledgeBase.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge base: %w", err)
	}

	// Filter by similarity threshold
	var relevant []knowledge.KnowledgeItem
	for _, item := range items {
		similarity := cl.calculateSimilarity(data.Data, item.Data)
		if similarity >= cl.config.SimilarityThreshold {
			relevant = append(relevant, item)
		}
	}

	return relevant, nil
}

// computeAttentionWeights calculates attention weights for different data aspects
func (cl *ContextLayer) computeAttentionWeights(data *stream.StreamData, knowledge []knowledge.KnowledgeItem) map[string]float64 {
	weights := make(map[string]float64)

	// Base attention weights from data content
	for key, value := range data.Data {
		// Simple content-based attention
		if str, ok := value.(string); ok {
			weights[key] = float64(len(str)) / 100.0 // Normalize by length
		} else {
			weights[key] = 0.5 // Default weight for non-string data
		}
	}

	// Boost attention for items matching knowledge
	for _, item := range knowledge {
		for key := range data.Data {
			if _, exists := item.Data[key]; exists {
				weights[key] *= 1.5 // Boost attention for known concepts
			}
		}
	}

	// Normalize weights
	return cl.normalizeWeights(weights)
}

// generateInterpretation creates a contextualized interpretation
func (cl *ContextLayer) generateInterpretation(
	data *stream.StreamData,
	knowledge []knowledge.KnowledgeItem,
	attention map[string]float64,
) map[string]interface{} {
	interpretation := make(map[string]interface{})

	// Start with original data
	for key, value := range data.Data {
		weight := attention[key]
		if weight > cl.config.AttentionThreshold {
			interpretation[key] = value
		}
	}

	// Enrich with knowledge
	knowledgeContext := cl.buildKnowledgeContext(knowledge)
	interpretation["knowledge_context"] = knowledgeContext

	// Add temporal context
	interpretation["temporal_context"] = cl.buildTemporalContext(data)

	// Add similarity metrics
	interpretation["similarity_scores"] = cl.calculateKnowledgeSimilarity(data.Data, knowledge)

	return interpretation
}

// buildKnowledgeContext creates a knowledge context from relevant items
func (cl *ContextLayer) buildKnowledgeContext(knowledge []knowledge.KnowledgeItem) map[string]interface{} {
	context := make(map[string]interface{})

	// Aggregate knowledge by type
	typeGroups := make(map[string][]knowledge.KnowledgeItem)
	for _, item := range knowledge {
		typeGroups[item.Type] = append(typeGroups[item.Type], item)
	}

	// Create weighted context from each type
	for itemType, items := range typeGroups {
		typeContext := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			typeContext = append(typeContext, item.Data)
		}
		context[itemType] = typeContext
	}

	return context
}

// buildTemporalContext creates temporal context information
func (cl *ContextLayer) buildTemporalContext(data *stream.StreamData) map[string]interface{} {
	context := map[string]interface{}{
		"timestamp":   data.CreatedAt,
		"age_seconds": time.Since(data.CreatedAt).Seconds(),
		"time_of_day": data.CreatedAt.Hour(),
		"day_of_week": data.CreatedAt.Weekday().String(),
	}

	// Add processing timeline
	if len(data.ProcessingLog) > 0 {
		context["processing_duration"] = time.Since(data.ProcessingLog[0].Timestamp).Seconds()
		context["processing_steps"] = len(data.ProcessingLog)
	}

	return context
}

// calculateKnowledgeSimilarity computes similarity scores with knowledge items
func (cl *ContextLayer) calculateKnowledgeSimilarity(
	data map[string]interface{},
	knowledge []knowledge.KnowledgeItem,
) map[string]float64 {
	similarities := make(map[string]float64)

	for _, item := range knowledge {
		similarity := cl.calculateSimilarity(data, item.Data)
		similarities[item.ID] = similarity
	}

	return similarities
}

// calculateSimilarity computes similarity between two data maps
func (cl *ContextLayer) calculateSimilarity(data1, data2 map[string]interface{}) float64 {
	if len(data1) == 0 || len(data2) == 0 {
		return 0.0
	}

	commonKeys := 0
	totalKeys := 0

	// Count common keys and calculate similarity
	for key, value1 := range data1 {
		totalKeys++
		if value2, exists := data2[key]; exists {
			commonKeys++
			// Additional similarity scoring based on value similarity
			if fmt.Sprintf("%v", value1) == fmt.Sprintf("%v", value2) {
				commonKeys++ // Boost for exact matches
			}
		}
	}

	for key := range data2 {
		if _, exists := data1[key]; !exists {
			totalKeys++
		}
	}

	if totalKeys == 0 {
		return 0.0
	}

	return float64(commonKeys) / float64(totalKeys)
}

// updateWorkingMemory updates the working memory with new information
func (cl *ContextLayer) updateWorkingMemory(data *stream.StreamData, interpretation map[string]interface{}) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Add interpretation to working memory
	memoryKey := fmt.Sprintf("interpretation_%s", data.ID)
	cl.workingMemory[memoryKey] = interpretation

	// Maintain memory size limit
	if len(cl.workingMemory) > cl.config.MaxMemoryItems {
		// Remove oldest items (simple FIFO for now)
		count := 0
		for key := range cl.workingMemory {
			if count >= cl.config.MaxMemoryItems/2 {
				break
			}
			delete(cl.workingMemory, key)
			count++
		}
	}
}

// normalizeWeights normalizes attention weights to sum to 1.0
func (cl *ContextLayer) normalizeWeights(weights map[string]float64) map[string]float64 {
	total := 0.0
	for _, weight := range weights {
		total += weight
	}

	if total == 0.0 {
		return weights
	}

	normalized := make(map[string]float64)
	for key, weight := range weights {
		normalized[key] = weight / total
	}

	return normalized
}

// Calculate computes confidence score for an interpretation
func (cc *ConfidenceCalculator) Calculate(
	interpretation map[string]interface{},
	knowledge []knowledge.KnowledgeItem,
	data *stream.StreamData,
) float64 {
	// Knowledge-based confidence
	knowledgeConf := cc.calculateKnowledgeConfidence(knowledge)

	// Novelty-based confidence (inverse of novelty)
	noveltyConf := cc.calculateNoveltyConfidence(interpretation, knowledge)

	// Consistency-based confidence
	consistencyConf := cc.calculateConsistencyConfidence(interpretation)

	// Weighted combination
	totalConf := (knowledgeConf * cc.knowledgeWeight) +
		(noveltyConf * cc.noveltyWeight) +
		(consistencyConf * cc.consistencyWeight)

	// Ensure confidence is in [0, 1] range
	return math.Max(0.0, math.Min(1.0, totalConf))
}

// calculateKnowledgeConfidence computes confidence based on available knowledge
func (cc *ConfidenceCalculator) calculateKnowledgeConfidence(knowledge []knowledge.KnowledgeItem) float64 {
	if len(knowledge) == 0 {
		return 0.1 // Low confidence without knowledge
	}

	// Higher confidence with more relevant knowledge
	confidence := math.Log(float64(len(knowledge)+1)) / math.Log(11) // Normalize to [0,1]
	return math.Min(confidence, 1.0)
}

// calculateNoveltyConfidence computes confidence based on novelty
func (cc *ConfidenceCalculator) calculateNoveltyConfidence(
	interpretation map[string]interface{},
	knowledge []knowledge.KnowledgeItem,
) float64 {
	if len(knowledge) == 0 {
		return 0.5 // Medium confidence for novel situations
	}

	// Calculate how much the interpretation matches existing knowledge
	matchScore := 0.0
	for _, item := range knowledge {
		// Simple matching - could be more sophisticated
		for key := range interpretation {
			if _, exists := item.Data[key]; exists {
				matchScore += 1.0
			}
		}
	}

	// Normalize by total possible matches
	totalPossible := float64(len(interpretation) * len(knowledge))
	if totalPossible == 0 {
		return 0.5
	}

	novelty := 1.0 - (matchScore / totalPossible)
	return 1.0 - novelty // High novelty = low confidence
}

// calculateConsistencyConfidence computes confidence based on internal consistency
func (cc *ConfidenceCalculator) calculateConsistencyConfidence(interpretation map[string]interface{}) float64 {
	// Simple consistency check - could be more sophisticated
	if len(interpretation) == 0 {
		return 0.0
	}

	// For now, return a base confidence that could be enhanced with
	// more sophisticated consistency checking
	return 0.7 // Default consistency confidence
}
