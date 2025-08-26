// Package graph - State management and persistence implementation
// 包 graph - 状态管理和持久化实现
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ================================
// In-Memory State Manager 内存状态管理器
// ================================

// MemoryStateManager provides in-memory state management.
// MemoryStateManager 提供内存状态管理。
type MemoryStateManager struct {
	// states stores the states in memory.
	states map[string]*State

	// lock protects concurrent access to states.
	lock sync.RWMutex

	// maxStates is the maximum number of states to keep in memory.
	maxStates int

	// cleanup tracks state access times for LRU eviction.
	cleanup map[string]time.Time
}

// NewMemoryStateManager creates a new in-memory state manager.
// NewMemoryStateManager 创建一个新的内存状态管理器。
func NewMemoryStateManager(maxStates int) *MemoryStateManager {
	if maxStates <= 0 {
		maxStates = 1000 // Default maximum
	}

	return &MemoryStateManager{
		states:    make(map[string]*State),
		maxStates: maxStates,
		cleanup:   make(map[string]time.Time),
	}
}

// Save implements the StateManager interface.
// Save 实现 StateManager 接口。
func (msm *MemoryStateManager) Save(ctx context.Context, state *State) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}
	if state.ID == "" {
		return fmt.Errorf("state ID cannot be empty")
	}

	msm.lock.Lock()
	defer msm.lock.Unlock()

	// Store the state
	msm.states[state.ID] = state.Clone()
	msm.cleanup[state.ID] = time.Now()

	// Perform cleanup if necessary
	if len(msm.states) > msm.maxStates {
		msm.evictOldest()
	}

	return nil
}

// Load implements the StateManager interface.
// Load 实现 StateManager 接口。
func (msm *MemoryStateManager) Load(ctx context.Context, id string) (*State, error) {
	if id == "" {
		return nil, fmt.Errorf("state ID cannot be empty")
	}

	msm.lock.Lock()
	defer msm.lock.Unlock()

	state, exists := msm.states[id]
	if !exists {
		return nil, fmt.Errorf("state %s not found", id)
	}

	// Update access time
	msm.cleanup[id] = time.Now()

	return state.Clone(), nil
}

// Delete implements the StateManager interface.
// Delete 实现 StateManager 接口。
func (msm *MemoryStateManager) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("state ID cannot be empty")
	}

	msm.lock.Lock()
	defer msm.lock.Unlock()

	delete(msm.states, id)
	delete(msm.cleanup, id)

	return nil
}

// evictOldest removes the oldest accessed state.
// evictOldest 移除最久未访问的状态。
func (msm *MemoryStateManager) evictOldest() {
	var oldestID string
	var oldestTime time.Time

	for id, accessTime := range msm.cleanup {
		if oldestID == "" || accessTime.Before(oldestTime) {
			oldestID = id
			oldestTime = accessTime
		}
	}

	if oldestID != "" {
		delete(msm.states, oldestID)
		delete(msm.cleanup, oldestID)
	}
}

// GetStats returns statistics about the state manager.
// GetStats 返回状态管理器的统计信息。
func (msm *MemoryStateManager) GetStats() map[string]interface{} {
	msm.lock.RLock()
	defer msm.lock.RUnlock()

	return map[string]interface{}{
		"total_states": len(msm.states),
		"max_states":   msm.maxStates,
	}
}

// Clear removes all states from memory.
// Clear 从内存中移除所有状态。
func (msm *MemoryStateManager) Clear() {
	msm.lock.Lock()
	defer msm.lock.Unlock()

	msm.states = make(map[string]*State)
	msm.cleanup = make(map[string]time.Time)
}

// ================================
// File-Based State Manager 基于文件的状态管理器
// ================================

// FileStateManager provides file-based state persistence.
// FileStateManager 提供基于文件的状态持久化。
type FileStateManager struct {
	// baseDir is the directory where state files are stored.
	baseDir string

	// lock protects concurrent access to files.
	lock sync.RWMutex

	// compression indicates whether to compress state files.
	compression bool
}

// NewFileStateManager creates a new file-based state manager.
// NewFileStateManager 创建一个新的基于文件的状态管理器。
func NewFileStateManager(baseDir string) (*FileStateManager, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FileStateManager{
		baseDir:     baseDir,
		compression: false,
	}, nil
}

