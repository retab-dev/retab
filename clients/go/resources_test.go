package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type capturedRequest struct {
	Method string
	Path   string
	Query  map[string][]string
	Body   Resource
}

func newCaptureServer(t *testing.T, response any) (*httptest.Server, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Method = r.Method
		captured.Path = r.URL.Path
		captured.Query = r.URL.Query()
		if r.Body != nil {
			defer r.Body.Close()
			_ = json.NewDecoder(r.Body).Decode(&captured.Body)
		}
		w.Header().Set("Content-Type", "application/json")
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
	return server, captured
}

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestResourceCreateRequestShapes(t *testing.T) {
	document := MIMEData{Filename: "invoice.pdf", Content: "data", MIMEType: "application/pdf"}
	tests := []struct {
		name       string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		assertBody func(t *testing.T, body Resource)
	}{
		{
			name: "extractions create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Extractions.Create(ctx, ExtractionCreateRequest{
					Document:           document,
					JSONSchema:         Resource{"type": "object"},
					Model:              "retab-small",
					ImageResolutionDPI: 192,
					NConsensus:         2,
					Instructions:       "read totals",
					Metadata:           map[string]string{"source": "test"},
					AdditionalMessages: []Resource{{"role": "user", "content": "extra"}},
					BustCache:          true,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/extractions",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "model", "retab-small")
				assertBodyNumber(t, body, "image_resolution_dpi", 192)
				assertBodyNumber(t, body, "n_consensus", 2)
				assertBodyBool(t, body, "bust_cache", true)
				assertNestedString(t, body, "document", "filename", "invoice.pdf")
				assertNestedString(t, body, "json_schema", "type", "object")
			},
		},
		{
			name: "splits create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Splits.Create(ctx, SplitCreateRequest{
					Document: document,
					Subdocuments: []SplitSubdocument{{
						Name:                   "invoice",
						Description:            "invoice pages",
						AllowMultipleInstances: true,
					}},
					Model:        "retab-small",
					NConsensus:   3,
					BustCache:    true,
					Instructions: "split carefully",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/splits",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "model", "retab-small")
				assertBodyNumber(t, body, "n_consensus", 3)
				subdocuments, ok := body["subdocuments"].([]any)
				if !ok || len(subdocuments) != 1 {
					t.Fatalf("subdocuments = %#v", body["subdocuments"])
				}
			},
		},
		{
			name: "classifications create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Classifications.Create(ctx, ClassificationCreateRequest{
					Document: document,
					Categories: []ClassificationCategory{{
						Name:        "invoice",
						Description: "invoice document",
					}},
					Model:        "retab-small",
					NConsensus:   2,
					FirstNPages:  4,
					Instructions: "choose one",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/classifications",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "model", "retab-small")
				assertBodyNumber(t, body, "first_n_pages", 4)
				assertBodyNumber(t, body, "n_consensus", 2)
			},
		},
		{
			name: "parses create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Parses.Create(ctx, ParseCreateRequest{
					Document:           document,
					Model:              "retab-small",
					TableParsingFormat: "markdown",
					ImageResolutionDPI: 192,
					Instructions:       "tables",
					BustCache:          true,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/parses",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "table_parsing_format", "markdown")
				assertBodyBool(t, body, "bust_cache", true)
			},
		},
		{
			name: "edits create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Edits.Create(ctx, EditCreateRequest{
					Instructions: "fill the form",
					Document:     document,
					Model:        "retab-small",
					Color:        "#000080",
					BustCache:    true,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/edits",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "instructions", "fill the form")
				assertNestedString(t, body, "config", "color", "#000080")
			},
		},
		{
			name: "schemas generate",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Schemas.Generate(ctx, GenerateSchemaRequest{
					Documents: []any{document},
					Model:     "retab-small",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/schemas/generate",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "model", "retab-small")
				documents, ok := body["documents"].([]any)
				if !ok || len(documents) != 1 {
					t.Fatalf("documents = %#v", body["documents"])
				}
			},
		},
		{
			name: "jobs create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Jobs.Create(ctx, JobCreateRequest{
					Endpoint: "/v1/extractions",
					Request:  Resource{"model": "retab-small"},
					Metadata: map[string]string{"owner": "go-test"},
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/jobs",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "endpoint", "/v1/extractions")
				assertNestedString(t, body, "request", "model", "retab-small")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, captured := newCaptureServer(t, Resource{"id": "res_123", "status": "completed"})
			defer server.Close()
			client := newTestClient(t, server)

			if err := test.call(context.Background(), client); err != nil {
				t.Fatal(err)
			}
			if captured.Method != test.wantMethod {
				t.Fatalf("method = %s", captured.Method)
			}
			if captured.Path != test.wantPath {
				t.Fatalf("path = %s", captured.Path)
			}
			test.assertBody(t, captured.Body)
		})
	}
}

