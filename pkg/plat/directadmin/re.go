// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package directadmin

import "regexp"

var (
	// reTarget validates a system user's home dir.
	// https://docs.directadmin.com/directadmin/backup-restore-migration/migration-to-da.html
	// https://systemd.io/USER_NAMES/#:~:text=A%20size%20limit%20is%20enforced,ambiguity%20with%20login%20accounting)%20and
	reTarget = regexp.MustCompile(`^/home/(?P<target>[a-z\d-]{1,31})(?:/.*)?$`)
)
