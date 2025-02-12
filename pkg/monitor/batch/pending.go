// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package batch

import (
	"sync"
)

// Pending represents concurrent safe batch of paths to scan.
type Pending struct {
	mtx  sync.Mutex
	vals map[string]struct{}
}

// NewPending returns a batch for pending scan paths.
func NewPending() *Pending {
	return &Pending{
		vals: map[string]struct{}{},
	}
}

// Get returns a given pending batch's paths. Paths are then cleared.
func (batch *Pending) Get() []string {
	batch.mtx.Lock()
	defer batch.mtx.Unlock()

	tmp := []string{}

	for path := range batch.vals {
		tmp = append(tmp, path)
	}

	// Reset
	clear(batch.vals)

	return tmp
}

// Add adds a given path to a pending batch.
func (batch *Pending) Add(path string) {
	batch.mtx.Lock()
	defer batch.mtx.Unlock()

	batch.vals[path] = struct{}{}
}
