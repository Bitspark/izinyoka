package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// KnowledgeItem represents a single knowledge item in the repository
type KnowledgeItem struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Tags         []string               `json:"tags,omitempty"`
	Data         map[string]interface{} `json:"data"`
	Confidence   float64                `json:"confidence"`
	CreatedAt    time.Time              `json:"createdAt"`
	LastModified time.Time              `json:"lastModified"`
	Source       string                 `json:"source,omitempty"`
}

// KnowledgeQuery defines parameters for querying the knowledge base
type KnowledgeQuery struct {
	ID       string   `json:"id,omitempty"`
	Type     string   `json:"type,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Source   string   `json:"source,omitempty"`
	MinConf  float64  `json:"minConfidence,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	SortBy   string   `json:"sortBy,omitempty"`
	SortDesc bool     `json:"sortDesc,omitempty"`
}

// KnowledgeBaseConfig contains configuration for the knowledge base
type KnowledgeBaseConfig struct {
	StoragePath   string `json:"storagePath"`
	BackupEnabled bool   `json:"backupEnabled"`
	BackupPath    string `json:"backupPath"`
	MaxItems      int    `json:"maxItems"`
}

// KnowledgeBase provides access to stored knowledge
type KnowledgeBase struct {
	config     KnowledgeBaseConfig
	items      map[string]KnowledgeItem
	mutex      sync.RWMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewKnowledgeBase creates a new knowledge repository
func NewKnowledgeBase(ctx context.Context, config KnowledgeBaseConfig) (*KnowledgeBase, error) {
	// Ensure storage directory exists
	if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create knowledge storage directory: %w", err)
	}

	// If backup is enabled, ensure backup directory exists
	if config.BackupEnabled && config.BackupPath != "" {
		if err := os.MkdirAll(config.BackupPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create knowledge backup directory: %w", err)
		}
	}

	// Create the knowledge base with cancellable context
	kbCtx, cancelFunc := context.WithCancel(ctx)
	kb := &KnowledgeBase{
		config:     config,
		items:      make(map[string]KnowledgeItem),
		ctx:        kbCtx,
		cancelFunc: cancelFunc,
	}

	// Load existing knowledge from storage
	if err := kb.loadFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load knowledge from disk: %w", err)
	}

	// Start background tasks if needed
	go kb.backgroundTasks()

	return kb, nil
}

// Query searches the knowledge base according to the provided criteria
func (kb *KnowledgeBase) Query(query *KnowledgeQuery) ([]KnowledgeItem, error) {
	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	results := []KnowledgeItem{}

	// Handle direct ID lookup
	if query.ID != "" {
		if item, exists := kb.items[query.ID]; exists {
			return []KnowledgeItem{item}, nil
		}
		return results, nil
	}

	// Filter items based on query parameters
	for _, item := range kb.items {
		if query.Type != "" && item.Type != query.Type {
			continue
		}

		if query.Source != "" && item.Source != query.Source {
			continue
		}

		if query.MinConf > 0 && item.Confidence < query.MinConf {
			continue
		}

		if len(query.Tags) > 0 && !containsAllTags(item.Tags, query.Tags) {
			continue
		}

		results = append(results, item)
	}

	// Apply limit if specified
	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results, nil
}

// Store adds or updates a knowledge item in the repository
func (kb *KnowledgeBase) Store(item KnowledgeItem) error {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()

	// Generate an ID if not provided
	if item.ID == "" {
		item.ID = generateID()
	}

	// Update timestamps
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.LastModified = now

	// Store the item
	kb.items[item.ID] = item

	// Persist to disk
	return kb.saveToDisk()
}

// Update modifies an existing knowledge item
func (kb *KnowledgeBase) Update(id string, updated KnowledgeItem) error {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()

	// Check if item exists
	existing, exists := kb.items[id]
	if !exists {
		return fmt.Errorf("knowledge item with ID %s not found", id)
	}

	// Preserve creation time and ID
	updated.CreatedAt = existing.CreatedAt
	updated.ID = id
	updated.LastModified = time.Now()

	// Update the item
	kb.items[id] = updated

	// Persist to disk
	return kb.saveToDisk()
}

// Delete removes a knowledge item from the repository
func (kb *KnowledgeBase) Delete(id string) error {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()

	// Check if item exists
	if _, exists := kb.items[id]; !exists {
		return fmt.Errorf("knowledge item with ID %s not found", id)
	}

	// Remove the item
	delete(kb.items, id)

	// Persist to disk
	return kb.saveToDisk()
}

// Shutdown cleanly stops the knowledge base operations
func (kb *KnowledgeBase) Shutdown() error {
	kb.cancelFunc()
	return kb.saveToDisk()
}

// Private helper methods

// loadFromDisk loads knowledge items from disk storage
func (kb *KnowledgeBase) loadFromDisk() error {
	files, err := filepath.Glob(filepath.Join(kb.config.StoragePath, "*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read knowledge file %s: %w", file, err)
		}

		var item KnowledgeItem
		if err := json.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("failed to parse knowledge file %s: %w", file, err)
		}

		kb.items[item.ID] = item
	}

	return nil
}

// saveToDisk persists the knowledge base to disk
func (kb *KnowledgeBase) saveToDisk() error {
	// Create backup if enabled
	if kb.config.BackupEnabled && kb.config.BackupPath != "" {
		backupTime := time.Now().Format("20060102-150405")
		backupDir := filepath.Join(kb.config.BackupPath, backupTime)

		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return fmt.Errorf("failed to create backup directory: %w", err)
		}

		// Copy current files to backup
		files, err := filepath.Glob(filepath.Join(kb.config.StoragePath, "*.json"))
		if err == nil {
			for _, file := range files {
				data, err := os.ReadFile(file)
				if err == nil {
					backupFile := filepath.Join(backupDir, filepath.Base(file))
					if err := os.WriteFile(backupFile, data, 0644); err != nil {
						return fmt.Errorf("failed to write backup file: %w", err)
					}
				}
			}
		}
	}

	// Clear storage directory
	files, err := filepath.Glob(filepath.Join(kb.config.StoragePath, "*.json"))
	if err == nil {
		for _, file := range files {
			os.Remove(file)
		}
	}

	// Save all items
	for id, item := range kb.items {
		data, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize knowledge item %s: %w", id, err)
		}

		filename := filepath.Join(kb.config.StoragePath, id+".json")
		if err := os.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("failed to write knowledge file %s: %w", filename, err)
		}
	}

	return nil
}

// backgroundTasks performs periodic maintenance operations
func (kb *KnowledgeBase) backgroundTasks() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-kb.ctx.Done():
			return
		case <-ticker.C:
			// Perform periodic backup
			if kb.config.BackupEnabled {
				kb.mutex.Lock()
				kb.saveToDisk()
				kb.mutex.Unlock()
			}
		}
	}
}

// Helper functions

// containsAllTags checks if an item contains all the specified tags
func containsAllTags(itemTags, queryTags []string) bool {
	tagMap := make(map[string]bool)
	for _, tag := range itemTags {
		tagMap[tag] = true
	}

	for _, tag := range queryTags {
		if !tagMap[tag] {
			return false
		}
	}

	return true
}

// generateID creates a unique ID for a knowledge item
func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}
