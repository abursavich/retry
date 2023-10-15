// SPDX-License-Identifier: MIT
//
// Copyright 2022 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

// Package retry provides backoff algorithms for retryable processes.
package retry

import (
	"context"
	"errors"
	"time"
)

// Policy is a policy for retrying a function.
type Policy interface {
	// Next returns the backoff duration to wait before the next attempt
	// and a bool indicating if a retry should be attempted.
	Next(err error, start, now time.Time, attempt int) (backoff time.Duration, retry bool)
}

// NewPermanentError returns a new error that wraps err and signals that the function should not be retried.
// If err is nil or is a permanent error already, it's return unchanged.
//
// It's provided as a convenience for simple use cases, but in complex use cases it's probably better
// to implement permanent error detection as a custom Policy layer.
func NewPermanentError(err error) error {
	if err == nil || isPermErr(err) {
		return err
	}
	return &permanentError{err}
}

var permErr error = &permanentError{}

func isPermErr(err error) bool { return errors.Is(err, permErr) }

type permanentError struct{ err error }

func (e *permanentError) Error() string { return e.err.Error() }

func (e *permanentError) Unwrap() error { return e.err }

func (e *permanentError) Is(err error) bool { return err == e || err == permErr }

// Do executes the retriable function according to the given policy.
//
// If fn returns a permanent error, the error will be returned without additional retry attempts.
//
// If ctx has a deadline before the next retry attempt would be scheduled it will return the
// last error without waiting for the deadline.
func Do(ctx context.Context, policy Policy, fn func() error) error {
	var t *time.Timer
	start := time.Now()
	deadline, hasDeadline := ctx.Deadline()
	for retry := 1; ; retry++ {
		err := fn()
		if err == nil || isPermErr(err) {
			// We don't return a permanentError's inner error because the permanentError
			// may be in the middle of a chain of errors and we don't want to drop any
			// errors that are wrapping it.
			return err
		}

		now := time.Now()
		next, ok := policy.Next(err, start, now, retry)
		if !ok {
			return err
		}
		if hasDeadline && deadline.Before(time.Now().Add(next)) {
			return err // TODO: context.DeadlineExceeded ?
		}

		if t == nil {
			t = time.NewTimer(next)
		} else {
			resetTimer(t, next)
		}
		select {
		case <-ctx.Done():
			t.Stop()
			return err // TODO: ctx.Err() ?
		case <-t.C:
		}
	}
}

func resetTimer(t *time.Timer, d time.Duration) {
	t.Stop()
	select {
	case <-t.C:
	default:
	}
	t.Reset(d)
}
