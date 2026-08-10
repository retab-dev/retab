package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Round-9 bug hunt. Each test pins a fix corroborated by reading the code and
// verified by independent review agents.

// A fresh-install scoped login (`auth login --env X --api-key ...`) legitimately
// claims DefaultEnvironment so plain commands work. But branch 5 of
// resolveCredential outranks the top-level api_key (branch 8), the access token
// (6), and the OAuth session (7) — and the top-level login paths only wiped the
// OTHER credential slots, never DefaultEnvironment. So a later unscoped
// `auth login --api-key ...` stored the new key but the stale default profile
// kept winning: the login reported success while plain `retab ...` silently used
// the old key. The unscoped login paths now clear DefaultEnvironment.
func TestUnscopedAPIKeyLoginClearsFreshInstallDefault(t *testing.T) {
	isolateHome(t)

	// Fresh install: a scoped login claims the default slot.
	captureStd(t, func() {
		if err := runAPIKeyLogin("rt_test_aaaa", "", "test"); err != nil {
			t.Fatalf("scoped login: %v", err)
		}
	})
	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultEnvironment != "test" {
		t.Fatalf("fresh-install scoped login should claim default; got %q", cfg.DefaultEnvironment)
	}

	// A later unscoped api-key login must take effect for plain commands.
	captureStd(t, func() {
		if err := runAPIKeyLogin("rt_live_bbbb", "", ""); err != nil {
			t.Fatalf("unscoped login: %v", err)
		}
	})
	cfg, err = loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultEnvironment != "" {
		t.Errorf("unscoped login left default_environment=%q; the stale profile still shadows the new key", cfg.DefaultEnvironment)
	}

	cred, err := resolveCredential(newTestRootCmd())
	if err != nil {
		t.Fatalf("resolveCredential: %v", err)
	}
	if cred.Source != sourceLegacyKey || cred.APIKey != "rt_live_bbbb" {
		t.Errorf("plain invocation resolved source=%q key=%q; the new top-level key must win", cred.Source, cred.APIKey)
	}

	// The named profile is untouched and still reachable via --env.
	selector := newTestRootCmd()
	_ = selector.PersistentFlags().Set("env", "test")
	if cred, err = resolveCredential(selector); err != nil {
		t.Fatalf("--env test: %v", err)
	}
	if cred.APIKey != "rt_test_aaaa" {
		t.Errorf("--env test resolved %q, want the preserved profile key", cred.APIKey)
	}
}

// Same shadowing, via the access-token login path.
func TestUnscopedAccessTokenLoginClearsFreshInstallDefault(t *testing.T) {
	isolateHome(t)

	captureStd(t, func() {
		if err := runAPIKeyLogin("rt_test_aaaa", "", "test"); err != nil {
			t.Fatalf("scoped login: %v", err)
		}
	})
	captureStd(t, func() {
		if err := runAccessTokenLogin("acctk_zzzz", ""); err != nil {
			t.Fatalf("access-token login: %v", err)
		}
	})

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultEnvironment != "" {
		t.Errorf("access-token login left default_environment=%q; it shadows the new token", cfg.DefaultEnvironment)
	}

	cred, err := resolveCredential(newTestRootCmd())
	if err != nil {
		t.Fatalf("resolveCredential: %v", err)
	}
	if cred.Source != sourceAccessToken {
		t.Errorf("plain invocation resolved %q; the stored access token must win", cred.Source)
	}
}

// loadConfig returns an error for a corrupt config file. The login paths used to
// discard it (`cfg, _ := loadConfig()`), then saveConfig(zeroCfg) would clobber
// ~/.retab/config.json — dropping every other stored environment profile. The
// login paths now treat a load failure as fatal, matching `env switch`.
func TestLoginDoesNotClobberCorruptConfig(t *testing.T) {
	home := isolateHome(t)
	// A pre-existing profile store, then a corrupting stray byte appended.
	if err := saveConfig(retabConfig{Environments: map[string]*environmentProfile{
		"staging": {Name: "staging", APIKey: "rt_test_keep"},
	}}); err != nil {
		t.Fatal(err)
	}
	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	corrupt := append([]byte(nil), original...)
	corrupt = append(corrupt, []byte("}}garbage")...)
	if err := os.WriteFile(path, corrupt, 0o600); err != nil {
		t.Fatal(err)
	}

	captureStd(t, func() {
		if err := runAPIKeyLogin("rt_live_new", "", ""); err == nil {
			t.Fatal("login on a corrupt config must fail, not silently reset it")
		}
	})

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, corrupt) {
		t.Errorf("login rewrote the corrupt config (home=%s); the stored profiles were at risk of being clobbered", home)
	}
	if !strings.Contains(string(after), "rt_test_keep") {
		t.Error("the pre-existing profile key was lost")
	}
}

// stageStdinUpload filepath.Joins --filename into a temp dir and writes stdin to
// it. A Windows reserved device name (CON, NUL, "aux.pdf", ...) resolves to the
// device in any directory, so the write is discarded and the later read returns
// 0 bytes — a silent 0-byte upload. The stdin staging path now rejects them,
// mirroring the download-side guard, on every platform for consistent behavior.
func TestStageStdinUploadRejectsReservedWindowsName(t *testing.T) {
	for _, name := range []string{"nul", "CON", "aux.pdf", "com1.txt", "NUL ", "lpt9"} {
		cmd := &cobra.Command{}
		cmd.Flags().String("filename", "", "")
		_ = cmd.Flags().Set("filename", name)
		cmd.SetIn(strings.NewReader("payload bytes"))

		_, cleanup, err := stageStdinUpload(cmd)
		if cleanup != nil {
			cleanup()
		}
		if err == nil {
			t.Errorf("stageStdinUpload(--filename %q) = nil error; a reserved device name must be rejected", name)
		}
	}

	// A normal name still stages successfully.
	cmd := &cobra.Command{}
	cmd.Flags().String("filename", "", "")
	_ = cmd.Flags().Set("filename", "invoice.pdf")
	cmd.SetIn(strings.NewReader("payload bytes"))
	staged, cleanup, err := stageStdinUpload(cmd)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		t.Fatalf("stageStdinUpload(invoice.pdf): %v", err)
	}
	if data, rerr := os.ReadFile(staged); rerr != nil || string(data) != "payload bytes" {
		t.Fatalf("staged file = %q, err %v; want the piped bytes", data, rerr)
	}
}

// The hand-rolled `tables query --csv` renderer wrote cells straight to the
// csv.Writer without the formula-injection guard that the shared writeCSV core
// applies, so a cell beginning `=`/`@`/`+`/`-` executed when the CSV opened in a
// spreadsheet. renderWorkflowTableRowsCSV now routes each cell through
// sanitizeCSVCell, matching `tables list --output csv`.
func TestTablesQueryCSVSanitizesFormulaInjection(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id": "tbl_x",
			"columns":  []map[string]any{{"name": "formula"}},
			"rows": []map[string]any{
				{"id": "r0", "position": 0, "data": map[string]any{"formula": `=HYPERLINK("http://evil","x")`}},
			},
			"row_count": 1, "filtered_row_count": 1, "offset": 0, "limit": 1, "has_more": false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "query", "tbl_x", "--output", "csv"); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})
	if strings.Contains(stdout, "\n=HYPERLINK") || strings.HasPrefix(stdout, "=HYPERLINK") {
		t.Errorf("formula cell emitted unsanitized:\n%s", stdout)
	}
	if !strings.Contains(stdout, `'=HYPERLINK`) {
		t.Errorf("formula cell was not neutralized with a leading quote:\n%s", stdout)
	}
}
