// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
)

type input struct {
	err error
	fn  fn
}

type want struct {
	err  error
	fn   fn
	help error
}

var mock = input{
	err: fmt.Errorf("mock"),
	fn:  fn(nil),
}

func Mock(min int, sub Sub) *Cmd {
	return &Cmd{
		Help: mock.err,
		Min:  min,
		Sub:  sub,
		Fn:   mock.fn,
	}
}

func TestRoute(t *testing.T) {
	tests := map[string]struct {
		input []string
		sub   Sub
		want  *want
	}{
		"none": {
			input: []string{
				"malwatch",
			},

			sub: Sub{},

			want: &want{
				err: ErrArgNone,
			},
		},

		"single": {
			input: []string{
				"malwatch",
				"mock",
			},

			sub: Sub{
				"mock": {
					Help: mock.err,
					Fn:   mock.fn,
				},
			},

			want: &want{
				err:  nil,
				fn:   mock.fn,
				help: mock.err,
			},
		},

		"single-under": {
			input: []string{
				"malwatch",
				"mock",
			},

			sub: Sub{
				"mock": Mock(
					1,

					Sub{},
				),
			},

			want: &want{
				err:  mock.err,
				fn:   mock.fn,
				help: mock.err,
			},
		},

		"nested": {
			input: []string{
				"malwatch",
				"mock-a",
				"mock-b",
			},

			sub: Sub{
				"mock-a": Mock(
					1,

					Sub{
						"mock-b": {
							Help: mock.err,
							Min:  0,
							Fn:   mock.fn,
						},
					},
				),
			},

			want: &want{
				err:  nil,
				fn:   mock.fn,
				help: mock.err,
			},
		},

		"nested-under": {
			input: []string{
				"malwatch",
				"mock-a",
				"mock-b",
			},

			sub: Sub{
				"mock-a": Mock(
					1,

					Sub{
						"mock-b": {
							Help: mock.err,
							Min:  1,
							Fn:   mock.fn,
						},
					},
				),
			},

			want: &want{
				err:  mock.err,
				fn:   mock.fn,
				help: mock.err,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Args = test.input
			flag.Parse()

			result, err := test.sub.Route()
			if !errors.Is(err, test.want.err) {
				t.Errorf("unexpected route error: %v, want %v", err, test.want.err)
			}

			if result == nil {
				return
			}

			if fmt.Sprintf("%v", result.Fn) != fmt.Sprintf("%v", test.want.fn) {
				t.Errorf("unexpected route fn result %v, want %v", result.Fn, test.want.fn)
			}

			if !errors.Is(result.Help, test.want.help) {
				t.Errorf("unexpected route help result %v, want %v", result.Help, test.want.help)
			}
		})
	}
}

func TestRouteErrs(t *testing.T) {
	sub := Sub{
		"actions": Mock(
			1,

			Sub{
				"get": {
					Help: mock.err,
					Min:  1,
					Fn:   mock.fn,
				},
			},
		),
	}

	os.Args = []string{
		"malwatch",
		t.Name(),
	}

	flag.Parse()

	if _, err := sub.Route(); !errors.Is(err, ErrArgInvalid) {
		t.Errorf("unexpected route error %v, want %v", err, ErrArgInvalid)
	}
}

func TestRun(t *testing.T) {
	input := Sub{
		"install": {
			Help: mock.err,

			Fn: func(*env.Env, []string) error {
				return nil
			},
		},
	}

	os.Args = []string{
		"malwatch",
		"install",
	}

	flag.Parse()

	if _, err := Run(input); err != nil {
		t.Errorf("run error: %v", err)
	}
}

func TestPrint(t *testing.T) {
	input := Sub{
		"1": &Cmd{
			Help: errors.New("desc-1"),
		},

		"2": &Cmd{
			Help: errors.New("desc-2"),
		},
	}

	if err := input.Print(); err != nil {
		t.Errorf("print error: %v", err)
	}
}
