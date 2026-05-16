// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestAlertVerb(t *testing.T) {
	var (
		input = &Alerter{
			verb: VerbAlert,
		}

		want = VerbAlert
	)

	if got := input.Verb(); got != want {
		t.Errorf("unexpected verb result %v, want %v", got, want)
	}
}

func TestActAlert(t *testing.T) {
	tests := map[string]struct {
		input *state.Result
		want  error
	}{
		"single": {
			input: state.NewResult(
				"",

				state.Paths{
					t.Name(): {},
				},
			),

			want: nil,
		},

		"multi": {
			input: state.NewResult(
				"",

				state.Paths{
					t.Name():        {},
					t.Name() + "-b": {},
				},
			),

			want: nil,
		},

		"none": {
			input: state.NewResult(
				"",

				state.Paths{},
			),

			want: nil,
		},
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %v", err)
	}

	alerter := NewAlerter(env)

	for _, alerter := range alerter.senders {
		path := alerter.Cfg().Path()

		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatalf("mkdir err %s", err)
		}

		if _, err := os.Create(path); err != nil {
			t.Fatalf("file create err %s", err)
		}
	}

	if err := alerter.Load(env.Paths.Install.Root); err != nil {
		t.Fatalf("alerter load err %v", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := alerter.Act(test.input); got != nil &&
				!strings.HasSuffix(err.Error(), "connection refused") {
				t.Errorf("act err %v", got)
			}
		})
	}
}

func TestAlerterLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock err %s", err)
	}

	input := NewAlerter(env)

	for _, alerter := range input.senders {
		path := alerter.Cfg().Path()

		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatalf("mkdir err %s", err)
		}

		if _, err := os.Create(path); err != nil {
			t.Fatalf("file create err %s", err)
		}
	}

	if got := input.Load(env.Paths.Install.Root); got != nil {
		t.Errorf("alerter load err %v", got)
	}
}

func TestAlerterLoadErrs(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock err %s", err)
	}

	env.Paths.Alerts.Dir = "/dev/null"

	var (
		input = NewAlerter(env)
		want  = ErrCfgLoad
	)

	if got := input.Load(env.Paths.Install.Root); !errors.Is(got, want) {
		t.Errorf("unexpected alerter load err %s, want %v", got, want)
	}
}
