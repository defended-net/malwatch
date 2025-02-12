// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"sync"
)

// Errs represents an error store.
type Errs struct {
	Mtx  sync.Mutex
	Vals []error
}

// Get returns the store's errors. Store is then cleared.
func (errs *Errs) Get() []error {
	errs.Mtx.Lock()
	defer errs.Mtx.Unlock()

	tmp := make([]error, len(errs.Vals))

	copy(tmp, errs.Vals)

	errs.Vals = nil

	return tmp
}

// Add adds a given error. Same error is returned.
func (errs *Errs) Add(err error) {
	errs.Mtx.Lock()
	defer errs.Mtx.Unlock()

	errs.Vals = append(errs.Vals, err)
}
