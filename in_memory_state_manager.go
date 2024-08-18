// Copyright (c) 2024 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package saga

import "sync"

// InMemoryStateManager is an implementation of the
// StateManager interface that stores the state of each step
// in memory using a map.
type InMemoryStateManager struct {
	state map[int]bool
	mu    sync.RWMutex
}

// NewInMemoryStateManager creates a new instance of InMemoryStateManager.
func NewInMemoryStateManager() *InMemoryStateManager {
	return &InMemoryStateManager{
		state: make(map[int]bool),
	}
}

func (m *InMemoryStateManager) SetStepState(stepIndex int, success bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state[stepIndex] = success
	return nil
}

func (m *InMemoryStateManager) StepState(stepIndex int) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, exists := m.state[stepIndex]
	if !exists {
		return false, nil
	}
	return state, nil
}
