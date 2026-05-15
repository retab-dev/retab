package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestWorkflowsArtifactsRejectInvalidOperationBeforeRequest(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "server should not be reached", http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

	cases := []struct {
		name string
		cmd  func() error
	}{
		{
			name: "get positional operation",
			cmd: func() error {
				return workflowsArtifactsGetCmd.RunE(workflowsArtifactsGetCmd, []string{"extract", "art_123"})
			},
		},
		{
			name: "list filter operation",
			cmd: func() error {
				if err := workflowsArtifactsListCmd.Flags().Set("operation", "extract"); err != nil {
					return err
				}
				t.Cleanup(func() { _ = workflowsArtifactsListCmd.Flags().Set("operation", "") })
				return workflowsArtifactsListCmd.RunE(workflowsArtifactsListCmd, []string{"run_123"})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := hits.Load()
			var err error
			_, stderr := captureStd(t, func() {
				err = tc.cmd()
			})
			if err == nil {
				t.Fatal("expected invalid operation error")
			}
			if !strings.Contains(stderr, "invalid operation") {
				t.Fatalf("stderr %q does not mention invalid operation", stderr)
			}
			if got := hits.Load(); got != before {
				t.Fatalf("server was hit %d time(s), want 0", got-before)
			}
		})
	}
}

func TestWorkflowsArtifactsExamplesUseBackendOperationNames(t *testing.T) {
	for _, example := range []string{workflowsArtifactsCmd.Example, workflowsArtifactsGetCmd.Example, workflowsArtifactsListCmd.Example} {
		if strings.Contains(example, " get extract ") {
			t.Fatalf("artifact examples should use extraction, got:\n%s", example)
		}
	}
	if !strings.Contains(workflowsArtifactsGetCmd.Example, "get extraction") {
		t.Fatalf("artifact get example should include extraction operation, got:\n%s", workflowsArtifactsGetCmd.Example)
	}
}
