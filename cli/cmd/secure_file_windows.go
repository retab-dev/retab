//go:build windows

package cmd

import "golang.org/x/sys/windows"

// secureConfigFile restricts path to the current user only: an owner-only DACL
// with inheritance disabled. On Windows os.Chmod cannot express this (it only
// toggles the read-only bit), so a config file holding the OAuth refresh_token
// and API keys would otherwise inherit the parent directory's ACL and be
// readable by other local accounts.
func secureConfigFile(path string) error {
	// Resolve the current user's SID from the process token.
	token := windows.GetCurrentProcessToken()
	user, err := token.GetTokenUser()
	if err != nil {
		return err
	}
	sid := user.User.Sid

	// A single ACE granting the current user full control; nothing else.
	entries := []windows.EXPLICIT_ACCESS{
		{
			AccessPermissions: windows.GENERIC_ALL,
			AccessMode:        windows.GRANT_ACCESS,
			Inheritance:       windows.NO_INHERITANCE,
			Trustee: windows.TRUSTEE{
				TrusteeForm:  windows.TRUSTEE_IS_SID,
				TrusteeType:  windows.TRUSTEE_IS_USER,
				TrusteeValue: windows.TrusteeValueFromSID(sid),
			},
		},
	}
	dacl, err := windows.ACLFromEntries(entries, nil)
	if err != nil {
		return err
	}

	// Apply the DACL and mark it PROTECTED so inherited ACEs (e.g. a broad
	// "Users" grant from the parent directory) are stripped.
	return windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, dacl, nil,
	)
}
