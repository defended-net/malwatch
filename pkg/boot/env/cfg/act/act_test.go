// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env/path"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
)

var (
	cfg = &Cfg{
		Default: []string{"alert"},

		Paths: map[string]map[string][]string{},
	}

	sigs  = map[string][]string{}
	paths = map[string]map[string][]string{}

	verbs = []string{
		"alert",
		"quarantine",
		"exile",
		"clean",
	}

	acts = []acter.Acter{
		acter.Mock(verbs[0], true),
		acter.Mock(verbs[1], true),
		acter.Mock(verbs[2], true),
		acter.Mock(verbs[3], true),
	}
)

func TestNew(t *testing.T) {
	var (
		want = t.TempDir()
		got  = New(want)
	)

	if got.path != want {
		t.Errorf("unexpected new cfg result %v, want %v", got.path, want)
	}
}

func TestPath(t *testing.T) {
	got := &Cfg{
		path: t.Name(),
	}

	if got.Path() != t.Name() {
		t.Errorf("unexpected cfg path result %v", got.Path())
	}
}

func TestNewVerbs(t *testing.T) {
	tests := map[string]struct {
		cfg  *Cfg
		path string
		sig  string
		want []string
	}{
		"asterisk": {
			cfg: &Cfg{
				Default: cfg.Default,

				Signatures: sigs,

				Paths: map[string]map[string][]string{
					"/target/index.php": {
						"*": {
							verbs[1],
						},
					},
				},
			},

			path: "/target/index.php",
			sig:  "",

			want: []string{
				verbs[1],
			},
		},

		"path-sig-wlist": {
			cfg: &Cfg{
				Paths: map[string]map[string][]string{
					"/target/index.php": {
						"eicar": {},
					},
				},
			},

			path: "/target/index.php",
			sig:  "eicar",

			want: []string{},
		},

		"path-sig": {
			cfg: &Cfg{
				Default: cfg.Default,

				Signatures: sigs,

				Paths: map[string]map[string][]string{
					"/target/index.php": {
						"eicar": {
							verbs[1],
						},
					},
				},
			},

			path: "/target/index.php",
			sig:  "eicar",

			want: []string{
				verbs[1],
			},
		},

		"basepath-rule": {
			cfg: &Cfg{
				Default: cfg.Default,

				Signatures: sigs,

				Paths: map[string]map[string][]string{
					"/target": {
						"eicar": {
							verbs[1],
						},
					},
				},
			},

			path: "/target/index.php",
			sig:  "eicar",

			want: []string{
				verbs[1],
			},
		},

		"basepath-wlist": {
			cfg: &Cfg{
				Paths: map[string]map[string][]string{
					"/target": {
						"eicar": {},
					},
				},
			},

			path: "/target/index.php",
			sig:  "eicar",

			want: []string{},
		},

		"basepath-star-rule": {
			cfg: &Cfg{
				Paths: map[string]map[string][]string{
					"/target": {
						"*": {
							verbs[0],
						},
					},
				},
			},

			path: "/target/index.php",
			sig:  "eicar",

			want: []string{
				verbs[0],
			},
		},

		"sig": {
			cfg: &Cfg{
				Default: cfg.Default,

				Signatures: map[string][]string{
					"eicar": {
						verbs[1],
					},
				},

				Paths: paths,
			},

			path: "/target/index.php",
			sig:  "eicar",

			want: []string{
				verbs[1],
			},
		},

		"sig-wlist": {
			cfg: &Cfg{
				Signatures: map[string][]string{
					"eicar": {},
				},
			},

			sig: "eicar",

			want: []string{},
		},

		"default": {
			cfg: &Cfg{
				Default: cfg.Default,

				Signatures: map[string][]string{
					"none": {
						verbs[1],
					},
				},

				Paths: map[string]map[string][]string{},
			},

			path: "/target/no-match.php",
			sig:  "eicar",

			want: []string{
				verbs[0],
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.cfg.NewVerbs(test.path, test.sig)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected new verbs result %v, want %v", got, test.want)
			}
		})
	}
}