func TestResourceGetDeleteAndFilePaths(t *testing.T) {
	tests := []struct {
		name       string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
	}{
		{"files get", func(ctx context.Context, client *Client) error {
			_, err := client.Files.Get(ctx, "file_123")
			return err
		}, http.MethodGet, "/files/file_123"},
		{"files download link", func(ctx context.Context, client *Client) error {
			_, err := client.Files.GetDownloadLink(ctx, "file_123")
			return err
		}, http.MethodGet, "/files/file_123/download-link"},
		{"files create upload", func(ctx context.Context, client *Client) error {
			_, err := client.Files.CreateUpload(ctx, PrepareUploadRequest{
				Filename:    "invoice.pdf",
				ContentType: "application/pdf",
				SizeBytes:   10,
				SHA256:      "abc",
			})
			return err
		}, http.MethodPost, "/files/upload"},
		{"files complete upload", func(ctx context.Context, client *Client) error {
			_, err := client.Files.CompleteUpload(ctx, "file_123", "abc")
			return err
		}, http.MethodPost, "/files/upload/file_123/complete"},
		{"extractions get", func(ctx context.Context, client *Client) error {
			_, err := client.Extractions.Get(ctx, "ext_123")
			return err
		}, http.MethodGet, "/extractions/ext_123"},
		{"extractions sources", func(ctx context.Context, client *Client) error {
			_, err := client.Extractions.Sources(ctx, "ext_123")
			return err
		}, http.MethodGet, "/extractions/ext_123/sources"},
		{"splits delete", func(ctx context.Context, client *Client) error {
			return client.Splits.Delete(ctx, "split_123")
		}, http.MethodDelete, "/splits/split_123"},
		{"classifications get", func(ctx context.Context, client *Client) error {
			_, err := client.Classifications.Get(ctx, "cls_123")
			return err
		}, http.MethodGet, "/classifications/cls_123"},
		{"parses delete", func(ctx context.Context, client *Client) error {
			return client.Parses.Delete(ctx, "parse_123")
		}, http.MethodDelete, "/parses/parse_123"},
		{"edits get", func(ctx context.Context, client *Client) error {
			_, err := client.Edits.Get(ctx, "edit_123")
			return err
		}, http.MethodGet, "/edits/edit_123"},
		{"jobs cancel", func(ctx context.Context, client *Client) error {
			_, err := client.Jobs.Cancel(ctx, "job_123")
			return err
		}, http.MethodPost, "/jobs/job_123/cancel"},
		{"jobs retry", func(ctx context.Context, client *Client) error {
			_, err := client.Jobs.Retry(ctx, "job_123")
			return err
		}, http.MethodPost, "/jobs/job_123/retry"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, captured := newCaptureServer(t, Resource{"id": "res_123", "object": "job", "status": "completed"})
			defer server.Close()
			client := newTestClient(t, server)

			if err := test.call(context.Background(), client); err != nil {
				t.Fatal(err)
			}
			if captured.Method != test.wantMethod {
				t.Fatalf("method = %s", captured.Method)
			}
			if captured.Path != test.wantPath {
				t.Fatalf("path = %s", captured.Path)
			}
		})
	}
}

