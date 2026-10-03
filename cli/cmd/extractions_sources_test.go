//go:build !retab_oagen_cli_extractions

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractionSourcesModesOnWire(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-sources-key")
	for _, tc := range []struct {
		name                string
		flags               map[string]string
		method, query, mode string
	}{
		{"legacy", nil, "GET", "", ""},
		{"located", map[string]string{"mode": "located", "background": "true"}, "POST", "", "located"},
		{"cited", map[string]string{"mode": "cited", "background": "true"}, "POST", "", "cited"},
		{"poll", map[string]string{"mode": "located", "poll": "true", "job-id": "src_123"}, "GET", "job_id=src_123&mode=located", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != tc.method || r.URL.Path != "/v1/extractions/extr_123/sources" || r.URL.RawQuery != tc.query {
					t.Errorf("request: %s %s", r.Method, r.URL)
				}
				if tc.mode != "" {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["mode"] != tc.mode || body["background"] != true {
						t.Errorf("body: %v %v", body, err)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"object":"extraction.sources","job":{"id":"src_123","status":"pending"}}`))
			}))
			defer server.Close()
			t.Setenv("RETAB_API_BASE_URL", server.URL)
			cmd := extractionsSourcesCmd
			cmd.SetOut(&strings.Builder{})
			for name, value := range tc.flags {
				if err := cmd.Flags().Set(name, value); err != nil {
					t.Fatal(err)
				}
			}
			t.Cleanup(func() {
				for name := range tc.flags {
					flag := cmd.Flags().Lookup(name)
					_ = cmd.Flags().Set(name, flag.DefValue)
					flag.Changed = false
				}
			})
			if err := cmd.RunE(cmd, []string{"extr_123"}); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("requests=%d", calls)
			}
		})
	}
}
