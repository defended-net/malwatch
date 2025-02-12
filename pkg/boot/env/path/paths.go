// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package path

// Paths represents paths.
type Paths struct {
	Install *Install
	Cfg     *Cfg
	Plat    *Plat
	Sigs    *Sigs
	Alerts  *Alerts
}

// Install represents install paths.
type Install struct {
	Dir  string
	Bin  string
	Path string
	Tmp  string
	Db   string
	Log  string
}

// Cfg represents cfg paths.
type Cfg struct {
	Dir     string
	Base    string
	Secrets string
	Acts    string
}

// Sigs represents sig paths.
type Sigs struct {
	Dir string
	Src string
	Idx string
	Yrc string
	Tmp string
}

// Plat represents plat cfg paths.
type Plat struct {
	Dir string
	Cfg string
}

// Alerts represents alert cfg paths.
type Alerts struct {
	Dir string
}