func TestAddRule(t *testing.T) {
	var (
		tmp  = t.TempDir()
		name = t.Name()

		tests = map[string]struct {
			path  string
			sig   string
			verbs []string
			want  *Cfg
		}{
			"alert": {
				sig: "eicar",

				verbs: []string{
					verbs[0],
				},

				want: &Cfg{
					path: name,

					Signatures: map[string][]string{
						"eicar": {
							verbs[0],
						},
					},

					Paths: map[string]map[string][]string{},
				},
			},

			"wlist": {
				sig: "eicar",

				verbs: []string{""},

				want: &Cfg{
					path: name,

					Signatures: map[string][]string{
						"eicar": {},
					},

					Paths: map[string]map[string][]string{},
				},
			},
		}

		root, err = os.OpenRoot(tmp)
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			input := &Cfg{
				path: test.want.path,

				Signatures: map[string][]string{},
				Paths:      map[string]map[string][]string{},
			}

			if got := input.AddSigVerbs(root, acts, test.sig, test.verbs); got != nil {
				t.Fatalf("add sig verbs err %v", got)
			}

			if !reflect.DeepEqual(input, test.want) {
				t.Errorf("unexpected add sig verbs result %v, want %v", input, test.want)
			}
		})
	}
}

func TestAddRuleErrs(t *testing.T) {
	tests := map[string]struct {
		path  string
		rule  string
		verbs []string
		want  error
	}{
		"invalid-char": {
			rule: "eicar%",

			verbs: []string{
				verbs[0],
			},

			want: ErrInvalidChars,
		},

		"no-verbs": {
			rule: "eicar",

			verbs: []string{},

			want: ErrUnknownVerb,
		},

		"star": {
			rule: "*",

			verbs: []string{
				verbs[0],
			},

			want: ErrStarSkipNotAllowed,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				input = &Cfg{
					path:       t.Name(),
					Signatures: map[string][]string{},
				}

				got = input.AddSigVerbs(nil, acts, test.rule, test.verbs)
			)

			if !errors.Is(got, test.want) {
				t.Errorf("unexpected add sig verbs err %v, want %v", got, test.want)
			}
		})
	}
}

func TestAddPath(t *testing.T) {
	var (
		tmp  = t.TempDir()
		name = t.Name()

		tests = map[string]struct {
			path  string
			sig   string
			verbs []string
			want  map[string]map[string][]string
		}{
			"alert": {
				path: "/target/index.php",
				sig:  "eicar",

				verbs: []string{
					verbs[0],
				},

				want: map[string]map[string][]string{
					"/target/index.php": {
						"eicar": []string{
							verbs[0],
						},
					},
				},
			},
		}

		root, err = os.OpenRoot(tmp)
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	for subname, test := range tests {
		t.Run(subname, func(t *testing.T) {
			var (
				input = &Cfg{
					path:  name,
					Paths: map[string]map[string][]string{},
				}

				got = input.AddPathVerbs(root, acts, test.path, test.sig, test.verbs)
			)

			if got != nil {
				t.Fatalf("add path verbs err %v", got)
			}

			if !reflect.DeepEqual(input.Paths, test.want) {
				t.Errorf("unexpected add path verbs result %v, want %v", input.Paths, test.want)
			}
		})
	}
}

func TestAddPathErrs(t *testing.T) {
	tests := map[string]struct {
		path  string
		rule  string
		verbs []string
		want  error
	}{
		"not-abs": {
			path: "target/index.php",

			verbs: []string{
				verbs[0],
			},

			want: fsys.ErrPathNotAbs,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				input = &Cfg{
					path:  t.Name(),
					Paths: paths,
				}

				got = input.AddPathVerbs(nil, acts, test.path, test.rule, test.verbs)
			)

			if !errors.Is(got, test.want) {
				t.Errorf("unexpected add path verbs err %v, want %v", got, test.want)
			}
		})
	}
}

