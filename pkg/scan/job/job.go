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

// Job represents scan job. Each target can have exactly one job.
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

// New returns new job.
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

// Walk traverses paths for given job.
func (job *Job) Walk(skips *act.Skips, sz int) <-chan string {
	queue := make(chan string, sz)

	go func() {
		defer close(queue)

		for _, entry := range append(job.paths.Dirs, job.paths.Files...) {
			if err := filepath.WalkDir(entry, func(path string, info os.DirEntry, err error) error {
				switch {
				case err != nil:
					job.State.AddErr(err)

					return nil

				case fsys.IsRel(path, skips.Dirs...):
					return filepath.SkipDir

				case skips.HasFile(path):
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

// Start starts given job.
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

// Stop stops given job by flushing hits.
func (job *Job) Stop() {
	hits := make([]*state.Hit, 0, job.batchSz)

	for hit := range job.State.Hits {
		hits = append(hits, hit)

		if len(hits) >= job.batchSz {
			job.flush(hits)
			hits = hits[:0]
		}
	}

	job.flush(hits)

	for _, err := range job.State.Errs() {
		slog.Error(err.Error())
	}
}

// flush first performs acts and then tasks for given hits.
func (job *Job) flush(hits []*state.Hit) {
	if len(hits) == 0 {
		return
	}

	for _, result := range state.Group(job.Target, hits) {
		job.Acts(result)
		job.Tasks(result)
	}
}

// Acts performs acts for given result.
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

// Tasks performs tasks for given result.
func (job *Job) Tasks(result *state.Result) {
	for _, task := range job.tasks {
		if err := task(result); err != nil {
			result.AddErr(err)
		}
	}
}

// FilterAct returns grouped hits for given act verb.
func FilterAct(result *state.Result, verb string) *state.Result {
	filtered := state.NewResult(result.Target, state.Paths{})

	for path, meta := range result.Paths {
		if slices.Contains(meta.Acts, verb) {
			filtered.Paths[path] = meta
		}
	}

	return filtered
}
