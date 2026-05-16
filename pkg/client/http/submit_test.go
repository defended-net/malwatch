// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestSubmit(t *testing.T) {
	var (
		svc = httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, rdr *http.Request) {
			if rdr.Header.Get("Authorization") != "Bearer key" {
				t.Errorf("unexpected auth header %v", rdr.Header.Get("Authorization"))
			}

			if rdr.Header.Get("Content-Type") != "text/plain" {
				t.Errorf("unexpected content type %v", rdr.Header.Get("Content-Type"))
			}

			body, err := io.ReadAll(rdr.Body)
			if err != nil {
				t.Errorf("read err %v", err)
			}

			if string(body) != t.Name() {
				t.Errorf("unexpected body %v, want %v", string(body), t.Name())
			}

			wr.WriteHeader(http.StatusOK)
		}))

		path = filepath.Join(t.TempDir(), t.Name())

		input = &secret.Submit{
			Endpoint: svc.URL,
			Key:      "key",
		}
	)
	defer svc.Close()

	if err := os.WriteFile(path, []byte(t.Name()), 0600); err != nil {
		t.Fatalf("file write err %v", err)
	}

	if got := Submit(input, path); got != nil {
		t.Errorf("submit err %v", got)
	}
}

func TestSubmitErrs(t *testing.T) {
	tests := map[string]struct {
		fn   func(t *testing.T) (string, *secret.Submit)
		want error
	}{
		"dot-dots": {
			fn: func(t *testing.T) (string, *secret.Submit) {
				t.Helper()

				return t.TempDir() + "/../" + t.Name(), &secret.Submit{
					Endpoint: "https://" + t.Name(),
					Key:      "key",
				}
			},

			want: fsys.ErrPathTraverse,
		},

		"not-abs": {
			fn: func(t *testing.T) (string, *secret.Submit) {
				t.Helper()

				return "rel/path", &secret.Submit{
					Endpoint: "https://" + t.Name(),
					Key:      "key",
				}
			},

			want: fsys.ErrPathNotAbs,
		},

		"not-exist": {
			fn: func(t *testing.T) (string, *secret.Submit) {
				t.Helper()

				return filepath.Join(t.TempDir(), "not-exist"), &secret.Submit{
					Endpoint: "https://" + t.Name(),
					Key:      "key",
				}
			},

			want: fsys.ErrFileOpen,
		},

		"endpoint-invalid": {
			fn: func(t *testing.T) (string, *secret.Submit) {
				t.Helper()

				path := filepath.Join(t.TempDir(), strings.ReplaceAll(t.Name(), "/", "-"))

				if err := os.WriteFile(path, []byte(t.Name()), 0600); err != nil {
					t.Fatalf("write file error %v", err)
				}

				return path, &secret.Submit{
					Endpoint: "https://" + t.Name(),
					Key:      "key",
				}
			},

			want: ErrReqDo,
		},

		"status-invalid": {
			fn: func(t *testing.T) (string, *secret.Submit) {
				t.Helper()

				var (
					svc = httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
						wr.WriteHeader(http.StatusInternalServerError)
					}))

					path = filepath.Join(t.TempDir(), strings.ReplaceAll(t.Name(), "/", "-"))

					input = &secret.Submit{
						Endpoint: svc.URL,
						Key:      "key",
					}
				)

				t.Cleanup(svc.Close)

				if err := os.WriteFile(path, []byte(t.Name()), 0600); err != nil {
					t.Fatalf("file write err %v", err)
				}

				return path, input
			},

			want: ErrBadStatus,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			path, secrets := test.fn(t)

			if got := Submit(secrets, path); !errors.Is(got, test.want) {
				t.Errorf("unexpected submit err %v, want %v", got, test.want)
			}
		})
	}
}
