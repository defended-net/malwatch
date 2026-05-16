// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

// rss returns rss from /proc/self/statm unit is kib.
func rss() (int64, error) {
	statm, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0, err
	}

	parts := strings.Fields(string(statm))
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid statm format %v", string(statm))
	}

	pages, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return pages * int64(os.Getpagesize()) / 1024, nil
}

func TestSetLeak(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	before, err := rss()
	if err != nil {
		t.Fatalf("rss err %v", err)
	}

	for idx := range 1000 {
		if err := Set(env.Paths.Sigs.Yrc, uint64(idx)); err != nil {
			t.Fatalf("set err  %v", err)
		}

		sigs, err := Acquire()
		if err != nil {
			t.Fatalf("acquire err  %v", err)
		}

		sigs.Release()
	}

	runtime.GC()
	debug.FreeOSMemory()

	after, err := rss()
	if err != nil {
		t.Fatalf("rss err %v", err)
	}

	// 4MiB cushion for alloc noise.
	if got := after - before; got > 4096 {
		t.Errorf("rss grew by %d kib", got)
	}
}

func TestSetLeakConcurrent(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	before, err := rss()
	if err != nil {
		t.Fatalf("rss err %v", err)
	}

	var wg sync.WaitGroup

	for idx := range 100 {
		if err := Set(env.Paths.Sigs.Yrc, uint64(idx)); err != nil {
			t.Fatalf("set err %v", err)
		}

		for range 500 {
			wg.Add(1)

			go func() {
				defer wg.Done()

				sigs, err := Acquire()
				if err != nil {
					return
				}
				defer sigs.Release()

				_ = sigs.Rev
			}()
		}
	}

	wg.Wait()

	runtime.GC()
	debug.FreeOSMemory()

	after, err := rss()
	if err != nil {
		t.Fatalf("rss err %v", err)
	}

	// 4MiB cushion for alloc noise.
	if got := after - before; got > 4096 {
		t.Errorf("rss grew by %d kib", got)
	}
}

func TestAcquireNil(t *testing.T) {
	curr.Store(nil)

	if _, got := Acquire(); got == nil {
		t.Errorf("unexpected acquire success")
	}
}

func TestSetErr(t *testing.T) {
	if got := Set("/dev/null/not-exist", 0); got == nil {
		t.Errorf("unexpected set success")
	}
}
