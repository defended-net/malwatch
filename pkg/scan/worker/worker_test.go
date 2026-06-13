// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package worker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/sig"
)

var sample = `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`

func TestNew(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := sig.Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	if _, got := New(env.Cfg); got != nil {
		t.Errorf("create worker err %v", got)
	}
}

func TestWork(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	worker, err := Mock(env)
	if err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	queue := make(chan string)

	go func() {
		defer close(queue)

		queue <- filepath.Join(t.TempDir(), t.Name())
	}()

	state := state.NewJob()

	state.WGrp.Add(1)

	worker.Work(context.Background(), state, queue)
}

func TestScan(t *testing.T) {
	var (
		path  = filepath.Join(t.TempDir(), t.Name())
		input = []byte(sample)
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	worker, err := Mock(env)
	if err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	if err := os.WriteFile(path, input, 0600); err != nil {
		t.Errorf("file write err %v", err)
	}

	state := state.NewJob()

	go func(worker *Worker) {
		defer close(state.Hits)

		worker.Scan(path, state)
	}(worker)
}

func TestScanOpenErr(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	worker, err := Mock(env)
	if err != nil {
		t.Fatalf("worker mock err %v", err)
	}

	job := state.NewJob()

	worker.Scan(filepath.Join(t.TempDir(), "not-exist"), job)

	if len(job.Errs()) == 0 {
		t.Errorf("unexpected empty errs")
	}
}

func TestWorkCtxCancel(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	worker, err := Mock(env)
	if err != nil {
		t.Fatalf("worker mock err %v", err)
	}

	queue := make(chan string, 1)
	queue <- filepath.Join(t.TempDir(), t.Name())
	close(queue)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	job := state.NewJob()
	job.WGrp.Add(1)

	worker.Work(ctx, job, queue)
}

func TestScanMatch(t *testing.T) {
	var (
		path  = filepath.Join(t.TempDir(), t.Name())
		input = []byte(sample)
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	worker, err := Mock(env)
	if err != nil {
		t.Fatalf("worker mock err %v", err)
	}

	if err := os.WriteFile(path, input, 0600); err != nil {
		t.Fatalf("file write err %v", err)
	}

	job := state.NewJob()

	go func() {
		defer close(job.Hits)

		worker.Scan(path, job)
	}()

	for hit := range job.Hits {
		if hit.Path != path {
			t.Errorf("unexpected hit path %v, want %v", hit.Path, path)
		}
	}
}

func TestRefresh(t *testing.T) {
	_env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	worker, err := Mock(_env)
	if err != nil {
		t.Fatalf("worker mock err %v", err)
	}

	if got := worker.Refresh(); got != nil {
		t.Errorf("refresh err %v", got)
	}
}

func TestMock(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if _, got := Mock(env); got != nil {
		t.Errorf("sig mock err %v", got)
	}
}
