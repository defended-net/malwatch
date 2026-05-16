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
		t.Fatalf("file write err %v", err)
	}

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat err %v", err)
	}

	var (
		want = &Meta{
			Rules: []string{"eicar"},
			Attr:  fsys.NewAttr(stat),
			Acts:  []string{"alert"},
		}

		got = NewMeta(fsys.NewAttr(stat), []string{"eicar"}, "alert")
	)

	want.Time = got.Time

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected save file hit result %v, want %v", got, want)
	}
}

func TestSaveZeroLength(t *testing.T) {
	input := &History{
		Target: "target",
		Paths:  Paths{},
	}

	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	if got := input.Save(db); got != nil {
		t.Errorf("save err %v", got)
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
		t.Fatalf("db mock err %s", err)
	}

	if err := input.Save(db); err != nil {
		t.Fatalf("save err %v", err)
	}

	got, err := SelectAll(db)
	if err != nil {
		t.Fatalf("select all err %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected select all result %v, want %v", got, want)
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
		t.Fatalf("db mock err %s", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, hit := range test.input {
				if err := hit.Save(db); err != nil {
					t.Fatalf("db save err %v", err)
				}
			}

			got, err := SelectTarget(db, name)
			if err != nil {
				t.Fatalf("select target err %v", err)
			}

			if len(got) != test.want {
				t.Errorf("unexpected select target result %v, want %v", len(got), test.want)
			}
		})
	}
}

func TestSelectAllNoDb(t *testing.T) {
	if _, got := SelectAll(nil); got == nil {
		t.Errorf("unexpected select all success")
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
			want:  fsys.ErrPathTraverse,
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
			if got := fsys.HasDotDots(test.input); !errors.Is(got, test.want) {
				t.Errorf("unexpected has dot dots result %v, want %v", got, test.want)
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
		t.Fatalf("db mock err %s", err)
	}

	for _, hit := range input {
		if got := hit.Save(db); got != nil {
			t.Errorf("update err %v", got)
		}
	}
}

func TestDelTarget(t *testing.T) {
	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	hits := &History{
		Target: "target",

		Paths: Paths{
			t.TempDir(): {},
		},
	}

	if err := hits.Save(db); err != nil {
		t.Errorf("save err %v", err)
	}

	if got := DelTarget(db, "target"); got != nil {
		t.Errorf("del err %v", got)
	}
}

func TestDelPath(t *testing.T) {
	var (
		tmp = t.TempDir()

		hits = &History{
			Target: "target",

			Paths: Paths{
				tmp: {},
			},
		}
	)

	db, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	if err := hits.Save(db); err != nil {
		t.Fatalf("save err %v", err)
	}

	if got := DelPath(db, "target", tmp); got != nil {
		t.Errorf("del err %v", got)
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
		t.Fatalf("db mock err %s", err)
	}

	if err := hits.Save(db); err != nil {
		t.Fatalf("save err %v", err)
	}

	if err := update.Save(db); err != nil {
		t.Fatalf("save err %v", err)
	}

	got, err := SelectLast(db, "/path")
	if err != nil {
		t.Fatalf("select last hit err %v", err)
	}

	if !reflect.DeepEqual(got.Rules, want.Rules) {
		t.Errorf("unexpected select last hit result %v, want %v", got.Rules, want.Rules)
	}
}

func TestSelectLastHitNone(t *testing.T) {
	input, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	if _, got := SelectLast(input, "/path"); got != nil {
		t.Errorf("unexpected select last hit err %v", got)
	}
}

func TestSelectLastHitErrs(t *testing.T) {
	want := fsys.ErrPathNotAbs

	input, err := orm.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("db mock err %s", err)
	}

	if _, got := SelectLast(input, "../path"); !errors.Is(got, want) {
		t.Errorf("unexpected select last hit err %v, want %v", got, want)
	}
}

func TestRestore(t *testing.T) {
	var (
		srcDir        = t.TempDir()
		srcPath       = filepath.Join(srcDir, t.Name())
		quarantineDir = t.TempDir()
	)

	if err := os.MkdirAll(filepath.Join(quarantineDir, srcDir), 0700); err != nil {
		t.Errorf("quarantine dir create err %v", err)
	}

	dst, err := os.Create(filepath.Join(quarantineDir, srcDir, t.Name()+"-quarantined"))
	if err != nil {
		t.Errorf("file create err %v", err)
	}

	hit := &History{
		Paths: Paths{
			srcPath: {
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

	if got := hit.Paths[srcPath][0].Restore(quarantineDir, srcPath); got != nil {
		t.Errorf("restore err %v", got)
	}
}

func TestRestoreErrs(t *testing.T) {
	var (
		input = &Meta{
			Status: "/path",
			Attr:   &fsys.Attr{},
		}

		want = fsys.ErrPathTraverse
	)

	if got := input.Restore("../", filepath.Join(t.TempDir(), "../")); !errors.Is(got, want) {
		t.Errorf("unexpected restore err %v, want % v", got, want)
	}
}

func TestPathsToSlice(t *testing.T) {
	var (
		tmp = t.TempDir()

		input = Paths{
			tmp: []*Meta{
				{},
			},
		}

		got = input.ToSlice()

		want = [][]string{
			input[tmp][0].ToSlice(tmp),
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected paths to slice result %v, want %v", got, want)
	}
}

func TestMetaToSlice(t *testing.T) {
	var (
		tmp  = t.TempDir()
		time = time.Now()

		input = &Meta{
			Time: time,
			Attr: &fsys.Attr{
				CTime: time,
				MTime: time,
			},
		}

		got = input.ToSlice(tmp)

		want = []string{
			tmp,
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
		tmp = t.TempDir()

		input = &History{
			Target: "target",

			Paths: Paths{
				tmp: []*Meta{
					{},
				},
			},
		}

		got = input.ToSlice()

		want = [][]string{
			{
				tmp,
				strings.Join([]string{}, "\n"),
				strings.Join([]string{}, ",")},
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected history to slice result %v, want %v", got, want)
	}
}
