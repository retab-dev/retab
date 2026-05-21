package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNodeParityRootNamespacesAreInstalled(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	if client.Partitions == nil {
		t.Fatalf("missing root namespaces: partitions=%v", client.Partitions)
	}
	if client.Edits.Templates == nil {
		t.Fatalf("missing edits.templates namespace")
	}
	if client.Workflows.Artifacts == nil || client.Workflows.Blocks == nil || client.Workflows.Edges == nil || client.Workflows.Tests == nil || client.Workflows.Tests.Runs == nil {
		t.Fatalf("missing workflow nested namespaces")
	}
}

func TestRequestOptionsMergeParamsHeadersAndBody(t *testing.T) {
	var header string
	var rawQuery string
	var body Resource
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header = r.Header.Get("X-Test")
		rawQuery = r.URL.RawQuery
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Resource{"id": "partition_123"})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	params := url.Values{}
	params.Set("debug", "1")
	_, err := client.Partitions.Create(context.Background(), PartitionCreateRequest{
		Document:     MIMEData{Filename: "invoice.pdf", URL: "data:application/pdf;base64,AAAA"},
		Key:          "vendor",
		Instructions: "partition by vendor",
		Model:        "retab-small",
		AllowOverlap: true,
	}, WithRequestParams(params), WithRequestHeader("X-Test", "yes"), WithRequestBody(map[string]any{"model": "retab-large"}))
	if err != nil {
		t.Fatal(err)
	}
	if header != "yes" {
		t.Fatalf("header = %q", header)
	}
	if rawQuery != "debug=1" {
		t.Fatalf("rawQuery = %q", rawQuery)
	}
	if body["model"] != "retab-large" {
		t.Fatalf("body = %#v", body)
	}
	if body["allow_overlap"] != true {
		t.Fatalf("body = %#v", body)
	}
}

func TestNodePaginationEnvelope(t *testing.T) {
	var list PaginatedList[Resource]
	if err := json.Unmarshal([]byte(`{"data":[{"id":"x"}],"list_metadata":{"before":"b","after":"a"}}`), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Data) != 1 || list.ListMetadata.Before != "b" || list.ListMetadata.After != "a" {
		t.Fatalf("list = %#v", list)
	}
}

