package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestParsesCreateRejectsInvalidTableParsingFormatBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if err := parsesCreateCmd.Flags().Set("table-parsing-format", "banana"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = parsesCreateCmd.Flags().Set("table-parsing-format", "") })

	var err error
	_, stderr := captureStd(t, func() {
		err = parsesCreateCmd.RunE(parsesCreateCmd, nil)
	})
	if err == nil {
		t.Fatal("expected invalid table parsing format error")
	}
	if !strings.Contains(stderr, "invalid --table-parsing-format") {
		t.Fatalf("stderr %q does not mention invalid table parsing format", stderr)
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("server was hit %d time(s), want no requests", got)
	}
}