func TestSetSigVerbs(t *testing.T) {
	var (
		tmp  = t.TempDir()
		name = t.Name()

		tests = map[string]struct {
			sig   string
			verbs []string
			want  map[string][]string
		}{
			"alert": {
				sig: "eicar",

				verbs: []string{
					verbs[0],
				},

				want: map[string][]string{
					"eicar": {
						verbs[0],
					},
				},
			},
		}

		root, err = os.OpenRoot(tmp)
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	for subname, test := range tests {
		t.Run(subname, func(t *testing.T) {
			var (
				input = &Cfg{
					path:       name,
					Signatures: map[string][]string{},
				}

				got = input.SetSigVerbs(root, acts, test.sig, test.verbs)
			)

			if got != nil {
				t.Fatalf("set sig verbs err %v", err)
			}

			if !reflect.DeepEqual(input.Signatures, test.want) {
				t.Errorf("unexpected set sig verbs result %v, want %v", input.Signatures, test.want)
			}
		})
	}
}

func TestSetSigVerbsErrs(t *testing.T) {
	tests := map[string]struct {
		sig   string
		verbs []string
		want  error
	}{
		"alert": {
			sig: "eicar%",

			verbs: []string{
				verbs[0],
			},

			want: ErrInvalidChars,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				input = &Cfg{}
				got   = input.SetSigVerbs(nil, []acter.Acter{}, test.sig, test.verbs)
			)

			if !errors.Is(got, test.want) {
				t.Errorf("unexpected set sig verbs err %v, want %v", got, test.want)
			}
		})
	}
}

func TestSetPathVerbs(t *testing.T) {
	var (
		tmp  = t.TempDir()
		name = t.Name()

		tests = map[string]struct {
			input *Cfg
			path  string
			sig   string
			verbs []string
			want  map[string]map[string][]string
		}{
			"alert": {
				input: &Cfg{
					path: name,

					Paths: map[string]map[string][]string{},
				},

				path: "/target/index.php",
				sig:  "eicar",

				verbs: []string{
					verbs[0],
				},

				want: map[string]map[string][]string{
					"/target/index.php": {
						"eicar": []string{
							verbs[0],
						},
					},
				},
			},
		}

		root, err = os.OpenRoot(tmp)
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	for subname, test := range tests {
		t.Run(subname, func(t *testing.T) {
			if got := test.input.SetPathVerbs(
				root,
				acts,
				test.path,
				test.sig,
				test.verbs,
			); got != nil {
				t.Fatalf("set path verbs err %v", got)
			}

			if !reflect.DeepEqual(test.input.Paths, test.want) {
				t.Errorf("unexpected set path verbs result %v, want %v", test.input.Paths, test.want)
			}
		})
	}
}

func TestSetPathVerbsErrs(t *testing.T) {
	tests := map[string]struct {
		path  string
		rule  string
		verbs []string
		want  error
	}{
		"not-abs": {
			path: "target/index.php",

			verbs: []string{
				verbs[0],
			},

			want: fsys.ErrPathNotAbs,
		},

		"no-verbs": {
			path:  "/target/index.php",
			rule:  "eicar%",
			verbs: []string{},
			want:  ErrInvalidChars,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &Cfg{
				Signatures: map[string][]string{},
				Paths:      map[string]map[string][]string{},
			}

			if got := cfg.SetPathVerbs(nil, []acter.Acter{}, test.path, test.rule, test.verbs); !errors.Is(got, test.want) {
				t.Errorf("unexpected set path verb err %v, want %v", got, test.want)
			}
		})
	}
}

