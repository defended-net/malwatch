// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package slack

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	client "github.com/defended-net/malwatch/pkg/client/http"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Sender represents the sender.
type Sender struct {
	cfg        *Cfg
	client     *http.Client
	identifier string
	secrets    *secret.Slack
}

// Message represents Slack webhook message.
type Message struct {
	Text     string  `json:"text"`
	Username string  `json:"username,omitempty"`
	Channel  string  `json:"channel,omitempty"`
	Blocks   []Block `json:"blocks,omitempty"`
}

// Block represents block.
type Block struct {
	Type string `json:"type"`
	Text Text   `json:"text"`
}

// Text represents text.
type Text struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

var statuses = []int{
	http.StatusOK,
	http.StatusAccepted,
	http.StatusCreated,
	http.StatusNoContent,
}

// New returns new sender.
func New(env *env.Env) *Sender {
	return &Sender{
		cfg: NewCfg(filepath.Join(env.Paths.Alerts.Dir, "slack.toml")),

		client: &http.Client{
			Timeout: 5 * time.Second,
		},

		identifier: env.Cfg.Identifier,
		secrets:    env.Cfg.Secrets.Alerts.Slack,
	}
}

// Load loads alerter.
func (sender *Sender) Load(root *os.Root) error {
	if err := sender.cfg.Load(root); err != nil {
		return err
	}

	if sender.secrets.Webhook == "" {
		return acter.ErrDisabled
	}

	return nil
}

// Alert sends alert.
func (sender *Sender) Alert(result *state.Result) error {
	slog.Info("sending alert", "transport", "slack")

	payload, err := sender.NewAlert(result)
	if err != nil {
		return err
	}

	hdrs := http.Header{}
	hdrs.Set("Content-Type", "application/json")

	return client.Post(sender.client, hdrs, nil, sender.secrets.Webhook, bytes.NewBuffer(payload), statuses...)
}

// NewAlert creates alert.
func (sender *Sender) NewAlert(result *state.Result) ([]byte, error) {
	hits, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}

	var (
		title = "Malwatch Scan Report - " + sender.identifier

		msg = &Message{
			Text:     title,
			Username: sender.cfg.User,
			Channel:  sender.cfg.Channel,

			Blocks: []Block{
				{
					Type: "section",

					Text: Text{
						Type:  "plain_text",
						Text:  ":warning: " + title,
						Emoji: true,
					},
				},

				{
					Type: "section",

					Text: Text{
						Type: "plain_text",
						Text: limText(string(hits)),
					},
				},
			},
		}
	)

	return json.Marshal(msg)
}

// Cfg returns cfg.
func (sender *Sender) Cfg() plat.Cfg {
	return sender.cfg
}

func limText(input string) string {
	const lim = 3000

	if len([]rune(input)) <= lim {
		return input
	}

	return string([]rune(input)[:lim]) + "\n... truncated"
}

// Mock mocks sender.
func Mock(name string, dir string) (*Sender, error) {
	env, err := env.Mock(name, dir)
	if err != nil {
		return nil, err
	}

	return New(env), nil
}
