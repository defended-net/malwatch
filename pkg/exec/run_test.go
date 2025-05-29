// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package exec

import (
	"errors"
	"strings"
	"testing"
)

type want struct {
	out string
	err error
	has []string
}

func TestRun(t *testing.T) {
	tests := []struct {
		name string
		bin  string
		args []string
		want *want
	}{
		{
			name: "echo-single",
			bin:  "echo",
			args: []string{"malwatch"},

			want: &want{
				out: "malwatch\n",
				err: nil,
			},
		},
		{
			name: "echo-multi",
			bin:  "echo",
			args: []string{
				"mal",
				"watch",
			},

			want: &want{
				out: "mal watch\n",
				err: nil,
			},
		},
		{
			name: "newline",
			bin:  "echo",

			want: &want{
				out: "\n",
				err: nil,
			},
		},
		{
			name: "empty-stdout",
			bin:  "true",

			want: &want{
				err: nil,
			},
		},
		{
			name: "quotes",
			bin:  "sh",
			args: []string{
				"-c",
				"echo 'spaced args'",
			},

			want: &want{
				out: "spaced args\n",
				err: nil,
			},
		},
		{
			name: "status-code",
			bin:  "sh",
			args: []string{
				"-c",
				"exit 0",
			},

			want: &want{
				err: nil,
			},
		},
		{
			name: "status-code-err",
			bin:  "sh",
			args: []string{
				"-c",
				"exit 1",
			},

			want: &want{
				err: ErrRun,
				has: []string{
					"exit status 1",
				},
			},
		},
		{
			name: "metachars-bin",
			bin:  "echo;",
			args: []string{
				"hello",
			},

			want: &want{
				err: ErrMetaChars,
			},
		},
		{
			name: "metachars-args",
			bin:  "echo",
			args: []string{
				`insecure | args`,
			},

			want: &want{
				err: ErrMetaChars,
			},
		},
		{
			name: "not-found",
			bin:  "not-found",

			want: &want{
				err: ErrRun,
				has: []string{
					"executable file not found",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := Run(test.bin, test.args...)

			if test.want.err != nil {
				if err == nil {
					t.Fatalf("expected run error %v, got nil", test.want.err.Error())
					return
				}

				if !errors.Is(err, test.want.err) {
					t.Fatalf("unexpected run error %T %v, got %v", err, err, test.want.err)
				}

				for _, substr := range test.want.has {
					if !strings.Contains(err.Error(), substr) {
						t.Fatalf("missing substring for run error %v %v", err.Error(), substr)
					}
				}

				return
			}

			if err != nil {
				t.Fatalf("run error: %v", err)
				return
			}

			if out != test.want.out {
				t.Errorf("unexpected run output %v, got %v", out, test.want.out)
			}
		})
	}
}
