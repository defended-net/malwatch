// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package sig

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/client/git"
)

func TestUpdate(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := Mock(env); err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	if err := Update(env); err != nil {
		t.Errorf("update error: %v", err)
	}
}

func TestInstall(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := Mock(env); err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	if err := os.MkdirAll(env.Paths.Sigs.Tmp, 0750); err != nil {
		t.Fatalf("sig dir make error: %v", err)
	}

	yrSrc := filepath.Join(env.Paths.Sigs.Tmp, "src.yara")

	file, err := os.Create(yrSrc)
	if err != nil {
		t.Fatalf("src file create error: %v", err)
	}
	defer file.Close()

	update := &update{
		srcs: map[string][]string{
			t.Name(): {
				filepath.Base(yrSrc),
			},
		},

		paths: &path.Sigs{
			Tmp: env.Paths.Sigs.Tmp,
			Src: env.Paths.Sigs.Src,
		},
	}

	if err := update.install(nil); err != nil {
		t.Errorf("update error: %v", err)
	}
}

func TestInstallNoSrc(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	update := &update{
		srcs: map[string][]string{
			t.Name(): {"missing.yara"},
		},

		paths: &path.Sigs{
			Tmp: env.Paths.Sigs.Tmp,
			Src: env.Paths.Sigs.Src,
		},
	}

	if err := update.install(nil); err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestCloneErrs(t *testing.T) {
	tests := map[string]struct {
		input *secret.Cfg
		want  error
	}{
		"auth-required": {
			input: &secret.Cfg{
				Git: []*secret.Repo{
					{
						URL: "https://github.com/defended-net/auth",
					},
				},
			},

			want: git.ErrClone,
		},

		"no-repo": {
			input: &secret.Cfg{
				Git: []*secret.Repo{
					{
						URL: "https://github.com/defended-net",
					},
				},
			},

			want: git.ErrClone,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			update := &update{
				paths: &path.Sigs{
					Tmp: t.TempDir(),
				},

				secrets: test.input,
			}

			if err := update.clone(test.input.Git[0]); !errors.Is(err, test.want) {
				t.Errorf("unexpected clone error: %v want %v", err, test.want)
			}
		})
	}
}

func TestMock(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %v", err)
	}

	if err := Mock(env); err != nil {
		t.Errorf("sig mock error: %v", err)
	}
}