func TestDelSig(t *testing.T) {
	var (
		tmp  = t.TempDir()
		name = t.Name()

		tests = map[string]struct {
			input *Cfg
			path  string
			sig   string
			want  map[string][]string
		}{
			"alert": {
				input: &Cfg{
					path: name,

					Signatures: map[string][]string{
						"eicar": {
							verbs[0],
						},
					},
				},

				sig: "eicar",

				want: map[string][]string{},
			},
		}

		root, err = os.OpenRoot(tmp)
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	for subname, test := range tests {
		t.Run(subname, func(t *testing.T) {
			if got := test.input.DelSigVerbs(root, test.sig); got != nil {
				t.Fatalf("del sig err %v", got)
			}

			if !reflect.DeepEqual(test.input.Signatures, test.want) {
				t.Errorf("unexpected del sig result %v, want %v", test.input, test.want)
			}
		})
	}
}

func TestDelSigErrs(t *testing.T) {
	var (
		input = &Cfg{
			path: t.Name(),

			Signatures: map[string][]string{},
			Paths:      map[string]map[string][]string{},
		}

		want = ErrNoActs
	)

	if got := input.DelSigVerbs(nil, "eicar"); !errors.Is(got, want) {
		t.Errorf("unexpected del sig err %v, want %v", got, want)
	}
}

func TestDelPath(t *testing.T) {
	var (
		tmp  = t.TempDir()
		name = t.Name()

		tests = map[string]struct {
			input *Cfg
			path  string
			sig   string
			want  map[string]map[string][]string
		}{
			"alert": {
				input: &Cfg{
					path: name,

					Paths: map[string]map[string][]string{
						"/target/index.php": {
							"eicar": []string{
								verbs[0],
							},
						},
					},
				},

				path: "/target/index.php",

				want: map[string]map[string][]string{},
			},
		}

		root, err = os.OpenRoot(tmp)
	)

	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	for subname, test := range tests {
		t.Run(subname, func(t *testing.T) {
			if got := test.input.DelPathVerbs(root, test.path); got != nil {
				t.Fatalf("del path err %v", got)
			}

			if !reflect.DeepEqual(test.input.Paths, test.want) {
				t.Errorf("unexpected del path result %v, want %v", test.input, test.want)
			}
		})
	}
}

func TestDelPathErrs(t *testing.T) {
	tests := map[string]struct {
		input *Cfg
		path  string
		sig   string
		want  error
	}{
		"alert": {
			input: &Cfg{
				Paths: map[string]map[string][]string{
					"/target/index.php": {
						"eicar": []string{
							verbs[0],
						},
					},
				},
			},

			path: "/target-b/index.php",
			sig:  "eicar",

			want: ErrNoActs,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := test.input.DelPathVerbs(nil, test.path); !errors.Is(got, test.want) {
				t.Errorf("unexpected del path err %v, want %v", got, test.want)
			}
		})
	}
}

