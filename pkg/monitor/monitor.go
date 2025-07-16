// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.etcd.io/bbolt"
	"golang.org/x/sys/unix"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/monitor/batch"
	"github.com/defended-net/malwatch/pkg/scan"
	"github.com/defended-net/malwatch/pkg/scan/job"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/scan/worker"
	"github.com/defended-net/malwatch/pkg/sig"
	"github.com/defended-net/malwatch/third_party/fan"
	"github.com/defended-net/malwatch/third_party/yr"
)

// Monitor represents a monitor.
type Monitor struct {
	job      *job.Job
	notify   *fan.Notify
	scanner  *yr.Scanner
	interval time.Duration
	db       *bbolt.DB
	workers  []*worker.Worker
	skips    *act.Skips
	queue    chan string
	hits     *batch.Hits
}

// New returns a monitor from given env.
func New(env *env.Env) (*Monitor, error) {
	rules, err := yr.LoadRules(env.Paths.Sigs.Yrc)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", sig.ErrYrRulesLoad, err)
	}

	scanner, err := yr.NewScanner(rules)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", sig.ErrYrScanner, err)
	}

	var (
		tasks = []func(result *state.Result) error{
			func(result *state.Result) error { return result.Save(env.Db) },
			func(result *state.Result) error { return result.Log() },
		}

		job = job.New("", nil, 0, env.Cfg.Scans.BatchSz, env.Plat.Acters(), tasks, false)
	)

	monitor := &Monitor{
		job:      job,
		scanner:  scanner,
		interval: time.Duration(env.Cfg.Scans.Monitor.Timeout) * time.Second,
		db:       env.Db,
		skips:    act.GetSkips(env.Cfg.Acts, env.Paths),
		queue:    make(chan string, env.Cfg.Scans.BatchSz),
		hits:     &batch.Hits{},
	}

	if err := monitor.addPaths(env); err != nil {
		return nil, err
	}

	for thread := 1; thread <= env.Cfg.Threads; thread++ {
		// File expiration is irrelevant for monitor.
		worker, err := worker.New(env.Cfg, rules)
		if err != nil {
			return nil, err
		}

		monitor.workers = append(monitor.workers, worker)
	}

	return monitor, nil
}

// Run starts a monitor while orchestrating timed batches.
func Run(env *env.Env) error {
	ctx, cancel := context.WithCancel(context.Background())
	env.State.AddCancel(cancel)

	monitor, err := New(env)
	if err != nil {
		return err
	}

	var (
		timer   = time.NewTimer(monitor.interval)
		hits    []*state.Hit
		grouped []*state.Result
	)

	go monitor.listen(ctx)

	for _, worker := range monitor.workers {
		monitor.job.State.WGrp.Add(1)
		go worker.Work(ctx, monitor.job.State, monitor.queue)
	}

	for {
		select {
		case <-ctx.Done():
			// avoid memory leak.
			if !timer.Stop() {
				<-timer.C
			}

			return ctx.Err()

		case <-timer.C:
			hits = monitor.hits.Get(true)
			if len(hits) == 0 {
				timer.Reset(monitor.interval)
				continue
			}

			grouped = state.Group("", hits)

			for _, result := range grouped {
				monitor.job.Acts(result)
				monitor.job.Tasks(result)
			}

			timer.Reset(monitor.interval)

		case hit := <-monitor.job.State.Hits:
			monitor.hits.Add(hit)
		}
	}
}

// listen waits for notify events then validates before adding each path to the batch.
func (monitor *Monitor) listen(ctx context.Context) {
	var (
		path   string
		target string
		err    error
	)

	for {
		path, err = update(monitor.notify)
		if err != nil {
			monitor.job.State.AddErr(err)
		}

		// Cheaper than re, check first.
		if path == "" {
			continue
		}

		target = re.Target(path)
		if target == "fs" {
			continue
		}

		// Check file skips. Do early because cheaper than dir check.
		if _, ok := monitor.skips.Files[path]; ok {
			continue
		}

		// Check dir skips.
		if fsys.IsRel(path, monitor.skips.Dirs...) {
			continue
		}

		select {
		case <-ctx.Done():
			return

		default:
			monitor.queue <- path
		}
	}
}

// update gets ecents from a given notify.
func update(notify *fan.Notify) (string, error) {
	meta, err := notify.GetEvent()
	if err != nil {
		return "", err
	}
	defer meta.Close()

	if meta == nil {
		return "", nil
	}

	switch {
	case meta.MatchMask(unix.FAN_CLOSE_WRITE):
		path, err := meta.GetPath()
		if err != nil {
			return "", err
		}

		return path, nil

	default:
		return "", fmt.Errorf("%w, %v", ErrUnknownEvent, meta)
	}
}

// addPaths adds the notify before comparing scan paths with filesystem mounts to mark them.
func (monitor *Monitor) addPaths(env *env.Env) error {
	notify, err := fan.NewNotify(
		unix.FAN_CLOEXEC|
			unix.FAN_CLASS_NOTIF|
			unix.FAN_UNLIMITED_QUEUE|
			unix.FAN_UNLIMITED_MARKS,
		os.O_RDONLY|
			unix.O_LARGEFILE|
			unix.O_CLOEXEC,
	)
	if err != nil {
		return err
	}

	monitor.notify = notify

	paths, err := scan.Glob(env.Cfg.Scans.Paths)
	if err != nil {
		return err
	}

	for _, path := range paths {
		mount, err := fsys.MntPoint(path)
		if err != nil {
			return fmt.Errorf("%w, %v, %v", ErrUnknownMnt, err, mount)
		}

		if err = notify.Mark(
			unix.FAN_MARK_ADD|
				unix.FAN_MARK_MOUNT,
			unix.FAN_CLOSE_WRITE,
			unix.AT_FDCWD,
			mount,
		); err != nil {
			return fmt.Errorf("%w, %v", ErrNotifierMark, err)
		}
	}

	return nil
}
