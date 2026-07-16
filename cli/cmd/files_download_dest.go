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
	// separator (or a Windows drive-relative remnant like "C:") just in case.
	if strings.ContainsAny(base, `/\`) {
		return ""
	}
	return base
}
