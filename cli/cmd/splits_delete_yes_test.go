//go:build !retab_oagen_cli_splits

package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestSplitsDeleteWithYesFlagProceedsWithoutPrompt(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var sawDelete atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/v1/splits/split_to_delete" {
			sawDelete.Add(1)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := splitsDeleteCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = splitsDeleteCmd.Flags().Set("yes", "false") })

	if err := splitsDeleteCmd.RunE(splitsDeleteCmd, []string{"split_to_delete"}); err != nil {
		t.Fatalf("delete with --yes: %v", err)
	}
	if sawDelete.Load() != 1 {
		t.Fatalf("expected one DELETE call, got %d", sawDelete.Load())
	}
}

func TestSplitsDeleteWithoutYesAndNonTTYStdinRefuses(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := splitsDeleteCmd.Flags().Set("yes", "false"); err != nil {
		t.Fatal(err)
	}

	err := splitsDeleteCmd.RunE(splitsDeleteCmd, []string{"split_keep"})
	if err == nil {
		t.Fatal("expected refusal when stdin is not a TTY")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("error %q does not mention --yes", err.Error())
	}
	if hits.Load() != 0 {
		t.Fatalf("server was hit %d time(s), want 0", hits.Load())
	}
}
