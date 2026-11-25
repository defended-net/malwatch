// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/scan"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/sig"
	"github.com/defended-net/malwatch/third_party/fan"
	"golang.org/x/sys/unix"
)

var sample = `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`

func TestMain(m *testing.M) {
	if os.Getuid() != 0 {
		fmt.Println("monitor: tests require root")

		return
	}

	re.SetTargets(regexp.MustCompile(`^/(?P<target>[^/]+)`))

	m.Run()
}

func TestNewErrs(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Paths.Sigs.Yrc = ""

	want := sig.ErrYrcGet

	if _, got := New(env); !errors.Is(got, want) {
		t.Errorf("unexpected new monitor err %v, want %v", got, want)
	}
}

func TestNewErr(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := sig.Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	env.Cfg.Scans.Paths = []string{"["}

	if _, got := New(env); got == nil {
		t.Error("unexpected monitor create success")
	}
}

func TestRefresh(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error %v", err)
	}

	if err := sig.Mock(env, true); err != nil {
		t.Fatalf("sig mock error %v", err)
	}

	monitor, err := New(env)
	if err != nil {
		t.Fatalf("monitor error %v", err)
	}

	tests := map[string]struct {
		rev  uint64
		want uint64
	}{
		"rev-1": {
			rev:  1,
			want: 1,
		},

		"rev-2": {
			rev:  2,
			want: 2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := monitor.refresh(test.rev); err != nil {
				t.Errorf("refresh err %v", err)
			}

			if monitor.rev != test.want {
				t.Errorf("wanted rev %d, want %d", monitor.rev, test.want)
			}
		})
	}
}

func TestMark(t *testing.T) {
	var (
		dir = t.TempDir()

		tests = map[string]struct {
			path string
			kind uint
			want error
		}{
			"fs": {
				path: dir,
				kind: unix.FAN_MARK_FILESYSTEM,
			},

			"mnt": {
				path: dir,
				kind: unix.FAN_MARK_MOUNT,
			},

			"not-exist": {
				path: filepath.Join(dir, "not-exist"),
				kind: unix.FAN_MARK_MOUNT,
				want: unix.ENOENT,
			},
		}
	)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			notify, err := fan.NewNotify(
				unix.FAN_CLOEXEC|
					unix.FAN_CLASS_NOTIF,

				os.O_RDONLY|
					unix.O_CLOEXEC,
			)
			if err != nil {
				t.Fatalf("notify create error %v", err)
			}

			t.Cleanup(func() {
				if err := notify.File.Close(); err != nil {
					t.Fatalf("notify close error %v", err)
				}
			})

			if got := mark(notify, test.path, test.kind); !errors.Is(got, test.want) {
				t.Fatalf("unexpected mark err %v, want %v", got, test.want)
			}
		})
	}
}

func TestAddPaths(t *testing.T) {
	var (
		env, err = env.Mock(t.Name(), t.TempDir())

		first  = t.TempDir()
		second = t.TempDir()
	)

	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	t.Cleanup(func() {
		if err := env.Paths.Install.Root.Close(); err != nil {
			t.Errorf("close install root err %v", err)
		}
	})

	tests := map[string]struct {
		paths []string
		want  error
	}{
		"one": {
			paths: []string{first},
		},

		"many": {
			paths: []string{
				first,
				second,
			},
		},

		"not-exist": {
			paths: []string{"not-exist"},
			want:  scan.ErrNoScanPaths,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env.Cfg.Scans.Paths = test.paths

			var (
				input  Monitor
				notify = input.notify
				got    = input.addPaths(env)
			)

			if notify != nil {
				t.Cleanup(func() {
					if err := notify.File.Close(); err != nil {
						t.Errorf("notify close err %v", err)
					}
				})
			}

			switch {
			case !errors.Is(got, test.want):
				t.Fatalf("unexpected add paths err %v, want %v", got, test.want)

			case got == nil && input.notify == nil:
				t.Fatalf("nil notify")
			}
		})
	}
}

func TestTick(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := sig.Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	monitor, err := New(env)
	if err != nil {
		t.Fatalf("monitor err %v", err)
	}

	monitor.rev = 1

	var (
		hits    []*state.Hit
		grouped []*state.Result
	)

	if err := monitor.tick(&hits, &grouped); err != nil {
		t.Errorf("tick err %v", err)
	}

	if monitor.rev != 0 {
		t.Errorf("unexpected rev %d, want 0", monitor.rev)
	}
}

func TestRun(t *testing.T) {
	var (
		tmp  = t.TempDir()
		path = filepath.Join(tmp, t.Name())
		want = context.Canceled
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := sig.Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	if _, err := os.Create(path); err != nil {
		t.Fatalf("file create err %v", err)
	}

	go func() {
		time.Sleep(3 * time.Second)

		for _, file := range []struct {
			path string
			data []byte
		}{
			{
				path, []byte(sample),
			},

			{
				path, []byte(sample),
			},

			{
				filepath.Join(tmp, "clean"), []byte(`clean`),
			},
		} {
			if err := os.WriteFile(file.path, file.data, 0600); err != nil {
				t.Errorf("file write err %v", err)
			}
		}

		time.Sleep(10 * time.Second)

		env.State.CancelAll()
	}()

	if got := Run(env); !errors.Is(got, want) {
		t.Errorf("unexpected run err %v, want %v", got, want)
	}
}
