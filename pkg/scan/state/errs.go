// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"sync"
)

// Errs represents an error store.
type Errs struct {
	mtx  sync.Mutex
	Vals []error
}

// Get returns the store's errors. Store is then cleared.
func (errs *Errs) Get() []error {
	errs.mtx.Lock()
	defer errs.mtx.Unlock()

	tmp := errs.Vals
	errs.Vals = nil

	return tmp
}

// Add adds a given error. Same error is returned.
func (errs *Errs) Add(err error) {
	errs.mtx.Lock()
	defer errs.mtx.Unlock()

	errs.Vals = append(errs.Vals, err)
}