func TestFilesPrepareUploadRequestsMatchNode(t *testing.T) {
	client := &Client{}
	uploadRequest := client.Files
	if uploadRequest != nil {
		t.Fatalf("zero-value client should not have files service installed")
	}

	retabClient, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	prepared := retabClient.Files.PrepareUpload("invoice.pdf", "application/pdf", 10, "abc")
	if prepared.URL != "/files/upload" || prepared.Method != http.MethodPost {
		t.Fatalf("prepared upload = %#v", prepared)
	}
	body, ok := prepared.Body.(map[string]any)
	if !ok || body["sha256"] != "abc" || body["size_bytes"] != int64(10) {
		t.Fatalf("prepared upload body = %#v", prepared.Body)
	}

	complete := retabClient.Files.PrepareCompleteUpload("file_123", "abc")
	if complete.URL != "/files/upload/file_123/complete" || complete.Method != http.MethodPost {
		t.Fatalf("prepared complete upload = %#v", complete)
	}
}

func TestFilesUploadRequestShape(t *testing.T) {
	uploadMethod := ""
	uploadPath := ""
	uploadHeader := ""
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadMethod = r.Method
		uploadPath = r.URL.Path
		uploadHeader = r.Header.Get("x-upload-token")
		w.WriteHeader(http.StatusOK)
	}))
	defer uploadServer.Close()

	var paths []string
	var bodies []Resource
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		body := Resource{}
		if r.Body != nil {
			defer r.Body.Close()
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/files/upload":
			_ = json.NewEncoder(w).Encode(Resource{
				"fileId":        "file_123",
				"uploadUrl":     uploadServer.URL + "/direct-upload",
				"uploadMethod":  "PUT",
				"uploadHeaders": map[string]string{"x-upload-token": "secret"},
				"expiresAt":     "2026-01-01T00:00:00Z",
			})
		case "/files/upload/file_123/complete":
			_ = json.NewEncoder(w).Encode(MIMEData{
				Filename: "invoice.pdf",
				URL:      "https://files.example/invoice.pdf",
				MIMEType: "application/pdf",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer apiServer.Close()
	client := newTestClient(t, apiServer)

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invoice.pdf")
	if err := os.WriteFile(filePath, []byte("invoice"), 0o600); err != nil {
		t.Fatal(err)
	}
	mimeData, err := client.Files.Upload(context.Background(), filePath)
	if err != nil {
		t.Fatal(err)
	}
	if mimeData.URL != "https://files.example/invoice.pdf" {
		t.Fatalf("url = %s", mimeData.URL)
	}
	if strings.Join(paths, ",") != "/files/upload,/files/upload/file_123/complete" {
		t.Fatalf("paths = %#v", paths)
	}
	assertBodyString(t, bodies[0], "filename", "invoice.pdf")
	assertBodyString(t, bodies[0], "content_type", "application/pdf")
	assertBodyNumber(t, bodies[0], "size_bytes", 7)
	if _, ok := bodies[0]["sha256"].(string); !ok {
		t.Fatalf("sha256 = %#v", bodies[0]["sha256"])
	}
	if uploadMethod != http.MethodPut {
		t.Fatalf("upload method = %s", uploadMethod)
	}
	if uploadPath != "/direct-upload" {
		t.Fatalf("upload path = %s", uploadPath)
	}
	if uploadHeader != "secret" {
		t.Fatalf("upload header = %s", uploadHeader)
	}
	if len(bodies[1]) != 1 {
		t.Fatalf("complete body = %#v", bodies[1])
	}
}

func TestListQueryShapes(t *testing.T) {
	fromDate := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	toDate := time.Date(2026, 1, 3, 4, 5, 6, 0, time.UTC)
	includeRequest := true
	tests := []struct {
		name     string
		call     func(context.Context, *Client) error
		wantPath string
		assert   func(t *testing.T, query map[string][]string)
	}{
		{
			name: "extractions list metadata",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Extractions.List(ctx, &ListExtractionsParams{
					ListParams: ListParams{
						Before:   "before",
						After:    "after",
						Limit:    7,
						Order:    "asc",
						Filename: "invoice.pdf",
						FromDate: &fromDate,
						ToDate:   &toDate,
					},
					OriginType: "api",
					OriginID:   "origin_123",
					Metadata:   map[string]string{"tenant": "acme"},
				})
				return err
			},
			wantPath: "/extractions",
			assert: func(t *testing.T, query map[string][]string) {
				assertQuery(t, query, "limit", "7")
				assertQuery(t, query, "order", "asc")
				assertQuery(t, query, "filename", "invoice.pdf")
				assertQuery(t, query, "from_date", "2026-01-02T03:04:05Z")
				assertQuery(t, query, "metadata", `{"tenant":"acme"}`)
			},
		},
		{
			name: "jobs retrieve",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Jobs.Retrieve(ctx, "job_123", &JobRetrieveParams{
					IncludeRequest:  true,
					IncludeResponse: true,
				})
				return err
			},
			wantPath: "/jobs/job_123",
			assert: func(t *testing.T, query map[string][]string) {
				assertQuery(t, query, "include_request", "true")
				assertQuery(t, query, "include_response", "true")
			},
		},
		{
			name: "jobs list filters",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Jobs.List(ctx, &ListJobsParams{
					Limit:          3,
					Status:         "completed",
					Endpoint:       "/v1/parses",
					DocumentType:   []string{"invoice", "receipt"},
					Metadata:       map[string]string{"tenant": "acme"},
					IncludeRequest: &includeRequest,
				})
				return err
			},
			wantPath: "/jobs",
			assert: func(t *testing.T, query map[string][]string) {
				assertQuery(t, query, "limit", "3")
				assertQuery(t, query, "status", "completed")
				assertQuery(t, query, "endpoint", "/v1/parses")
				assertQuery(t, query, "metadata", `{"tenant":"acme"}`)
				assertQuery(t, query, "include_request", "true")
				if got := strings.Join(query["document_type"], ","); got != "invoice,receipt" {
					t.Fatalf("document_type = %q", got)
				}
			},
		},
		{
			// Regression: a typed-nil map made it past the `value == nil`
			// guard in addJSONQuery and serialised as `metadata=null`,
			// which the API then rejected with HTTP 400. The CLI hit this
			// the moment a user ran `retab jobs list` with no flags.
			// Pin the fix: an unset Metadata must NOT appear in the URL.
			name: "jobs list omits metadata when unset",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Jobs.List(ctx, &ListJobsParams{Limit: 1})
				return err
			},
			wantPath: "/jobs",
			assert: func(t *testing.T, query map[string][]string) {
				assertQuery(t, query, "limit", "1")
				if _, ok := query["metadata"]; ok {
					t.Fatalf("metadata should not appear when unset, got %q", query["metadata"])
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, captured := newCaptureServer(t, Resource{
				"data":     []Resource{{"id": "res_123", "object": "job", "status": "completed"}},
				"has_more": false,
			})
			defer server.Close()
			client := newTestClient(t, server)

			if err := test.call(context.Background(), client); err != nil {
				t.Fatal(err)
			}
			if captured.Path != test.wantPath {
				t.Fatalf("path = %s", captured.Path)
			}
			test.assert(t, captured.Query)
		})
	}
}

func assertBodyString(t *testing.T, body Resource, key string, want string) {
	t.Helper()
	if got, _ := body[key].(string); got != want {
		t.Fatalf("%s = %#v", key, body[key])
	}
}

func assertBodyNumber(t *testing.T, body Resource, key string, want float64) {
	t.Helper()
	if got, _ := body[key].(float64); got != want {
		t.Fatalf("%s = %#v", key, body[key])
	}
}

func assertBodyBool(t *testing.T, body Resource, key string, want bool) {
	t.Helper()
	if got, _ := body[key].(bool); got != want {
		t.Fatalf("%s = %#v", key, body[key])
	}
}

func assertNestedString(t *testing.T, body Resource, parent string, key string, want string) {
	t.Helper()
	nested, ok := body[parent].(map[string]any)
	if !ok {
		t.Fatalf("%s = %#v", parent, body[parent])
	}
	if got, _ := nested[key].(string); got != want {
		t.Fatalf("%s.%s = %#v", parent, key, nested[key])
	}
}

func assertQuery(t *testing.T, query map[string][]string, key string, want string) {
	t.Helper()
	values := query[key]
	if len(values) != 1 || values[0] != want {
		t.Fatalf("%s = %#v", key, values)
	}
}
