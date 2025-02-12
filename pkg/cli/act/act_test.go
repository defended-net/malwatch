// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"reflect"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	cfg "github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

func TestGet(t *testing.T) {
	tests := map[string]struct {
		input []string
	}{
		"path-rule-verb": {
			input: []string{
				"/target/file.php",
			},
		},
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := Get(env, test.input); err != nil {
				t.Fatalf("get error: %s", err)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  *cfg.Cfg
	}{
		"path-rule-verb": {
			input: []string{
				"/target/file.php",
				"rule_name",
				"alert",
			},

			want: &cfg.Cfg{
				Signatures: map[string][]string{},

				Paths: map[string]map[string][]string{
					"/target/file.php": {
						"rule_name": {
							"alert",
						},
					},
				},
			},
		},

		"rule-verb": {
			input: []string{
				"rule_name",
				"alert",
			},

			want: &cfg.Cfg{
				Signatures: map[string][]string{
					"rule_name": {
						"alert",
					},
				},

				Paths: map[string]map[string][]string{},
			},
		},

		"rule-verb-verb": {
			input: []string{
				"rule_name",
				"alert",
				"quarantine",
			},

			want: &cfg.Cfg{
				Signatures: map[string][]string{
					"rule_name": {
						"alert",
						"quarantine",
					},
				},

				Paths: map[string]map[string][]string{},
			},
		},

		"rule-duperule-verb": {
			input: []string{
				"rule_name",
				"alert",
				"alert",
			},

			want: &cfg.Cfg{
				Signatures: map[string][]string{
					"rule_name": {
						"alert",
					},
				},

				Paths: map[string]map[string][]string{},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env, err := env.Mock(t.Name(), t.TempDir())
			if err != nil {
				t.Fatalf("env mock error: %s", err)
			}

			env.Plat = plat.Mock([]acter.Acter{
				act.NewExiler(env),
				act.NewQuarantiner(env),
				act.NewCleaner(env),
				act.NewAlerter(env),
			}...)

			if err := env.Plat.Load(); err != nil {
				t.Fatalf("plat load error: %s", err)
			}

			if err := Set(env, test.input); err != nil {
				t.Fatalf("set error: %s", err)
			}

			if !reflect.DeepEqual(env.Cfg.Acts.Paths, test.want.Paths) {
				t.Errorf("unexpected set path result %v, want %v", env.Cfg.Acts.Paths, test.want.Paths)
			}

			if !reflect.DeepEqual(env.Cfg.Acts.Signatures, test.want.Signatures) {
				t.Errorf("unexpected set rule result %v, want %v", env.Cfg.Acts.Signatures, test.want.Signatures)
			}
		})
	}
}

func TestDel(t *testing.T) {
	tests := map[string]struct {
		input []string
		cfg   *cfg.Cfg
		want  *cfg.Cfg
	}{
		"rule": {
			input: []string{
				"rule_name",
			},

			cfg: &cfg.Cfg{
				Signatures: map[string][]string{
					"rule_name": {
						"alert",
					},
				},

				Paths: map[string]map[string][]string{},
			},

			want: &cfg.Cfg{
				Signatures: map[string][]string{},

				Paths: map[string]map[string][]string{},
			},
		},

		"path": {
			input: []string{
				"/target/file.php",
			},

			cfg: &cfg.Cfg{
				Signatures: map[string][]string{},

				Paths: map[string]map[string][]string{
					"/target/file.php": {
						"rule_name": []string{
							"alert",
						},
					},
				},
			},

			want: &cfg.Cfg{
				Signatures: map[string][]string{},

				Paths: map[string]map[string][]string{},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env, err := env.Mock(t.Name(), t.TempDir())
			if err != nil {
				t.Fatalf("env mock error: %s", err)
			}

			env.Cfg.Acts.Paths = test.cfg.Paths
			env.Cfg.Acts.Signatures = test.cfg.Signatures

			if err := Del(env, test.input); err != nil {
				t.Fatalf("del error: %s", err)
			}

			if !reflect.DeepEqual(env.Cfg.Acts.Paths, test.want.Paths) {
				t.Errorf("unexpected add path result %v, want %v", env.Cfg.Acts.Paths, test.want.Paths)
			}

			if !reflect.DeepEqual(env.Cfg.Acts.Signatures, test.want.Signatures) {
				t.Errorf("unexpected add rule result %v, want %v", env.Cfg.Acts.Signatures, test.want.Signatures)
			}
		})
	}
}

func TestDelErrs(t *testing.T) {
	tests := map[string]struct {
		input []string
	}{
		"path-rule-verb": {
			input: []string{
				"/target/file.php",
				"rule_name",
				"alert",
			},
		},

		"rule-verb": {
			input: []string{
				"rule_name",
				"alert",
			},
		},
	}

	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	//	env.Cfg.Acts.SetPath(env.Paths.Cfg.Acts)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := Del(env, test.input); err == nil {
				t.Fatalf("unexpected del success")
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := map[string]struct {
		input []string
		want  *Parsed
	}{
		"path-rule-verb": {
			input: []string{
				"/target/file.php",
				"rule_name",
				"alert",
			},

			want: &Parsed{
				key:  "/target/file.php",
				path: "/target/file.php",
				rule: "rule_name",

				verbs: []string{
					"alert",
				},
			},
		},

		"rule-wlist": {
			input: []string{
				"/target/file.php",
				"rule_name",
				"",
			},

			want: &Parsed{
				key:  "/target/file.php",
				path: "/target/file.php",
				rule: "rule_name",

				verbs: []string{
					"",
				},
			},
		},

		"path-skip": {
			input: []string{
				"/target/file.php",
			},

			want: &Parsed{
				key:  "/target/file.php",
				path: "/target/file.php",
				rule: "*",

				verbs: []string{
					"",
				},
			},
		},

		"rule-verb": {
			input: []string{
				"rule_name",
				"alert",
			},

			want: &Parsed{
				key:  "rule_name",
				rule: "rule_name",

				verbs: []string{
					"alert",
				},
			},
		},

		"rule-skip": {
			input: []string{
				"rule_name",
			},

			want: &Parsed{
				key:  "rule_name",
				rule: "rule_name",

				verbs: []string{
					"",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := Parse(test.input)

			if !reflect.DeepEqual(result, test.want) {
				t.Errorf("unexpected parse result %v, want %v", result, test.want)
			}
		})
	}
}
