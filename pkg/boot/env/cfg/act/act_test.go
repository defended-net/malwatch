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
	input := &Cfg{
		path: t.Name(),
	}

	if input.Path() != t.Name() {
		t.Errorf("unexpected cfg path result %v", input.Path())
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
			result := test.cfg.NewVerbs(test.path, test.sig)

			if !reflect.DeepEqual(result, test.want) {
				t.Errorf("unexpected new verbs result %v, want %v", result, test.want)
			}
		})
	}
}

func TestAddRule(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	tests := map[string]struct {
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
				path: file.Name(),

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
				path: file.Name(),

				Signatures: map[string][]string{
					"eicar": {},
				},

				Paths: map[string]map[string][]string{},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &Cfg{
				path: file.Name(),

				Signatures: map[string][]string{},
				Paths:      map[string]map[string][]string{},
			}

			if err := cfg.AddSigVerbs(acts, test.sig, test.verbs); err != nil {
				t.Fatalf("add sig error: %v", err)
			}

			if !reflect.DeepEqual(cfg, test.want) {
				t.Errorf("unexpected add rule verb result %v, want %v", cfg, test.want)
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

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &Cfg{
				path: file.Name(),

				Signatures: map[string][]string{},
			}

			if err := cfg.AddSigVerbs(acts, test.rule, test.verbs); !errors.Is(err, test.want) {
				t.Errorf("unexpected add sig error: %v, want %v", err, test.want)
			}
		})
	}
}

func TestAddPath(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	tests := map[string]struct {
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

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &Cfg{
				path: file.Name(),

				Paths: map[string]map[string][]string{},
			}

			if err := cfg.AddPathVerbs(acts, test.path, test.sig, test.verbs); err != nil {
				t.Fatalf("add path error: %v", err)
			}

			if !reflect.DeepEqual(cfg.Paths, test.want) {
				t.Errorf("unexpected add path verb result %v, want %v", cfg.Paths, test.want)
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

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &Cfg{
				path: file.Name(),

				Paths: paths,
			}

			if err := cfg.AddPathVerbs(acts, test.path, test.rule, test.verbs); !errors.Is(err, test.want) {
				t.Errorf("unexpected add path verb error: %v, want %v", err, test.want)
			}
		})
	}
}

func TestSetSig(t *testing.T) {
	tests := map[string]struct {
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

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &Cfg{
				path:       file.Name(),
				Signatures: map[string][]string{},
			}

			if err := cfg.SetSigVerbs(acts, test.sig, test.verbs); err != nil {
				t.Fatalf("set sig error: %v", err)
			}

			if !reflect.DeepEqual(cfg.Signatures, test.want) {
				t.Errorf("unexpected set rule result %v, want %v", cfg.Signatures, test.want)
			}
		})
	}
}

func TestSetSigErrs(t *testing.T) {
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
			cfg := &Cfg{}

			if err := cfg.SetSigVerbs([]acter.Acter{}, test.sig, test.verbs); !errors.Is(err, test.want) {
				t.Errorf("unexpected set rule error: %v, want %v", err, test.want)
			}
		})
	}
}