func TestMIMEDataMatchesNodeJSONAndStorageID(t *testing.T) {
	mimeData := MIMEData{
		Filename: "invoice.pdf",
		URL:      "https://storage.retab.com/org_123/file_123.pdf",
		Content:  "not-serialized",
		MIMEType: "application/pdf",
	}
	encoded, err := json.Marshal(mimeData)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"filename":"invoice.pdf","url":"https://storage.retab.com/org_123/file_123.pdf"}` {
		t.Fatalf("encoded = %s", encoded)
	}
	if mimeData.ID() != "file_123" {
		t.Fatalf("id = %q", mimeData.ID())
	}
	if (MIMEData{URL: "https://storage.retab.com/org_123/file_123.pdf?x=1"}).ID() != "" {
		t.Fatalf("storage URLs with query params must not infer an id")
	}
}

// InferMIMEData accepts whatever http.DetectContentType can classify
// (text, images, PDFs, Office zips, etc.) — broader than the previous
// 4-format whitelist, which contradicted the CLI's marketing copy
// promising Excel / email / text support. Only genuinely-opaque bytes
// (those sniffed as `application/octet-stream`) still fail.
func TestInferMIMEDataAcceptsTextRejectsOpaqueBytes(t *testing.T) {
	// Plain text was rejected pre-fix; must succeed now.
	got, err := InferMIMEData([]byte("plain text is supposed to be detected"))
	if err != nil {
		t.Fatalf("plain text should now be accepted, got: %v", err)
	}
	if got.URL == "" || got.Filename == "" {
		t.Errorf("expected populated MIMEData, got %+v", got)
	}

	// Genuinely structureless bytes — short enough that no sniffer can
	// classify them — must still fail rather than guessing wrong.
	if _, err := InferMIMEData([]byte{0x00, 0x01, 0x02}); err == nil {
		t.Fatalf("expected opaque bytes to fail MIME inference")
	}
}

func TestFilesUploadRejectsNonLocalPaths(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Files.Upload(context.Background(), "https://example.com/invoice.pdf"); err == nil {
		t.Fatalf("expected URL upload rejection")
	}
	if _, err := client.Files.Upload(context.Background(), "data:application/pdf;base64,AAAA"); err == nil {
		t.Fatalf("expected data URI upload rejection")
	}
}

func TestFileUploadContentTypeMatchesNode(t *testing.T) {
	cases := map[string]string{
		"file.pdf":  "application/pdf",
		"file.png":  "image/png",
		"file.jpg":  "image/jpeg",
		"file.jpeg": "image/jpeg",
		"file.txt":  "text/plain",
		"file.csv":  "application/octet-stream",
	}
	for filename, want := range cases {
		if got := contentTypeForFilename(filename); got != want {
			t.Fatalf("%s content type = %q, want %q", filename, got, want)
		}
	}
}

func TestWorkflowRunCreateMaterializesDocumentsLikeNode(t *testing.T) {
	var body Resource
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/wf_123/run" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(workflowRunResponse("run_123", "wf_123", "running"))
	}))
	defer server.Close()
	client := newTestClient(t, server)

	_, err := client.Workflows.Runs.Create(context.Background(), CreateWorkflowRunRequest{
		WorkflowID: "wf_123",
		Documents: map[string]any{
			"start-1": MIMEData{Filename: "invoice.pdf", URL: "data:application/pdf;base64,AAAA"},
		},
		JSONInputs: map[string]any{"vendor": "Retab"},
	})
	if err != nil {
		t.Fatal(err)
	}
	documents, ok := body["documents"].(map[string]any)
	if !ok {
		t.Fatalf("documents = %#v", body["documents"])
	}
	startDocument, ok := documents["start-1"].(map[string]any)
	if !ok {
		t.Fatalf("start document = %#v", documents["start-1"])
	}
	if startDocument["filename"] != "invoice.pdf" || startDocument["content"] != "AAAA" || startDocument["mime_type"] != "application/pdf" {
		t.Fatalf("start document = %#v", startDocument)
	}
	if body["version"] != "production" {
		t.Fatalf("version = %#v", body["version"])
	}
}

func TestWorkflowRunCreatePreservesURLBackedDocuments(t *testing.T) {
	var body Resource
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/wf_123/run" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(workflowRunResponse("run_123", "wf_123", "running"))
	}))
	defer server.Close()
	client := newTestClient(t, server)

	_, err := client.Workflows.Runs.Create(context.Background(), CreateWorkflowRunRequest{
		WorkflowID: "wf_123",
		Documents: map[string]any{
			"start-1": MIMEData{Filename: "invoice.pdf", URL: "https://storage.retab.com/org_1/file_123.pdf"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	documents, ok := body["documents"].(map[string]any)
	if !ok {
		t.Fatalf("documents = %#v", body["documents"])
	}
	startDocument, ok := documents["start-1"].(map[string]any)
	if !ok {
		t.Fatalf("start document = %#v", documents["start-1"])
	}
	if startDocument["filename"] != "invoice.pdf" || startDocument["url"] != "https://storage.retab.com/org_1/file_123.pdf" {
		t.Fatalf("start document = %#v", startDocument)
	}
	if _, ok := startDocument["content"]; ok {
		t.Fatalf("url-backed workflow document should not be materialized inline: %#v", startDocument)
	}
}

func TestWorkflowRunCreateAcceptsJSONDocumentDescriptors(t *testing.T) {
	var body Resource
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/wf_123/run" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(workflowRunResponse("run_123", "wf_123", "running"))
	}))
	defer server.Close()
	client := newTestClient(t, server)

	_, err := client.Workflows.Runs.Create(context.Background(), CreateWorkflowRunRequest{
		WorkflowID: "wf_123",
		Documents: map[string]any{
			"start-1": map[string]any{
				"filename": "invoice.pdf",
				"url":      "https://storage.retab.com/org_1/file_123.pdf",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	documents, ok := body["documents"].(map[string]any)
	if !ok {
		t.Fatalf("documents = %#v", body["documents"])
	}
	startDocument, ok := documents["start-1"].(map[string]any)
	if !ok {
		t.Fatalf("start document = %#v", documents["start-1"])
	}
	if startDocument["filename"] != "invoice.pdf" || startDocument["url"] != "https://storage.retab.com/org_1/file_123.pdf" {
		t.Fatalf("start document = %#v", startDocument)
	}
}

func TestListDefaultsMatchNode(t *testing.T) {
	requests := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests[r.URL.Path] = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Resource{"data": []Resource{}})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	if _, err := client.Workflows.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Runs.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Jobs.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	if requests["/workflows"] != "limit=10&order=desc" {
		t.Fatalf("workflow defaults = %q", requests["/workflows"])
	}
	if requests["/workflows/runs"] != "limit=20&order=desc" {
		t.Fatalf("workflow run defaults = %q", requests["/workflows/runs"])
	}
	if requests["/jobs"] != "limit=20&order=desc" {
		t.Fatalf("job defaults = %q", requests["/jobs"])
	}
}

// TestWorkflowsListPropagatesFieldsParam pins the SDK serialisation of
// ListWorkflowsParams.Fields into the `?fields=` query parameter. A
// dogfood session against the production API surfaced that the response
// still includes every field — this test confirms the SDK side is fine,
// so the failure is server-side (the API ignores the query param).
func TestWorkflowsListPropagatesFieldsParam(t *testing.T) {
	var rawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Resource{"data": []Resource{}})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	params := &ListWorkflowsParams{Fields: "id,name,updated_at"}
	if _, err := client.Workflows.List(context.Background(), params); err != nil {
		t.Fatal(err)
	}
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		t.Fatalf("parse query %q: %v", rawQuery, err)
	}
	if got := values.Get("fields"); got != "id,name,updated_at" {
		t.Fatalf("fields query param = %q, want %q (raw=%q)", got, "id,name,updated_at", rawQuery)
	}
}

func TestFetchJSONRequiresJSONContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(`{"id":"file_123"}`))
	}))
	defer server.Close()
	client := newTestClient(t, server)

	_, err := client.Files.Get(context.Background(), "file_123")
	if err == nil {
		t.Fatalf("expected content-type error")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Message != "Response is not JSON" {
		t.Fatalf("err = %#v", err)
	}
}

func TestWorkflowNodeParitySubclientsUseNodePaths(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := r.Method + " " + r.URL.Path
		if r.URL.RawQuery != "" {
			request = r.Method + "?" + r.URL.RawQuery + " " + r.URL.Path
		}
		requests = append(requests, request)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/workflows/blocks":
			_ = json.NewEncoder(w).Encode(Resource{
				"data": []Resource{{
					"id": "block_1", "workflow_id": "wf_123", "organization_id": "org", "type": "start-document",
				}},
				"list_metadata": Resource{"before": nil, "after": nil},
			})
		case "/workflows/edges":
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			_ = json.NewEncoder(w).Encode(Resource{
				"data":          []Resource{},
				"list_metadata": Resource{"before": nil, "after": nil},
			})
		case "/workflows/tests":
			_ = json.NewEncoder(w).Encode(Resource{"data": []Resource{{"id": "test_1"}}})
		case "/workflows/tests/runs":
			_ = json.NewEncoder(w).Encode(Resource{
				"data": []Resource{{
					"id":        "testrun_1",
					"workflow":  Resource{"workflow_id": "wf_123", "version_id": "draft_1", "name_at_run_time": "Workflow"},
					"trigger":   Resource{"type": "api"},
					"lifecycle": Resource{"status": "completed"},
					"timing":    Resource{"created_at": "2026-05-18T10:00:00Z"},
				}},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	client := newTestClient(t, server)

	if _, err := client.Workflows.Blocks.List(context.Background(), "wf_123"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Edges.List(context.Background(), "wf_123", nil); err != nil {
		t.Fatal(err)
	}
	if err := client.Workflows.Edges.DeleteAll(context.Background(), "wf_123"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Tests.List(context.Background(), ListWorkflowTestsRequest{WorkflowID: "wf_123"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Tests.Runs.List(context.Background(), ListWorkflowTestRunsParams{
		WorkflowID: "wf_123",
		TestID:     "test_1",
		Limit:      10,
	}); err != nil {
		t.Fatal(err)
	}

	want := "GET?workflow_id=wf_123 /workflows/blocks,GET?workflow_id=wf_123 /workflows/edges,DELETE?workflow_id=wf_123 /workflows/edges,GET?limit=50&workflow_id=wf_123 /workflows/tests,GET?limit=10&test_id=test_1&workflow_id=wf_123 /workflows/tests/runs"
	if strings.Join(requests, ",") != want {
		t.Fatalf("requests = %s", strings.Join(requests, ","))
	}
}