// Save implements the StateManager interface.
// Save 实现 StateManager 接口。
func (fsm *FileStateManager) Save(ctx context.Context, state *State) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}
	if state.ID == "" {
		return fmt.Errorf("state ID cannot be empty")
	}

	fsm.lock.Lock()
	defer fsm.lock.Unlock()

	// Serialize state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	// Write to file
	filename := fsm.getFilename(state.ID)
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// Load implements the StateManager interface.
// Load 实现 StateManager 接口。
func (fsm *FileStateManager) Load(ctx context.Context, id string) (*State, error) {
	if id == "" {
		return nil, fmt.Errorf("state ID cannot be empty")
	}

	fsm.lock.RLock()
	defer fsm.lock.RUnlock()

	// Read file
	filename := fsm.getFilename(id)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("state %s not found", id)
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Deserialize state
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to deserialize state: %w", err)
	}

	return &state, nil
}

// Delete implements the StateManager interface.
// Delete 实现 StateManager 接口。
func (fsm *FileStateManager) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("state ID cannot be empty")
	}

	fsm.lock.Lock()
	defer fsm.lock.Unlock()

	filename := fsm.getFilename(id)
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete state file: %w", err)
	}

	return nil
}

// getFilename returns the filename for a given state ID.
// getFilename 返回给定状态ID的文件名。
func (fsm *FileStateManager) getFilename(id string) string {
	return filepath.Join(fsm.baseDir, fmt.Sprintf("state_%s.json", id))
}

// ListStates returns a list of all stored state IDs.
// ListStates 返回所有存储的状态ID列表。
func (fsm *FileStateManager) ListStates() ([]string, error) {
	fsm.lock.RLock()
	defer fsm.lock.RUnlock()

	files, err := ioutil.ReadDir(fsm.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var stateIDs []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			// Extract state ID from filename
			name := file.Name()
			if len(name) > 11 && name[:6] == "state_" {
				stateID := name[6 : len(name)-5] // Remove "state_" prefix and ".json" suffix
				stateIDs = append(stateIDs, stateID)
			}
		}
	}

	return stateIDs, nil
}

// Cleanup removes old state files based on age.
// Cleanup 根据年龄移除旧的状态文件。
func (fsm *FileStateManager) Cleanup(maxAge time.Duration) error {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()

	files, err := ioutil.ReadDir(fsm.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			if file.ModTime().Before(cutoff) {
				filename := filepath.Join(fsm.baseDir, file.Name())
				os.Remove(filename) // Ignore errors
			}
		}
	}

	return nil
}

// ================================
// Redis State Manager Redis状态管理器
// ================================

// RedisStateManager provides Redis-based state persistence.
// Note: This is a placeholder implementation. In a real scenario,
// you would use a Redis client library like go-redis.
// RedisStateManager 提供基于Redis的状态持久化。
// 注意：这是一个占位符实现。在实际场景中，
// 你会使用Redis客户端库，如go-redis。
type RedisStateManager struct {
	// client would be a Redis client instance
	// client 应该是Redis客户端实例
	client interface{}

	// keyPrefix is the prefix for Redis keys
	keyPrefix string

	// ttl is the time-to-live for stored states
	ttl time.Duration
}

// NewRedisStateManager creates a new Redis-based state manager.
// NewRedisStateManager 创建一个新的基于Redis的状态管理器。
func NewRedisStateManager(client interface{}, keyPrefix string, ttl time.Duration) *RedisStateManager {
	return &RedisStateManager{
		client:    client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}
}

// Save implements the StateManager interface.
// Save 实现 StateManager 接口。
func (rsm *RedisStateManager) Save(ctx context.Context, state *State) error {
	// This is a placeholder implementation
	// In a real implementation, you would:
	// 1. Serialize the state to JSON
	// 2. Store it in Redis with the appropriate key and TTL
	return fmt.Errorf("Redis state manager not implemented")
}

// Load implements the StateManager interface.
// Load 实现 StateManager 接口。
func (rsm *RedisStateManager) Load(ctx context.Context, id string) (*State, error) {
	// This is a placeholder implementation
	// In a real implementation, you would:
	// 1. Retrieve the state from Redis using the key
	// 2. Deserialize it from JSON
	return nil, fmt.Errorf("Redis state manager not implemented")
}

// Delete implements the StateManager interface.
// Delete 实现 StateManager 接口。
func (rsm *RedisStateManager) Delete(ctx context.Context, id string) error {
	// This is a placeholder implementation
	// In a real implementation, you would delete the key from Redis
	return fmt.Errorf("Redis state manager not implemented")
}

