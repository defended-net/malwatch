// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import (
	"errors"
	"log/slog"
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
		t.Fatalf("env mock error: %v", err)
	}

	if err := Load(env); err != nil {
		t.Errorf("logger load error: %v", err)
	}
}

func TestLoadErrs(t *testing.T) {
	env := &env.Env{
		Paths: &path.Paths{
			Install: &path.Install{
				Log: "/dev/null/log",
			},
		},

		Cfg: &base.Cfg{
			Log: &logger.Cfg{
				Dir: "/dev/null/test.cfg",
			},
		},
	}

	if err := Load(env); !errors.Is(err, fsys.ErrDirCreate) {
		t.Errorf("logger load error: %v", err)
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
				Value: slog.StringValue("/dev/null/src.go"),
			},

			want: slog.Attr{
				Key:   slog.SourceKey,
				Value: slog.StringValue("src.go"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			attr := rewriteAttrs(nil, test.input)

			if !reflect.DeepEqual(attr, test.want) {
				t.Errorf("unexpected logger attr: %v, want %v", test.input, test.want)
			}
		})
	}
}
