// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package history

import (
	"errors"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/db"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestGet(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  error
	}{
		"empty": {
			input: []string{},

			want: nil,
		},

		"target": {
			input: []string{
				"target",
			},

			want: nil,
		},

		"path": {
			input: []string{
				t.TempDir(),
			},

			want: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env, err := env.Mock(name, t.TempDir())
			if err != nil {
				t.Fatalf("env mock err %s", err)
			}

			if err := db.Load(env); err != nil {
				t.Fatalf("db load err %s", err)
			}

			if got := Get(env, test.input); !errors.Is(got, test.want) {
				t.Errorf("unexpected get result, err %v, want %v", got, test.want)
			}
		})
	}
}

func TestDel(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  error
	}{
		"target": {
			input: []string{
				"history",
			},

			want: nil,
		},

		"path": {
			input: []string{
				t.TempDir(),
			},

			want: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env, err := env.Mock(name, t.TempDir())
			if err != nil {
				t.Fatalf("env mock err %s", err)
			}

			if err := db.Load(env); err != nil {
				t.Fatalf("db load err %s", err)
			}

			if got := Del(env, test.input); !errors.Is(got, test.want) {
				t.Errorf("unexpected del result, err %v, want %v", got, test.want)
			}
		})
	}
}

func TestGetWithHistory(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %s", err)
	}

	if err := db.Load(env); err != nil {
		t.Fatalf("db load err %s", err)
	}

	hist := &hit.History{
		Target: "target",
		Paths: hit.Paths{
			"/target/test.php": {
				{
					Rules:  []string{"eicar"},
					Status: "/quarantine/test.php",
					Attr:   &fsys.Attr{},
				},
			},
		},
	}

	if err := hist.Save(env.Db); err != nil {
		t.Fatalf("save err %v", err)
	}

	if err := Get(env, []string{}); err != nil {
		t.Errorf("get err %v", err)
	}

	if err := Get(env, []string{"/target/test.php"}); err != nil {
		t.Errorf("get path err %v", err)
	}

	if err := Get(env, []string{"target"}); err != nil {
		t.Errorf("get target err %v", err)
	}
}
