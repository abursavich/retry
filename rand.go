// SPDX-License-Identifier: MIT
//
// Copyright 2022 Andrew Bursavich. All rights reserved.
// Use of this source code is governed by The MIT License
// which can be found in the LICENSE file.

package retry

import (
	"math/rand"
	"sync"
	"time"
)

var globalRand = rand.New(newLockedSource(time.Now().UnixNano()))

type lockedSource struct {
	mu  sync.Mutex
	src rand.Source64
}

func newLockedSource(seed int64) rand.Source64 {
	return &lockedSource{src: rand.NewSource(seed).(rand.Source64)}
}

func (src *lockedSource) Int63() int64 {
	src.mu.Lock()
	defer src.mu.Unlock()
	return src.src.Int63()
}

func (src *lockedSource) Uint64() uint64 {
	src.mu.Lock()
	defer src.mu.Unlock()
	return src.src.Uint64()
}

func (r *lockedSource) Seed(seed int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.src.Seed(seed)
}
