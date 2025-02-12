// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package job

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/scan/worker"
	"github.com/defended-net/malwatch/pkg/sig"
	"github.com/defended-net/malwatch/third_party/yr"
)

func TestStart(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	rules, err := yr.LoadRules(env.Paths.Sigs.Yrc)
	if err != nil {
		t.Fatalf("yara load error: %v", err)
	}

	job := New("target", &Paths{}, 0, 1, []acter.Acter{}, []func(*state.Result) error{}, true)

	worker, err := worker.New(env.Cfg, rules, 0)
	if err != nil {
		t.Fatalf("worker create error: %v", err)
	}

	job.Start(context.Background(), &act.Skips{}, worker)
}

func TestStopBatch(t *testing.T) {
	var (
		path = filepath.Join(t.TempDir(), t.Name())

		task = []func(*state.Result) error{
			func(*state.Result) error {
				_, err := os.Create(path)
				return err
			},
		}

		job = New("target", &Paths{}, time.Second, 1, []acter.Acter{}, task, true)
	)

	go func() {
		defer close(job.State.Hits)

		for idx := 0; idx < 2; idx++ {
			job.State.Hits <- &state.Hit{
				Meta: &hit.Meta{},
			}
		}
	}()

	job.Stop()

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("task flag file not found: %v", path)
		}
	}
}

func TestStopNoHit(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	rules, err := yr.LoadRules(env.Paths.Sigs.Yrc)
	if err != nil {
		t.Fatalf("yara load error: %v", err)
	}

	var (
		path = filepath.Join(t.TempDir(), t.Name())

		task = []func(*state.Result) error{
			func(*state.Result) error {
				_, err := os.Create(path)
				return err
			},
		}

		job = New("target", &Paths{}, time.Second, 3, []acter.Acter{}, task, true)
	)

	worker, err := worker.New(env.Cfg, rules, 0)
	if err != nil {
		t.Fatalf("worker create error: %v", err)
	}

	job.Start(context.Background(), &act.Skips{}, worker)

	job.Stop()

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("task flag file not found: %v", path)
		}
	}
}

func TestWalk(t *testing.T) {
	var (
		dir  = t.TempDir()
		path = filepath.Join(dir, t.Name())

		input = &Paths{
			Files: []string{path},
			Dirs:  []string{dir},
		}
	)

	file, err := os.Create(path)
	if err != nil {
		t.Errorf("file create error: %v", err)
	}
	defer file.Close()

	job := New("target", input, 0, 1, []acter.Acter{}, nil, true)

	job.Walk(&act.Skips{}, 1)
}

func TestNew(t *testing.T) {
	got := New("target", &Paths{}, 0, 1, []acter.Acter{}, []func(*state.Result) error{}, true)

	want := &Job{
		Target:  "target",
		State:   got.State,
		paths:   got.paths,
		timeout: 0,
		batchSz: 1,
		db:      got.db,
		spinner: got.spinner,
		acters:  []acter.Acter{},
		tasks:   []func(*state.Result) error{},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected new job result %v, want %v", got, want)
	}
}

func TestActs(t *testing.T) {
	var (
		mock   = acter.Mock(t.Name())
		result = &state.Result{}
	)

	job := New("target", &Paths{}, 0, 1, []acter.Acter{mock}, []func(*state.Result) error{}, true)

	job.Acts(result)
}

func TestTasks(t *testing.T) {
	var (
		result = &state.Result{
			Errs: &state.Errs{},
		}

		task = func(*state.Result) error {
			return nil
		}
	)

	job := New("target", &Paths{}, 0, 1, []acter.Acter{}, []func(*state.Result) error{task}, true)

	job.Tasks(result)

	got := result.Errs.Get()

	if len(got) > 0 {
		t.Errorf("unexpected tasks errors %v", got)
	}
}

func TestTasksErr(t *testing.T) {
	var (
		result = &state.Result{
			Errs: &state.Errs{},
		}

		task = []func(*state.Result) error{
			func(*state.Result) error {
				return io.EOF
			},
		}
	)

	job := New("target", &Paths{}, 0, 1, []acter.Acter{}, task, true)

	job.Tasks(result)

	var (
		got  = result.Errs.Get()
		want = []error{io.EOF}
	)

	if !slices.Equal(got, []error{io.EOF}) {
		t.Errorf("unexpected task errors result %v, want %v", got, want)
	}
}

func TestFilterAct(t *testing.T) {
	want := &state.Result{
		Paths: state.Paths{
			t.TempDir(): &hit.Meta{
				Acts: []string{
					t.Name(),
				},
			},
		},
	}

	got := FilterAct(want, t.Name())

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected filter act result %v, want %v", got, want)
	}
}
