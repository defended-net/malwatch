// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package batch

import (
	"sync"

	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Hits represents concurrent safe hits.
type Hits struct {
	mtx  sync.Mutex
	vals []*state.Hit
}

// Get returns hits. Optional slice clear.
func (hits *Hits) Get(clear bool) []*state.Hit {
	hits.mtx.Lock()
	defer hits.mtx.Unlock()

	if clear {
		tmp := hits.vals
		hits.vals = nil

		return tmp
	}

	tmp := make([]*state.Hit, len(hits.vals))
	copy(tmp, hits.vals)

	return tmp
}

// Add adds a given path to a batch.
func (hits *Hits) Add(hit *state.Hit) {
	hits.mtx.Lock()
	defer hits.mtx.Unlock()

	hits.vals = append(hits.vals, hit)
}
