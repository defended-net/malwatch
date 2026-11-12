#!/bin/sh
# deb postinst / rpm %post: restart malwatch-monitor if running.
if [ -d /run/systemd/system ] && systemctl is-active --quiet malwatch-monitor; then
	systemctl try-restart malwatch-monitor ||
		echo "malwatch: failed to restart malwatch-monitor, see: systemctl status malwatch-monitor" >&2
fi