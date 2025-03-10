// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package job

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"time"

	"go.etcd.io/bbolt"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/scan/worker"
	"github.com/defended-net/malwatch/pkg/tui"
)

// Job represents a scan job. Each target can have exactly one job.
type Job struct {
	Target  string
	State   *state.Job
	timeout time.Duration
	batchSz int
	db      *bbolt.DB
	spinner *tui.Spinner
	paths   *Paths
	acters  []acter.Acter
	tasks   []func(*state.Result) error
}

// Paths represents scan paths.
type Paths struct {
	Files []string
	Dirs  []string
}

// New returns a new job.
func New(target string, paths *Paths, timeout time.Duration, batchSz int, acters []acter.Acter, tasks []func(*state.Result) error, ticker bool) *Job {
	job := &Job{
		Target:  target,
		State:   state.NewJob(),
		timeout: timeout,
		batchSz: batchSz,
		paths:   paths,
		acters:  acters,
		tasks:   tasks,
	}

	if ticker {
		job.spinner = tui.NewSpinner(125*time.Millisecond, "Reading Files")
	}

	return job
}

// Walk traverses paths.
// File counting done here in single thread to avoid lock contention from workers handling it.
func (job *Job) Walk(skips *act.Skips, sz int) chan string {
	queue := make(chan string, sz)

	go func() {
		defer close(queue)

		for _, entry := range append(job.paths.Dirs, job.paths.Files...) {
			if err := filepath.WalkDir(entry, func(path string, info os.DirEntry, err error) error {
				switch {
				case err != nil:
					return nil

				case fsys.IsRel(path, skips.Dirs...):
					return filepath.SkipDir

				case skips.Files[path] != struct{}{}:
					return nil

				case !info.Type().IsRegular():
					return nil
				}

				queue <- path

				return nil
			}); err != nil {
				job.State.AddErr(fmt.Errorf("%v, %w", fsys.ErrWalk, err))
			}
		}
	}()

	return queue
}

// Start starts a job.
func (job *Job) Start(ctx context.Context, skips *act.Skips, workers ...*worker.Worker) {
	go job.spinner.Start()
	queue := job.Walk(skips, job.batchSz)

	for _, worker := range workers {
		job.State.WGrp.Add(1)
		go worker.Work(ctx, job.State, queue)
	}

	go func() {
		defer job.spinner.Stop()
		defer close(job.State.Hits)

		job.State.WGrp.Wait()
	}()
}

// Stop stops a job.
func (job *Job) Stop() {
	var (
		idx  int
		hits = []*state.Hit{}
	)

	for hit := range job.State.Hits {
		idx++
		hits = append(hits, hit)

		if idx > job.batchSz {
			grouped := state.Group(job.Target, hits)

			for _, result := range grouped {
				job.Acts(result)
				job.Tasks(result)
			}

			hits = nil
			idx = 0
		}
	}

	grouped := state.Group(job.Target, hits)

	for _, result := range grouped {
		job.Acts(result)
		job.Tasks(result)
	}

	for _, err := range job.State.Errs() {
		slog.Error(err.Error())
	}
}

// Acts performs actions with a given result.
func (job *Job) Acts(result *state.Result) {
	for _, acter := range job.acters {
		filtered := FilterAct(result, acter.Verb())

		if len(filtered.Paths) == 0 {
			continue
		}

		if err := acter.Act(filtered); err != nil {
			result.AddErr(err)
		}
	}
}

// Tasks performs tasks with a given result.
func (job *Job) Tasks(result *state.Result) {
	for _, task := range job.tasks {
		if err := task(result); err != nil {
			result.AddErr(err)
		}
	}
}

// FilterAct returns grouped hits from given act verb.
func FilterAct(result *state.Result, verb string) *state.Result {
	filtered := state.NewResult(result.Target, state.Paths{})

	for path, meta := range result.Paths {
		if slices.Contains(meta.Acts, verb) {
			filtered.Paths[path] = meta
		}
	}

	return filtered
}
