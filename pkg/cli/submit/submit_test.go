// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package submit

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/db"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestDo(t *testing.T) {
	var (
		svc   = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
		input = filepath.Join(t.TempDir(), t.Name())
	)

	defer svc.Close()

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	env.Cfg.Secrets.Submit.Endpoint = svc.URL

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %s", err)
	}

	file, err := os.Create(input)
	if err != nil {
		t.Fatal("file create err", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	if got := Do(env, []string{input}); got != nil {
		t.Errorf("do err %v", got)
	}
}

func TestDoInvalidPath(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  error
	}{
		"invalid": {
			input: []string{
				"\\",
			},

			want: fsys.ErrPathNotAbs,
		},

		"rel": {
			input: []string{
				"target/index.php",
			},

			want: fsys.ErrPathNotAbs,
		},

		"file": {
			input: []string{
				"index.php",
			},

			want: fsys.ErrPathNotAbs,
		},

		"space": {
			input: []string{
				" ",
			},

			want: fsys.ErrPathNotAbs,
		},

		"none": {
			input: []string{
				"",
			},

			want: fsys.ErrPathNotAbs,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env, err := env.Mock(t.Name(), t.TempDir())
			if err != nil {
				t.Errorf("env mock err %v", err)
			}

			if got := Do(env, test.input); !errors.Is(got, test.want) {
				t.Errorf("unexpected do result %v, want %v", got, test.want)
			}
		})
	}
}

func TestDoErrs(t *testing.T) {
	var (
		input = []string{filepath.Join(t.TempDir(), t.Name())}
		want  = fs.ErrNotExist
	)

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %v", err)
	}

	if got := Do(env, input); !errors.Is(got, want) {
		t.Errorf("unexpected submit err %v, want %v", got, want)
	}
}
