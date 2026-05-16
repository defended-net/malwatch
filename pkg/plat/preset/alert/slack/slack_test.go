// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package slack

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	client "github.com/defended-net/malwatch/pkg/client/http"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

func TestNewAlert(t *testing.T) {
	var (
		msg        = Message{}
		input, err = Mock(t.Name(), t.TempDir())
		want       = 2
	)

	if err != nil {
		t.Fatalf("sender mock error %v", err)
	}

	got, err := input.NewAlert(&state.Result{})
	if err != nil {
		t.Fatalf("alert create error %v", err)
	}

	if err := json.Unmarshal(got, &msg); err != nil {
		t.Fatalf("alert invalid json %v", err)
	}

	switch {
	case msg.Text == "":
		t.Errorf("no fallback text")

	case len(msg.Blocks) != 2:
		t.Fatalf("unexpected block count %v, want %v", len(msg.Blocks), want)

	case msg.Blocks[0].Text.Type != "plain_text":
		t.Fatalf("unexpected block count %v, want %v", len(msg.Blocks), want)

	case !msg.Blocks[0].Text.Emoji:
		t.Errorf("no emoji support")
	}
}

func TestLoadDisabled(t *testing.T) {
	var (
		tmp        = t.TempDir()
		input, err = Mock(t.Name(), tmp)
	)

	if err != nil {
		t.Fatalf("sender mock error %v", err)
	}

	input.cfg.path = "slack.toml"
	input.secrets.Webhook = ""

	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("open root err %v", err)
	}

	defer func() {
		// lint
		_ = root.Close()
	}()

	if got := input.Load(root); !errors.Is(got, acter.ErrDisabled) {
		t.Errorf("unexpected slack load error %v, want %v", got, acter.ErrDisabled)
	}
}

func TestAlert(t *testing.T) {
	var (
		wantMethod      = http.MethodPost
		wantContentType = "application/json"
	)

	tests := map[string]struct {
		status int
		want   error
	}{
		"ok": {
			status: http.StatusOK,
			want:   nil,
		},

		"accepted": {
			status: http.StatusAccepted,
			want:   nil,
		},

		"created": {
			status: http.StatusCreated,
			want:   nil,
		},

		"no-content": {
			status: http.StatusNoContent,
			want:   nil,
		},

		"invalid-status": {
			status: http.StatusInternalServerError,
			want:   client.ErrBadStatus,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
				msg := Message{}

				if req.Method != wantMethod {
					t.Errorf("unexpected method %v, want %v", req.Method, wantMethod)
				}

				if got := req.Header.Get("Content-Type"); got != wantContentType {
					t.Errorf("unexpected content type %v, want %v", got, wantContentType)
				}

				if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
					t.Errorf("request body is not valid slack message json %v", err)
				}

				if msg.Text == "" {
					t.Errorf("missing fallback text")
				}

				wr.WriteHeader(test.status)
			}))
			defer svc.Close()

			input, err := Mock(t.Name(), t.TempDir())
			if err != nil {
				t.Fatalf("sender mock err %v", err)
			}

			input.secrets.Webhook = svc.URL

			if got := input.Alert(&state.Result{}); !errors.Is(got, test.want) {
				t.Errorf("unexpected alert err %v, want %v", got, test.want)
			}
		})
	}
}

func TestAlertErrs(t *testing.T) {
	var (
		input, err = Mock(t.Name(), t.TempDir())
		want       = client.ErrReqDo
	)

	if err != nil {
		t.Fatalf("mock error %v", err)
	}

	input.secrets.Webhook = "http://127.0.0.1:0"

	if got := input.Alert(&state.Result{}); !errors.Is(got, want) {
		t.Errorf("unexpected alert error %v, want %v", got, want)
	}
}

func TestLimText(t *testing.T) {
	const lim = 3000

	tests := map[string]struct {
		input string
		want  string
	}{
		"under": {
			input: "short",
			want:  "short",
		},

		"ok": {
			input: strings.Repeat("a", lim),
			want:  strings.Repeat("a", lim),
		},

		"over": {
			input: strings.Repeat("a", lim+1),
			want:  strings.Repeat("a", lim) + "\n... truncated",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := limText(test.input); got != test.want {
				t.Errorf("unexpected text result %v, want %v", len(got), len(test.want))
			}
		})
	}
}

func TestLimTextUnicode(t *testing.T) {
	var (
		lim   = 3000
		input = strings.Repeat("⚠", lim+1)
		want  = strings.Repeat("⚠", lim) + "\n... truncated"
	)

	if got := limText(input); got != want {
		t.Errorf("unexpected unicode result %v want %v", got, want)
	}
}
