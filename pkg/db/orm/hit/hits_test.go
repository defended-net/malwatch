// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package hit

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/defended-net/malwatch/pkg/db/orm"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestNewMeta(t *testing.T) {
	file, err := os.OpenFile(filepath.Join(t.TempDir(), t.Name()), os.O_CREATE, 0600)
	if err != nil {
		t.Fatalf("file write error: %v", err)
	}

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat error: %v", err)
	}

	want := &Meta{
		Rules: []string{"eicar"},
		Attr:  fsys.NewAttr(stat),
		Acts:  []string{"alert"},
	}

	result := NewMeta(fsys.NewAttr(stat), []string{"eicar"}, "alert")

	want.Time = result.Time

	if !reflect.DeepEqual(result, want) {
		t.Errorf("unexpected save file hit result %v, want %v", result, want)
	}
}

func TestSaveZeroLength(t *testing.T) {
	input := &History{
		Target: "target",

		Paths: Paths{},
	}

	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	if err := input.Save(db); err != nil {
		t.Errorf("save error: %v", err)
	}
}

func TestSelectAll(t *testing.T) {
	input := &History{
		Target: "target",

		Paths: Paths{
			"/target/select-all.php": {
				{
					Rules:  []string{"eicar"},
					Status: "/quarantine/select-all.php",
					Attr:   &fsys.Attr{},
				},
			},

			"/target/select-all-b.php": {
				{
					Rules:  []string{"eicar"},
					Status: "/quarantine/select-all-b.php",
					Attr:   &fsys.Attr{},
				},
			},
		},
	}

	want := []*History{
		{
			Target: "target",

			Paths: Paths{
				"/target/select-all.php": {
					{
						Rules:  []string{"eicar"},
						Status: "/quarantine/select-all.php",
						Attr:   &fsys.Attr{},
					},
				},

				"/target/select-all-b.php": {
					{
						Rules:  []string{"eicar"},
						Status: "/quarantine/select-all-b.php",
						Attr:   &fsys.Attr{},
					},
				},
			},
		},
	}

	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	if err := input.Save(db); err != nil {
		t.Fatalf("save file hit error: %v", err)
	}

	hits, err := SelectAll(db)
	if err != nil {
		t.Fatalf("select all hits error: %v", err)
	}

	if !reflect.DeepEqual(hits, want) {
		t.Errorf("unexpected select all hits result %v, want %v", hits, want)
	}
}

func TestSelectTarget(t *testing.T) {
	tests := map[string]struct {
		input []*History
		want  int
	}{
		"target": {
			input: []*History{
				{
					Target: "target",

					Paths: Paths{
						"/target/select-target.php": {
							{
								Rules:  []string{"eicar"},
								Status: "/quarantine/sh",
								Attr:   &fsys.Attr{},
							},
						},
					},
				},
			},

			want: 1,
		},

		"fs": {
			input: []*History{
				{
					Target: "fs",

					Paths: Paths{
						"/bin/sh": {
							{
								Rules:  []string{"eicar"},
								Status: "/quarantine/sh",
								Attr:   &fsys.Attr{},
							},
						},
					},
				},
			},

			want: 1,
		},
	}

	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, hit := range test.input {
				if err := hit.Save(db); err != nil {
					t.Fatalf("db save file hit error: %v", err)
				}
			}

			result, err := SelectTarget(db, name)
			if err != nil {
				t.Fatalf("select target hits lookup error: %v", err)
			}

			if len(result) != test.want {
				t.Errorf("unexpected select target hits result %v, want %v", len(result), test.want)
			}
		})
	}
}

func TestSelectAllNoDb(t *testing.T) {
	if _, err := SelectAll(nil); err == nil {
		t.Errorf("unexpected select all no db success: %v", err)
	}
}

func TestHasValidPaths(t *testing.T) {
	tests := map[string]struct {
		input string
		meta  *Meta
		want  error
	}{
		"not-abs": {
			input: "../path",
			want:  fsys.ErrPathNotAbs,
		},

		"root-path": {
			input: "/",
			want:  fsys.ErrPathRoot,
		},

		"root-meta": {
			input: "/",
			want:  fsys.ErrPathRoot,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := fsys.HasDotDots(test.input); !errors.Is(err, test.want) {
				t.Errorf("unexpected has valid paths result %v, want %v", err, test.want)
			}
		})
	}
}

func TestUpdateHits(t *testing.T) {
	input := []*History{
		{
			Target: "target-a",

			Paths: Paths{
				"/target-a/sort-a-1.php": {},
				"/target-a/sort-a-2.php": {},
			},
		},

		{
			Target: "target-b",

			Paths: Paths{
				"/target-b/sort-b-1.php": {},
			},
		},
	}

	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	for _, hit := range input {
		if err := hit.Save(db); err != nil {
			t.Errorf("update error: %v", err)
		}
	}
}

