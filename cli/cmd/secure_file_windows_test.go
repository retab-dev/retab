//go:build windows

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

// secureConfigFile must restrict the file to the current user without locking
// the owner out — i.e. the owner can still read and write it afterward.
func TestSecureConfigFileWindowsOwnerStillHasAccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"version":2}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := secureConfigFile(path); err != nil {
		t.Fatalf("secureConfigFile: %v", err)
	}
	// Owner must still be able to read it.
	if _, err := os.ReadFile(path); err != nil {
		t.Fatalf("owner can no longer read secured file: %v", err)
	}
	// ...and write it (the save path overwrites/renames over it).
	if err := os.WriteFile(path, []byte(`{"version":2,"api_key":"x"}`), 0o600); err != nil {
		t.Fatalf("owner can no longer write secured file: %v", err)
	}

	// Confirm the ACL is actually restrictive (not a silent no-op): the DACL
	// must be PROTECTED (inheritance stripped) and hold exactly one ACE — the
	// single owner grant we set.
	sd, err := windows.GetNamedSecurityInfo(
		path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatalf("GetNamedSecurityInfo: %v", err)
	}
	control, _, err := sd.Control()
	if err != nil {
		t.Fatalf("Control: %v", err)
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		t.Errorf("DACL is not protected; inherited ACEs were not stripped")
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		t.Fatalf("DACL: %v", err)
	}
	if dacl == nil {
		t.Fatal("nil DACL (everyone allowed)")
	}
	if got := dacl.AceCount; got != 1 {
		t.Errorf("DACL has %d ACEs, want exactly 1 (owner-only)", got)
	}
}
