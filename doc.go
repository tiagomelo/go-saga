// Copyright (c) 2024 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

// Package saga provides an implementation of the Saga pattern
// for managing distributed transactions.
// A Saga is a sequence of steps that either all succeed or all
// are compensated (rolled back) in case of failure.
// This package supports in-memory state management by default
// but can be extended to use external state storage.
package saga
