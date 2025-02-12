// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package history

import (
	"errors"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/db"
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
				t.Fatalf("env mock error: %s", err)
			}

			if err := db.Load(env); err != nil {
				t.Fatalf("db load error: %s", err)
			}

			if result := Get(env, test.input); !errors.Is(result, test.want) {
				t.Errorf("unexpected get result, error: %v, want %v", result, test.want)
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
				t.Fatalf("env mock error: %s", err)
			}

			if err := db.Load(env); err != nil {
				t.Fatalf("db load error: %s", err)
			}

			if result := Del(env, test.input); !errors.Is(result, test.want) {
				t.Errorf("unexpected del result, error: %v, want %v", result, test.want)
			}
		})
	}
}
