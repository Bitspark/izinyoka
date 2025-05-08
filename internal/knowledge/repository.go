package knowledge

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// KnowledgeBase manages storage and retrieval of knowledge
type KnowledgeBase struct {
	items    map[string]KnowledgeItem
	basePath string
	indexer  *Indexer
	mu       sync.RWMutex
	isInit   bool
}

// KnowledgeItem represents a unit of stored knowledge
type KnowledgeItem struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Content   interface{}            `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      []string               `json:"tags,omitempty"`
}

// KnowledgeQuery defines search parameters for retrieving knowledge
type KnowledgeQuery struct {
	ID         string                 `json:"id,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	TimeRange  *TimeRange             `json:"time_range,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Offset     int                    `json:"offset,omitempty"`
	Sort       string                 `json:"sort,omitempty"`
}

// TimeRange specifies a time window for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Indexer provides fast lookup for knowledge items
type Indexer struct {
	typeIndex map[string]map[string]bool
	tagIndex  map[string]map[string]bool
	timeIndex map[string]time.Time
}

// NewKnowledgeBase creates a new knowledge base
func NewKnowledgeBase(basePath string) (*KnowledgeBase, error) {
	if basePath == "" {
		return nil, errors.New("base path cannot be empty")
	}

	// Create indexer
	indexer := &Indexer{
		typeIndex: make(map[string]map[string]bool),
		tagIndex:  make(map[string]map[string]bool),
		timeIndex: make(map[string]time.Time),
	}

	return &KnowledgeBase{
		items:    make(map[string]KnowledgeItem),
		basePath: basePath,
		indexer:  indexer,
		isInit:   false,
	}, nil
}

// Initialize sets up the knowledge base storage
func (kb *KnowledgeBase) Initialize() error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(kb.basePath, 0755); err != nil {
		return err
	}

	// Create subdirectories for different knowledge types
	subdirs := []string{"models", "patterns", "heuristics"}
	for _, dir := range subdirs {
		if err := os.MkdirAll(filepath.Join(kb.basePath, dir), 0755); err != nil {
			return err
		}
	}

	// Initialize with some basic knowledge if empty
	if len(kb.items) == 0 {
		// Add some basic seed knowledge
		seeds := []KnowledgeItem{
			{
				ID:        "base-pattern-1",
				Type:      "pattern",
				Content:   map[string]interface{}{"pattern_type": "sequence", "values": []int{1, 2, 3}},
				Timestamp: time.Now(),
				Tags:      []string{"base", "seed"},
			},
			{
				ID:        "base-heuristic-1",
				Type:      "heuristic",
				Content:   map[string]interface{}{"if": "value > 10", "then": "prioritize"},
				Timestamp: time.Now(),
				Tags:      []string{"base", "seed"},
			},
		}

		for _, seed := range seeds {
			kb.items[seed.ID] = seed
			kb.indexItem(seed)
		}
	}

	kb.isInit = true
	return nil
}

// Store adds or updates a knowledge item
func (kb *KnowledgeBase) Store(item KnowledgeItem) error {
	if item.ID == "" {
		return errors.New("item ID cannot be empty")
	}

	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Check if knowledge base is initialized
	if !kb.isInit {
		return errors.New("knowledge base not initialized")
	}

	// Set timestamp if not set
	if item.Timestamp.IsZero() {
		item.Timestamp = time.Now()
	}

	// Initialize metadata if nil
	if item.Metadata == nil {
		item.Metadata = make(map[string]interface{})
	}

	// Update internal map
	kb.items[item.ID] = item

	// Update indices
	kb.indexItem(item)

	// Persist to storage
	return kb.persistItem(item)
}