// ================================
// Composite State Manager 复合状态管理器
// ================================

// CompositeStateManager combines multiple state managers with different strategies.
// CompositeStateManager 组合多个状态管理器，使用不同的策略。
type CompositeStateManager struct {
	// primary is the primary state manager
	primary StateManager

	// secondary is the secondary state manager (fallback)
	secondary StateManager

	// writeThrough indicates whether to write to both managers
	writeThrough bool

	// readStrategy determines the read strategy
	readStrategy ReadStrategy
}

// ReadStrategy determines how to read from multiple state managers.
// ReadStrategy 确定如何从多个状态管理器读取。
type ReadStrategy int

const (
	// ReadPrimaryFirst tries primary first, then secondary
	ReadPrimaryFirst ReadStrategy = iota
	// ReadSecondaryFirst tries secondary first, then primary
	ReadSecondaryFirst
	// ReadPrimaryOnly only reads from primary
	ReadPrimaryOnly
	// ReadSecondaryOnly only reads from secondary
	ReadSecondaryOnly
)

// NewCompositeStateManager creates a new composite state manager.
// NewCompositeStateManager 创建一个新的复合状态管理器。
func NewCompositeStateManager(primary, secondary StateManager, writeThrough bool, readStrategy ReadStrategy) *CompositeStateManager {
	return &CompositeStateManager{
		primary:      primary,
		secondary:    secondary,
		writeThrough: writeThrough,
		readStrategy: readStrategy,
	}
}

// Save implements the StateManager interface.
// Save 实现 StateManager 接口。
func (csm *CompositeStateManager) Save(ctx context.Context, state *State) error {
	// Always save to primary
	if err := csm.primary.Save(ctx, state); err != nil {
		return fmt.Errorf("failed to save to primary: %w", err)
	}

	// Save to secondary if write-through is enabled
	if csm.writeThrough && csm.secondary != nil {
		if err := csm.secondary.Save(ctx, state); err != nil {
			// Log error but don't fail the operation
			// In a real implementation, you might want to log this error
		}
	}

	return nil
}

// Load implements the StateManager interface.
// Load 实现 StateManager 接口。
func (csm *CompositeStateManager) Load(ctx context.Context, id string) (*State, error) {
	switch csm.readStrategy {
	case ReadPrimaryFirst:
		if state, err := csm.primary.Load(ctx, id); err == nil {
			return state, nil
		}
		if csm.secondary != nil {
			return csm.secondary.Load(ctx, id)
		}
		return nil, fmt.Errorf("state %s not found", id)

	case ReadSecondaryFirst:
		if csm.secondary != nil {
			if state, err := csm.secondary.Load(ctx, id); err == nil {
				return state, nil
			}
		}
		return csm.primary.Load(ctx, id)

	case ReadPrimaryOnly:
		return csm.primary.Load(ctx, id)

	case ReadSecondaryOnly:
		if csm.secondary != nil {
			return csm.secondary.Load(ctx, id)
		}
		return nil, fmt.Errorf("secondary state manager not available")

	default:
		return csm.primary.Load(ctx, id)
	}
}

// Delete implements the StateManager interface.
// Delete 实现 StateManager 接口。
func (csm *CompositeStateManager) Delete(ctx context.Context, id string) error {
	// Delete from primary
	if err := csm.primary.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete from primary: %w", err)
	}

	// Delete from secondary if available
	if csm.secondary != nil {
		csm.secondary.Delete(ctx, id) // Ignore errors
	}

	return nil
}

// ================================
// State Checkpoint Manager 状态检查点管理器
// ================================

// CheckpointManager manages state checkpoints for recovery.
// CheckpointManager 管理状态检查点用于恢复。
type CheckpointManager struct {
	// stateManager is the underlying state manager
	stateManager StateManager

	// checkpointInterval is how often to create checkpoints
	checkpointInterval time.Duration

	// maxCheckpoints is the maximum number of checkpoints to keep
	maxCheckpoints int

	// checkpoints tracks checkpoint metadata
	checkpoints map[string][]CheckpointInfo

	// lock protects concurrent access
	lock sync.RWMutex
}

