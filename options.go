// Copyright (c) 2024 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package saga

// Option defines a function type that applies a
// configuration option to a Saga instance.
type Option func(*saga)

// WithStateManager option allows the Saga to use a
// custom StateManager for tracking the state of each step,
// replacing the default in-memory state manager.
func WithStateManager(sm StateManager) Option {
	return func(s *saga) {
		s.stateManager = sm
	}
}
