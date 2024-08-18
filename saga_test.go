// Copyright (c) 2024 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type sampleState struct {
	X int
}

func TestExecute(t *testing.T) {
	ss := &sampleState{}
	testCases := []struct {
		name             string
		steps            []Step
		mockStateManager func() StateManager
		expectedValue    int
		expectedError    error
	}{
		{
			name: "two steps, both succeed",
			steps: []Step{
				NewStep("step1",
					func(ctx context.Context) error {
						ss.X = 1
						return nil
					},
					func(ctx context.Context) error {
						ss.X = 0
						return nil
					},
				),
				NewStep("step2",
					func(ctx context.Context) error {
						ss.X = 2
						return nil
					},
					func(ctx context.Context) error {
						ss.X -= 1
						return nil
					},
				),
			},
			expectedValue: 2,
		},
		{
			name: "two steps, first fails, compensate",
			steps: []Step{
				NewStep("step1",
					func(ctx context.Context) error {
						return errors.New("step1 error")
					},
					func(ctx context.Context) error {
						ss.X = 0
						return nil
					},
				),
				NewStep("step2",
					func(ctx context.Context) error {
						ss.X = 2
						return nil
					},
					func(ctx context.Context) error {
						ss.X -= 1
						return nil
					},
				),
			},
			expectedValue: 0,
			expectedError: errors.New("executing step step1: step1 error"),
		},
		{
			name: "two steps, second fails, compensate",
			steps: []Step{
				NewStep("step1",
					func(ctx context.Context) error {
						ss.X = 1
						return nil
					},
					func(ctx context.Context) error {
						ss.X = 0
						return nil
					},
				),
				NewStep("step2",
					func(ctx context.Context) error {
						return errors.New("step2 error")
					},
					func(ctx context.Context) error {
						ss.X -= 1
						return nil
					},
				),
			},
			expectedValue: 0,
			expectedError: errors.New("executing step step2: step2 error"),
		},
		{
			name: "two steps, second fails, failed to compensate",
			steps: []Step{
				NewStep("step1",
					func(ctx context.Context) error {
						ss.X = 1
						return nil
					},
					func(ctx context.Context) error {
						ss.X = 0
						return nil
					},
				),
				NewStep("step2",
					func(ctx context.Context) error {
						return errors.New("step2 error")
					},
					func(ctx context.Context) error {
						return errors.New("step2 compensate error")
					},
				),
			},
			expectedValue: 0,
			expectedError: errors.New("compensating after failure in step step2: step2 error: compensation failed with errors: [step2 compensate error]"),
		},
		{
			name: "error when getting step state",
			mockStateManager: func() StateManager {
				return &mockStateManager{
					stepStateErr: errors.New("step state error"),
				}
			},
			steps: []Step{
				NewStep("step1",
					func(ctx context.Context) error {
						ss.X = 1
						return nil
					},
					func(ctx context.Context) error {
						ss.X = 0
						return nil
					},
				),
				NewStep("step2",
					func(ctx context.Context) error {
						ss.X = 2
						return nil
					},
					func(ctx context.Context) error {
						ss.X -= 1
						return nil
					},
				),
			},
			expectedError: errors.New("retrieving state for step step1: step state error"),
		},
		{
			name: "two steps, first fails, error when setting step state",
			steps: []Step{
				NewStep("step1",
					func(ctx context.Context) error {
						return errors.New("step1 error")
					},
					func(ctx context.Context) error {
						ss.X = 0
						return nil
					},
				),
				NewStep("step2",
					func(ctx context.Context) error {
						ss.X = 2
						return nil
					},
					func(ctx context.Context) error {
						ss.X -= 1
						return nil
					},
				),
			},
			mockStateManager: func() StateManager {
				return &mockStateManager{
					setStepStateErr: errors.New("set step state error"),
				}
			},
			expectedValue: 0,
			expectedError: errors.New("setting state for step step1: set step state error"),
		},
		{
			name: "error when setting step state",
			mockStateManager: func() StateManager {
				return &mockStateManager{
					setStepStateErr: errors.New("set step state error"),
				}
			},
			steps: []Step{
				NewStep("step1",
					func(ctx context.Context) error {
						ss.X = 1
						return nil
					},
					func(ctx context.Context) error {
						ss.X = 0
						return nil
					},
				),
				NewStep("step2",
					func(ctx context.Context) error {
						ss.X = 2
						return nil
					},
					func(ctx context.Context) error {
						ss.X -= 1
						return nil
					},
				),
			},
			expectedValue: 1,
			expectedError: errors.New("setting state for step step1: set step state error"),
		},
	}
	for _, tc := range testCases {
		defer func() {
			ss.X = 0
		}()
		t.Run(tc.name, func(t *testing.T) {
			var saga Saga
			if tc.mockStateManager != nil {
				saga = New(WithStateManager(tc.mockStateManager()))
			} else {
				saga = New()
			}
			for _, step := range tc.steps {
				saga.AddStep(step)
			}
			err := saga.Execute(context.Background())
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf("expected no error, got %v", err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
				require.Equal(t, tc.expectedValue, ss.X)
			} else {
				if tc.expectedError != nil {
					t.Fatalf("expected error %v, got nil", tc.expectedError)
				}
				require.Equal(t, tc.expectedValue, ss.X)
			}
		})
	}
}

func TestSaga_RetryPartialExecution(t *testing.T) {
	ss := &sampleState{}
	step1 := NewStep("step1",
		func(ctx context.Context) error {
			ss.X = 1
			return nil
		},
		func(ctx context.Context) error {
			// let's not compensate for this test.
			return nil
		},
	)
	step2 := NewStep("step2",
		func(ctx context.Context) error {
			return errors.New("step2 error")
		},
		func(ctx context.Context) error {
			// let's not compensate for this test.
			return nil
		},
	)

	saga := New()
	saga.AddStep(step1)
	saga.AddStep(step2)

	expectedError := errors.New("executing step step2: step2 error")
	err := saga.Execute(context.Background())

	// Step 1 should be executed successfully.
	// Because we don't have a compensation for neither steps, x should remain 1.
	require.NotNil(t, err)
	require.Equal(t, expectedError.Error(), err.Error())
	require.Equal(t, 1, ss.X)

	// Let's retry the saga execution.
	// Step 1 should be skipped, step 2 should be executed.
	// Step 2 should fail again.
	// Now we're assigning x to be 5, and since Step 1 is skipped, x will keep the new value.
	ss.X = 5
	err = saga.Execute(context.Background())
	require.NotNil(t, err)
	require.Equal(t, expectedError.Error(), err.Error())
	require.Equal(t, 5, ss.X)
}

type mockStateManager struct {
	setStepStateErr error
	stepState       bool
	stepStateErr    error
}

func (m *mockStateManager) SetStepState(stepIndex int, success bool) error {
	return m.setStepStateErr
}

func (m *mockStateManager) StepState(stepIndex int) (bool, error) {
	return m.stepState, m.stepStateErr
}
