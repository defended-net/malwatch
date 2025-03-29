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
	svc := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer svc.Close()

	if err := Post(client, nil, nil, svc.URL, bytes.NewReader([]byte(t.Name())), 200); err != nil {
		t.Errorf("post error: %s", err)
	}
}

func TestPostRespCodes(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer svc.Close()

	if err := Post(client, nil, nil, svc.URL, bytes.NewReader([]byte(t.Name())), 404); !errors.Is(err, ErrBadStatus) {
		t.Errorf("unexpected post resp code error: %v, want %v", err, ErrBadStatus)
	}
}

func TestPostErrs(t *testing.T) {
	if err := Post(client, nil, nil, "https://"+t.Name(), bytes.NewReader([]byte(t.Name())), 200); !errors.Is(err, ErrReqDo) {
		t.Errorf("unexpected post error: %v, want %v", err, ErrReqDo)
	}
}

func TestPostWithHeaders(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected header: %v, want %v", r.Header.Get("Content-Type"), "application/json")
		}
	}))
	defer svc.Close()

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	if err := Post(client, headers, nil, svc.URL, bytes.NewReader([]byte(t.Name())), 200); err != nil {
		t.Errorf("post error: %s", err)
	}
}

func TestPostWithAuth(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "user" || pass != "pass" {
			t.Errorf("unexpected auth: user=%v, pass=%v", user, pass)
		}
	}))
	defer svc.Close()

	htpasswd := &Passwd{
		User: "user",
		Pass: "pass",
	}

	if err := Post(client, nil, htpasswd, svc.URL, bytes.NewReader([]byte(t.Name())), 200); err != nil {
		t.Errorf("post error: %s", err)
	}
}
