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
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestUpdate(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	if got := Update(env); got != nil {
		t.Errorf("update err %v", got)
	}
}

func TestInstall(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if err := Mock(env, true); err != nil {
		t.Fatalf("sig mock err %v", err)
	}

	if err := os.MkdirAll(env.Paths.Sigs.Tmp, 0700); err != nil {
		t.Fatalf("mkdir err %v", err)
	}

	yr := filepath.Join(env.Paths.Sigs.Tmp, "src.yr")

	file, err := os.Create(yr)
	if err != nil {
		t.Fatalf("file create err %v", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	input := &update{
		srcs: map[string][]string{
			t.Name(): {
				filepath.Base(yr),
			},
		},

		paths: &path.Sigs{
			Tmp:  env.Paths.Sigs.Tmp,
			Src:  env.Paths.Sigs.Src,
			Root: env.Paths.Install.Root,
		},
	}

	if got := input.install(nil); got != nil {
		t.Errorf("install err %v", got)
	}
}

func TestInstallNoSrc(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	input := &update{
		srcs: map[string][]string{
			t.Name(): {"missing.yr"},
		},

		paths: &path.Sigs{
			Tmp:  env.Paths.Sigs.Tmp,
			Src:  env.Paths.Sigs.Src,
			Root: env.Paths.Install.Root,
		},
	}

	if got := input.install(nil); got == nil {
		t.Errorf("unexpected install success")
	}
}

func TestInstallNilRoot(t *testing.T) {
	var (
		input = &update{
			srcs: map[string][]string{
				t.Name(): {"missing.yr"},
			},

			paths: &path.Sigs{
				Tmp: "/tmp",
				Src: "/tmp",
			},
		}

		want = fsys.ErrPathRoot
	)

	if got := input.install(nil); !errors.Is(got, want) {
		t.Errorf("unexpected install err %v, want %v", got, want)
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
			var (
				dir  = t.TempDir()
				sigs = filepath.Join(dir, "sigs")
				tmp  = filepath.Join(dir, "tmp", "sigs")
			)

			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatalf("open root err %v", err)
			}

			defer func() {
				// lint
				_ = root.Close()
			}()

			if err := os.MkdirAll(tmp, 0700); err != nil {
				t.Fatalf("mkdir err %v", err)
			}

			input := &update{
				paths: &path.Sigs{
					Dir:  sigs,
					Tmp:  tmp,
					Root: root,
				},

				secrets: test.input,
			}

			if got := input.clone(test.input.Git[0]); !errors.Is(got, test.want) {
				t.Errorf("unexpected clone err %v want %v", got, test.want)
			}
		})
	}
}

func TestMock(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if got := Mock(env, true); got != nil {
		t.Errorf("sig mock err %v", got)
	}
}

func TestUpdateNoRepos(t *testing.T) {
	want := ErrNoRepos

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	env.Cfg.Secrets.Git = nil

	if got := Update(env); !errors.Is(got, want) {
		t.Errorf("unexpected update err %v, want %v", got, want)
	}
}

func TestCloneNoRepoOwner(t *testing.T) {
	var (
		repo = &secret.Repo{URL: "/"}

		input = &update{
			paths: &path.Sigs{
				Tmp: t.TempDir(),
			},
		}

		want = ErrNoRepoOwner
	)

	if got := input.clone(repo); !errors.Is(got, want) {
		t.Errorf("unexpected clone err %v, want %v", got, want)
	}
}

func TestWalk(t *testing.T) {
	var (
		tmp  = t.TempDir()
		yr   = filepath.Join(tmp, "rule.yr")
		data = []byte("rule x { condition: true }")

		input = &update{
			paths: &path.Sigs{
				Tmp: tmp,
			},
		}
	)

	if err := os.WriteFile(yr, data, 0600); err != nil {
		t.Fatalf("file write err %v", err)
	}

	if err := input.walk(tmp); err != nil {
		t.Errorf("walk err %v", err)
	}

	if len(input.srcs) == 0 {
		t.Errorf("unexpected empty srcs")
	}
}

func TestWalkErrs(t *testing.T) {
	var (
		dir = filepath.Join(t.TempDir(), "/not-exist")

		input = &update{
			paths: &path.Sigs{},
		}
	)

	if got := input.walk(dir); got == nil {
		t.Errorf("unexpected walk success")
	}
}

func TestInstallParentTraverse(t *testing.T) {
	tmp := t.TempDir()

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	input := &update{
		paths: &path.Sigs{
			Tmp:  filepath.Join(tmp, "tmp"),
			Src:  filepath.Join(tmp, "src"),
			Root: root,
		},

		srcs: map[string][]string{
			"../escape": {"file.yr"},
		},
	}

	if got := input.install(nil); got == nil {
		t.Errorf("unexpected install success")
	}
}

func TestInstallSrcTraverse(t *testing.T) {
	tmp := t.TempDir()

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	input := &update{
		paths: &path.Sigs{
			Tmp:  filepath.Join(tmp),
			Src:  filepath.Join(tmp, "src"),
			Root: root,
		},

		srcs: map[string][]string{
			"sub": {"../escape.yr"},
		},
	}

	if got := input.install(nil); got == nil {
		t.Errorf("unexpected install success")
	}
}
