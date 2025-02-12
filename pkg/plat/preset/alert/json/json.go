// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package json

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	client "github.com/defended-net/malwatch/pkg/client/http"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Sender represents the sender.
type Sender struct {
	client     *http.Client
	cfg        *Cfg
	secrets    *secret.JSON
	identifier string
}

// Alert represents an alert.
type Alert struct {
	Payload []byte `json:"payload"`
}

// statuses stores successful http resp codes.
var statuses = []int{
	http.StatusOK,
	http.StatusAccepted,
	http.StatusCreated,
	http.StatusNoContent,
}

// New returns a new sender.
func New(env *env.Env) *Sender {
	sender := &Sender{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},

		cfg:        NewCfg(filepath.Join(env.Paths.Alerts.Dir, "json.toml")),
		secrets:    env.Cfg.Secrets.Alerts.JSON,
		identifier: env.Cfg.Identifier,
	}

	return sender
}

// Alert sends an alert.
func (sender *Sender) Alert(result *state.Result) error {
	slog.Info("sending alert", "transport", "json")

	alert, err := sender.NewAlert(result)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	return client.Post(sender.client, nil, sender.secrets, sender.cfg.Endpoint, payload, statuses)
}

// NewAlert creates an alert.
func (sender *Sender) NewAlert(result *state.Result) ([]byte, error) {
	hits, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	alert := &Alert{
		Payload: hits,
	}

	payload, err := json.Marshal(alert)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// Load loads the alerter.
func (sender *Sender) Load() error {
	return sender.cfg.Load()
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