func TestSetPathVerbs(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	tests := map[string]struct {
		input *Cfg
		path  string
		sig   string
		verbs []string
		want  map[string]map[string][]string
	}{
		"alert": {
			input: &Cfg{
				path: file.Name(),

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

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.input.SetPathVerbs(
				acts,
				test.path,
				test.sig,
				test.verbs); err != nil {
				t.Fatalf("add path verbs error: %v", err)
			}

			if !reflect.DeepEqual(test.input.Paths, test.want) {
				t.Errorf("unexpected set path verb result %v, want %v", test.input.Paths, test.want)
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

			if err := cfg.SetPathVerbs([]acter.Acter{}, test.path, test.rule, test.verbs); !errors.Is(err, test.want) {
				t.Errorf("unexpected set path verb error: %v, want %v", err, test.want)
			}
		})
	}
}

func TestDelSig(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	tests := map[string]struct {
		input *Cfg
		path  string
		sig   string
		want  map[string][]string
	}{
		"alert": {
			input: &Cfg{
				path: file.Name(),

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

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.input.DelSigVerbs(test.sig); err != nil {
				t.Fatalf("del sig error: %v", err)
			}

			if !reflect.DeepEqual(test.input.Signatures, test.want) {
				t.Errorf("unexpected del sig result %v, want %v", test.input, test.want)
			}
		})
	}
}

func TestDelSigErrs(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	cfg := &Cfg{
		path: file.Name(),

		Signatures: map[string][]string{},
		Paths:      map[string]map[string][]string{},
	}

	if err := cfg.DelSigVerbs("eicar"); !errors.Is(err, ErrNoActs) {
		t.Errorf("unexpected del sig error: %v, want %v", err, ErrNoActs)
	}
}

func TestDelPath(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	tests := map[string]struct {
		input *Cfg
		path  string
		sig   string
		want  map[string]map[string][]string
	}{
		"alert": {
			input: &Cfg{
				path: file.Name(),

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

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.input.DelPathVerbs(test.path); err != nil {
				t.Fatalf("del path error: %v", err)
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
			if err := test.input.DelPathVerbs(test.path); !errors.Is(err, test.want) {
				t.Errorf("unexpected del path error: %v, want %v", err, test.want)
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
			if err := test.input.Compact(test.path, test.sig); err != nil {
				t.Fatalf("compact error: %v", err)
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
	)

	got := input.Get("/target/index.php")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected get result %v, want %v", got, want)
	}

}

func TestSkips(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("file create error: %v", err)
	}
	defer file.Close()

	tests := map[string]struct {
		input map[string]map[string][]string
		want  *Skips
	}{
		"skip-file": {
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

		"not-skip-file": {
			input: map[string]map[string][]string{
				file.Name(): {
					"rule": {},
				},
			},

			want: &Skips{
				Files: map[string]struct{}{},
			},
		},

		"not-skip-file-verbs": {
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

		"not-skip-file-composite": {
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

	paths := &path.Paths{
		Install: &path.Install{},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &Cfg{
				Paths:      test.input,
				Quarantine: &Quarantine{},
			}

			result := GetSkips(cfg, paths)

			if !reflect.DeepEqual(result, test.want) {
				t.Errorf("unexpected skip files result %v, want %v", result, test.want)
			}
		})
	}
}

func TestSkipDirs(t *testing.T) {
	dir := t.TempDir()

	tests := map[string]struct {
		input map[string]map[string][]string
		want  *Skips
	}{
		"skip-dir": {
			input: map[string]map[string][]string{
				dir: {
					"*": {},
				},
			},

			want: &Skips{
				Dirs: []string{
					dir,
				},

				Files: map[string]struct{}{},
			},
		},

		"not-skip-dir": {
			input: map[string]map[string][]string{
				dir: {
					"rule": {},
				},
			},

			want: &Skips{
				Files: map[string]struct{}{},
			},
		},

		"not-skip-dir-verbs": {
			input: map[string]map[string][]string{
				dir: {
					"rule": []string{
						verbs[0],
					},
				},
			},

			want: &Skips{
				Files: map[string]struct{}{},
			},
		},

		"not-skip-dir-composite": {
			input: map[string]map[string][]string{
				dir: {
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

	paths := &path.Paths{
		Install: &path.Install{},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &Cfg{
				Paths:      test.input,
				Quarantine: &Quarantine{},
			}

			result := GetSkips(cfg, paths)

			if !reflect.DeepEqual(result, test.want) {
				t.Errorf("unexpected skip dirs result %v, want %v", result, test.want)
			}
		})
	}
}

func TestSkipsErrs(t *testing.T) {
	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Errorf("file create error: %v", err)
	}
	defer file.Close()

	cfg := &Cfg{
		Paths: map[string]map[string][]string{
			file.Name(): {
				"*": []string{},
			},
		},

		Quarantine: &Quarantine{},
	}

	paths := &path.Paths{
		Install: &path.Install{},
	}

	GetSkips(cfg, paths)
}

func TestMock(t *testing.T) {
	Mock(t.TempDir())
}
