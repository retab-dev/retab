package cmd

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCLIUploadsLargeWorkflowDocumentsBeforeSubmission(t *testing.T) {
	for _, sizes := range [][]int{{25 * 1024 * 1024}, {12 * 1024 * 1024, 12 * 1024 * 1024}} {
		t.Run(string(rune('0'+len(sizes)))+" documents", func(t *testing.T) {
			uploads, finalized, submitted := 0, 0, false
			storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut || r.Header.Get("Authorization") != "" {
					t.Errorf("storage request leaked credentials or used wrong method")
				}
				n, err := io.Copy(io.Discard, r.Body)
				if err != nil || n != int64(sizes[uploads]) {
					t.Errorf("upload bytes=%d error=%v", n, err)
				}
				uploads++
			}))
			defer storage.Close()
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/v1/files/upload":
					_ = json.NewEncoder(w).Encode(map[string]any{"fileId": "file_large", "uploadUrl": storage.URL, "uploadMethod": "PUT"})
				case "/v1/files/upload/file_large/complete":
					finalized++
					_, _ = io.WriteString(w, `{"filename":"large.pdf","url":"https://storage.retab.com/file_large"}`)
				case "/v1/workflows/runs":
					body, _ := io.ReadAll(io.LimitReader(r.Body, 32*1024*1024+1))
					if len(body) > 32*1024*1024 || finalized != len(sizes) {
						t.Errorf("submitted before uploads or with oversized body: finalized=%d bytes=%d", finalized, len(body))
					}
					var request struct {
						Documents map[string]struct{ URL string }
					}
					if err := json.Unmarshal(body, &request); err != nil {
						t.Error(err)
					}
					for _, document := range request.Documents {
						if document.URL != "https://storage.retab.com/file_large" {
							t.Error("workflow document is still inline")
						}
					}
					submitted = true
					_, _ = io.WriteString(w, `{"id":"run_1"}`)
				default:
					t.Errorf("unexpected API request %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer api.Close()
			t.Setenv("RETAB_API_KEY", "rt_test_key")
			t.Setenv("RETAB_API_BASE_URL", api.URL)
			cmd := &cobra.Command{Use: "upload-regression"}
			rootCmd.AddCommand(cmd)
			t.Cleanup(func() { rootCmd.RemoveCommand(cmd) })
			documents := map[string]any{}
			for i, size := range sizes {
				documents[string(rune('a'+i))] = map[string]string{"filename": "large.pdf", "url": "data:application/pdf;base64," + base64.StdEncoding.EncodeToString([]byte(strings.Repeat("A", size)))}
			}
			_, err := cliJSONRequest(cmd, http.MethodPost, "/v1/workflows/runs", nil, map[string]any{"workflow_id": "wf_1", "documents": documents})
			if err != nil || !submitted || uploads != len(sizes) {
				t.Fatalf("submission=%t uploads=%d error=%v", submitted, uploads, err)
			}
		})
	}
}
