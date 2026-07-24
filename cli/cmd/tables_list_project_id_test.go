//go:build !retab_oagen_cli_tables

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// addColumnSchemaOverridesField must treat an empty/whitespace value — even one
// passed explicitly as "" — as "not provided": it must neither register a bogus
// empty multipart field nor (on `tables replace`, which keys schema preservation
// off the field's presence) silently disable that default. Non-empty valid JSON
// is registered; invalid JSON is rejected before any upload.
func TestAddColumnSchemaOverridesFieldEmptyIsAbsent(t *testing.T) {
	newCmd := func(value string, changed bool) *cobra.Command {
		c := &cobra.Command{Use: "x"}
		c.Flags().String("column-schema-overrides", "", "")
		if changed {
			_ = c.Flags().Set("column-schema-overrides", value)
		}
		return c
	}

	// Explicitly-passed empty string: must not register the key.
	fields := map[string]string{}
	if err := addColumnSchemaOverridesField(newCmd("", true), fields); err != nil {
		t.Fatalf("explicit empty: unexpected error %v", err)
	}
	if _, ok := fields["column_schema_overrides"]; ok {
		t.Fatalf("explicit empty value registered a bogus field: %v", fields)
	}

	// Whitespace-only: also treated as absent.
	fields = map[string]string{}
	if err := addColumnSchemaOverridesField(newCmd("   ", true), fields); err != nil {
		t.Fatalf("whitespace: unexpected error %v", err)
	}
	if _, ok := fields["column_schema_overrides"]; ok {
		t.Fatalf("whitespace value registered a bogus field: %v", fields)
	}

	// Flag not set at all: absent.
	fields = map[string]string{}
	if err := addColumnSchemaOverridesField(newCmd("", false), fields); err != nil {
		t.Fatalf("unset: unexpected error %v", err)
	}
	if _, ok := fields["column_schema_overrides"]; ok {
		t.Fatalf("unset flag registered a field: %v", fields)
	}

	// Valid JSON object: registered verbatim.
	fields = map[string]string{}
	valid := `{"code":{"type":"string"}}`
	if err := addColumnSchemaOverridesField(newCmd(valid, true), fields); err != nil {
		t.Fatalf("valid JSON: unexpected error %v", err)
	}
	if fields["column_schema_overrides"] != valid {
		t.Fatalf("valid JSON not registered: %v", fields)
	}

	// Invalid JSON: rejected locally.
	if err := addColumnSchemaOverridesField(newCmd("not json", true), map[string]string{}); err == nil {
		t.Fatal("expected invalid JSON to be rejected")
	}
}

// Invalid --column-schema-overrides must be rejected locally with a clear error
// (mirroring --filters validation), before any upload, instead of surfacing as a
// confusing server-side error. The local JSON check must fire before the network.
func TestTablesCreateRejectsInvalidColumnSchemaOverrides(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	csvPath := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(csvPath, []byte("name,amount\nAlpha,100\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// rootCmd is shared across tests and cobra retains parsed flag state; reset
	// the overrides flag so this test is order-independent.
	if f := tablesCreateCmd.Flags().Lookup("column-schema-overrides"); f != nil {
		_ = f.Value.Set("")
		f.Changed = false
	}
	t.Cleanup(func() {
		if f := tablesCreateCmd.Flags().Lookup("column-schema-overrides"); f != nil {
			_ = f.Value.Set("")
			f.Changed = false
		}
	})

	// The validation error originates inside RunE, so executeRoot converts it to
	// errSilent (message already printed to stderr). Assert on stderr + non-nil.
	args := []string{"tables", "create", "--name", "demo", "--file", csvPath, "--project-id", "proj_abc123", "--column-schema-overrides", "not json"}
	var err error
	_, stderr := captureStd(t, func() {
		err = runRootForTest(t, args...)
	})
	if err == nil {
		t.Fatalf("expected an error for invalid --column-schema-overrides")
	}
	if !strings.Contains(stderr, "column-schema-overrides") {
		t.Fatalf("stderr = %q, want it to mention --column-schema-overrides", stderr)
	}
}

// The backend gates GET /v1/tables on FGA project_from_query_project_id, which
// returns 400 "Project ID is required" without a project_id query param (listing
// is authorized as project:view on the named project). The CLI must require
// --project-id locally instead of letting the user hit a confusing server 400.
func TestTablesListRequiresProjectID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
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
