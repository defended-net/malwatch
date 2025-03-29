// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
)

// Submit uploads a malware sample.
func Submit(secrets *secret.Submit, path string) error {
	var (
		rdr, wr = io.Pipe()

		sess = &http.Client{
			Timeout: 5 * time.Second,
		}

		hdrs = http.Header{
			"Authorization": {
				"Bearer " + secrets.Key,
			},

			"Content-Type": {
				"text/plain",
			},
		}
	)

	go func() {
		defer wr.Close()

		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()

		if _, err = io.Copy(wr, file); err != nil {
			return
		}
	}()

	return Post(
		sess,
		hdrs,
		nil,
		secrets.Endpoint,
		rdr,
		200,
	)
}
