package cmd

import (
	"os"
	"path/filepath"
	"strings"
)

// resolveDirDest handles the case where dest points at an existing directory:
// the file is written inside it under the server filename (falling back to the
// file id when the server stored no name), curl -O style. Without this, the
// atomic temp-then-rename in streamDownloadToFile renames a temp file over a
// directory and fails with a confusing "file exists". A non-directory dest
// (including a not-yet-existing path) is returned unchanged.
//
// Untagged so both the default build (files.go) and the generated-overlay
// build (files_transfer_oagen_overlay.go) share one implementation — the
// overlay previously drifted and skipped these protections entirely.
func resolveDirDest(dest, serverName, fileID string) string {
	info, err := os.Stat(dest)
	if err != nil || !info.IsDir() {
		return dest
	}
	name := safeDownloadName(serverName)
	if name == "" {
		name = fileID
	}
	return filepath.Join(dest, name)
}

// safeDownloadName reduces a server-supplied filename to a single, safe path
// component. The server-recorded name is untrusted input (a file may have been
// uploaded with a name like "../evil.pdf" or "sub/report.pdf" — create-upload
// does not reject path separators), and `files download` promises to write it
// "in the current directory". Without this, filepath.Join(dir, name) would let
// a crafted name escape the target directory. Returns "" when the name has no
// usable base component, so callers fall back to the file id.
func safeDownloadName(serverName string) string {
	base := filepath.Base(serverName)
	switch base {
	case ".", "..", string(filepath.Separator), "":
		return ""
	}
	// filepath.Base strips both / and \ separators; guard against any residual
	// separator just in case. Also reject ':' — on the Windows target it is
	// the NTFS alternate-data-stream separator, so a server name like
	// "report:2024.pdf" (no valid drive letter for filepath.VolumeName to
	// strip) would otherwise write into a hidden stream on a zero-length
	// "report" file instead of a normal file. ':' is invalid in a Windows
	// filename component regardless, so falling back to the file id is safe.
	if strings.ContainsAny(base, `/\:`) {
		return ""
	}
	// Reject Windows reserved device names (CON, NUL, COM1, ...). Windows
	// resolves these to devices even with an extension — "nul.pdf" is the NUL
	// device, "con" is the console — so streamDownloadToFile's rename onto such
	// a base name silently discards the downloaded bytes (or fails a confusing
	// rename) instead of writing the file the user asked for. Fall back to the
	// file id. Applied on every platform (like the ':' guard above) so a file
	// downloaded on Unix and Windows lands the same way.
	if isReservedWindowsBaseName(base) {
		return ""
	}
	return base
}

// isReservedWindowsBaseName reports whether name's stem (the part before the
// first '.') is a Windows reserved device name, matched case-insensitively.
// Windows treats "NUL", "CON", "COM1".."COM9", "LPT1".."LPT9", "PRN", "AUX"
// as devices regardless of any extension, so "nul", "NUL.pdf", and "com1.txt"
// are all reserved.
func isReservedWindowsBaseName(name string) bool {
	stem := name
	if i := strings.IndexByte(stem, '.'); i >= 0 {
		stem = stem[:i]
	}
	// Windows strips trailing spaces (and dots) from a filename component before
	// resolving it, so "NUL " still names the NUL device. Trailing dots are
	// already handled by splitting on the first '.' above; trim trailing spaces
	// here so that variant can't slip past the match.
	stem = strings.TrimRight(stem, " ")
	switch strings.ToUpper(stem) {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return true
	}
	return false
}
