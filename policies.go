// SPDX-License-Identifier: MIT
//
// Copyright 2022 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package retry

import (
	"math"
	"math/rand"
	"time"
)

// Default policy values.
const (
	DefaultMinBackoff   = 150 * time.Millisecond
	DefaultMaxBackoff   = 15 * time.Second
	DefaultGrowthFactor = 1.5
	DefaultJitterFactor = 0.5
)

var defaultPolicy = WithDefaultRandomJitter(DefaultExponentialBackoff())

// DefaultBackoff returns a Policy using the default exponential backoff
// with the default random jitter.
//
// This results in the following behavior:
//
//	Attempt         Backoff                  Total
//	      1     [0.075s,  0.225s]     [ 0.075s,   0.225s]
//	      2     [0.113s,  0.338s]     [ 0.188s,   0.562s]
//	      3     [0.169s,  0.506s]     [ 0.356s,   1.069s]
//	      4     [0.253s,  0.759s]     [ 0.609s,   1.828s]
//	      5     [0.380s,  1.139s]     [ 0.989s,   2.967s]
//	      6     [0.570s,  1.709s]     [ 1.559s,   4.676s]
//	      7     [0.854s,  2.563s]     [ 2.413s,   7.239s]
//	      8     [1.281s,  3.844s]     [ 3.694s,  11.083s]
//	      9     [1.922s,  5.767s]     [ 5.617s,  16.850s]
//	     10     [2.883s,  8.650s]     [ 8.500s,  25.499s]
//	     11     [4.325s, 12.975s]     [12.825s,  38.474s]
//	     12     [6.487s, 19.462s]     [19.312s,  57.936s]
//	     13     [7.500s, 22.500s]     [26.812s,  80.436s]
//	     14     [7.500s, 22.500s]     [34.312s, 102.936s]
//	     15     [7.500s, 22.500s]     [41.812s, 125.436s]
//	    ...            ...                    ...
func DefaultBackoff() Policy {
	return defaultPolicy
}

var never = WithMaxRetries(nil, 0)

// Never returns a Policy that doesn't allow any retry attempts.
func Never() Policy {
	return never
}

var immediately = ConstantBackoff(0)

// Immediately returns a Policy that retries with no backoff.
func Immediately() Policy {
	return immediately
}

// ConstantBackoff returns a Policy that uses a constant backoff duration.
func ConstantBackoff(backoff time.Duration) Policy {
	return constantBackoff{backoff}
}

type constantBackoff struct {
	backoff time.Duration
}

func (p constantBackoff) Next(start, now time.Time, attempt int) (time.Duration, bool) {
	return time.Duration(p.backoff), true
}

// ExponentialBackoff returns a Policy in which the backoff grows exponentially.
// The backoff will start at the min and will be scaled by the growth factor
// for each successive attempt until it's capped at the max.
func ExponentialBackoff(min, max time.Duration, factor float64) Policy {
	if min <= 0 {
		min = DefaultMinBackoff
	}
	if max <= 0 {
		max = DefaultMaxBackoff
	}
	if factor <= 1 {
		factor = DefaultGrowthFactor
	}
	return &exponentialBackoff{
		min:    min,
		max:    max,
		factor: factor,
	}
}

// DefaultExponentialBackoff returns an ExponentialBackoff Policy with the default values
// of min 150ms, max 15s, and factor 150%.
//
// This results in the following behavior:
//
//	Attempt     Backoff      Total
//	      1      0.150s      0.150s
//	      2      0.225s      0.375s
//	      3      0.338s      0.713s
//	      4      0.506s      1.219s
//	      5      0.759s      1.978s
//	      6      1.139s      3.117s
//	      7      1.709s      4.826s
//	      8      2.563s      7.389s
//	      9      3.844s     11.233s
//	     10      5.767s     17.000s
//	     11      8.650s     25.649s
//	     12     12.975s     38.624s
//	     13     15.000s     53.624s
//	     13     15.000s     53.624s
//	     14     15.000s     68.624s
//	     15     15.000s     83.624s
//	    ...         ...         ...
func DefaultExponentialBackoff() Policy {
	return ExponentialBackoff(DefaultMinBackoff, DefaultMaxBackoff, DefaultGrowthFactor)
}

type exponentialBackoff struct {
	min    time.Duration
	max    time.Duration
	factor float64
}

func (p *exponentialBackoff) Next(start, now time.Time, attempt int) (time.Duration, bool) {
	growthFactor := math.Pow(p.factor, float64(attempt-1))
	backoff := time.Duration(growthFactor * float64(p.min))
	if backoff > p.max {
		backoff = p.max
	}
	return backoff, true
}

// WithRandomJitter returns a Policy that wraps the parent Policy and adds random jitter
// as a plus or minus factor of its backoff. For example, with a factor of 0.5 and a parent
// backoff of 10s, the randomized backoff would be in the interval of [5s, 15s].
func WithRandomJitter(parent Policy, rand *rand.Rand, factor float64) Policy {
	if rand == nil {
		rand = globalRand
	}
	if factor <= 0 || factor > 1 {
		factor = DefaultJitterFactor
	}
	return &withRandomJitter{parent: parent, factor: factor, rand: rand}
}

// WithDefaultRandomJitter returns a Policy that wraps the parent Policy with random jitter
// using the default values of a globally shared source of randomness and a factor of 50%.
func WithDefaultRandomJitter(parent Policy) Policy {
	return WithRandomJitter(parent, globalRand, DefaultJitterFactor)
}

type withRandomJitter struct {
	parent Policy
	factor float64
	rand   *rand.Rand
}

func (p *withRandomJitter) Next(start, now time.Time, attempt int) (time.Duration, bool) {
	b, allow := p.parent.Next(start, now, attempt)
	if !allow {
		return 0, false
	}
	r := p.rand.Float64()
	j := p.factor
	// r = [0, 1)
	// 2*r = [0, 2)
	// 2*r - 1 = [-1, 1)
	// j*(2*r - 1) = [-j, j)
	// 1 + j*(2*r - 1) = [1 - j, 1 + j)
	// b*(1 + j*(2*r - 1)) = [b - j*b, b + j*b)
	return time.Duration(float64(b) * (1 + (j * (2*r - 1)))), true
}

// WithMaxRetries returns a Policy that wraps the parent Policy and sets a limit
// for the total number of retry attempts.
func WithMaxRetries(parent Policy, limit int) Policy {
	return &maxRetries{parent, limit}
}

type maxRetries struct {
	parent Policy
	limit  int
}

func (p *maxRetries) Next(start, now time.Time, attempt int) (time.Duration, bool) {
	if attempt > p.limit {
		return 0, false
	}
	return p.Next(start, now, attempt)
}

// WithMaxElapsedDuration returns a Policy that wraps the parent Policy and sets a limit
// for the total elapsed duration in which retries are allowed.
func WithMaxElapsedDuration(parent Policy, limit time.Duration) Policy {
	return &maxElapsed{parent, limit}
}

type maxElapsed struct {
	parent Policy
	limit  time.Duration
}

func (p *maxElapsed) Next(start, now time.Time, attempt int) (time.Duration, bool) {
	d, ok := p.parent.Next(start, now, attempt)
	if start.Add(p.limit).Before(now.Add(d)) {
		return 0, false
	}
	return d, ok
}
