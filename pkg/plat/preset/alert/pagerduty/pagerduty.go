// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package pagerduty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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
	Identifier string
	secrets    *secret.PagerDuty
}

// Alert represents an alert.
type Alert struct {
	Payload     *Payload `json:"payload"`
	RoutingKey  string   `json:"routing_key"`
	DedupKey    string   `json:"dedup_key"`
	EventAction string   `json:"event_action"`
}

// Payload represents an event payload.
type Payload struct {
	Summary  string        `json:"summary"`
	Source   string        `json:"source"`
	Severity string        `json:"severity"`
	Details  *state.Result `json:"custom_details"`
}

var statuses = []int{
	http.StatusOK,
	http.StatusAccepted,
	http.StatusCreated,
	http.StatusNoContent,
}

// New returns a new transport.
func New(env *env.Env) *Sender {
	return &Sender{
		cfg: NewCfg(filepath.Join(env.Paths.Alerts.Dir, "pagerduty.toml")),

		client: &http.Client{
			Timeout: 5 * time.Second,
		},

		Identifier: env.Cfg.Identifier,
		secrets:    env.Cfg.Secrets.Alerts.PagerDuty,
	}
}

// Load loads alerter cfg files.
func (sender *Sender) Load() error {
	if err := sender.cfg.Load(); err != nil {
		return err
	}

	if sender.cfg.Endpoint == "" {
		return acter.ErrDisabled
	}

	return nil
}

// Alert sends an alert.
func (sender *Sender) Alert(result *state.Result) error {
	slog.Info("sending alert", "transport", "pagerduty")

	payload, err := sender.NewAlert(result)
	if err != nil {
		return err
	}

	return client.Post(sender.client, nil, nil, sender.cfg.Endpoint, bytes.NewBuffer(payload), statuses...)
}

// NewAlert creates an alert.
func (sender *Sender) NewAlert(result *state.Result) ([]byte, error) {
	alert := &Alert{
		Payload: &Payload{
			Summary:  fmt.Sprintf("Malwatch Scan Report - %v", sender.Identifier),
			Source:   sender.Identifier,
			Severity: sender.cfg.Severity,
			Details:  result,
		},

		RoutingKey:  sender.secrets.Token,
		EventAction: "trigger",
	}

	payload, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// Cfg returns the cfg.
func (sender *Sender) Cfg() plat.Cfg {
	return sender.cfg
}

// Mock mocks a sender.
func Mock(name string, dir string) (*Sender, error) {
	env, err := env.Mock(name, dir)
	if err != nil {
		return nil, err
	}

	return New(env), nil
}
