<pre>
              __            __      __
  __ _  ___ _/ /    _____ _/ /_____/ /
 /    \/ _ `/ / |/|/ / _ `/ __/ __/ _ \
/_/_/_/\_,_/_/|__,__/\_,_/\__/\__/_//_/
                           defended.net
</pre>

Malwatch is a fast and lightweight malware scanner written in `go` that is ideal for Linux based web server environments. It is capable of scaling to any requirements and is currently used with some of the internet's largest deployments.

Besides excellent detection rates, key design considerations are low resource usage and high performance. A powerful and easy to understand api is provided to cover your alerting requirements and platform integration.

The web hosting industry is in need for a modern open source solution that is done properly. Our objective is to offer a comparably better solution to commercial products, thereby empowering everyone to have good malware protection.

Real time monitoring is offered by `malwatch-monitor`. The number of files present on the system or volume being created / changed does not influence startup time, memory or cpu usage because a queue design ensures there would always be the same resource usage as an equivalent ondemand file scan. The result is an extremely efficient monitoring presence which would otherwise be impractical due to resource usage.

A complete malware signature set is included with updates provided by our [signature repo](https://github.com/defended-net/malwatch-signatures).

There is tremendous value if malwatch is elected to replace your fleet's existing commercial solution. Please consider sponsoring to help us maintain this and other projects.

# Primary Features

- On Demand and Real Time scans.
- Leading performance with resource usage profile delivered in ~22 MiB memory alongside low cpu footprint even under full load.
- Comprehensive malware signature set.
- Simple yet powerful API to integrate with your backend and platform. Ideal to improve threat intelligence.
- Flexible alerting capability with built-in support for PagerDuty, e - mail and custom JSON. Use the API to easily add your own!
- Complete control over the outcome of malware with the help of `actions`. We include `alert`, `quarantine`, `clean` and `exile` but custom ones can be defined.
- Intuitive and completely transparent signature management. Commit changes to a `git` repo for secure delivery.
- Structured logs entirely formatted as JSON.
- ACID compliant database for record keeping.

# Installation

Create a directory anywhere you prefer, software is meant to be portable. A common choice is `/opt/malwatch`. Then extract the binary there from the downloaded archive:

    wget https://github.com/defended-net/malwatch/releases/download/v1.0.0/malwatch_1.0.0_linux_amd64.tar.gz
    mkdir /opt/malwatch
    tar -C /opt/malwatch -xzvf malwatch_1.0.0_linux_amd64.tar.gz

It would be recommended to set up your `PATH`:

    export PATH=$PATH:/opt/malwatch

Config files will automatically be built in the binary's path upon execution. Let's try:

    malwatch

If you are using automation (such as Ansible) and want a clean exit, then `malwatch install` can be used.

Optional config files have the `.disabled` file extension. These can be renamed to `.toml` to enable.

## Real Time Scanning

Real time malware scanning is possible with `malwatch-monitor`. A `systemd` unit can automatically be created with `malwatch install systemd`. This is optional - feel free to set up your own or use any preferred setup such as a foreground process or even `screen`. We believe software should be flexible.

Using `systemd`, it is necessary to enable followed by starting it:

    systemctl enable malwatch-monitor
    systemctl start malwatch-monitor

## Cron

`cron` can be used to schedule scans at preferred interval. It is not recommended to use scheduled scans if real time scanning with `malwatch-monitor` is already being used.

The command `crontab -e` is used to add or modify cron jobs. The command field can specofy the absolute path `/opt/malwatch/malwatch scan` to automatically scan all targets. An example configuration to scan each day at 01:00 AM is as follows:

    0 1 * * * /opt/malwatch/malwatch scan

# Setup

Some basic config variables are needed for basic operation and must be defined in the file `cfg/cfg.toml`.

Once configured, you are ready to perform your first scan!

    malwatch scan /var/www/html

# Targets

The term `target` means a group of paths which share a common parent. This is accomplished using regex. The default value is `Targets = ["^/var/www/(?P<target>[^/]+)"]`

In the cfg file `(?P<target>.*)` is a **capture group**. Any paths under the directory `/var/www` will be considered the `target`. The default `Paths` config variable is `/var/www/html`, which means any detections will be associated as target `html`.

We could then scan a path `/var/www/images` but any detections there would then be assigned the target `images`.

A scan of path `/var/static` does not match with the regex and thus is assigned the target `fs`, which is the *catchall* target for all detections which do not match the `Targets` regex.

Targets are essential to grouping detections, especially when sending alerts and saving detections to the database.

Let's now go over the other config variables to better understand their role.

Variable  | Description
------------- | -------------
Identifier  | Custom identifier, useful for alerts and logging. Defaults to system hostname.
Cores  | Limit execution based on processor core count. `0` disables limit.
Threads  | Limit execution based on thread count. `0` disables limit.
Timeout  | Time limit (min) of scans _per target_. `0` disables limit.
Monitor.Timeout  | Interval (sec) for real time monitoring scans to scan the queue.
Targets  | Regex to determine a path's `target` classification.
Paths  | List of paths to scan. Multiple entries is possible `["/path/a", "/path/b"]`
MaxAge  | Maximum age of files to scan (days).
BlkSz  | Chunk size (KiB) per read of each file.
BatchSz  | Maximum number of detections per alert before sending the next alert.
Verbose  | Enables extra trace information in logs.

# Actions

Actions are configured in the file `cfg/actions.toml`

An outcome which occurs as a result from a detection is an `action`. Actions are so powerful because they can easily be customised at a granular level. Custom actions can even be created

Each `action` comprises of a `verb` paired with `acter`. Each detection can have multiple `actions`. The lack of any `verb` for a detection means no actions will occur, this can be considered the same as a traditional "whitelist".

The following verbs are bundled with malwatch:

# Verbs

Verb  | Outcome
------------- | -------------
`alert` | Notification by means of one or more alerters. We bundle PagerDuty, e - mail and custom JSON.
`quarantine` | Move detection to a the quarantine path (defined in `cfg/cfg.toml`).
`exile` | Uploads detection to your `s3` bucket. File is removed after succcessful upload, unless `quarantine` is included as the actions.
`clean` | Sed based expressions which remove malware automatically. Basic `base64` encoded malware removal expressions are included.

# Platform Integrations

- cPanel

cPanel is included as a drop - in platform integration. It can be enabled by renaming the file `cfg/platform/cpanel.disabled` to `cfg/platform/cpanel.toml`. No further configuration is needed for it to work.

# Documentation

Quick Start guide, advanced usage, changelog and other resources for both users and developers are available at our [Documentation](https://docs.defended.net/malwatch)

# Community

Join our [Discord](https://discord.com/invite/pnAGEGgRjX) We would love to hear from you!

# Website

https://defended.net

# Acknowledgements

A special thank you to the following projects which are used by `malwatch`:

- Yara (https://github.com/VirusTotal/yara)
- Go-Yara (https://github.com/hillu/go-yara)
- bbolt (https://go.etcd.io/bbolt)
- minio-go (https://github.com/minio/minio-go)
- go-git (https://github.com/go-git/go-git)
- Go.Sed (https://github.com/rwtodd/Go.Sed)
- simpletable (https://github.com/alexeyco/simpletable)
- toml (https://github.com/BurntSushi/toml)

Hopefully this software can somehow bring a bit of peace to this troubled world.