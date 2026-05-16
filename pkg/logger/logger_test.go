// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/base"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/logger"
	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	if got := Load(env); got != nil {
		t.Errorf("logger load err %v", got)
	}
}

func TestLoadErrs(t *testing.T) {
	var (
		tmp     = t.TempDir()
		logPath = filepath.Join(tmp, t.Name())
		blocker = filepath.Join(tmp, "blocker")
	)

	if err := os.MkdirAll(logPath, 0700); err != nil {
		t.Fatalf("mkdir err %v", err)
	}

	if _, err := os.Create(blocker); err != nil {
		t.Fatalf("file create err %v", err)
	}

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	tests := map[string]struct {
		env  *env.Env
		want error
	}{
		"dir": {
			env: &env.Env{
				Opts: &env.Opts{},

				Paths: &path.Paths{
					Install: &path.Install{
						Log:  filepath.Join(blocker, "child", "log"),
						Root: root,
					},
				},

				Cfg: &base.Cfg{
					Log: &logger.Cfg{},
				},
			},

			want: fsys.ErrDirCreate,
		},

		"file": {
			env: &env.Env{
				Opts: &env.Opts{},

				Paths: &path.Paths{
					Install: &path.Install{
						Log: logPath,
					},
				},

				Cfg: &base.Cfg{
					Log: &logger.Cfg{},
				},
			},

			want: fsys.ErrPathRoot,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Load(test.env); !errors.Is(got, test.want) {
				t.Errorf("unexpected load err %v, want %v", got, test.want)
			}
		})
	}
}

func TestRewriteAttrs(t *testing.T) {
	tests := map[string]struct {
		input slog.Attr
		want  slog.Attr
	}{
		"any": {
			input: slog.Attr{
				Key:   "any",
				Value: slog.Value{},
			},

			want: slog.Attr{
				Key:   "any",
				Value: slog.Value{},
			},
		},

		"none": {
			input: slog.Attr{
				Value: slog.Value{},
			},

			want: slog.Attr{
				Value: slog.Value{},
			},
		},

		"src": {
			input: slog.Attr{
				Key:   slog.SourceKey,
				Value: slog.StringValue("/dev/null/file"),
			},

			want: slog.Attr{
				Key:   slog.SourceKey,
				Value: slog.StringValue("file"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := rewriteAttrs(nil, test.input); !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected logger attr %v, want %v", test.input, test.want)
			}
		})
	}
}
