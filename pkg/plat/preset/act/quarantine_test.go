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
		t.Fatalf("env mock err %s", err)
	}

	var (
		got = NewQuarantiner(env)

		want = &Quarantiner{
			verb: VerbQuarantine,
			dir:  env.Cfg.Acts.Quarantine.Dir,
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected create quarantiner result %v, want %v", got, want)
	}
}

func TestQuarantineLoad(t *testing.T) {
	input := &Quarantiner{
		dir: t.TempDir(),
	}

	if got := input.Load(nil); got != nil {
		t.Errorf("load err %v", got)
	}
}

func TestQuarantineVerb(t *testing.T) {
	var (
		input = &Quarantiner{
			verb: VerbQuarantine,
		}

		want = VerbQuarantine
	)

	if got := input.Verb(); got != want {
		t.Errorf("unexpected verb result %v, want %v", got, want)
	}
}

func TestQuarantine(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create err %v", err)
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %v", err)
	}

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat err %v", err)
	}

	var (
		hit = state.NewResult(
			"",

			state.Paths{
				file.Name(): hit.NewMeta(
					fsys.NewAttr(stat),

					[]string{
						t.Name(),
					},

					"quarantine"),
			},
		)

		acter = &Quarantiner{
			dir: env.Cfg.Acts.Quarantine.Dir,
		}
	)

	if err := acter.Load(nil); err != nil {
		t.Fatalf("quarantiner load err %v", err)
	}

	if got := acter.Act(hit); got != nil {
		t.Errorf("act err %v", got)
	}
}

func TestQuarantineDisabled(t *testing.T) {
	var (
		input = &Quarantiner{}
		want  = acter.ErrDisabled
	)

	if got := input.Load(nil); !errors.Is(got, want) {
		t.Errorf("unexpected quarantiner load err %v, want %v", got, want)
	}
}

func TestQuarantineNoDir(t *testing.T) {
	var (
		input = &Quarantiner{}
		want  = ErrQuarantineNoDir
	)

	if got := input.Act(nil); !errors.Is(got, want) {
		t.Errorf("unexpected act err %v, want %v", got, want)
	}
}

func TestQuarantineErrs(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %v", err)
	}

	tests := map[string]struct {
		input string
		want  error
	}{
		"fs-err": {
			input: "/dev/null/err",
			want:  ErrQuarantineMv,
		},

		"rel": {
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

			var (
				result = state.NewResult(
					"",

					state.Paths{
						test.input: meta,
					},
				)

				input = Quarantiner{
					dir: env.Cfg.Acts.Quarantine.Dir,
				}
			)

			if err := input.Load(nil); err != nil {
				t.Fatalf("load err %v", err)
			}

			if err := input.Act(result); err != nil {
				t.Fatalf("act err %v", err)
			}

			if got := result.Errs()[0]; !errors.Is(got, test.want) {
				t.Errorf("unexpected act err %v, want %v", got, test.want)
			}
		})
	}
}
