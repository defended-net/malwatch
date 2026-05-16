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
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/client/s3"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestNewExiler(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %s", err)
	}

	var (
		got = NewExiler(env)

		want = &Exiler{
			verb:          VerbExile,
			secrets:       env.Cfg.Secrets.S3,
			quarantineDir: env.Cfg.Acts.Quarantine.Dir,
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected create exiler result %v, want %v", got, want)
	}
}

func TestExileLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %s", err)
	}

	input := NewExiler(env)

	if got := input.Load(nil); got != nil {
		t.Errorf("exiler load err %v", got)
	}
}

func TestExileDisabled(t *testing.T) {
	var (
		input = &Exiler{
			secrets: &secret.S3{},
		}

		want = acter.ErrDisabled
	)

	if got := input.Load(nil); !errors.Is(got, want) {
		t.Errorf("unexpected exiler load error %v, want %v", got, want)
	}
}

func TestExileNoRegion(t *testing.T) {
	var (
		input = &Exiler{
			secrets: &secret.S3{},
		}

		want = ErrExileNoRegion
	)

	if got := input.Act(nil); !errors.Is(got, want) {
		t.Errorf("unexpected act error %v, want %v", got, want)
	}
}

func TestExileVerb(t *testing.T) {
	var (
		input = &Exiler{
			verb: VerbExile,
		}

		got  = input.Verb()
		want = VerbExile
	)

	if got != want {
		t.Errorf("unexpected verb result %v, want %v", got, want)
	}
}

func TestExile(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %s", err)
	}

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create err %v", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat err %v", err)
	}

	result := state.NewResult("",
		state.Paths{
			file.Name(): hit.NewMeta(
				fsys.NewAttr(stat),

				[]string{
					t.Name(),
				},

				"exile",
			),
		})

	transport, err := s3.New(env.Cfg.Secrets.S3)
	if err != nil {
		t.Fatalf("transport create err %v", err)
	}

	input := &Exiler{
		secrets:   env.Cfg.Secrets.S3,
		transport: transport,
	}

	if err := input.Load(nil); err != nil {
		t.Fatalf("exiler load err %v", err)
	}

	if got := input.Act(result); got != nil {
		t.Errorf("act err %v", got)
	}
}

func TestExileRemoved(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %s", err)
	}

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create err %v", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat err %v", err)
	}

	hit := state.NewResult(
		"",

		state.Paths{
			file.Name(): hit.NewMeta(
				fsys.NewAttr(stat),

				[]string{
					t.Name(),
				},

				[]string{
					"exile",
					"quarantine",
				}...,
			),
		},
	)

	input := &Exiler{
		secrets: env.Cfg.Secrets.S3,
	}

	if err := input.Load(nil); err != nil {
		t.Fatalf("exiler load err %v", err)
	}

	if got := input.Act(hit); got != nil {
		t.Errorf("act err %v", got)
	}
}

func TestExileSingleNoAttr(t *testing.T) {
	var (
		input  = &Exiler{}
		result = state.NewResult("", state.Paths{})
	)

	input.Single(result, t.TempDir(), &hit.Meta{})

	if len(result.Errs()) == 0 {
		t.Errorf("unexpected empty errs")
	}
}

func TestExileSingleOpenErrs(t *testing.T) {
	var (
		input  = &Exiler{}
		result = state.NewResult("", state.Paths{})

		meta = &hit.Meta{
			Attr: &fsys.Attr{},
		}
	)

	input.Single(result, "/dev/null/not-exist", meta)

	if len(result.Errs()) == 0 {
		t.Errorf("unexpected empty errs")
	}
}
