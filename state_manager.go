// Copyright (c) 2024 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package saga

// StateManager defines the interface for managing
// the state of each step in a Saga.
// Implementations of this interface can store state in-memory,
// in a database, or any other storage mechanism.
type StateManager interface {
	// SetStepState records the completion state (success or failure)
	// of a specific step in the Saga.
	// stepIndex indicates the step's position in the Saga, and success
	// indicates whether the step completed successfully.
	SetStepState(stepIndex int, success bool) error

	// StepState retrieves the completion state of a specific step in the Saga.
	// It returns true if the step was successfully completed,
	// false otherwise, and any error encountered during retrieval.
	StepState(stepIndex int) (bool, error)
}
