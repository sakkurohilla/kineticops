package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// Manager manages agent state persistence
type Manager struct {
	stateDir string
	registry map[string]interface{}
	mutex    sync.RWMutex
}

// NewManager creates a new state manager
func NewManager(stateDir string) (*Manager, error) {
	// Create state directory if it doesn't exist
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	manager := &Manager{
		stateDir: stateDir,
		registry: make(map[string]interface{}),
	}

	// Load existing state
	if err := manager.load(); err != nil {
		// If load fails, start with empty state
		manager.registry = make(map[string]interface{})
	}

	return manager, nil
}

// GetOffset gets the last read offset for a file
func (m *Manager) GetOffset(filePath string) int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := fmt.Sprintf("filebeat.inputs.%s.offset", filePath)
	if offset, exists := m.registry[key]; exists {
		if offsetInt, ok := offset.(float64); ok {
			return int64(offsetInt)
		}
	}
	return 0
}

// SetOffset sets the last read offset for a file
func (m *Manager) SetOffset(filePath string, offset int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := fmt.Sprintf("filebeat.inputs.%s.offset", filePath)
	m.registry[key] = offset

	// Save state asynchronously
	go m.save()
}

// GetState gets a state value
func (m *Manager) GetState(key string) interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.registry[key]
}

// SetState sets a state value
func (m *Manager) SetState(key string, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.registry[key] = value

	// Save state asynchronously
	go m.save()
}

// load loads state from disk
func (m *Manager) load() error {
	registryPath := filepath.Join(m.stateDir, "registry.json")
	
	data, err := ioutil.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing state
		}
		return fmt.Errorf("failed to read registry: %w", err)
	}

	if err := json.Unmarshal(data, &m.registry); err != nil {
		return fmt.Errorf("failed to unmarshal registry: %w", err)
	}

	return nil
}

// save saves state to disk
func (m *Manager) save() error {
	m.mutex.RLock()
	registryCopy := make(map[string]interface{})
	for k, v := range m.registry {
		registryCopy[k] = v
	}
	m.mutex.RUnlock()

	data, err := json.MarshalIndent(registryCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	registryPath := filepath.Join(m.stateDir, "registry.json")
	tempPath := registryPath + ".tmp"

	// Write to temp file first
	if err := ioutil.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp registry: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, registryPath); err != nil {
		return fmt.Errorf("failed to rename registry: %w", err)
	}

	return nil
}

// Close closes the state manager
func (m *Manager) Close() error {
	return m.save()
}