// Query retrieves knowledge items matching the query
func (kb *KnowledgeBase) Query(query *KnowledgeQuery) ([]KnowledgeItem, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	// Check if knowledge base is initialized
	if !kb.isInit {
		return nil, errors.New("knowledge base not initialized")
	}

	// Default limit
	limit := 100
	if query.Limit > 0 {
		limit = query.Limit
	}

	// Filter by direct ID if specified
	if query.ID != "" {
		if item, ok := kb.items[query.ID]; ok {
			return []KnowledgeItem{item}, nil
		}
		return []KnowledgeItem{}, nil
	}

	// Start with all items or type-filtered items
	candidates := make(map[string]KnowledgeItem)
	if query.Type != "" {
		// Filter by type
		if typeMap, ok := kb.indexer.typeIndex[query.Type]; ok {
			for id := range typeMap {
				if item, ok := kb.items[id]; ok {
					candidates[id] = item
				}
			}
		}
	} else {
		// Start with all items
		for id, item := range kb.items {
			candidates[id] = item
		}
	}

	// Filter by tags
	if len(query.Tags) > 0 {
		for id := range candidates {
			item := candidates[id]
			if !kb.itemHasTags(item, query.Tags) {
				delete(candidates, id)
			}
		}
	}

	// Filter by time range
	if query.TimeRange != nil && !query.TimeRange.Start.IsZero() {
		for id := range candidates {
			timestamp := kb.indexer.timeIndex[id]
			if timestamp.Before(query.TimeRange.Start) ||
				(!query.TimeRange.End.IsZero() && timestamp.After(query.TimeRange.End)) {
				delete(candidates, id)
			}
		}
	}

	// Filter by attributes
	if len(query.Attributes) > 0 {
		for id := range candidates {
			item := candidates[id]
			if !kb.itemMatchesAttributes(item, query.Attributes) {
				delete(candidates, id)
			}
		}
	}

	// Convert to slice
	result := make([]KnowledgeItem, 0, len(candidates))
	for _, item := range candidates {
		result = append(result, item)
	}

	// Sort results if requested
	if query.Sort != "" {
		// Implement sorting logic here
		// ...
	}

	// Apply offset and limit
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	if offset >= len(result) {
		return []KnowledgeItem{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// Update modifies an existing knowledge item
func (kb *KnowledgeBase) Update(id string, item KnowledgeItem) error {
	if id == "" {
		return errors.New("ID cannot be empty")
	}

	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Check if knowledge base is initialized
	if !kb.isInit {
		return errors.New("knowledge base not initialized")
	}

	// Check if item exists
	_, exists := kb.items[id]
	if !exists {
		return errors.New("item not found")
	}

	// Ensure ID consistency
	item.ID = id

	// Update timestamp
	item.Timestamp = time.Now()

	// Update indices
	kb.removeItemFromIndices(id)
	kb.indexItem(item)

	// Update in memory
	kb.items[id] = item

	// Persist to storage
	return kb.persistItem(item)
}

// Delete removes a knowledge item
func (kb *KnowledgeBase) Delete(id string) error {
	if id == "" {
		return errors.New("ID cannot be empty")
	}

	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Check if knowledge base is initialized
	if !kb.isInit {
		return errors.New("knowledge base not initialized")
	}

	// Check if item exists
	_, exists := kb.items[id]
	if !exists {
		return errors.New("item not found")
	}

	// Remove from indices
	kb.removeItemFromIndices(id)

	// Remove from memory
	delete(kb.items, id)

	// Remove from storage
	return kb.removeItemFromStorage(id)
}

// Helper methods

// indexItem adds an item to all indices
func (kb *KnowledgeBase) indexItem(item KnowledgeItem) {
	// Index by type
	if _, ok := kb.indexer.typeIndex[item.Type]; !ok {
		kb.indexer.typeIndex[item.Type] = make(map[string]bool)
	}
	kb.indexer.typeIndex[item.Type][item.ID] = true

	// Index by tags
	for _, tag := range item.Tags {
		if _, ok := kb.indexer.tagIndex[tag]; !ok {
			kb.indexer.tagIndex[tag] = make(map[string]bool)
		}
		kb.indexer.tagIndex[tag][item.ID] = true
	}

	// Index by time
	kb.indexer.timeIndex[item.ID] = item.Timestamp
}

// removeItemFromIndices removes an item from all indices
func (kb *KnowledgeBase) removeItemFromIndices(id string) {
	item, exists := kb.items[id]
	if !exists {
		return
	}

	// Remove from type index
	if typeMap, ok := kb.indexer.typeIndex[item.Type]; ok {
		delete(typeMap, id)
	}

	// Remove from tag index
	for _, tag := range item.Tags {
		if tagMap, ok := kb.indexer.tagIndex[tag]; ok {
			delete(tagMap, id)
		}
	}

	// Remove from time index
	delete(kb.indexer.timeIndex, id)
}

// itemHasTags checks if an item has all the specified tags
func (kb *KnowledgeBase) itemHasTags(item KnowledgeItem, tags []string) bool {
	if len(tags) == 0 {
		return true
	}

	itemTagSet := make(map[string]bool)
	for _, tag := range item.Tags {
		itemTagSet[tag] = true
	}

	for _, tag := range tags {
		if !itemTagSet[tag] {
			return false
		}
	}

	return true
}

// itemMatchesAttributes checks if an item's content matches the specified attributes
func (kb *KnowledgeBase) itemMatchesAttributes(item KnowledgeItem, attrs map[string]interface{}) bool {
	contentMap, ok := item.Content.(map[string]interface{})
	if !ok {
		return false
	}

	for k, v := range attrs {
		contentValue, exists := contentMap[k]
		if !exists || contentValue != v {
			return false
		}
	}

	return true
}

// persistItem saves an item to the filesystem
func (kb *KnowledgeBase) persistItem(item KnowledgeItem) error {
	// Determine the storage path based on item type
	var dirPath string
	switch item.Type {
	case "model":
		dirPath = filepath.Join(kb.basePath, "models")
	case "pattern":
		dirPath = filepath.Join(kb.basePath, "patterns")
	case "heuristic":
		dirPath = filepath.Join(kb.basePath, "heuristics")
	default:
		dirPath = kb.basePath
	}

	// Create file path
	filePath := filepath.Join(dirPath, item.ID+".json")

	// Marshal item to JSON
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filePath, data, 0644)
}

// removeItemFromStorage deletes an item's file
func (kb *KnowledgeBase) removeItemFromStorage(id string) error {
	// Search for the file in all possible directories
	dirs := []string{
		kb.basePath,
		filepath.Join(kb.basePath, "models"),
		filepath.Join(kb.basePath, "patterns"),
		filepath.Join(kb.basePath, "heuristics"),
	}

	for _, dir := range dirs {
		filePath := filepath.Join(dir, id+".json")
		if _, err := os.Stat(filePath); err == nil {
			return os.Remove(filePath)
		}
	}

	return nil
}

// LoadAllItems loads all stored knowledge items
func (kb *KnowledgeBase) LoadAllItems() error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Check if base path exists
	if _, err := os.Stat(kb.basePath); os.IsNotExist(err) {
		return errors.New("knowledge base path does not exist")
	}

	// Walk all directories and load JSON files
	return filepath.Walk(kb.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process JSON files
		if filepath.Ext(path) != ".json" {
			return nil
		}

		// Read file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Unmarshal item
		var item KnowledgeItem
		if err := json.Unmarshal(data, &item); err != nil {
			return err
		}

		// Add to memory and index
		kb.items[item.ID] = item
		kb.indexItem(item)

		return nil
	})
}