func TestDelTarget(t *testing.T) {
	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	hits := &History{
		Target: "target",

		Paths: Paths{
			t.TempDir(): {},
		},
	}

	if err := hits.Save(db); err != nil {
		t.Errorf("save error: %v", err)
	}

	if err := DelTarget(db, "target"); err != nil {
		t.Errorf("del error: %v", err)
	}
}

func TestDelPath(t *testing.T) {
	var (
		path = t.TempDir()

		hits = &History{
			Target: "target",

			Paths: Paths{
				path: {},
			},
		}
	)

	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	if err := hits.Save(db); err != nil {
		t.Fatalf("save error: %v", err)
	}

	if err := DelPath(db, "target", path); err != nil {
		t.Errorf("del error: %v", err)
	}
}

func TestSelectLast(t *testing.T) {
	var (
		hits = &History{
			Target: "fs",

			Paths: Paths{
				"/path": {
					{
						Time: time.Now(),

						Rules: []string{
							"eicar",
						},
					},
				},
			},
		}

		want = &Meta{
			Time: time.Now(),

			Rules: []string{
				"shell",
			},
		}

		update = &History{
			Target: "fs",

			Paths: Paths{
				"/path": {
					want,
				},
			},
		}
	)

	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	if err := hits.Save(db); err != nil {
		t.Fatalf("save error: %v", err)
	}

	if err := update.Save(db); err != nil {
		t.Fatalf("save error: %v", err)
	}

	result, err := SelectLast(db, "/path")
	if err != nil {
		t.Fatalf("select last hit error: %v", err)
	}

	if !reflect.DeepEqual(result.Rules, want.Rules) {
		t.Errorf("unexpected select last hit result %v, want %v", result.Rules, want.Rules)
	}
}

func TestSelectLastHitNone(t *testing.T) {
	input, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	if _, err = SelectLast(input, "/path"); err != nil {
		t.Errorf("unexpected select last hit error: %v", err)
	}
}

func TestSelectLastHitErrs(t *testing.T) {
	input, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock error: %s", err)
	}

	if _, err := SelectLast(input, "../path"); !errors.Is(err, fsys.ErrPathNotAbs) {
		t.Errorf("select last hit error: %v", err)
	}
}

func TestRestore(t *testing.T) {
	var (
		origDir       = t.TempDir()
		quarantineDir = t.TempDir()
		src           = filepath.Join(origDir, t.Name())
	)

	if err := os.MkdirAll(filepath.Join(quarantineDir, origDir), 0750); err != nil {
		t.Errorf("quarantine dir create error: %v", err)
	}

	dst, err := os.Create(filepath.Join(quarantineDir, origDir, t.Name()+"-quarantined"))
	if err != nil {
		t.Errorf("file create error: %v", err)
	}

	hit := &History{
		Paths: Paths{
			src: {
				{
					Status: filepath.Base(dst.Name()),

					Attr: &fsys.Attr{
						UID: os.Getuid(),
						GID: os.Getgid(),
					},
				},
			},
		},
	}

	if err := hit.Paths[src][0].Restore(quarantineDir, src); err != nil {
		t.Errorf("restore error: %v", err)
	}
}

func TestRestoreErrs(t *testing.T) {
	input := &Meta{
		Status: "/path",
		Attr:   &fsys.Attr{},
	}

	if err := input.Restore("../", filepath.Join(t.TempDir(), "../")); !errors.Is(err, fsys.ErrPathNotAbs) {
		t.Errorf("restore error: %v", err)
	}
}

func TestPathsToSlice(t *testing.T) {
	var (
		dir = t.TempDir()

		input = Paths{
			dir: []*Meta{
				{},
			},
		}

		got = input.ToSlice()

		want = [][]string{
			input[dir][0].ToSlice(dir),
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected paths to slice result %v, want %v", got, want)
	}
}

func TestMetaToSlice(t *testing.T) {
	var (
		dir  = t.TempDir()
		time = time.Now()

		input = &Meta{
			Time: time,
			Attr: &fsys.Attr{
				CTime: time,
				MTime: time,
			},
		}

		got = input.ToSlice(dir)

		want = []string{
			dir,
			time.UTC().Format(tzFmt),
			strings.Join([]string{}, "\n"),
			time.Format(tzFmt),
			time.Format(tzFmt),
			strconv.FormatUint(uint64(0), 10),
			strconv.FormatUint(uint64(0), 10),
			"",
			strings.Join([]string{}, ","),
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected meta to slice result %v, want %v", got, want)
	}
}

func TestHistoryToSlice(t *testing.T) {
	var (
		dir = t.TempDir()

		input = &History{
			Target: "target",

			Paths: Paths{
				dir: []*Meta{
					{},
				},
			},
		}

		got = input.ToSlice()

		want = [][]string{
			{
				dir,
				strings.Join([]string{}, "\n"),
				strings.Join([]string{}, ",")},
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected history to slice result %v, want %v", got, want)
	}
}
