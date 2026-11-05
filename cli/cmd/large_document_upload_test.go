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

func TestCLIUploadsLargeDocumentsBeforeSubmission(t *testing.T) {
	for _, tc := range []struct {
		name    string
		sizes   []int
		asArray bool
	}{
		{name: "single workflow document", sizes: []int{25 * 1024 * 1024}},
		{name: "workflow document map", sizes: []int{12 * 1024 * 1024, 12 * 1024 * 1024}},
		{name: "schema document array", sizes: []int{15 * 1024 * 1024, 15 * 1024 * 1024}, asArray: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sizes := tc.sizes
			endpoint := "/v1/workflows/runs"
			if tc.asArray {
				endpoint = "/v1/schemas/generate"
			}
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
				case endpoint:
					body, _ := io.ReadAll(io.LimitReader(r.Body, 32*1024*1024+1))
					if len(body) > 32*1024*1024 || finalized != len(sizes) {
						t.Errorf("submitted before uploads or with oversized body: finalized=%d bytes=%d", finalized, len(body))
					}
					var request struct {
						Documents json.RawMessage
					}
					if err := json.Unmarshal(body, &request); err != nil {
						t.Error(err)
					}
					var references []struct{ URL string }
					if tc.asArray {
						if err := json.Unmarshal(request.Documents, &references); err != nil {
							t.Error(err)
						}
					} else {
						var mapped map[string]struct{ URL string }
						if err := json.Unmarshal(request.Documents, &mapped); err != nil {
							t.Error(err)
						}
						for _, reference := range mapped {
							references = append(references, reference)
						}
					}
					if len(references) != len(sizes) {
						t.Errorf("document count changed: %d", len(references))
					}
					for _, document := range references {
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
			array := make([]map[string]string, 0, len(sizes))
			for i, size := range sizes {
				document := map[string]string{"filename": "large.pdf", "url": "data:application/pdf;base64," + base64.StdEncoding.EncodeToString([]byte(strings.Repeat("A", size)))}
				documents[string(rune('a'+i))] = document
				array = append(array, document)
			}
			body := map[string]any{"workflow_id": "wf_1", "documents": documents}
			if tc.asArray {
				body = map[string]any{"documents": array}
			}
			original, _ := json.Marshal(body)
			_, err := cliJSONRequest(cmd, http.MethodPost, endpoint, nil, body)
			if err != nil || !submitted || uploads != len(sizes) {
				t.Fatalf("submission=%t uploads=%d error=%v", submitted, uploads, err)
			}
			after, _ := json.Marshal(body)
			if string(after) != string(original) {
				t.Error("upload mutated caller document data")
			}
		})
	}
}
