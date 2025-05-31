package metabolic

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager coordinates all metabolic components of the system
type Manager struct {
	glycolyticCycle *GlycolyticCycle
	dreamingModule  *DreamingModule
	lactateCycle    *LactateCycle
	mutex           sync.RWMutex
	logger          *logrus.Logger
}

// NewManager creates a new metabolic component manager
func NewManager(gc *GlycolyticCycle, dm *DreamingModule, lc *LactateCycle) *Manager {
	return &Manager{
		glycolyticCycle: gc,
		dreamingModule:  dm,
		lactateCycle:    lc,
		logger:          logrus.New(),
	}
}

// GetGlycolyticCycle returns the glycolytic cycle component
func (m *Manager) GetGlycolyticCycle() *GlycolyticCycle {
	return m.glycolyticCycle
}

// GetDreamingModule returns the dreaming module component
func (m *Manager) GetDreamingModule() *DreamingModule {
	return m.dreamingModule
}

// GetLactateCycle returns the lactate cycle component
func (m *Manager) GetLactateCycle() *LactateCycle {
	return m.lactateCycle
}

// HandleResourceEvent processes a system resource event
func (m *Manager) HandleResourceEvent(ctx context.Context, resourceType string, event string, data map[string]interface{}) error {
	// Log the event
	m.logger.WithFields(logrus.Fields{
		"resourceType": resourceType,
		"event":        event,
	}).Debug("Resource event received")

	switch resourceType {
	case "memory":
		return m.handleMemoryEvent(ctx, event, data)
	case "computation":
		return m.handleComputationEvent(ctx, event, data)
	case "storage":
		return m.handleStorageEvent(ctx, event, data)
	default:
		m.logger.WithField("resourceType", resourceType).Warn("Unknown resource type in event")
		return nil
	}
}

// handleMemoryEvent processes memory-related events
func (m *Manager) handleMemoryEvent(ctx context.Context, event string, data map[string]interface{}) error {
	switch event {
	case "pressure_high":
		// When memory pressure is high, we need to reduce memory usage
		// by clearing partial computations and limiting task parallelism

		// Reduce active dreams
		if m.dreamingModule != nil {
			// We would ideally pause or reduce dream processing
			// This would require additional API in the dreaming module
		}

		// Clear non-essential partial computations
		if m.lactateCycle != nil {
			// Get current stats
			stats := m.lactateCycle.GetStorageStats()
			activePartials, _ := stats["activePartials"].(int)

			// If we have many active partials, evict some
			// Implementation detail: We'd need to add this method to LactateCycle
			if activePartials > 10 {
				// Ideally: m.lactateCycle.EvictNonEssentialPartials(activePartials / 2)
			}
		}

		// Reduce task parallelism in glycolytic cycle
		if m.glycolyticCycle != nil {
			// Ideally: m.glycolyticCycle.ReduceParallelism(0.5) // reduce by 50%
		}

	case "pressure_normal":
		// When memory pressure returns to normal, we can resume normal operation

		// Resume normal dream processing
		if m.dreamingModule != nil {
			// Ideally: m.dreamingModule.ResumeNormalOperation()
		}

		// Resume normal task parallelism in glycolytic cycle
		if m.glycolyticCycle != nil {
			// Ideally: m.glycolyticCycle.ResumeNormalParallelism()
		}
	}

	return nil
}

// handleComputationEvent processes computation-related events
func (m *Manager) handleComputationEvent(ctx context.Context, event string, data map[string]interface{}) error {
	switch event {
	case "overload":
		// System is computationally overloaded, need to reduce load

		// Pause dreaming
		if m.dreamingModule != nil {
			// Ideally: m.dreamingModule.PauseAllDreams()
		}

		// Prioritize high-priority tasks only
		if m.glycolyticCycle != nil {
			// Ideally: m.glycolyticCycle.ProcessOnlyHighPriorityTasks()
		}

	case "underutilized":
		// System is underutilized, can increase computational load

		// Increase dream processing
		if m.dreamingModule != nil {
			// Ideally: m.dreamingModule.IncreaseDreamingActivity()
		}

		// Process more lower-priority tasks
		if m.glycolyticCycle != nil {
			// Ideally: m.glycolyticCycle.ProcessMoreTasks()
		}

	case "partial_failure":
		// Some computation failed and needs to be stored for recovery
		if data != nil && m.lactateCycle != nil {
			computationType, _ := data["type"].(string)
			computationData, _ := data["data"].(map[string]interface{})
			progress, _ := data["progress"].(float64)

			if computationType != "" && computationData != nil {
				partial := &PartialComputation{
					Type:          computationType,
					Data:          computationData,
					Progress:      progress,
					Intermediate:  make(map[string]interface{}),
					CreatedAt:     time.Now(),
					LastUpdatedAt: time.Now(),
					ExpiresAt:     time.Now().Add(24 * time.Hour),
					Recoverable:   true,
					Priority:      5, // Medium priority
				}

				// Store the partial computation
				m.lactateCycle.StorePartial(partial)
			}
		}
	}

	return nil
}

// handleStorageEvent processes storage-related events
func (m *Manager) handleStorageEvent(ctx context.Context, event string, data map[string]interface{}) error {
	switch event {
	case "capacity_warning":
		// Storage capacity is becoming a concern

		// Reduce retention of partial computations
		if m.lactateCycle != nil {
			// Ideally: m.lactateCycle.ReduceRetentionTime(0.5) // cut TTL in half
		}

		// Prioritize cleanup of low-value dreams
		if m.dreamingModule != nil {
			// Ideally: m.dreamingModule.CleanupLowValueDreams()
		}

	case "io_contention":
		// IO operations are experiencing contention

		// Reduce parallel storage operations
		if m.lactateCycle != nil {
			// Ideally: m.lactateCycle.ReduceParallelIO()
		}
	}

	return nil
}

// Shutdown gracefully stops all metabolic components
func (m *Manager) Shutdown() {
	// Shutdown in reverse order of dependency

	// First, stop the dreaming module
	if m.dreamingModule != nil {
		m.dreamingModule.Shutdown()
	}

	// Next, stop the glycolytic cycle
	// No explicit shutdown method, as it uses context cancellation

	// Finally, shut down the lactate cycle
	if m.lactateCycle != nil {
		m.lactateCycle.Shutdown()
	}

	m.logger.Info("Metabolic components shut down")
}
