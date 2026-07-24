//go:build !retab_oagen_cli_tables

package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTablesCommandsAreRegistered(t *testing.T) {
	for _, path := range [][]string{
		{"tables", "create"},
		{"tables", "delete"},
		{"tables", "download"},
		{"tables", "get"},
		{"tables", "list"},
		{"tables", "profile"},
		{"tables", "query"},
		{"tables", "replace"},
		{"tables", "schema"},
		{"tables", "validate"},
	} {
		cmd, _, err := rootCmd.Find(path)
		if err != nil {
			t.Fatalf("retab %v is not registered: %v", path, err)
		}
		if cmd == nil || cmd.Name() != path[len(path)-1] {
			t.Fatalf("retab %v resolved to %v", path, cmd)
		}
	}
}

func TestTablesCommandsHonorOutputTable(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	csvPath := filepath.Join(t.TempDir(), "bank-holidays.csv")
	if err := os.WriteFile(csvPath, []byte("countrycode,holidaystartdate\nGB,2026-01-01\nFR,2026-01-02\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/tables":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(tableListFixture())
		case r.Method == http.MethodPut && r.URL.Path == "/v1/tables/tbl_bank":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(tableListFixture())
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tables":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(tableListFixture())
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tables/tbl_bank":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"table": tableFixture()})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tables/tbl_bank/schema":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"table_id": "tbl_bank",
				"columns":  tableFixture()["columns"],
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tables/tbl_bank/profile":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"table_id":  "tbl_bank",
				"row_count": 2,
				"columns": []map[string]any{{
					"name":           "countrycode",
					"json_schema":    map[string]any{"type": "string"},
					"row_count":      2,
					"null_count":     0,
					"empty_count":    0,
					"distinct_count": 2,
					"sample_values":  []string{"GB", "FR"},
				}},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/tables/tbl_bank/validate":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"table_id":    "tbl_bank",
				"has_errors":  false,
				"diagnostics": []any{},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tables/tbl_bank/download":
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write([]byte("countrycode,holidaystartdate\nGB,2026-01-01\nFR,2026-01-02\n"))
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/tables/tbl_bank":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.RequestURI())
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	cases := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "create",
			args: []string{"tables", "create", "--name", "bank_holidays", "--file", csvPath, "--project-id", "proj_bank", "--output", "table"},
			want: []string{"id", "name", "filename", "rows", "columns", "tbl_bank", "bank_holidays"},
		},
		{
			name: "replace",
			args: []string{"tables", "replace", "tbl_bank", "--file", csvPath, "--output", "table"},
			want: []string{"id", "name", "filename", "rows", "columns", "tbl_bank", "bank_holidays"},
		},
		{
			name: "list",
			args: []string{"tables", "list", "--project-id", "proj_bank", "--output-table"},
			want: []string{"id", "name", "filename", "rows", "columns", "tbl_bank", "bank_holidays"},
		},
		{
			name: "get",
			args: []string{"tables", "get", "tbl_bank", "--output", "table"},
			want: []string{"id", "name", "filename", "rows", "columns", "tbl_bank", "bank_holidays"},
		},
		{
			name: "schema",
			args: []string{"tables", "schema", "tbl_bank", "--output", "table"},
			want: []string{"name", "type", "format", "required", "unique", "sample_values", "countrycode"},
		},
		{
			name: "profile",
			args: []string{"tables", "profile", "tbl_bank", "--output", "table"},
			want: []string{"name", "type", "rows", "nulls", "empty", "distinct", "countrycode"},
		},
		{
			name: "validate",
			args: []string{"tables", "validate", "tbl_bank", "--required", "countrycode", "--output", "table"},
			want: []string{"table_id", "status", "diagnostics", "tbl_bank", "ok"},
		},
		{
			name: "download",
			args: []string{"tables", "download", "tbl_bank", "--output", "table"},
			want: []string{"countrycode", "holidaystartdate", "GB", "2026-01-01", "FR"},
		},
		{
			name: "delete",
			args: []string{"tables", "delete", "tbl_bank", "--yes", "--output", "table"},
			want: []string{"id", "status", "tbl_bank", "deleted"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr := captureStd(t, func() {
				if err := runRootForTest(t, tc.args...); err != nil {
					t.Fatalf("retab %v: %v", tc.args, err)
				}
			})
			if stderr != "" {
				t.Fatalf("stderr = %q, want empty", stderr)
			}
			for _, want := range tc.want {
				if !strings.Contains(stdout, want) {
					t.Fatalf("stdout missing %q:\n%s", want, stdout)
				}
			}
			if strings.Contains(stdout, "{") || strings.Contains(stdout, "\"tables\"") || strings.Contains(stdout, "\"table_id\"") {
				t.Fatalf("--output table should not render JSON for %s:\n%s", tc.name, stdout)
			}
		})
	}
}