func TestCompact(t *testing.T) {
	tests := map[string]struct {
		input *Cfg
		path  string
		sig   string
		want  *Cfg
	}{
		"sigs": {
			input: &Cfg{
				Signatures: map[string][]string{
					"eicar": {
						verbs[0],
						verbs[0],
						verbs[0],
					},
				},

				Paths: map[string]map[string][]string{},
			},

			path: "/target/index.php",
			sig:  "eicar",

			want: &Cfg{
				Signatures: map[string][]string{
					"eicar": {
						verbs[0],
					},
				},

				Paths: map[string]map[string][]string{},
			},
		},

		"paths": {
			input: &Cfg{
				Signatures: map[string][]string{},

				Paths: map[string]map[string][]string{
					"/target/index.php": {
						"eicar": []string{
							verbs[0],
							verbs[0],
							verbs[0],
						},
					},
				},
			},

			path: "/target/index.php",
			sig:  "eicar",

			want: &Cfg{
				Signatures: map[string][]string{},

				Paths: map[string]map[string][]string{
					"/target/index.php": {
						"eicar": []string{
							verbs[0],
						},
					},
				},
			},
		},

		"nils": {
			input: &Cfg{
				Signatures: nil,

				Paths: map[string]map[string][]string{
					"/target/index.php": nil,
				},
			},

			path: "/target/index.php",
			sig:  "eicar",

			want: &Cfg{
				Signatures: nil,

				Paths: map[string]map[string][]string{
					"/target/index.php": nil,
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := test.input.Compact(test.path, test.sig); got != nil {
				t.Fatalf("compact err %v", got)
			}

			if !reflect.DeepEqual(test.input, test.want) {
				t.Errorf("unexpected compact result %v, want %v", test.input, test.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	var (
		input = Cfg{
			Paths: map[string]map[string][]string{
				"/target/index.php": {
					"eicar": {
						"alert",
					},
				},
			},
		}

		want = []*Loadout{
			{
				Rule: "eicar",

				Actions: []string{
					"alert",
				},
			},
		}

		got = input.Get("/target/index.php")
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected get result %v, want %v", got, want)
	}

}

func TestGetSkips(t *testing.T) {
	var (
		file, err = os.Create(filepath.Join(t.TempDir(), t.Name()))

		tests = map[string]struct {
			input map[string]map[string][]string
			want  *Skips
		}{
			"file": {
				input: map[string]map[string][]string{
					file.Name(): {
						"*": []string{},
					},
				},

				want: &Skips{
					Files: map[string]struct{}{
						file.Name(): {},
					},
				},
			},

			"not-file": {
				input: map[string]map[string][]string{
					file.Name(): {
						"rule": {},
					},
				},

				want: &Skips{
					Files: map[string]struct{}{},
				},
			},

			"not-file-verbs": {
				input: map[string]map[string][]string{
					file.Name(): {
						"rule": []string{
							verbs[0],
						},
					},
				},

				want: &Skips{
					Files: map[string]struct{}{},
				},
			},

			"not-file-composite": {
				input: map[string]map[string][]string{
					file.Name(): {
						"*": []string{
							verbs[0],
						},
					},
				},

				want: &Skips{
					Files: map[string]struct{}{},
				},
			},
		}

		paths = &path.Paths{
			Install: &path.Install{},
		}
	)

	if err != nil {
		t.Fatalf("file create err %v", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				cfg = &Cfg{
					Paths:      test.input,
					Quarantine: &Quarantine{},
				}

				got = GetSkips(cfg, paths)
			)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected get skips result %v, want %v", got, test.want)
			}
		})
	}
}

func TestGetSkipsDirs(t *testing.T) {
	var (
		tmp = t.TempDir()

		tests = map[string]struct {
			input map[string]map[string][]string
			want  *Skips
		}{
			"dir": {
				input: map[string]map[string][]string{
					tmp: {
						"*": {},
					},
				},

				want: &Skips{
					Dirs: []string{
						tmp,
					},

					Files: map[string]struct{}{},
				},
			},

			"not-dir": {
				input: map[string]map[string][]string{
					tmp: {
						"rule": {},
					},
				},

				want: &Skips{
					Files: map[string]struct{}{},
				},
			},

			"not-dir-verbs": {
				input: map[string]map[string][]string{
					tmp: {
						"rule": []string{
							verbs[0],
						},
					},
				},

				want: &Skips{
					Files: map[string]struct{}{},
				},
			},

			"not-dir-composite": {
				input: map[string]map[string][]string{
					tmp: {
						"*": []string{
							verbs[0],
						},
					},
				},

				want: &Skips{
					Files: map[string]struct{}{},
				},
			},
		}

		paths = &path.Paths{
			Install: &path.Install{},
		}
	)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				cfg = &Cfg{
					Paths:      test.input,
					Quarantine: &Quarantine{},
				}

				got = GetSkips(cfg, paths)
			)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected get skips result %v, want %v", got, test.want)
			}
		})
	}
}

func TestGetSkipsErrs(t *testing.T) {
	var (
		file, err = os.Create(filepath.Join(t.TempDir(), t.Name()))

		input = &Cfg{
			Paths: map[string]map[string][]string{
				file.Name(): {
					"*": []string{},
				},
			},

			Quarantine: &Quarantine{},
		}

		paths = &path.Paths{
			Install: &path.Install{},
		}
	)

	if err != nil {
		t.Errorf("file create err %v", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	GetSkips(input, paths)
}

func TestHasFile(t *testing.T) {
	var (
		skips = &Skips{
			Files: map[string]struct{}{
				"/usr/bin/echo":  {},
				"/usr/bin/true":  {},
				"/usr/bin/false": {},
			},
		}

		tests = []struct {
			name  string
			input string
			want  bool
		}{
			{
				name:  "match",
				input: "/usr/bin/true",
				want:  true,
			},

			{
				name:  "match-other",
				input: "/usr/bin/false",
				want:  true,
			},

			{
				name:  "miss",
				input: "/etc/file",
				want:  false,
			},

			{
				name:  "empty",
				input: "",
				want:  false,
			},

			{
				name:  "pfx",
				input: "/etc",
				want:  false,
			},
		}
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := skips.HasFile(test.input); got != test.want {
				t.Errorf("unexpected has file result %v, want %v", got, test.want)
			}
		})
	}
}

func TestMock(t *testing.T) {
	Mock(t.TempDir())
}
