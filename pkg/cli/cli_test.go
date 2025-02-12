// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"errors"
	"flag"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/defended-net/malwatch/pkg/cli/act"
	"github.com/defended-net/malwatch/pkg/cli/help"
	"github.com/defended-net/malwatch/pkg/cli/info"
	"github.com/defended-net/malwatch/pkg/cli/install"
)

type want struct {
	err  error
	fn   string
	help error
}

func TestUnwrap(t *testing.T) {
	tests := map[string]struct {
		input  []string
		layout Layout
		want   *want
	}{
		"none": {
			input: []string{
				"malwatch",
			},

			layout: map[string]*Cmd{},

			want: &want{
				err: ErrArgMissing,
			},
		},

		"single": {
			input: []string{
				"malwatch",
				"info",
			},

			layout: map[string]*Cmd{
				"info": {
					Help: help.ErrInfo,
					Fn:   info.Do,
				},
			},

			want: &want{
				err:  nil,
				fn:   "info.Do",
				help: help.ErrInfo,
			},
		},

		"single-under": {
			input: []string{
				"malwatch",
				"info",
			},

			layout: map[string]*Cmd{
				"info": {
					Help: help.ErrInfo,
					Min:  1,
					Fn:   info.Do,
				},
			},

			want: &want{
				err:  help.ErrInfo,
				fn:   "info.Do",
				help: help.ErrInfo,
			},
		},

		"nested": {
			input: []string{
				"malwatch",
				"actions",
				"get",
			},

			layout: Layout{
				"actions": {
					Help: help.ErrActs,
					Min:  1,

					Layout: Layout{
						"get": {
							Help: help.ErrActs,
							Min:  0,
							Fn:   act.Get,
						},
					},
				},
			},

			want: &want{
				err:  nil,
				fn:   "act.Get",
				help: help.ErrActs,
			},
		},

		"nested-under": {
			input: []string{
				"malwatch",
				"actions",
				"get",
			},

			layout: Layout{
				"actions": {
					Help: help.ErrActs,
					Min:  1,

					Layout: Layout{
						"get": {
							Help: help.ErrActs,
							Min:  1,
							Fn:   act.Get,
						},
					},
				},
			},

			want: &want{
				err:  help.ErrActs,
				fn:   "act.Get",
				help: help.ErrActs,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			os.Args = test.input
			flag.Parse()

			result, err := test.layout.Unwrap()
			if !errors.Is(err, test.want.err) {
				t.Errorf("unexpected unwrapped error: %v, want %v", err, test.want.err)
			}

			if result == nil {
				return
			}

			fn := runtime.FuncForPC(reflect.ValueOf(result.Fn).Pointer()).Name()

			if !strings.HasSuffix(fn, test.want.fn) {
				t.Errorf("unexpected unwrapped fn result %v, want %v", fn, test.want)
			}

			if !errors.Is(result.Help, test.want.help) {
				t.Errorf("unexpected unwrapped help result %v, want %v", result.Help, test.want.help)
			}
		})
	}
}

func TestUnwrapErrs(t *testing.T) {
	layout := Layout{
		"actions": {
			Help: help.ErrActs,
			Min:  1,

			Layout: Layout{
				"get": {
					Help: help.ErrActs,
					Min:  1,
					Fn:   act.Get,
				},
			},
		},
	}

	os.Args = []string{
		"malwatch",
		t.Name(),
	}

	flag.Parse()

	if _, err := layout.Unwrap(); !errors.Is(err, ErrArgInvalid) {
		t.Errorf("unexpected unwrapped error %v, want %v", err, ErrArgInvalid)
	}
}

func TestRun(t *testing.T) {
	input := Layout{
		"install": {
			Help: help.ErrInstall,
			Fn:   install.Do,
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
	input := Layout{
		"1": &Cmd{
			Help: errors.New("1-desc"),
		},

		"2": &Cmd{
			Help: errors.New("2-desc"),
		},
	}

	if err := input.Print(); err != nil {
		t.Errorf("print error: %v", err)
	}
}
