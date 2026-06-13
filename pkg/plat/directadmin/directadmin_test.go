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
		t.Errorf("env mock err %v", err)
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
		t.Errorf("unexpected cfg result %v, want %v", got, want)
	}

	if !reflect.DeepEqual(got.endpoints, want.endpoints) {
		t.Errorf("unexpected endpoints result %v, want %v", got, want)
	}
}

func TestLoad(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("plat mock err %v", err)
	}

	svc := httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
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
			t.Fatalf("json marshal err %v", err)
		}
	}))

	plat.url = svc.URL

	if err = plat.Load(nil); err != nil {
		t.Errorf("load err %v", err)
	}
}

func TestLoadAuthed(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Errorf("plat mock err %v", err)
	}

	plat.url = "ftp://localhost"

	if err = plat.Load(nil); err != nil && !strings.Contains(err.Error(), "unsupported protocol scheme") {
		t.Errorf("load err %v", err)
	}
}

func TestDocRoots(t *testing.T) {
	tests := []struct {
		name string
		svc  *httptest.Server
		want []string
	}{

		{
			name: "single-user-single-dom",

			svc: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
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
					t.Fatalf("json marshal err %v", err)
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

			svc: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
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
					t.Fatalf("json marshal err %v", err)
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

			svc: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
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
					t.Fatalf("json marshal err %v", err)
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

			svc: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
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
					t.Fatalf("json marshal err %v", err)
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

			svc: httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
				data := Info{Users: map[string]users{}}
				if err := json.NewEncoder(wr).Encode(data); err != nil {
					t.Fatalf("json marshal err %v", err)
				}
			})),

			want: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := &Plat{
				endpoints: &endpoints{
					docroots: &endpoint{
						url: test.svc.URL,
					},
				},
			}

			got, err := input.DocRoots()
			if err != nil {
				t.Errorf("docroots err %v", err)
			}

			if !slices.Equal(got, test.want) {
				t.Errorf("unexpected docroots result %v, got %v", test.want, got)
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

func TestAuth(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("plat mock err %v", err)
	}

	plat.url = "ftp://localhost"

	if got := plat.Auth("user"); got != nil {
		t.Errorf("auth err %v", got)
	}
}

func TestAuthUserInvalid(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("plat mock err %v", err)
	}

	plat.url = ""

	if got := plat.Auth("../invalid"); got == nil {
		t.Errorf("unexpected auth success")
	}
}

func TestAuthBinInvalid(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("plat mock err %v", err)
	}

	plat.url = ""
	plat.bin = "/dev/null/not-exist"

	if got := plat.Auth("user"); got == nil {
		t.Errorf("unexpected auth success")
	}
}

func TestAuthSchemeInvalid(t *testing.T) {
	plat, err := Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("plat mock err %v", err)
	}

	plat.url = ""
	plat.bin = "/bin/echo"

	if got := plat.Auth("user"); got == nil {
		t.Errorf("unexpected auth success")
	}
}

func TestAddDocrootPathInvalid(t *testing.T) {
	if got := addDocroot(map[string]struct{}{}, "el/path"); got {
		t.Errorf("unexpected add docroot success")
	}
}

func TestAddDocrootDotDots(t *testing.T) {
	if got := addDocroot(map[string]struct{}{}, "/../etc/file"); got {
		t.Errorf("unexpected add docroot success")
	}
}
