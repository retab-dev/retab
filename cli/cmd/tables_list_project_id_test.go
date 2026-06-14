//go:build !retab_oagen_cli_tables

package cmd

import (
	"strings"
	"testing"
)

// The backend gates GET /v1/tables on FGA project_from_query_project_id, which
// returns 400 "Project ID is required" without a project_id query param (listing
// is authorized as project:view on the named project). The CLI must require
// --project-id locally instead of letting the user hit a confusing server 400.
func TestTablesListRequiresProjectID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	// rootCmd is shared across tests and cobra retains parsed flag state, so a
	// prior `tables list --project-id ...` would leak in here and mark the flag
	// Changed. Reset to a pristine "not set" state so required-flag validation
	// fires, making this test order-independent.
	if f := tablesListCmd.Flags().Lookup("project-id"); f != nil {
		_ = f.Value.Set("")
		f.Changed = false
	}
	t.Cleanup(func() {
		if f := tablesListCmd.Flags().Lookup("project-id"); f != nil {
			_ = f.Value.Set("")
			f.Changed = false
		}
	})

	err := runRootForTest(t, "tables", "list")
	if err == nil {
		t.Fatalf("expected an error when --project-id is omitted")
	}
	if !strings.Contains(err.Error(), "project-id") {
		t.Fatalf("error = %q, want it to mention the required project-id flag", err.Error())
	}
}
