//go:build !retab_oagen_cli_tables

package cmd

import (
	"encoding/json"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The backend requires project_id as a multipart form field on POST /v1/tables
// (handlers.createHandler returns 422 "Project ID is required" without it). The
// CLI must expose a --project-id flag and forward it in the upload, otherwise
// `retab tables create` is impossible to use.
func TestTablesCreateSendsProjectID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	csvPath := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(csvPath, []byte("name,amount\nAlpha,100\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var gotProjectID string
	var sawProjectField bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/tables" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.RequestURI())
		}
		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatalf("parse content-type: %v", err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		form, err := mr.ReadForm(1 << 20)
		if err != nil {
			t.Fatalf("read multipart form: %v", err)
		}
		if vs, ok := form.Value["project_id"]; ok && len(vs) > 0 {
			sawProjectField = true
			gotProjectID = vs[0]
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tableListFixture())
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	args := []string{"tables", "create", "--name", "demo", "--file", csvPath, "--project-id", "proj_abc123"}
	if err := runRootForTest(t, args...); err != nil {
		t.Fatalf("retab %v: %v", args, err)
	}
	if !sawProjectField {
		t.Fatalf("multipart body did not include a project_id field")
	}
	if gotProjectID != "proj_abc123" {
		t.Fatalf("project_id = %q, want %q", gotProjectID, "proj_abc123")
	}
}

// project_id is required by the backend, so the CLI should reject a create that
// omits it with a clear local error instead of a server 422.
func TestTablesCreateRequiresProjectID(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	csvPath := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(csvPath, []byte("name,amount\nAlpha,100\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// rootCmd is shared across tests and cobra retains parsed flag state on it,
	// so a prior `tables create --project-id ...` would leak in here and mark
	// the flag Changed. Reset to a pristine "not set" state so cobra's required-
	// flag validation fires, making this test order-independent.
	if f := tablesCreateCmd.Flags().Lookup("project-id"); f != nil {
		_ = f.Value.Set("")
		f.Changed = false
	}
	t.Cleanup(func() {
		if f := tablesCreateCmd.Flags().Lookup("project-id"); f != nil {
			_ = f.Value.Set("")
			f.Changed = false
		}
	})

	args := []string{"tables", "create", "--name", "demo", "--file", csvPath}
	err := runRootForTest(t, args...)
	if err == nil {
		t.Fatalf("expected an error when --project-id is omitted")
	}
	if !strings.Contains(err.Error(), "project-id") {
		t.Fatalf("error = %q, want it to mention the required project-id flag", err.Error())
	}
}
