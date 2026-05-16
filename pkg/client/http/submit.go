// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/fsys"
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

	if err := fsys.HasDotDots(path); err != nil {
		return err
	}

	fd, _, err := fsys.Open(path)
	if err != nil {
		return err
	}

	file := os.NewFile(uintptr(fd), path)
	defer fsys.Close(file)

	go func() {
		defer fsys.Close(wr)

		if _, err := io.Copy(wr, file); err != nil {
			wr.CloseWithError(err)
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
