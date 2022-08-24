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

// Policy is a policy for retrying an operation.
type Policy interface {
	// Next returns the backoff duration to wait before the next attempt
	// and a bool indicating if a retry should be attempted.
	Next(start, now time.Time, attempt int) (backoff time.Duration, allow bool)
}

// A PermanentError signals that an operation is not retriable.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string { return e.Err.Error() }

func (e *PermanentError) Unwrap() error { return e.Err }

// Do executes the retriable function according to the given Policy.
//
// If fn returns a PermanentError, its inner error will be returned without follow-up retry attempts.
//
// If the context has a deadline before the next retry attempt would be scheduled it will return the
// last error without waiting.
func Do(ctx context.Context, p Policy, fn func() error) error {
	var (
		pe *PermanentError
		t  *time.Timer
	)
	start := time.Now()
	deadline, hasDeadline := ctx.Deadline()
	for retry := 1; ; retry++ {
		err := fn()
		if err == nil {
			return nil
		}
		if errors.As(err, &pe) {
			return pe.Err
		}

		now := time.Now()
		next, ok := p.Next(start, now, retry)
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
	if !t.Stop() {
		<-t.C
	}
	t.Reset(d)
}
