// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package state

import (
	"slices"
	"sync"

	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
)

// Job represents a job state.
type Job struct {
	WGrp  *sync.WaitGroup
	Count int       `json:"count"`
	Hits  chan *Hit `json:"-"`
	Errs  *Errs     `json:"-"`
}

// Hit represents a hit detection.
type Hit struct {
	Path string
	Meta *hit.Meta
}

// Result represents a target's hits.
type Result struct {
	Target string
	Paths  Paths
	Errs   *Errs `json:"-"`
}

// Paths represents hit meta per path.
type Paths map[string]*hit.Meta

// New returns a job state.
func New() *Job {
	return &Job{
		WGrp: &sync.WaitGroup{},
		Hits: make(chan *Hit),
		Errs: &Errs{},
	}
}

// Group returns a slice of results from given target and slice of hits.
// Empty target param means hits are from unknown target(s) and should apply re.Target.
func Group(target string, hits []*Hit, errs *Errs) []*Result {
	if len(hits) == 0 {
		return []*Result{
			{
				Target: target,
				Paths:  Paths{},
				Errs:   errs,
			},
		}
	}

	grouped := make(map[string]*Result)

	for _, hit := range hits {
		if target == "" {
			target = re.Target(hit.Path)
		}

		result, exists := grouped[target]
		if !exists {
			result = &Result{
				Target: target,
				Paths:  Paths{},
				Errs:   errs,
			}

			grouped[target] = result
		}

		meta, exists := result.Paths[hit.Path]
		if !exists {
			result.Paths[hit.Path] = hit.Meta
		} else {
			for _, rule := range hit.Meta.Rules {
				if !slices.Contains(meta.Rules, rule) {
					meta.Rules = append(meta.Rules, rule)
				}
			}
		}
	}

	results := make([]*Result, 0, len(grouped))

	for _, result := range grouped {
		results = append(results, result)
	}

	return results
}

// AddErr adds error to given job state.
func (state *Job) AddErr(err error) {
	state.Errs.Add(err)
}

// GetErrs returns given job state's errors.
func (state *Job) GetErrs() []error {
	return state.Errs.Get()
}
