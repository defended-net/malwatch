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
		t.Fatalf("env mock error: %s", err)
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
		t.Errorf("unexpected exiler result %v, want %v", got, want)
	}
}

func TestExileLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	input := NewExiler(env)

	if err := input.Load(); err != nil {
		t.Errorf("exiler load error: %v", err)
	}
}

func TestExileDisabled(t *testing.T) {
	var (
		input = &Exiler{
			secrets: &secret.S3{},
		}

		want = acter.ErrDisabled
	)

	if got := input.Load(); !errors.Is(got, want) {
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

	if err := input.Act(nil); !errors.Is(err, want) {
		t.Errorf("unexpected exile error %v, want %v", err, want)
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
		t.Fatalf("env mock error: %s", err)
	}

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat error: %v", err)
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
		t.Fatalf("transport create error: %v", err)
	}

	exiler := &Exiler{
		secrets:   env.Cfg.Secrets.S3,
		transport: transport,
	}

	if err := exiler.Load(); err != nil {
		t.Fatalf("exiler load error: %v", err)
	}

	if err := exiler.Act(result); err != nil {
		t.Errorf("exiler error: %v", err)
	}
}

func TestExileRemoved(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	stat := &unix.Stat_t{}

	if err := unix.Stat(file.Name(), stat); err != nil {
		t.Fatalf("stat error: %v", err)
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

	exiler := &Exiler{
		secrets: env.Cfg.Secrets.S3,
	}

	if err := exiler.Load(); err != nil {
		t.Fatalf("exiler load error: %v", err)
	}

	if err := exiler.Act(hit); err != nil {
		t.Errorf("exiler error: %v", err)
	}
}
