package fan

import "errors"

// FAN
var (
	// ErrInit means fan init error.
	ErrInit = errors.New("fan: init error")

	// ErrVerMismatch means ver does not match the system ver.
	ErrVerMismatch = errors.New("fan: mismatched fan ver")
)

// EVENTS
var (
	// ErrEventBinRead means event binary read error.
	ErrEventBinRead = errors.New("fan: event binary read error")

	// ErrEventClose means event close error.
	ErrEventClose = errors.New("fan: event close error")

	// ErrEventReadlink means event readlink error.
	ErrEventReadlink = errors.New("fan: event readlink error")
)
