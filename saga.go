// Copyright (c) 2024 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package saga

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

// Saga defines the interface for a Saga pattern implementation.
type Saga interface {
	// AddStep adds a new step to the Saga. Each step should define
	// its forward and compensation actions.
	AddStep(step Step)

	// Execute runs the Saga, executing each step in sequence.
	// If any step fails, the Saga triggers compensation
	// for all previously successful steps.
	Execute(ctx context.Context) error

	// Compensate rolls back all successfully executed steps if any
	// subsequent step fails during the Saga's execution.
	Compensate(ctx context.Context) error
}

// saga is the concrete implementation of the Saga interface.
type saga struct {
	steps        []Step
	currentStep  int
	stateManager StateManager
	mu           sync.Mutex
}

// new creates a new saga instance with the given options.
// It initializes the saga with an in-memory state manager
// by default, but this can be overridden with the provided options.
func new(options []Option) Saga {
	s := &saga{
		steps:        []Step{},
		stateManager: NewInMemoryStateManager(),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// New constructs a new Saga with the specified options.
// Options can include custom state managers or other configurations.
func New(options ...Option) Saga {
	return new(options)
}

func (s *saga) AddStep(step Step) {
	s.steps = append(s.steps, step)
}

func (s *saga) Execute(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for s.currentStep = 0; s.currentStep < len(s.steps); s.currentStep++ {
		step := s.steps[s.currentStep]

		// Skip steps that have already been completed.
		stepCompleted, err := s.stateManager.StepState(s.currentStep)
		if err != nil {
			return errors.Wrapf(err, "retrieving state for step %s", step.Name())
		}
		if stepCompleted {
			continue
		}

		// Try executing the current step.
		if err := step.ExecuteForward(ctx); err != nil {
			// Mark this step as failed.
			if err := s.stateManager.SetStepState(s.currentStep, false); err != nil {
				return errors.Wrapf(err, "setting state for step %s", step.Name())
			}

			// Trigger compensation for all previously successful steps.
			if errComp := s.Compensate(ctx); errComp != nil {
				return errors.Wrapf(errComp, "compensating after failure in step %s: %v", step.Name(), err)
			}

			// Return the original error.
			return errors.Wrapf(err, "executing step %s", step.Name())
		}

		// Mark this step as successfully completed.
		if err := s.stateManager.SetStepState(s.currentStep, true); err != nil {
			return errors.Wrapf(err, "setting state for step %s", step.Name())
		}
	}

	return nil
}

func (s *saga) Compensate(ctx context.Context) error {
	var compensationErrors []error

	// Compensate from the current step backwards.
	for i := s.currentStep; i >= 0; i-- {
		step := s.steps[i]
		if err := step.ExecuteCompensate(ctx); err != nil {
			compensationErrors = append(compensationErrors, err)
		}
	}

	if len(compensationErrors) > 0 {
		// Aggregate all compensation errors into a single error.
		return errors.Errorf("compensation failed with errors: %v", compensationErrors)
	}

	return nil
}
