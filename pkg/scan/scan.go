// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package scan

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/cmd"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/job"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/scan/worker"
	"github.com/defended-net/malwatch/pkg/sig"
	"github.com/defended-net/malwatch/third_party/yr"
)

// Scan represents a scan.
type Scan struct {
	opts    *env.Opts
	jobs    []*job.Job
	workers []*worker.Worker
	skips   *act.Skips
	acts    []acter.Acter
	timeout time.Duration
	cancel  *cmd.Cancel
}

// New returns a scan for given env and paths.
func New(env *env.Env, paths ...string) (*Scan, error) {
	rules, err := yr.LoadRules(env.Paths.Sigs.Yrc)
	if err != nil {
		return nil, fmt.Errorf("%w, %v, %v", sig.ErrYrRulesLoad, err, env.Paths.Sigs.Yrc)
	}

	scan := &Scan{
		opts:    env.Opts,
		cancel:  env.State.Cancel,
		skips:   act.GetSkips(env.Cfg.Acts, env.Paths),
		acts:    env.Plat.Acters(),
		timeout: time.Duration(env.Cfg.Scans.Timeout) * time.Minute,
	}

	scan.addJobs(env, paths)

	for thread := 1; thread <= env.Cfg.Threads; thread++ {
		worker, err := worker.New(env.Cfg, rules)
		if err != nil {
			return nil, err
		}

		scan.workers = append(scan.workers, worker)
	}

	return scan, nil
}

// addJobs adds jobs to a scan for given env and paths.
func (scan *Scan) addJobs(env *env.Env, paths []string) {
	tasks := []func(state *state.Result) error{
		func(result *state.Result) error { return result.Save(env.Db) },
		func(result *state.Result) error { return result.Log() },
	}

	if !env.Opts.Unattended {
		tasks = append(tasks, func(result *state.Result) error { return result.Print() })
	}

	for target, paths := range Group(paths) {
		job := job.New(target, paths, scan.timeout, env.Cfg.Scans.BatchSz, scan.acts, tasks, !scan.opts.NoTicker)

		scan.jobs = append(scan.jobs, job)
	}
}

// Run runs a scan.
func (scan *Scan) Run() error {
	for _, job := range scan.jobs {
		ctx, cancel := context.WithTimeout(context.Background(), scan.timeout)
		scan.cancel.Add(cancel)

		job.Start(ctx, scan.skips, scan.workers...)
		job.Stop()
	}

	return nil
}

// Glob returns globbed scan paths from given paths.
func Glob(paths []string) ([]string, error) {
	glob := []string{}

	for _, path := range paths {
		paths, err := filepath.Glob(path)
		if err != nil {
			return paths, fmt.Errorf("%v, %w, %v", ErrPathGlob, err, paths)
		}

		glob = append(glob, paths...)
	}

	if len(glob) == 0 {
		return glob, ErrNoScanPaths
	}

	return glob, nil
}

// Group sorts given paths to target:paths.
func Group(paths []string) map[string]*job.Paths {
	grouped := map[string]*job.Paths{}

	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			slog.Error(fsys.ErrStat.Error(), "path", path)
			continue
		}

		target := re.Target(path)

		if _, ok := grouped[target]; !ok {
			grouped[target] = &job.Paths{}
		}

		if stat.IsDir() {
			grouped[target].Dirs = append(grouped[target].Dirs, path)
			continue
		}

		grouped[target].Files = append(grouped[target].Files, path)
	}

	return grouped
}
