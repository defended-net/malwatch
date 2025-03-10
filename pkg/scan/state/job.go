// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import "sync"

// Job represents a job state.
type Job struct {
	WGrp  *sync.WaitGroup
	Count int       `json:"count"`
	Hits  chan *Hit `json:"-"`
	errs  *Errs     `json:"-"`
}

// NewJob returns a job.
func NewJob() *Job {
	return &Job{
		WGrp: &sync.WaitGroup{},
		Hits: make(chan *Hit),
		errs: &Errs{},
	}
}

// AddErr adds error to given job.
func (job *Job) AddErr(err error) {
	job.errs.Add(err)
}

// Errs returns given job's errs.
func (job *Job) Errs() []error {
	return job.errs.Get()
}
