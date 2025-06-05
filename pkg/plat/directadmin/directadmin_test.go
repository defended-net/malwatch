// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package directadmin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/plat/preset/act"
)

func TestNew(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("env mock error: %v", err)
	}

	var (
		want = &Plat{
			env: env,

			acters: act.Preset(env),

			cfg: &Cfg{
				path: filepath.Join(env.Paths.Plat.Dir, "directadmin.toml"),
			},

			// https://docs.directadmin.com/directadmin/general-usage/directadmin-binary.html
			bin: "/usr/local/directadmin/directadmin",

			endpoints: &endpoints{
				docroots: &endpoint{
					base: "/CMD_API_DOMAIN",

					params: map[string]string{
						"action": "document_root_all",
					},
				},
			},
		}

		got = New(env)
	)

	if !reflect.DeepEqual(got.cfg, want.cfg) {
		t.Errorf("unexpected cfg %v, want %v", got, want)
	}

	if !reflect.DeepEqual(got.endpoints, want.endpoints) {
		t.Errorf("unexpected endpoints %v, want %v", got, want)
	}
}

func TestLoad(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("mock error: %v", err)
	}

	serve := httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		data := Info{
			Users: map[string]users{
				"one": {
					"one": domains{
						"one.example": &docroot{
							PublicHTML:  "/home/one/domains/one.example/public_html",
							PrivateHTML: "/home/one/domains/one.example/private_html",

							Subdomains: domains{},
						},
					},
				},
			},
		}

		if err := json.NewEncoder(wr).Encode(data); err != nil {
			t.Fatalf("json marshal error: %v", err)
		}
	}))

	plat.url = serve.URL

	if err = plat.Load(); err != nil {
		t.Errorf("load error: %v", err)
	}
}

func TestLoadAuthed(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("mock error: %v", err)
	}

	if err = plat.Load(); err != nil && !strings.Contains(err.Error(), "unsupported protocol scheme") {
		t.Errorf("load error: %v", err)
	}
}