func TestTablesListEmptyPreservesTablesArray(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	resetProjectID := func() {
		if f := tablesListCmd.Flags().Lookup("project-id"); f != nil {
			_ = f.Value.Set("")
			f.Changed = false
		}
	}
	resetProjectID()
	t.Cleanup(resetProjectID)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/tables" || r.URL.Query().Get("project_id") != "proj_empty" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.RequestURI())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tables":[]}`))
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "list", "--project-id", "proj_empty", "--output", "json"); err != nil {
			t.Fatalf("tables list: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, `"tables": []`) {
		t.Fatalf("empty table list should preserve the public array field, got:\n%s", stdout)
	}
}

func tableListFixture() map[string]any {
	return map[string]any{"tables": []map[string]any{tableFixture()}}
}

func tableFixture() map[string]any {
	return map[string]any{
		"id":        "tbl_bank",
		"name":      "bank_holidays",
		"filename":  "bank-holidays.csv",
		"row_count": 2,
		"columns": []map[string]any{
			{
				"name":          "countrycode",
				"json_schema":   map[string]any{"type": "string"},
				"sample_values": []string{"GB", "FR"},
				"required":      false,
				"unique":        false,
			},
			{
				"name":          "holidaystartdate",
				"json_schema":   map[string]any{"type": "string", "format": "date"},
				"sample_values": []string{"2026-01-01"},
				"required":      false,
				"unique":        false,
			},
		},
		"updated_at": "2026-06-03T11:05:36Z",
	}
}

// normalizeCSVHeaders must never emit the same name twice — the names become
// map keys when CSV rows are rendered as a table, so a collision silently drops
// a column's data. A one-shot `name_N` suffix can itself collide with an
// existing header (["a","a","a_2"] -> "a_2" twice); the disambiguator has to
// keep probing until the result is genuinely unused.
func TestNormalizeCSVHeadersAreUnique(t *testing.T) {
	cases := [][]string{
		{"a", "a", "a_2"},
		{"a", "a", "a"},
		{"a", "a_2", "a"},
		{"", "", "column_1"},
		{"x", "y", "z"},
	}
	for _, headers := range cases {
		got := normalizeCSVHeaders(headers)
		if len(got) != len(headers) {
			t.Fatalf("normalizeCSVHeaders(%v) length = %d, want %d (%v)", headers, len(got), len(headers), got)
		}
		seen := map[string]bool{}
		for _, name := range got {
			if seen[name] {
				t.Fatalf("normalizeCSVHeaders(%v) produced duplicate %q: %v", headers, name, got)
			}
			seen[name] = true
		}
	}
}

func TestParseTableWhereBetweenRequiresTwoValues(t *testing.T) {
	// A well-formed range parses to a two-element value array.
	for _, raw := range []string{"amount between 100..200", "amount between 100,200"} {
		filter, err := parseTableWhereFlag(raw)
		if err != nil {
			t.Fatalf("parseTableWhereFlag(%q) unexpected error: %v", raw, err)
		}
		vals, ok := filter["value"].([]any)
		if !ok || len(vals) != 2 {
			t.Fatalf("parseTableWhereFlag(%q) value = %#v, want a 2-element array", raw, filter["value"])
		}
	}

	// Malformed ranges must all be rejected with a clear error instead of
	// silently sending a bare/one-sided value the server rejects with a 400:
	// a single value, too many comma values, a one-sided range, and an
	// over-long range.
	for _, raw := range []string{
		"amount between 100",
		"amount between 100,200,300",
		"amount between 100..",
		"amount between ..200",
		"amount between 1..2..3",
	} {
		if _, err := parseTableWhereFlag(raw); err == nil {
			t.Fatalf("parseTableWhereFlag(%q) = nil error, want a two-values-required error", raw)
		} else if !strings.Contains(err.Error(), "two values") {
			t.Fatalf("parseTableWhereFlag(%q) error = %q, want it to mention two values", raw, err.Error())
		}
	}
}

func TestTablesDoNotExposeCellLevelMutationCommands(t *testing.T) {
	for _, path := range [][]string{
		{"tables", "columns"},
		{"tables", "rows"},
		{"tables", "update"},
	} {
		cmd, _, err := rootCmd.Find(path)
		if err == nil && cmd != nil && cmd.Name() == path[len(path)-1] {
			t.Fatalf("retab %v should not be registered", path)
		}
	}

	queryCmd, _, err := rootCmd.Find([]string{"tables", "query"})
	if err != nil {
		t.Fatalf("retab tables query is not registered: %v", err)
	}
	if len(queryCmd.Commands()) != 0 {
		t.Fatalf("retab tables query should not expose mutation subcommands")
	}
}

func TestTablesCreateUploadsCSVAsMultipart(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	csvPath := filepath.Join(t.TempDir(), "bank-holidays.csv")
	if err := os.WriteFile(csvPath, []byte("countrycode;holidaystartdate\nGB;2026-01-01\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var sawCreate bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/tables" {
			t.Fatalf("path = %s, want /v1/tables", r.URL.Path)
		}
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data;") {
			t.Fatalf("Content-Type = %q, want multipart/form-data", contentType)
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		if got := r.FormValue("name"); got != "bank_holidays" {
			t.Fatalf("name form value = %q, want bank_holidays", got)
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("file form field: %v", err)
		}
		defer file.Close()
		if header.Filename != "bank-holidays.csv" {
			t.Fatalf("filename = %q, want bank-holidays.csv", header.Filename)
		}
		body, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("read uploaded file: %v", err)
		}
		if !strings.Contains(string(body), "GB;2026-01-01") {
			t.Fatalf("uploaded body = %q", string(body))
		}
		sawCreate = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tables": []map[string]any{
				{
					"id":   "tbl_123",
					"name": "bank_holidays",
				},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "create", "--name", "bank_holidays", "--file", csvPath); err != nil {
			t.Fatalf("tables create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "tbl_123") {
		t.Fatalf("stdout %q does not contain table id", stdout)
	}
	if !sawCreate {
		t.Fatal("server did not receive create request")
	}
}

func TestTablesCreateJSONPrintsCreatedTableObject(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	csvPath := filepath.Join(t.TempDir(), "bank-holidays.csv")
	if err := os.WriteFile(csvPath, []byte("countrycode,holidaystartdate\nGB,2026-01-01\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/tables" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.RequestURI())
		}
		w.Header().Set("Content-Type", "application/json")
		olderTable := tableFixture()
		olderTable["id"] = "tbl_older"
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tables": []map[string]any{
				tableFixture(),
				olderTable,
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		err := runRootForTest(t,
			"tables", "create",
			"--name", "bank_holidays",
			"--file", csvPath,
			"--project-id", "proj_abc123",
			"--output", "json",
		)
		if err != nil {
			t.Fatalf("tables create: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode stdout %q: %v", stdout, err)
	}
	if got["id"] != "tbl_bank" {
		t.Fatalf("id = %v, want tbl_bank; stdout=%s", got["id"], stdout)
	}
	if _, hasTables := got["tables"]; hasTables {
		t.Fatalf("create JSON should be a single table object, got list envelope: %s", stdout)
	}
}

func TestTablesQueryTableOutputRendersRowsAsDataGrid(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var sawQuery bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/tables/tbl_bank/query" {
			t.Fatalf("path = %s, want /v1/tables/tbl_bank/query", r.URL.Path)
		}
		sawQuery = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id": "tbl_bank",
			"columns": []map[string]any{
				{"name": "countrycode", "json_schema": map[string]any{"type": "string"}},
				{"name": "holidaystartdate", "json_schema": map[string]any{"type": "string"}},
			},
			"rows": []map[string]any{
				{
					"id":       "workflow_table_row_0",
					"position": 0,
					"data": map[string]any{
						"countrycode":      "GB",
						"holidaystartdate": "2026-01-01 00:00:00.00",
					},
				},
				{
					"id":       "workflow_table_row_1",
					"position": 1,
					"data": map[string]any{
						"countrycode":      "FR",
						"holidaystartdate": "2026-01-02 00:00:00.00",
					},
				},
			},
			"row_count":          468,
			"filtered_row_count": 468,
			"offset":             0,
			"limit":              2,
			"has_more":           false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "query", "tbl_bank", "--limit", "2", "--output", "table"); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "showing 2 of 468 rows") {
		t.Fatalf("stdout = %q, want summary footer", stdout)
	}
	if !sawQuery {
		t.Fatal("server did not receive query request")
	}
	for _, want := range []string{
		"countrycode",
		"holidaystartdate",
		"GB",
		"2026-01-01 00:00:00.00",
		"FR",
		"2026-01-02 00:00:00.00",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
	for _, unwanted := range []string{"table_id", "workflow_table_row_0", "\"rows\"", "{", "}"} {
		if strings.Contains(stdout, unwanted) {
			t.Fatalf("stdout should be a data table, but contains %q:\n%s", unwanted, stdout)
		}
	}
}

func TestTablesQueryTableOutputShowsPaginationHint(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id": "tbl_bank",
			"columns": []map[string]any{
				{"name": "countrycode"},
			},
			"rows": []map[string]any{
				{
					"id":       "workflow_table_row_3",
					"position": 3,
					"data":     map[string]any{"countrycode": "RO"},
				},
				{
					"id":       "workflow_table_row_4",
					"position": 4,
					"data":     map[string]any{"countrycode": "US"},
				},
			},
			"row_count":          10,
			"filtered_row_count": 10,
			"offset":             3,
			"limit":              2,
			"has_more":           true,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "query", "tbl_bank", "--offset", "3", "--limit", "2", "--output", "table"); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})

	if !strings.Contains(stdout, "RO") || !strings.Contains(stdout, "US") {
		t.Fatalf("expected table rows in stdout, got:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "more rows available; pass --offset 5 to continue") {
		t.Fatalf("expected pagination hint on stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "showing 2 of 10 rows") {
		t.Fatalf("expected summary footer on stdout, got %q", stdout)
	}
}

func TestTablesQueryTableOutputCleansComplexCells(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	longNote := strings.Repeat("0123456789", 12)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id": "tbl_bank",
			"columns": []map[string]any{
				{"name": "countrycode"},
				{"name": "note"},
				{"name": "metadata"},
				{"name": "missing"},
			},
			"rows": []map[string]any{
				{
					"id":       "workflow_table_row_0",
					"position": 0,
					"data": map[string]any{
						"countrycode": "GB",
						"note":        "line one\nline two\twith tabs " + longNote,
						"metadata": map[string]any{
							"region": "eu",
							"rank":   1,
						},
					},
				},
			},
			"row_count":          1,
			"filtered_row_count": 1,
			"offset":             0,
			"limit":              1,
			"has_more":           false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "query", "tbl_bank", "--limit", "1", "--output", "table"); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "showing 1 of 1 rows") {
		t.Fatalf("stdout = %q, want summary footer", stdout)
	}
	if !strings.Contains(stdout, `{"rank":1,"region":"eu"}`) {
		t.Fatalf("expected compact JSON object cell, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "line one line two with tabs") {
		t.Fatalf("expected multiline cell to be collapsed, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "\twith tabs") || strings.Contains(stdout, "\nline two") {
		t.Fatalf("expected no raw newline/tab in cell, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "GB") || !strings.Contains(stdout, " -") {
		t.Fatalf("expected ordinary values and missing-cell placeholder, got:\n%s", stdout)
	}
	if strings.Contains(stdout, longNote) {
		t.Fatalf("expected long cell to be truncated, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "...") {
		t.Fatalf("expected ellipsis for truncated cell, got:\n%s", stdout)
	}
}

func TestTablesQueryNativeFlagsBuildRequestBody(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/tables/tbl_bank/query" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id":            "tbl_bank",
			"columns":             []map[string]any{{"name": "countrycode"}, {"name": "holidaystartdate"}},
			"rows":                []map[string]any{},
			"row_count":           468,
			"filtered_row_count":  0,
			"offset":              0,
			"limit":               3,
			"has_more":            false,
			"next_cursor":         nil,
			"previous_cursor":     nil,
			"query_execution_ms":  nil,
			"snapshot_version_id": nil,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	_, _ = captureStd(t, func() {
		if err := runRootForTest(
			t,
			"tables", "query", "tbl_bank",
			"--where", "countrycode in GB,FR",
			"--where", "holidaystartdate>=2026-01-01",
			"--where", "note is-empty",
			"--where", "amount lt 40",
			"--select", "countrycode,holidaystartdate",
			"--search", "GB",
			"--search-columns", "countrycode,note",
			"--sort", "countrycode:asc",
			"--sort", "holidaystartdate:desc",
			"--limit", "3",
			"--count",
		); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})

	filters, ok := requestBody["filters"].([]any)
	if !ok || len(filters) != 4 {
		t.Fatalf("filters = %#v, want 4 filters", requestBody["filters"])
	}
	firstFilter := filters[0].(map[string]any)
	if firstFilter["column"] != "countrycode" || firstFilter["operator"] != "in" {
		t.Fatalf("first filter = %#v", firstFilter)
	}
	if got := firstFilter["value"].([]any); len(got) != 2 || got[0] != "GB" || got[1] != "FR" {
		t.Fatalf("first filter value = %#v", firstFilter["value"])
	}
	secondFilter := filters[1].(map[string]any)
	if secondFilter["operator"] != "gte" || secondFilter["value"] != "2026-01-01" {
		t.Fatalf("second filter = %#v", secondFilter)
	}
	thirdFilter := filters[2].(map[string]any)
	if thirdFilter["operator"] != "is_empty" {
		t.Fatalf("third filter = %#v", thirdFilter)
	}
	fourthFilter := filters[3].(map[string]any)
	if fourthFilter["operator"] != "lt" || fourthFilter["value"] != float64(40) {
		t.Fatalf("fourth filter = %#v", fourthFilter)
	}
	if got := requestBody["select"].([]any); len(got) != 2 || got[0] != "countrycode" || got[1] != "holidaystartdate" {
		t.Fatalf("select = %#v", requestBody["select"])
	}
	search := requestBody["search"].(map[string]any)
	if search["query"] != "GB" {
		t.Fatalf("search = %#v", search)
	}
	sortRules := requestBody["sort"].([]any)
	if len(sortRules) != 2 || sortRules[1].(map[string]any)["direction"] != "desc" {
		t.Fatalf("sort = %#v", sortRules)
	}
	if requestBody["count_only"] != true {
		t.Fatalf("count_only = %#v", requestBody["count_only"])
	}
}

func TestTablesQueryCSVOutputAndRowMetadata(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id": "tbl_bank",
			"columns":  []map[string]any{{"name": "countrycode"}, {"name": "note"}},
			"rows": []map[string]any{
				{
					"id":       "workflow_table_row_7",
					"position": 7,
					"data":     map[string]any{"countrycode": "GB", "note": "line one\nline two"},
				},
			},
			"row_count":          8,
			"filtered_row_count": 1,
			"offset":             7,
			"limit":              1,
			"has_more":           false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "query", "tbl_bank", "--output", "csv", "--show-row-id", "--show-position"); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})

	if !strings.Contains(stdout, "row_id,position,countrycode,note") {
		t.Fatalf("missing CSV header:\n%s", stdout)
	}
	if !strings.Contains(stdout, "workflow_table_row_7,7,GB,line one line two") {
		t.Fatalf("missing CSV row:\n%s", stdout)
	}
	if !strings.Contains(stderr, "showing 1 of 1 rows") {
		t.Fatalf("missing summary footer: %q", stderr)
	}
}

// Regression: CSV is meant to be a faithful, re-importable table, but the
// cell renderer applied the table grid's --max-width truncation (default 96)
// to CSV too — silently rewriting any long cell to "<prefix>..." and breaking
// round-trips. CSV must emit the full value regardless of --max-width.
func TestTablesQueryCSVDoesNotTruncateLongCells(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	longNote := strings.Repeat("abcd ", 40) // 200 chars, well past the 96 default
	longNote = strings.TrimSpace(longNote)
	expected := strings.Join(strings.Fields(longNote), " ")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id": "tbl_long",
			"columns":  []map[string]any{{"name": "note"}},
			"rows": []map[string]any{
				{"id": "r0", "position": 0, "data": map[string]any{"note": longNote}},
			},
			"row_count":          1,
			"filtered_row_count": 1,
			"offset":             0,
			"limit":              1,
			"has_more":           false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "query", "tbl_long", "--output", "csv"); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})

	if !strings.Contains(stdout, expected) {
		t.Fatalf("CSV cell should be the full untruncated value, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "...") {
		t.Fatalf("CSV must not truncate long cells with an ellipsis, got:\n%s", stdout)
	}
}

func TestTablesQueryCSVRendersNullAsEmptyCell(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id": "tbl_nulls",
			"columns":  []map[string]any{{"name": "name"}, {"name": "age"}, {"name": "salary"}},
			"rows": []map[string]any{
				{
					"id":       "workflow_table_row_5",
					"position": 5,
					"data":     map[string]any{"name": "Frank", "age": nil, "salary": nil},
				},
			},
			"row_count":          6,
			"filtered_row_count": 1,
			"offset":             5,
			"limit":              1,
			"has_more":           false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "query", "tbl_nulls", "--output", "csv"); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})

	// Null cells must be empty in CSV (a faithful, re-importable CSV), not the
	// human "-" placeholder used in table output.
	if !strings.Contains(stdout, "Frank,,") {
		t.Fatalf("null cells should render as empty CSV fields, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "-") {
		t.Fatalf("CSV output must not contain the table null placeholder '-':\n%s", stdout)
	}
}

// CSV must render JSON scalars faithfully and re-importably: booleans as
// lowercase true/false (not Go/Python-style True/False), integral numbers
// without a trailing .0, and fractional numbers at full precision.
func TestTablesQueryCSVRendersScalarsFaithfully(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id": "tbl_scalars",
			"columns":  []map[string]any{{"name": "active"}, {"name": "age"}, {"name": "salary"}},
			"rows": []map[string]any{
				{"id": "r0", "position": 0, "data": map[string]any{"active": true, "age": 30, "salary": 50000.5}},
				{"id": "r1", "position": 1, "data": map[string]any{"active": false, "age": 25, "salary": 42000}},
			},
			"row_count":          2,
			"filtered_row_count": 2,
			"offset":             0,
			"limit":              2,
			"has_more":           false,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, _ := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "query", "tbl_scalars", "--output", "csv"); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})

	if !strings.Contains(stdout, "active,age,salary") {
		t.Fatalf("missing CSV header:\n%s", stdout)
	}
	if !strings.Contains(stdout, "true,30,50000.5") {
		t.Fatalf("expected faithful row 'true,30,50000.5', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "false,25,42000") {
		t.Fatalf("expected faithful row 'false,25,42000' (no trailing .0), got:\n%s", stdout)
	}
	if strings.Contains(stdout, "True") || strings.Contains(stdout, "False") {
		t.Fatalf("booleans must be lowercase in CSV, got:\n%s", stdout)
	}
}

func TestTablesQueryAllFetchesEveryPage(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	offsets := []int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		offset := int(body["offset"].(float64))
		offsets = append(offsets, offset)
		hasMore := offset == 0
		rowID := "workflow_table_row_0"
		if offset > 0 {
			rowID = "workflow_table_row_1"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"table_id": "tbl_bank",
			"columns":  []map[string]any{{"name": "countrycode"}},
			"rows": []map[string]any{{
				"id":       rowID,
				"position": offset,
				"data":     map[string]any{"countrycode": rowID},
			}},
			"row_count":          2,
			"filtered_row_count": 2,
			"offset":             offset,
			"limit":              1,
			"has_more":           hasMore,
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := runRootForTest(t, "tables", "query", "tbl_bank", "--all", "--offset", "0", "--limit", "1", "--output", "table"); err != nil {
			t.Fatalf("tables query: %v", err)
		}
	})

	if len(offsets) != 2 || offsets[0] != 0 || offsets[1] != 1 {
		t.Fatalf("offsets = %#v, want [0 1]", offsets)
	}
	if !strings.Contains(stdout, "workflow_table_row_0") || !strings.Contains(stdout, "workflow_table_row_1") {
		t.Fatalf("missing merged rows:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "showing 2 of 2 rows") {
		t.Fatalf("missing all summary: %q", stdout)
	}
	if strings.Contains(stdout, "more rows available") {
		t.Fatalf("--all should consume pagination, got %q", stdout)
	}
}

func TestTablesSchemaProfileValidateCommandsUsePublicRoutes(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tables/tbl_bank/schema":
			_ = json.NewEncoder(w).Encode(map[string]any{"table_id": "tbl_bank", "columns": []map[string]any{{"name": "countrycode"}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tables/tbl_bank/profile":
			if got := r.URL.Query()["select"]; len(got) != 1 || got[0] != "countrycode,holidaystartdate" {
				t.Fatalf("profile query = %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"table_id": "tbl_bank", "row_count": 1, "columns": []map[string]any{}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/tables/tbl_bank/validate":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode validate body: %v", err)
			}
			if body["columns"].(map[string]any)["countrycode"].(map[string]any)["is_not_empty"] != true {
				t.Fatalf("validate body = %#v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"table_id": "tbl_bank", "diagnostics": []any{}, "has_errors": false})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.RequestURI())
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for _, args := range [][]string{
		{"tables", "schema", "tbl_bank"},
		{"tables", "profile", "tbl_bank", "--select", "countrycode, holidaystartdate"},
		{"tables", "validate", "tbl_bank", "--required", "countrycode", "--not-empty", "countrycode", "--unique", "countrycode"},
	} {
		_, _ = captureStd(t, func() {
			if err := runRootForTest(t, args...); err != nil {
				t.Fatalf("retab %v: %v", args, err)
			}
		})
	}
	if len(requests) != 3 {
		t.Fatalf("requests = %#v", requests)
	}
}
