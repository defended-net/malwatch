// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

// Result represents a target's hits.
type Result struct {
	Target string
	Paths  Paths
	errs   *Errs `json:"-"`
}

// NewResult returns a result. Errs is created.
func NewResult(target string, paths Paths) *Result {
	return &Result{
		Target: target,
		Paths:  paths,
		errs:   &Errs{},
	}
}

// AddErr adds error to given result.
func (result *Result) AddErr(err error) {
	result.errs.Add(err)
}

// Errs returns given result's errs.
func (result *Result) Errs() []error {
	return result.errs.Get()
}
