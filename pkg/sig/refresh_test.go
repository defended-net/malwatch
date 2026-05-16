// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestRefresh(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if _, err = os.Create(filepath.Join(env.Paths.Sigs.Dir, t.Name())); err != nil {
		t.Fatalf("file create err %v", err)
	}

	if _, err = os.Create(filepath.Join(env.Paths.Sigs.Dir, "index.yr")); err != nil {
		t.Fatalf("file create err %v", err)
	}

	if got := Refresh(env); got != nil {
		t.Errorf("refresh err %v", got)
	}
}

func TestRefreshErrs(t *testing.T) {
	want := fsys.ErrWalk

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Paths.Sigs.Src = filepath.Join("/dev/null", t.Name())

	if got := Refresh(env); !errors.Is(got, want) {
		t.Errorf("unexpected refresh err %v, want %v", got, want)
	}
}

func TestCompileErrs(t *testing.T) {
	want := fsys.ErrFileCreate

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Paths.Sigs.Idx = filepath.Join("/dev/null", t.Name())

	input := &update{
		paths: &path.Sigs{
			Tmp: t.TempDir(),
		},
	}

	if got := input.compile(); !errors.Is(got, want) {
		t.Errorf("unexpected compile err %v, want %v", got, want)
	}
}

func TestRefreshErr(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Paths.Sigs.Idx = "/not-exist"

	if got := Refresh(env); got == nil {
		t.Errorf("unexpected refresh success")
	}
}
