// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"golang.org/x/sys/unix"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/db"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestNewQuarantiner(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	var (
		got = NewQuarantiner(env)

		want = &Quarantiner{
			verb: VerbQuarantine,
			dir:  env.Cfg.Acts.Quarantine.Dir,
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected quarantine result %v, want %v", got, want)
	}
}

func TestQuarantineLoad(t *testing.T) {
	input := &Quarantiner{
		dir: t.TempDir(),
	}

	if err := input.Load(); err != nil {
		t.Errorf("quarantiner load error: %v", err)
	}
}

func TestQuarantineVerb(t *testing.T) {
	input := &Quarantiner{
		verb: VerbQuarantine,
	}

	if got := input.Verb(); got != VerbQuarantine {
		t.Errorf("unexpected verb result %v, want %v", got, VerbQuarantine)
	}
}

func TestQuarantine(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load error: %v", err)
	}

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat error: %v", err)
	}

	hit := state.NewResult("",
		state.Paths{
			file.Name(): hit.NewMeta(
				fsys.NewAttr(stat),

				[]string{
					t.Name(),
				},

				"quarantine"),
		})

	acter := &Quarantiner{
		dir: env.Cfg.Acts.Quarantine.Dir,
	}

	if err := acter.Act(hit); err != nil {
		t.Errorf("quarantine error: %v", err)
	}
}

func TestQuarantineDisabled(t *testing.T) {
	var (
		input = &Quarantiner{}
		want  = acter.ErrDisabled
	)

	if got := input.Load(); !errors.Is(got, want) {
		t.Errorf("unexpected quarantiner load error %v, want %v", got, want)
	}
}

func TestQuarantineNoDir(t *testing.T) {
	var (
		input = &Quarantiner{}
		want  = ErrQuarantineNoDir
	)

	if err := input.Act(nil); !errors.Is(err, want) {
		t.Errorf("unexpected quarantine error %v, want %v", err, want)
	}
}

func TestQuarantineErrs(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("tmp db create error: %v", err)
	}

	tests := map[string]struct {
		input string
		want  error
	}{
		"fs-err": {
			input: "/dev/null/err",
			want:  ErrQuarantineMv,
		},

		"relative": {
			input: "dev/null/err",
			want:  ErrQuarantineMv,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			meta := &hit.Meta{
				Status: test.input,

				Attr: &fsys.Attr{},

				Acts: []string{
					"quarantine",
				},
			}

			result := state.NewResult("",
				state.Paths{
					test.input: meta,
				})

			acter := Quarantiner{dir: env.Cfg.Acts.Quarantine.Dir}

			if err := acter.Act(result); err != nil {
				t.Fatalf("quarantine error %v", err)
			}

			err := result.Errs()[0]

			if !errors.Is(err, test.want) {
				t.Errorf("unexpected quarantine error %v, want %v", err, test.want)
			}
		})
	}
}
