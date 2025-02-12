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
		t.Fatalf("env mock error: %v", err)
	}

	if _, err = os.Create(filepath.Join(env.Paths.Sigs.Dir, t.Name())); err != nil {
		t.Fatalf("yr src write error: %v", err)
	}

	if _, err = os.Create(filepath.Join(env.Paths.Sigs.Dir, "index.yr")); err != nil {
		t.Fatalf("yr index write error: %v", err)
	}

	if err := Refresh(env); err != nil {
		t.Errorf("refresh error: %v", err)
	}
}

func TestRefreshErrs(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	env.Paths.Sigs.Src = filepath.Join("/dev/null", t.Name())

	if err := Refresh(env); !errors.Is(err, fsys.ErrWalk) {
		t.Errorf("unexpected refresh error: %v, want %v", err, fsys.ErrWalk)
	}
}

func TestCompileErrs(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	env.Paths.Sigs.Idx = filepath.Join("/dev/null", t.Name())

	update := &update{
		paths: &path.Sigs{
			Tmp: t.TempDir(),
		},
	}

	if err := update.compile(); !errors.Is(err, fsys.ErrFileCreate) {
		t.Errorf("unexpected compile error: %v, want %v", err, fsys.ErrFileCreate)
	}
}
