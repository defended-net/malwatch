// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/defended-net/malwatch/pkg/fsys"
)

// Passwd represents htpasswd.
type Passwd struct {
	User string
	Pass string
}

const (
	// UA stores user agent https://www.rfc-editor.org/rfc/rfc7231#section-5.5.3
	UA     = "malwatch (+https://github.com/defended-net/malwatch)"
	bodySz = 10 << 20
)

// Post sends a post request.
// statuses define happy status codes.
func Post(
	client *http.Client,
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

	req.Header.Set("User-Agent", UA)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w, %v", ErrReqDo, err)
	}
	defer fsys.Close(resp.Body)

	body, err := io.ReadAll(io.LimitReader(resp.Body, bodySz))
	if err != nil {
		return fmt.Errorf("%w, %v, %v, %v", ErrReqDo, err, "body", string(body))
	}

	if !slices.Contains(statuses, resp.StatusCode) {
		return fmt.Errorf("%w, %v, %v, %v", ErrBadStatus, strconv.Itoa(resp.StatusCode), "body", string(body))
	}

	return nil
}

// Clean strips auth component from given url.
func Clean(input string) string {
	url, err := url.Parse(input)
	if err != nil {
		return ""
	}

	url.User = nil

	return url.String()
}