func TestDocRoots(t *testing.T) {
	tests := []struct {
		name  string
		serve *httptest.Server
		want  []string
	}{

		{
			name: "single-user-single-dom",

			serve: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
				data := Info{
					Users: map[string]users{
						"one": {
							"one": domains{
								"one.example": &docroot{
									PublicHTML:  "/home/one/domains/one.example/public_html",
									PrivateHTML: "/home/one/domains/one.example/private_html",
								},
							},
						},
					},
				}

				if err := json.NewEncoder(wr).Encode(data); err != nil {
					t.Fatalf("json marshal error: %v", err)
				}
			})),

			want: []string{
				"/home/one/domains/one.example/private_html",
				"/home/one/domains/one.example/public_html",
				"/home/one/domains/one.example/tmp",
			},
		},

		{
			name: "single-user-single-sub",

			serve: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
				data := Info{
					Users: map[string]users{
						"one": {
							"one": domains{
								"one.example": &docroot{
									PublicHTML:  "/home/one/domains/one.example/public_html",
									PrivateHTML: "/home/one/domains/one.example/private_html",

									Subdomains: domains{
										"sub-1.domain1.com": &docroot{
											PublicHTML:  "/home/one/domains/one.example/sub-one/public_html",
											PrivateHTML: "/home/one/domains/one.example/sub-one/private_html",

											Subdomains: domains{},
										},
									},
								},
							},
						},
					},
				}

				if err := json.NewEncoder(wr).Encode(data); err != nil {
					t.Fatalf("json marshal error: %v", err)
				}
			})),

			want: []string{
				"/home/one/domains/one.example/private_html",
				"/home/one/domains/one.example/public_html",
				"/home/one/domains/one.example/sub-one/private_html",
				"/home/one/domains/one.example/sub-one/public_html",
				"/home/one/domains/one.example/tmp",
			},
		},

		{
			name: "single-user-multi-dom",

			serve: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
				data := Info{
					Users: map[string]users{
						"one": {
							"one": domains{
								"one.example": &docroot{
									PublicHTML:  "/home/one/domains/one.example/public_html",
									PrivateHTML: "/home/one/domains/one.example/private_html",

									Subdomains: domains{},
								},

								"two.example": &docroot{
									PublicHTML:  "/home/one/domains/two.example/public_html",
									PrivateHTML: "/home/one/domains/two.example/private_html",

									Subdomains: domains{},
								},
							},
						},
					},
				}

				if err := json.NewEncoder(wr).Encode(data); err != nil {
					t.Fatalf("json marshal error: %v", err)
				}
			})),

			want: []string{
				"/home/one/domains/one.example/private_html",
				"/home/one/domains/one.example/public_html",
				"/home/one/domains/one.example/tmp",

				"/home/one/domains/two.example/private_html",
				"/home/one/domains/two.example/public_html",
				"/home/one/domains/two.example/tmp",
			},
		},

		{
			name: "multi-user-multi-dom",

			serve: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
				data := Info{
					Users: map[string]users{
						"one": {
							"one": domains{
								"one.example": &docroot{
									PublicHTML:  "/home/one/domains/one.example/public_html",
									PrivateHTML: "/home/one/domains/one.example/private_html",

									Subdomains: domains{
										"sub-one.example": &docroot{
											PublicHTML:  "/home/one/domains/one.example/sub-one/public_html",
											PrivateHTML: "/home/one/domains/one.example/sub-one/private_html",

											Subdomains: domains{},
										},

										"sub-two.example": &docroot{
											PublicHTML:  "/home/one/domains/one.example/sub-two/public_html",
											PrivateHTML: "/home/one/domains/one.example/sub-two/private_html",

											Subdomains: domains{},
										},
									},
								},

								"two.example": &docroot{
									PublicHTML:  "/home/one/domains/two.example/public_html",
									PrivateHTML: "/home/one/domains/two.example/private_html",

									Subdomains: domains{},
								},
							},
						},

						"two": {
							"two": domains{
								"three.example": &docroot{
									PublicHTML:  "/home/two/domains/three.example/public_html",
									PrivateHTML: "/home/two/domains/three.example/private_html",

									Subdomains: domains{},
								},

								"four.example": &docroot{
									PublicHTML:  "/home/two/domains/four.example/public_html",
									PrivateHTML: "/home/two/domains/four.example/private_html",

									Subdomains: domains{},
								},
							},
						},
					},
				}

				if err := json.NewEncoder(wr).Encode(data); err != nil {
					t.Fatalf("json marshal error: %v", err)
				}
			})),

			want: []string{
				"/home/one/domains/one.example/private_html",
				"/home/one/domains/one.example/public_html",
				"/home/one/domains/one.example/sub-one/private_html",
				"/home/one/domains/one.example/sub-one/public_html",
				"/home/one/domains/one.example/sub-two/private_html",
				"/home/one/domains/one.example/sub-two/public_html",
				"/home/one/domains/one.example/tmp",

				"/home/one/domains/two.example/private_html",
				"/home/one/domains/two.example/public_html",
				"/home/one/domains/two.example/tmp",

				"/home/two/domains/four.example/private_html",
				"/home/two/domains/four.example/public_html",
				"/home/two/domains/four.example/tmp",

				"/home/two/domains/three.example/private_html",
				"/home/two/domains/three.example/public_html",
				"/home/two/domains/three.example/tmp",
			},
		},
		{
			name: "empty",

			serve: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
				data := Info{Users: map[string]users{}}
				if err := json.NewEncoder(wr).Encode(data); err != nil {
					t.Fatalf("json marshal error: %v", err)
				}
			})),

			want: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plat := &Plat{
				endpoints: &endpoints{
					docroots: &endpoint{
						url: test.serve.URL,
					},
				},
			}

			paths, err := plat.DocRoots()
			if err != nil {
				t.Errorf("docroots error: %v", err)
			}

			if !slices.Equal(paths, test.want) {
				t.Errorf("expected paths %v, got %v", test.want, paths)
			}
		})
	}
}

func TestCfg(t *testing.T) {
	var (
		plat = &Plat{
			cfg: &Cfg{
				User: t.Name(),
			},
		}

		got = plat.Cfg()
	)

	if !reflect.DeepEqual(got, plat.cfg) {
		t.Errorf("unexpected cfg result %v, want %v", got, plat.cfg)
	}
}

func TestActers(t *testing.T) {
	var (
		input = acter.Mock(act.VerbAlert, true)

		plat = &Plat{
			acters: []acter.Acter{
				input,
			},
		}

		got = plat.Acters()
	)

	if !reflect.DeepEqual(got, plat.acters) {
		t.Errorf("unexpected acters result %v, want %v", got, plat.acters)
	}
}
