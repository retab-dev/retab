//go:build windows

package cmd

import (
	"os"

	"golang.org/x/sys/windows"
)

// lockFileExclusive takes a blocking exclusive lock on f via LockFileEx.
// Without LOCKFILE_FAIL_IMMEDIATELY the call blocks until the lock is granted.
// A 1-byte range at offset 0 is sufficient for advisory whole-file locking.
func lockFileExclusive(f *os.File) error {
	return windows.LockFileEx(
		windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0, 1, 0,
		new(windows.Overlapped),
	)
}

// unlockFile releases the lock taken by lockFileExclusive.
func unlockFile(f *os.File) error {
	return windows.UnlockFileEx(
		windows.Handle(f.Fd()),
		0, 1, 0,
		new(windows.Overlapped),
	)
}
