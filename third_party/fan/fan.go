package fan

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/sys/unix"
)

const (
	// ProcFsFd stores procfs self.
	ProcFsFd = "/proc/self/fd"
)

// Notify represents a notify file handle.
type Notify struct {
	Fd     int
	File   *os.File
	Reader io.Reader
}

// EventMeta represents event metadata.
type EventMeta struct {
	unix.FanotifyEventMetadata
}

// Close closes a eventmeta.
func (meta *EventMeta) Close() error {
	if meta == nil {
		return nil
	}

	return unix.Close(int(meta.Fd))
}

// GetPath returns the path from given eventmetadata.
func (meta *EventMeta) GetPath() (string, error) {
	path, err := os.Readlink(
		filepath.Join(
			ProcFsFd,
			strconv.FormatUint(
				uint64(meta.Fd),
				10,
			),
		),
	)
	if err != nil {
		return "", fmt.Errorf("%w, %v", ErrEventReadlink, err)
	}

	return path, nil
}

// MatchMask checks if given eventmeta matches a given mask.
func (meta *EventMeta) MatchMask(mask int) bool {
	return (meta.Mask & uint64(mask)) == uint64(mask)
}

// File returns a *os.File from event meta.
func (meta *EventMeta) File() *os.File {
	// Prevent gc of fd.
	fd, err := unix.Dup(int(meta.Fd))
	if err != nil {
		return nil
	}

	return os.NewFile(uintptr(fd), "")
}

// NewNotify returns a notify.
func NewNotify(notifyFlags uint, openFlags int) (*Notify, error) {
	fd, err := unix.FanotifyInit(notifyFlags, uint(openFlags))
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrInit, err)
	}

	var (
		file   = os.NewFile(uintptr(fd), "")
		reader = bufio.NewReader(file)
	)

	return &Notify{
		Fd:     fd,
		File:   file,
		Reader: reader,
	}, nil
}

// Mark implements add / delete / modify for a mark.
func (notify *Notify) Mark(flags uint, mask uint64, dirFd int, path string) error {
	return unix.FanotifyMark(notify.Fd, flags, mask, dirFd, path)
}

// GetEvent returns an event.
func (notify *Notify) GetEvent() (*EventMeta, error) {
	event := new(EventMeta)

	if err := binary.Read(notify.Reader, binary.LittleEndian, event); err != nil {
		return nil, fmt.Errorf("%w, %v", ErrEventBinRead, err)
	}

	if event.Vers != unix.FANOTIFY_METADATA_VERSION {
		if err := event.Close(); err != nil {
			return nil, fmt.Errorf("%w, %v", ErrEventClose, err)
		}

		return nil, fmt.Errorf("%w, %v", ErrVerMismatch, event.Vers)
	}

	return event, nil
}
