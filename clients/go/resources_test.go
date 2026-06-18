package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestMIMEDataMarshalIncludesContent pins the content-only document contract:
// InferMIMEData accepts a {content, mime_type} descriptor with no url, so
// MarshalJSON must serialize those fields. A regression here silently sends an
// empty document ({"filename":"","url":""}) over the wire. The url-only case
// must stay clean (no empty content/mime_type keys) thanks to omitempty.
func TestMIMEDataMarshalIncludesContent(t *testing.T) {
	contentOnly, err := json.Marshal(MIMEData{Filename: "invoice.pdf", Content: "BASE64DATA", MIMEType: "application/pdf"})
	if err != nil {
		t.Fatalf("marshal content-only: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(contentOnly, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["content"] != "BASE64DATA" {
		t.Errorf("content dropped from marshaled MIMEData: %s", contentOnly)
	}
	if got["mime_type"] != "application/pdf" {
		t.Errorf("mime_type dropped from marshaled MIMEData: %s", contentOnly)
	}

	urlOnly, err := json.Marshal(MIMEData{Filename: "doc.pdf", URL: "https://example.com/doc.pdf"})
	if err != nil {
		t.Fatalf("marshal url-only: %v", err)
	}
	if strings.Contains(string(urlOnly), "content") || strings.Contains(string(urlOnly), "mime_type") {
		t.Errorf("url-only MIMEData leaked empty content/mime_type keys: %s", urlOnly)
	}
}

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
				_, err := client.Extractions.Create(ctx, &ExtractionsCreateParams{
					Document:           document,
					JSONSchema:         map[string]interface{}{"type": "object"},
					Model:              ptrTo("retab-small"),
					NConsensus:         ptrTo(2),
					Instructions:       ptrTo("read totals"),
					Metadata:           ptrTo(map[string]string{"source": "test"}),
					AdditionalMessages: []map[string]interface{}{{"role": "user", "content": "extra"}},
					BustCache:          ptrTo(true),
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/extractions",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "model", "retab-small")
				assertBodyNumber(t, body, "n_consensus", 2)
				assertBodyBool(t, body, "bust_cache", true)
				assertNestedString(t, body, "document", "filename", "invoice.pdf")
				assertNestedString(t, body, "json_schema", "type", "object")
			},
		},
		{
			name: "splits create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Splits.Create(ctx, &SplitsCreateParams{
					Document: document,
					Subdocuments: []*Subdocument{{
						Name:                   "invoice",
						Description:            ptrTo("invoice pages"),
						AllowMultipleInstances: ptrTo(true),
					}},
					Model:        ptrTo("retab-small"),
					NConsensus:   ptrTo(3),
					BustCache:    ptrTo(true),
					Instructions: ptrTo("split carefully"),
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/splits",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "model", "retab-small")
				assertBodyNumber(t, body, "n_consensus", 3)
				subdocuments, ok := body["subdocuments"].([]any)
				if !ok || len(subdocuments) != 1 {
					t.Fatalf("subdocuments = %#v", body["subdocuments"])
				}
				subdocument, ok := subdocuments[0].(map[string]any)
				if !ok {
					t.Fatalf("subdocuments[0] = %#v", subdocuments[0])
				}
				if got, _ := subdocument["allow_multiple_instances"].(bool); got != true {
					t.Fatalf("subdocuments[0].allow_multiple_instances = %#v", subdocument["allow_multiple_instances"])
				}
			},
		},
		{
			name: "classifications create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Classifications.Create(ctx, &ClassificationsCreateParams{
					Document: document,
					Categories: []*Category{{
						Name:        "invoice",
						Description: ptrTo("invoice document"),
					}},
					Model:        ptrTo("retab-small"),
					NConsensus:   ptrTo(2),
					FirstNPages:  ptrTo(4),
					Instructions: ptrTo("choose one"),
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/classifications",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "model", "retab-small")
				assertBodyNumber(t, body, "first_n_pages", 4)
				assertBodyNumber(t, body, "n_consensus", 2)
			},
		},
		{
			name: "parses create",
			call: func(ctx context.Context, client *Client) error {
				format := ParseRequestTableParsingFormatMarkdown
				_, err := client.Parses.Create(ctx, &ParsesCreateParams{
					Document:           document,
					Model:              ptrTo("retab-small"),
					TableParsingFormat: &format,
					Instructions:       ptrTo("tables"),
					BustCache:          ptrTo(true),
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/parses",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "table_parsing_format", "markdown")
				assertBodyBool(t, body, "bust_cache", true)
			},
		},
		{
			name: "edits create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Edits.Create(ctx, &EditsCreateParams{
					Instructions: "fill the form",
					Document:     document,
					Model:        ptrTo("retab-small"),
					Config:       &EditConfig{Color: ptrTo("#000080")},
					BustCache:    ptrTo(true),
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/edits",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "instructions", "fill the form")
				assertNestedString(t, body, "config", "color", "#000080")
			},
		},
		{
			name: "schemas generate",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Schemas.Generate(ctx, &SchemasGenerateParams{
					Documents: []MIMEData{{
						Filename: document.Filename,
						URL:      document.Content,
					}},
					Model: ptrTo("retab-small"),
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/schemas/generate",
			assertBody: func(t *testing.T, body Resource) {
				assertBodyString(t, body, "model", "retab-small")
				documents, ok := body["documents"].([]any)
				if !ok || len(documents) != 1 {
					t.Fatalf("documents = %#v", body["documents"])
				}
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
		}, http.MethodGet, "/v1/files/file_123"},
		{"files download link", func(ctx context.Context, client *Client) error {
			_, err := client.Files.GetDownloadLink(ctx, "file_123")
			return err
		}, http.MethodGet, "/v1/files/file_123/download-link"},
		{"files create upload", func(ctx context.Context, client *Client) error {
			contentType := "application/pdf"
			sha256 := "abc"
			_, err := client.Files.CreateUpload(ctx, &FilesCreateUploadParams{
				Filename:    "invoice.pdf",
				ContentType: &contentType,
				SizeBytes:   10,
				Sha256:      &sha256,
			})
			return err
		}, http.MethodPost, "/v1/files/upload"},
		{"files complete upload", func(ctx context.Context, client *Client) error {
			sha256 := "abc"
			_, err := client.Files.CompleteUpload(ctx, "file_123", &FilesCompleteUploadParams{Sha256: &sha256})
			return err
		}, http.MethodPost, "/v1/files/upload/file_123/complete"},
		{"extractions get", func(ctx context.Context, client *Client) error {
			_, err := client.Extractions.Get(ctx, "ext_123", nil)
			return err
		}, http.MethodGet, "/v1/extractions/ext_123"},
		{"extractions sources", func(ctx context.Context, client *Client) error {
			_, err := client.Extractions.Sources(ctx, "ext_123")
			return err
		}, http.MethodGet, "/v1/extractions/ext_123/sources"},
		{"splits delete", func(ctx context.Context, client *Client) error {
			return client.Splits.Delete(ctx, "split_123")
		}, http.MethodDelete, "/v1/splits/split_123"},
		{"classifications get", func(ctx context.Context, client *Client) error {
			_, err := client.Classifications.Get(ctx, "cls_123", nil)
			return err
		}, http.MethodGet, "/v1/classifications/cls_123"},
		{"parses delete", func(ctx context.Context, client *Client) error {
			return client.Parses.Delete(ctx, "parse_123")
		}, http.MethodDelete, "/v1/parses/parse_123"},
		{"edits get", func(ctx context.Context, client *Client) error {
			_, err := client.Edits.Get(ctx, "edit_123", nil)
			return err
		}, http.MethodGet, "/v1/edits/edit_123"},
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

func TestFilesDownloadLinkDecodesDurableMIMEData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/files/file_123/download-link" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"download_url": "https://storage.googleapis.com/bucket/org_1/file/file_123.pdf?signed=1",
			"expires_in":   "60 minutes",
			"filename":     "invoice.pdf",
			"mime_data": map[string]string{
				"filename": "invoice.pdf",
				"url":      "https://storage.retab.com/org_1/file_123.pdf",
			},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	link, err := client.Files.GetDownloadLink(context.Background(), "file_123")
	if err != nil {
		t.Fatal(err)
	}
	if link.MIMEData == nil {
		t.Fatal("expected mime_data")
	}
	if link.MIMEData.URL != "https://storage.retab.com/org_1/file_123.pdf" {
		t.Fatalf("mime url = %q", link.MIMEData.URL)
	}
	if link.MIMEData.ID() != "file_123" {
		t.Fatalf("mime id = %q", link.MIMEData.ID())
	}
}

func TestFilesUploadRequestShape(t *testing.T) {
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
		case "/v1/files/upload":
			_ = json.NewEncoder(w).Encode(Resource{
				"fileId":        "file_123",
				"uploadUrl":     "https://uploads.example/direct-upload",
				"uploadMethod":  "PUT",
				"uploadHeaders": map[string]string{"x-upload-token": "secret"},
				"expiresAt":     "2026-01-01T00:00:00Z",
			})
		case "/v1/files/upload/file_123/complete":
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

	contentType := "application/pdf"
	sha256 := "abc"
	upload, err := client.Files.CreateUpload(context.Background(), &FilesCreateUploadParams{
		Filename:    "invoice.pdf",
		ContentType: &contentType,
		SizeBytes:   7,
		Sha256:      &sha256,
	})
	if err != nil {
		t.Fatal(err)
	}
	mimeData, err := client.Files.CompleteUpload(context.Background(), upload.FileID, &FilesCompleteUploadParams{Sha256: &sha256})
	if err != nil {
		t.Fatal(err)
	}
	if mimeData.URL != "https://files.example/invoice.pdf" {
		t.Fatalf("url = %s", mimeData.URL)
	}
	if strings.Join(paths, ",") != "/v1/files/upload,/v1/files/upload/file_123/complete" {
		t.Fatalf("paths = %#v", paths)
	}
	assertBodyString(t, bodies[0], "filename", "invoice.pdf")
	assertBodyString(t, bodies[0], "content_type", "application/pdf")
	assertBodyNumber(t, bodies[0], "size_bytes", 7)
	if _, ok := bodies[0]["sha256"].(string); !ok {
		t.Fatalf("sha256 = %#v", bodies[0]["sha256"])
	}
	if len(bodies[1]) != 1 {
		t.Fatalf("complete body = %#v", bodies[1])
	}
}

func TestListQueryShapes(t *testing.T) {
	fromDate := "2026-01-02"
	toDate := "2026-01-03"
	tests := []struct {
		name     string
		call     func(context.Context, *Client) error
		wantPath string
		assert   func(t *testing.T, query map[string][]string)
	}{
		{
			name: "extractions list metadata",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Extractions.List(ctx, &ExtractionsListParams{
					PaginationParams: PaginationParams{
						Before: ptrTo("before"),
						After:  ptrTo("after"),
						Limit:  ptrTo(7),
						Order:  ptrTo("asc"),
					},
					Filename: ptrTo("invoice.pdf"),
					FromDate: &fromDate,
					ToDate:   &toDate,
					Metadata: ptrTo(`{"tenant":"acme"}`),
				})
				return err
			},
			wantPath: "/v1/extractions",
			assert: func(t *testing.T, query map[string][]string) {
				assertQuery(t, query, "limit", "7")
				assertQuery(t, query, "order", "asc")
				assertQuery(t, query, "filename", "invoice.pdf")
				assertQuery(t, query, "from_date", "2026-01-02")
				assertQuery(t, query, "metadata", `{"tenant":"acme"}`)
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

func ptrTo[T any](value T) *T {
	return &value
}

func assertQuery(t *testing.T, query map[string][]string, key string, want string) {
	t.Helper()
	values := query[key]
	if len(values) != 1 || values[0] != want {
		t.Fatalf("%s = %#v", key, values)
	}
}
