// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

func TestMonitor(t *testing.T) {
	want := 1

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := Mock(env, false); err != nil {
		t.Fatalf("sigs mock err %v", err)
	}

	if err := Monitor(env); err != nil {
		t.Fatalf("monitor err %v", err)
	}

	if got := len(env.State.GetCancels()); got != want {
		t.Errorf("unexpected get cancels result, got %v, want %v", got, want)
	}
}

func TestMonitorErrYrcRefresh(t *testing.T) {
	want := ErrYrcStat

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if got := Monitor(env); !errors.Is(got, want) {
		t.Errorf("unexpected monitor err %v, want %v", got, want)
	}
}

func TestMonitorDetect(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := Mock(env, true); err != nil {
		t.Fatalf("sigs mock err %v", err)
	}

	sigs, err := Acquire()
	if err != nil {
		t.Fatalf("acquire err %v", err)
	}

	initial := sigs.Rev
	sigs.Release()

	now := time.Now().Add(time.Second)

	if err := os.Chtimes(env.Paths.Sigs.Yrc, now, now); err != nil {
		t.Fatalf("chtime err %v", err)
	}

	time.Sleep(3 * time.Second)

	sigs, err = Acquire()
	if err != nil {
		t.Fatalf("acquire err %v", err)
	}

	got := sigs.Rev
	sigs.Release()

	if got <= initial {
		t.Errorf("unexpected rev result %v, want > %v", got, initial)
	}
}

func TestMonitorHandlesStatError(t *testing.T) {
	want := 1

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := Mock(env, false); err != nil {
		t.Fatalf("sigs mock err %v", err)
	}

	if err := Monitor(env); err != nil {
		t.Fatalf("monitor err %v", err)
	}

	yrc := env.Paths.Sigs.Yrc

	time.Sleep(1 * time.Second)

	if err := os.Remove(yrc); err != nil {
		t.Fatalf("del err %v", err)
	}

	time.Sleep(3 * time.Second)

	got := env.State.GetCancels()

	if got := len(got); got != want {
		t.Errorf("unexpected cancel count result %v want %v", got, want)
	}

	for _, cancel := range got {
		cancel()
	}
}

func TestMonitorInitialStatError(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := Mock(env, false); err != nil {
		t.Fatalf("sigs mock err %v", err)
	}

	yrc := env.Paths.Sigs.Yrc

	if err := Monitor(env); err != nil {
		t.Fatalf("monitor err %v", err)
	}

	if _, err := os.Stat(yrc); err != nil {
		t.Errorf("stat err %v", err)
	}

	for _, cancel := range env.State.GetCancels() {
		cancel()
	}
}

func TestMonitorRevAtomic(t *testing.T) {
	monitor := &monitor{}

	if got := monitor.rev.Load(); got != 0 {
		t.Fatalf("unexpected rev result %v, want 0", got)
	}

	monitor.rev.Store(1)

	if got := monitor.rev.Load(); got != 1 {
		t.Fatalf("unexpected rev result %v, want 1", got)
	}

	if got := monitor.rev.Add(1); got != 2 {
		t.Errorf("unexpected rev result %v, want 2", got)
	}
}

func TestMonitorRevConcurrent(t *testing.T) {
	var (
		want  = uint64(1000)
		got   uint64
		input = &monitor{}
		done  = make(chan struct{})
	)

	go func() {
		defer close(done)

		for range want {
			input.rev.Add(1)
		}
	}()

	for range want {
		input.rev.Load()
	}

	<-done

	if got = input.rev.Load(); got != want {
		t.Errorf("unexpected rev %v, want %v", got, want)
	}
}
