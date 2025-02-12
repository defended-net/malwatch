package fan

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/scan"
	"golang.org/x/sys/unix"
)

func TestNewNotify(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("monitor: root needed for tests, skipping")
	}

	_, err := NewNotify(
		unix.FAN_CLOEXEC|
			unix.FAN_CLASS_NOTIF|
			unix.FAN_UNLIMITED_QUEUE|
			unix.FAN_UNLIMITED_MARKS,
		os.O_RDONLY|
			unix.O_LARGEFILE|
			unix.O_CLOEXEC,
	)

	if err != nil {
		t.Fatalf("notify create error %v", err)
	}
}

func TestGetEvent(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("monitor: root needed for tests, skipping")
	}

	dir := t.TempDir()

	mount, err := fsys.MntPoint(dir)
	if err != nil {
		t.Fatalf("get mnt point error %v", err)
	}

	notify, err := NewNotify(
		unix.FAN_CLOEXEC|
			unix.FAN_CLASS_NOTIF|
			unix.FAN_UNLIMITED_QUEUE|
			unix.FAN_UNLIMITED_MARKS,
		os.O_RDONLY|
			unix.O_LARGEFILE|
			unix.O_CLOEXEC,
	)

	if err != nil {
		t.Fatalf("notify create error %v", err)
	}

	if err = notify.Mark(
		unix.FAN_MARK_ADD|
			unix.FAN_MARK_MOUNT,
		unix.FAN_CLOSE_WRITE,
		unix.AT_FDCWD,
		mount,
	); err != nil {
		t.Fatalf("mark error %v", err)
	}

	meta, err := notify.GetEvent()
	if err != nil {
		t.Fatalf("get event error %v", err)
	}
	defer meta.Close()

	if _, err := os.Create(filepath.Join(dir, t.Name())); err != nil {
		t.Errorf("file create error %v", err)
	}
}

func TestMark(t *testing.T) {
	if os.Getuid() != 0 {
		fmt.Println("monitor: tests require root")
		return
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error %v", err)
	}

	notify, err := NewNotify(
		unix.FAN_CLOEXEC|
			unix.FAN_CLASS_NOTIF|
			unix.FAN_UNLIMITED_QUEUE|
			unix.FAN_UNLIMITED_MARKS,
		os.O_RDONLY|
			unix.O_LARGEFILE|
			unix.O_CLOEXEC,
	)
	if err != nil {
		t.Fatalf("notify create error %v", err)
	}

	paths, err := scan.Glob(env.Cfg.Scans.Paths)
	if err != nil {
		t.Fatalf("scan path glob error %v", err)
	}

	for _, path := range paths {
		mount, err := fsys.MntPoint(path)
		if err != nil {
			t.Fatalf("get mnt point error %v", err)
		}

		if err = notify.Mark(
			unix.FAN_MARK_ADD|
				unix.FAN_MARK_MOUNT,
			unix.FAN_CLOSE_WRITE,
			unix.AT_FDCWD,
			mount,
		); err != nil {
			t.Errorf("mark error %v", err)
		}
	}
}

func TestFile(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	meta := &EventMeta{
		unix.FanotifyEventMetadata{
			Fd: int32(file.Fd()),
		},
	}

	result := meta.File()
	defer result.Close()

	if result == nil {
		t.Errorf("unexpected file nil result")
	}
}

func TestClose(t *testing.T) {
	input := filepath.Join(t.TempDir(), t.Name())
	file, err := os.Create(input)
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	meta := &EventMeta{
		unix.FanotifyEventMetadata{
			Fd: int32(file.Fd()),
		},
	}

	if result := meta.Close(); result != nil {
		t.Errorf("meta close error: %v", result)
	}
}

func TestGetPath(t *testing.T) {
	want := filepath.Join(t.TempDir(), t.Name())
	file, err := os.Create(want)
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	meta := &EventMeta{
		unix.FanotifyEventMetadata{
			Fd: int32(file.Fd()),
		},
	}

	result, err := meta.GetPath()
	if err != nil {
		t.Fatalf("get path error: %v", err)
	}

	if result != want {
		t.Errorf("unexpected path result %v, want %v", result, want)
	}
}

func TestMatchMask(t *testing.T) {
	tests := map[string]struct {
		input *EventMeta
		want  bool
	}{
		"true": {
			input: &EventMeta{
				unix.FanotifyEventMetadata{
					Mask: 1,
				},
			},

			want: true,
		},

		"false": {
			input: &EventMeta{
				unix.FanotifyEventMetadata{
					Mask: 2,
				},
			},

			want: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if result := test.input.MatchMask(1); result != test.want {
				t.Errorf("unexpected match mask result %v, want %v", result, test.want)
			}
		})
	}
}
