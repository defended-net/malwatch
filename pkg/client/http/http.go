// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"bytes"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
)

// Post sends a post request.
// statuses define happy status codes.
func Post(client *http.Client, hdrs http.Header, secrets *secret.JSON, endpoint string, payload []byte, statuses []int) error {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("%w, %v", ErrReqPrep, err)
	}

	if hdrs != nil {
		req.Header = hdrs
	}

	if secrets != nil {
		req.SetBasicAuth(secrets.User, secrets.Pass)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w, %v", ErrReqDo, err)
	}

	if !slices.Contains(statuses, resp.StatusCode) {
		return fmt.Errorf("%w, %v", ErrBadStatus, strconv.Itoa(resp.StatusCode))
	}

	return err
}
