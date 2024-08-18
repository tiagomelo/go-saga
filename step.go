// Copyright (c) 2024 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package saga

import "context"

// Step defines the interface for a step in the Saga pattern.
type Step interface {
	// ExecuteForward executes the main action of the step.
	// It takes a context for managing execution and returns
	// an error if the action fails.
	ExecuteForward(ctx context.Context) error

	// ExecuteCompensate executes the compensation action of the step.
	// This is called when the Saga needs to roll back
	// previous steps due to a failure.
	ExecuteCompensate(ctx context.Context) error

	// Name returns the name of the step, which can be used for
	// logging or debugging purposes.
	Name() string
}

// step is the concrete implementation of the Step interface.
type step struct {
	name       string
	forward    func(ctx context.Context) error
	compensate func(ctx context.Context) error
}

// NewStep creates a new Step instance with the provided name,
// forward action, and compensation action.
func NewStep(name string, forward, compensate func(ctx context.Context) error) Step {
	return &step{
		name:       name,
		forward:    forward,
		compensate: compensate,
	}
}

func (s *step) Name() string {
	return s.name
}

func (s *step) ExecuteForward(ctx context.Context) error {
	return s.forward(ctx)
}

func (s *step) ExecuteCompensate(ctx context.Context) error {
	return s.compensate(ctx)
}
