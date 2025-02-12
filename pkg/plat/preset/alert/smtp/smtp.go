// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package smtp

import (
	"encoding/json"
	"log/slog"
	"path/filepath"

	vars "github.com/caarlos0/env/v11"
	"github.com/wneessen/go-mail"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/plat"
	"github.com/defended-net/malwatch/pkg/scan/state"
)

// Sender represents the sender.
type Sender struct {
	cfg        *Cfg
	secrets    *secret.SMTP
	client     *mail.Client
	identifier string
}

// New returns a new transport.
func New(env *env.Env) *Sender {
	var (
		sender = &Sender{
			cfg:        NewCfg(filepath.Join(env.Paths.Alerts.Dir, "smtp.toml")),
			secrets:    env.Cfg.Secrets.Alerts.SMTP,
			identifier: env.Cfg.Identifier,
		}
	)

	return sender
}

// Load loads alerter cfg files.
func (sender *Sender) Load() error {
	if err := sender.cfg.Load(); err != nil {
		return err
	}

	client, err := mail.NewClient(sender.secrets.Hostname,
		mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithUsername(sender.secrets.User), mail.WithPassword(sender.secrets.Pass),
	)
	if err != nil {
		return err
	}

	sender.client = client

	return nil
}

// Alert sends an alert.
func (sender *Sender) Alert(result *state.Result) error {
	slog.Info("sending alert", "transport", "smtp")

	hits, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	message := mail.NewMsg()

	if err := message.From(sender.cfg.From); err != nil {
		return err
	}

	if err := message.To(sender.cfg.To...); err != nil {
		return err
	}

	message.Subject("malwatch alert - " + sender.identifier)
	message.SetBodyString(mail.TypeTextPlain, string(hits))

	return sender.client.DialAndSend(message)
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

	sender := New(env)

	if err := vars.Parse(sender.cfg); err != nil {
		return nil, err
	}

	client, err := mail.NewClient(sender.secrets.Hostname,
		mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithUsername(sender.secrets.User), mail.WithPassword(sender.secrets.Pass),
	)
	if err != nil {
		return nil, err
	}

	sender.client = client

	return sender, nil
}
