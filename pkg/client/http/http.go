// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
)

// Passwd represents htpasswd.
type Passwd struct {
	User string
	Pass string
}

// Post sends a post request.
// statuses define happy status codes.
func Post(client *http.Client,
	hdrs http.Header,
	auth *Passwd,
	endpoint string,
	rdr io.Reader,
	statuses ...int,
) error {
	req, err := http.NewRequest(http.MethodPost, endpoint, rdr)
	if err != nil {
		return fmt.Errorf("%w, %v", ErrReqPrep, err)
	}

	if hdrs != nil {
		req.Header = hdrs
	}

	if auth != nil {
		req.SetBasicAuth(auth.User, auth.Pass)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w, %v", ErrReqDo, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w, %v, %v, %v", ErrReqDo, err, "body", string(body))
	}

	if !slices.Contains(statuses, resp.StatusCode) {
		return fmt.Errorf("%w, %v, %v, %v", ErrBadStatus, strconv.Itoa(resp.StatusCode), "body", string(body))
	}

	return nil
}