// CheckpointInfo contains information about a checkpoint.
// CheckpointInfo 包含检查点的信息。
type CheckpointInfo struct {
	// ID is the checkpoint ID
	ID string `json:"id"`

	// StateID is the original state ID
	StateID string `json:"state_id"`

	// Timestamp is when the checkpoint was created
	Timestamp time.Time `json:"timestamp"`

	// NodeID is the node that was executing when checkpoint was created
	NodeID string `json:"node_id"`

	// StepCount is the number of steps executed
	StepCount int `json:"step_count"`
}

// NewCheckpointManager creates a new checkpoint manager.
// NewCheckpointManager 创建一个新的检查点管理器。
func NewCheckpointManager(stateManager StateManager, checkpointInterval time.Duration, maxCheckpoints int) *CheckpointManager {
	return &CheckpointManager{
		stateManager:       stateManager,
		checkpointInterval: checkpointInterval,
		maxCheckpoints:     maxCheckpoints,
		checkpoints:        make(map[string][]CheckpointInfo),
	}
}

// CreateCheckpoint creates a checkpoint of the current state.
// CreateCheckpoint 创建当前状态的检查点。
func (cm *CheckpointManager) CreateCheckpoint(ctx context.Context, state *State) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	cm.lock.Lock()
	defer cm.lock.Unlock()

	// Create checkpoint ID
	checkpointID := fmt.Sprintf("%s_checkpoint_%d", state.ID, time.Now().UnixNano())

	// Create checkpoint state
	checkpointState := state.Clone()
	checkpointState.ID = checkpointID

	// Save checkpoint
	if err := cm.stateManager.Save(ctx, checkpointState); err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}

	// Update checkpoint metadata
	checkpointInfo := CheckpointInfo{
		ID:        checkpointID,
		StateID:   state.ID,
		Timestamp: time.Now(),
		NodeID:    state.CurrentNode,
		StepCount: len(state.History),
	}

	checkpoints := cm.checkpoints[state.ID]
	checkpoints = append(checkpoints, checkpointInfo)

	// Remove old checkpoints if we exceed the maximum
	if len(checkpoints) > cm.maxCheckpoints {
		// Remove the oldest checkpoint
		oldestCheckpoint := checkpoints[0]
		cm.stateManager.Delete(ctx, oldestCheckpoint.ID)
		checkpoints = checkpoints[1:]
	}

	cm.checkpoints[state.ID] = checkpoints

	return nil
}

// RestoreFromCheckpoint restores state from a checkpoint.
// RestoreFromCheckpoint 从检查点恢复状态。
func (cm *CheckpointManager) RestoreFromCheckpoint(ctx context.Context, stateID string, checkpointIndex int) (*State, error) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	checkpoints, exists := cm.checkpoints[stateID]
	if !exists {
		return nil, fmt.Errorf("no checkpoints found for state %s", stateID)
	}

	if checkpointIndex < 0 || checkpointIndex >= len(checkpoints) {
		return nil, fmt.Errorf("invalid checkpoint index %d", checkpointIndex)
	}

	checkpointInfo := checkpoints[checkpointIndex]
	checkpointState, err := cm.stateManager.Load(ctx, checkpointInfo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	// Restore original state ID
	checkpointState.ID = stateID

	return checkpointState, nil
}

// GetCheckpoints returns all checkpoints for a state.
// GetCheckpoints 返回状态的所有检查点。
func (cm *CheckpointManager) GetCheckpoints(stateID string) []CheckpointInfo {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	checkpoints, exists := cm.checkpoints[stateID]
	if !exists {
		return nil
	}

	// Return a copy
	result := make([]CheckpointInfo, len(checkpoints))
	copy(result, checkpoints)
	return result
}

// CleanupCheckpoints removes old checkpoints.
// CleanupCheckpoints 移除旧的检查点。
func (cm *CheckpointManager) CleanupCheckpoints(ctx context.Context, maxAge time.Duration) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for stateID, checkpoints := range cm.checkpoints {
		var validCheckpoints []CheckpointInfo

		for _, checkpoint := range checkpoints {
			if checkpoint.Timestamp.After(cutoff) {
				validCheckpoints = append(validCheckpoints, checkpoint)
			} else {
				// Delete old checkpoint
				cm.stateManager.Delete(ctx, checkpoint.ID)
			}
		}

		if len(validCheckpoints) > 0 {
			cm.checkpoints[stateID] = validCheckpoints
		} else {
			delete(cm.checkpoints, stateID)
		}
	}

	return nil
}