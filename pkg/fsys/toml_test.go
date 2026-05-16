// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsys

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestReadTOML(t *testing.T) {
	type cfg struct {
		Example struct {
			Key string `toml:"key"`
		} `toml:"example"`

		Key string `toml:"key"`
	}

	tests := map[string]struct {
		input string
		path  func(dir, name string) string
		want  func(got cfg) string
	}{
		"rel": {
			input: `[example]
  key = "value"`,

			path: func(_, name string) string {
				return name
			},

			want: func(got cfg) string {
				return got.Example.Key
			},
		},

		"abs": {
			input: `key = "value"`,

			path: func(dir, name string) string {
				return filepath.Join(dir, name)
			},

			want: func(got cfg) string {
				return got.Key
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				tmp  = t.TempDir()
				file = "cfg.toml"
				path = filepath.Join(tmp, file)
				got  cfg
				want = "value"
			)

			if err := os.WriteFile(path, []byte(test.input), 0600); err != nil {
				t.Fatalf("file write err %v", err)
			}

			root, err := os.OpenRoot(tmp)
			if err != nil {
				t.Fatalf("open root err %v", err)
			}

			t.Cleanup(func() {
				// lint
				_ = root.Close()
			})

			if err := ReadTOML(root, test.path(tmp, file), &got); err != nil {
				t.Fatalf("read toml err %v", err)
			}

			if test.want(got) != want {
				t.Errorf("unexpected read toml result %v, want %v", test.want(got), want)
			}
		})
	}
}

func TestReadTOMLErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) (root *os.Root, name string)
		want error
	}{
		"nil-root": {
			fn: func(t *testing.T) (*os.Root, string) {
				return nil, t.Name()
			},

			want: ErrPathRoot,
		},

		"not-local": {
			fn: func(t *testing.T) (*os.Root, string) {
				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, "../escape"
			},

			want: ErrPathLocal,
		},

		"missing": {
			fn: func(t *testing.T) (*os.Root, string) {
				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, "missing"
			},

			want: ErrTOMLRead,
		},

		"decode": {
			fn: func(t *testing.T) (*os.Root, string) {
				var (
					tmp  = t.TempDir()
					name = "cfg.toml"
				)

				if err := os.WriteFile(filepath.Join(tmp, name), []byte("not = valid = toml"), 0600); err != nil {
					t.Fatalf("file write err %v", err)
				}

				root, err := os.OpenRoot(tmp)
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, name
			},

			want: ErrTOMLRead,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			root, path := test.fn(t)

			if got := ReadTOML(root, path, &struct{}{}); !errors.Is(got, test.want) {
				t.Errorf("unexpected read toml err %v, want %v", got, test.want)
			}
		})
	}
}

func TestInstallTOML(t *testing.T) {
	tests := map[string]struct {
		fn func(t *testing.T, dir, name string) (path string, want []byte)
	}{
		"basic": {
			fn: nil,
		},

		"disabled-empty": {
			fn: func(t *testing.T, dir, name string) (string, []byte) {
				path := filepath.Join(dir, name+".disabled")

				file, err := os.Create(path)
				if err != nil {
					t.Fatalf("file create err %v", err)
				}

				defer func() {
					// lint
					_ = file.Close()
				}()

				return path, nil
			},
		},

		"disabled-exist": {
			fn: func(t *testing.T, dir, name string) (string, []byte) {
				var (
					path = filepath.Join(dir, name+".disabled")
					want = []byte("preserved = true\n")
				)

				if err := os.WriteFile(path, want, 0600); err != nil {
					t.Fatalf("file write err %v", err)
				}

				return path, want
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				dir     = t.TempDir()
				file    = "cfg.toml"
				prePath string
				preWant []byte
			)

			if test.fn != nil {
				prePath, preWant = test.fn(t, dir, file)
			}

			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatalf("open root err %v", err)
			}

			t.Cleanup(func() {
				// lint
				_ = root.Close()
			})

			if err := InstallTOML(root, file, struct{}{}); err != nil {
				t.Errorf("install toml err %v", err)
			}

			if preWant == nil {
				return
			}

			got, err := os.ReadFile(prePath)
			if err != nil {
				t.Fatalf("file read err %v", err)
			}

			if string(got) != string(preWant) {
				t.Errorf("file truncated %q, want %q", got, preWant)
			}
		})
	}
}

func TestInstallTOMLErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) (root *os.Root, name string)
		want error
	}{
		"nil-root": {
			fn: func(t *testing.T) (*os.Root, string) {
				return nil, t.Name()
			},

			want: ErrPathRoot,
		},

		"not-local": {
			fn: func(t *testing.T) (*os.Root, string) {
				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, "../escape"
			},

			want: ErrPathLocal,
		},

		"exist": {
			fn: func(t *testing.T) (*os.Root, string) {
				var (
					tmp  = t.TempDir()
					name = "cfg.toml"
				)

				file, err := os.Create(filepath.Join(tmp, name))
				if err != nil {
					t.Fatalf("file create err %v", err)
				}

				defer func() {
					// lint
					_ = file.Close()
				}()

				root, err := os.OpenRoot(tmp)
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, name
			},

			want: fs.ErrExist,
		},

		"decode": {
			fn: func(t *testing.T) (*os.Root, string) {
				var (
					tmp   = t.TempDir()
					name  = "cfg.toml"
					input = []byte("not = valid = toml")
				)

				if err := os.WriteFile(filepath.Join(tmp, name), input, 0600); err != nil {
					t.Fatalf("file write err %v", err)
				}

				root, err := os.OpenRoot(tmp)
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, name
			},

			want: ErrTOMLRead,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			root, path := test.fn(t)

			if got := InstallTOML(root, path, &struct{}{}); !errors.Is(got, test.want) {
				t.Errorf("unexpected install toml err %v, want %v", got, test.want)
			}
		})
	}
}

func TestWriteTOML(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	t.Cleanup(func() {
		// lint
		_ = root.Close()
	})

	if got := WriteTOML(root, t.Name(), struct{}{}); got != nil {
		t.Errorf("write toml err %v", got)
	}
}

func TestWriteTOMLErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) (root *os.Root, name string)
		want error
	}{
		"nil-root": {
			fn: func(t *testing.T) (*os.Root, string) {
				return nil, t.Name()
			},

			want: ErrPathRoot,
		},

		"not-local": {
			fn: func(t *testing.T) (*os.Root, string) {
				root, err := os.OpenRoot(t.TempDir())
				if err != nil {
					t.Fatalf("open root err %v", err)
				}

				t.Cleanup(func() {
					// lint
					_ = root.Close()
				})

				return root, "../escape"
			},

			want: ErrPathLocal,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			root, path := test.fn(t)

			if got := WriteTOML(root, path, struct{}{}); !errors.Is(got, test.want) {
				t.Errorf("unexpected write toml err %v, want %v", got, test.want)
			}
		})
	}
}
