//go:build !windows

package cmd

import (
	"os"
	"syscall"
)

// lockFileExclusive takes a blocking exclusive advisory lock (flock LOCK_EX)
// on f. It blocks until the lock is available.
func lockFileExclusive(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// unlockFile releases the advisory lock held on f.
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
