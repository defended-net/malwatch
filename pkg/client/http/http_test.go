// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var client = &http.Client{
	Timeout: 5 * time.Second,
}

func TestPost(t *testing.T) {
	var (
		svc   = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
		input = bytes.NewReader([]byte(t.Name()))
		got   = Post(client, nil, nil, svc.URL, input, 200)
	)

	defer svc.Close()

	if got != nil {
		t.Errorf("post error %s", got)
	}
}

func TestPostRespCodes(t *testing.T) {
	var (
		svc   = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
		input = bytes.NewReader([]byte(t.Name()))
		got   = Post(client, nil, nil, svc.URL, input, 404)
		want  = ErrBadStatus
	)

	defer svc.Close()

	if !errors.Is(got, want) {
		t.Errorf("unexpected post resp code err %v, want %v", got, want)
	}
}

func TestPostErrs(t *testing.T) {
	var (
		input = bytes.NewReader([]byte(t.Name()))
		got   = Post(client, nil, nil, "https://"+t.Name(), input, 200)
		want  = ErrReqDo
	)

	if !errors.Is(got, want) {
		t.Errorf("unexpected post err %v, want %v", got, want)
	}
}

func TestPostHeaders(t *testing.T) {
	var (
		hdrs    = http.Header{}
		wantKey = "Content-Type"
		wantVal = "application / json"

		svc = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, rdr *http.Request) {
			if rdr.Header.Get(wantKey) != wantVal {
				t.Errorf("unexpected header %v, want %v", rdr.Header.Get(wantKey), wantVal)
			}
		}))

		input = bytes.NewReader([]byte(t.Name()))
	)

	defer svc.Close()

	hdrs.Set(wantKey, wantVal)

	if got := Post(client, hdrs, nil, svc.URL, input, 200); got != nil {
		t.Errorf("post err %s", got)
	}
}

func TestPostAuth(t *testing.T) {
	var (
		svc = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, rdr *http.Request) {
			user, pass, ok := rdr.BasicAuth()

			if !ok || user != "user" || pass != "pass" {
				t.Errorf("unexpected auth user %v, pass %v", user, pass)
			}
		}))

		htpasswd = &Passwd{
			User: "user",
			Pass: "pass",
		}

		input = bytes.NewReader([]byte(t.Name()))
	)

	defer svc.Close()

	if got := Post(client, nil, htpasswd, svc.URL, input, 200); got != nil {
		t.Errorf("post err %s", got)
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "none",
			input: "https://example.com/path",
			want:  "https://example.com/path",
		},

		{
			name:  "user-only",
			input: "https://user@example.com/path",
			want:  "https://example.com/path",
		},

		{
			name:  "user-pass",
			input: "https://user:pass@example.com/path",
			want:  "https://example.com/path",
		},

		{
			name:  "query",
			input: "https://user:pass@example.com/path?k=v",
			want:  "https://example.com/path?k=v",
		},

		{
			name:  "empty",
			input: "",
			want:  "",
		},

		{
			name:  "invalid",
			input: "://invalid",
			want:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Clean(test.input); got != test.want {
				t.Errorf("unexpected clean result %v, want %v", got, test.want)
			}
		})
	}
}

func TestPostUserAgent(t *testing.T) {
	var (
		hdrs    = http.Header{}
		wantKey = "User-Agent"

		svc = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, rdr *http.Request) {
			if got := rdr.Header.Get(wantKey); got != UA {
				t.Errorf("unexpected user agent %v, want %v", got, UA)
			}
		}))

		input = bytes.NewReader([]byte(t.Name()))
	)

	defer svc.Close()

	if got := Post(client, nil, nil, svc.URL, input, 200); got != nil {
		t.Errorf("post err %s", got)
	}

	hdrs.Set("Content-Type", "application/json")

	if got := Post(client, hdrs, nil, svc.URL, input, 200); got != nil {
		t.Errorf("post err %s", got)
	}
}
