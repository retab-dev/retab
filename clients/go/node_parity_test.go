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
	if client.EditTemplates == nil {
		t.Fatalf("missing edit templates namespace")
	}
	if client.WorkflowArtifacts == nil || client.WorkflowBlockExecutions == nil || client.WorkflowBlocks == nil || client.WorkflowEdges == nil || client.WorkflowTests == nil || client.WorkflowTestRuns == nil {
		t.Fatalf("missing workflow resource namespaces")
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
	model := "retab-small"
	allowOverlap := true
	_, err := client.Partitions.Create(context.Background(), &PartitionsCreateParams{
		Document:     MIMEData{Filename: "invoice.pdf", URL: "data:application/pdf;base64,AAAA"},
		Key:          "vendor",
		Instructions: "partition by vendor",
		Model:        &model,
		AllowOverlap: &allowOverlap,
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

func TestWorkflowRunCreateSendsMIMEDataDocumentShape(t *testing.T) {
	var body Resource
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs" {
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

	_, err := client.WorkflowRuns.Create(context.Background(), WithRequestBody(map[string]any{
		"workflow_id": "wf_123",
		"documents": map[string]any{
			"start-1": MIMEData{Filename: "invoice.pdf", Content: "AAAA", MIMEType: "application/pdf"},
		},
		"json_inputs": map[string]any{"vendor": "Retab"},
		"version":     "production",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if body["workflow_id"] != "wf_123" {
		t.Fatalf("workflow_id = %#v", body["workflow_id"])
	}
	documents, ok := body["documents"].(map[string]any)
	if !ok {
		t.Fatalf("documents = %#v", body["documents"])
	}
	startDocument, ok := documents["start-1"].(map[string]any)
	if !ok {
		t.Fatalf("start document = %#v", documents["start-1"])
	}
	if startDocument["filename"] != "invoice.pdf" {
		t.Fatalf("start document = %#v", startDocument)
	}
	if body["version"] != "production" {
		t.Fatalf("version = %#v", body["version"])
	}
}

func TestWorkflowRunCreatePreservesURLBackedDocuments(t *testing.T) {
	var body Resource
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs" {
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

	_, err := client.WorkflowRuns.Create(context.Background(), WithRequestBody(map[string]any{
		"workflow_id": "wf_123",
		"documents": map[string]any{
			"start-1": MIMEData{Filename: "invoice.pdf", URL: "https://storage.retab.com/org_1/file_123.pdf"},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if body["workflow_id"] != "wf_123" {
		t.Fatalf("workflow_id = %#v", body["workflow_id"])
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
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs" {
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

	_, err := client.WorkflowRuns.Create(context.Background(), WithRequestBody(map[string]any{
		"workflow_id": "wf_123",
		"documents": map[string]any{
			"start-1": map[string]any{
				"filename": "invoice.pdf",
				"url":      "https://storage.retab.com/org_1/file_123.pdf",
			},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if body["workflow_id"] != "wf_123" {
		t.Fatalf("workflow_id = %#v", body["workflow_id"])
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
	if _, err := client.WorkflowRuns.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Jobs.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}

	if requests["/v1/workflows"] != "" {
		t.Fatalf("workflow defaults = %q", requests["/v1/workflows"])
	}
	if requests["/v1/workflows/runs"] != "" {
		t.Fatalf("workflow run defaults = %q", requests["/v1/workflows/runs"])
	}
	if requests["/v1/jobs"] != "" {
		t.Fatalf("job defaults = %q", requests["/v1/jobs"])
	}
}

func TestWorkflowsListPassesBeforeAndAfterThrough(t *testing.T) {
	var rawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Resource{"data": []Resource{}})
	}))
	defer server.Close()
	client := newTestClient(t, server)

	_, err := client.Workflows.List(context.Background(), &WorkflowsListParams{
		PaginationParams: PaginationParams{Before: ptrString("wf_a"), After: ptrString("wf_b")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rawQuery != "after=wf_b&before=wf_a" {
		t.Fatalf("rawQuery = %q", rawQuery)
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
		case "/v1/workflows/blocks":
			_ = json.NewEncoder(w).Encode(Resource{
				"data": []Resource{{
					"id": "block_1", "workflow_id": "wf_123", "organization_id": "org", "type": "start_document",
				}},
				"list_metadata": Resource{"before": nil, "after": nil},
			})
		case "/v1/workflows/edges":
			_ = json.NewEncoder(w).Encode(Resource{
				"data":          []Resource{},
				"list_metadata": Resource{"before": nil, "after": nil},
			})
		case "/v1/workflows/tests":
			_ = json.NewEncoder(w).Encode(Resource{"data": []Resource{{"id": "test_1"}}})
		case "/v1/workflows/tests/runs":
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

	if _, err := client.WorkflowBlocks.List(context.Background(), &WorkflowBlocksListParams{WorkflowID: "wf_123"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WorkflowEdges.List(context.Background(), &WorkflowEdgesListParams{WorkflowID: "wf_123"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.WorkflowTests.List(context.Background(), &WorkflowTestsListParams{WorkflowID: "wf_123"}); err != nil {
		t.Fatal(err)
	}
	limit := 10
	if _, err := client.WorkflowTestRuns.List(context.Background(), &WorkflowTestRunsListParams{
		WorkflowID:       ptrString("wf_123"),
		TestID:           ptrString("test_1"),
		PaginationParams: PaginationParams{Limit: &limit},
	}); err != nil {
		t.Fatal(err)
	}

	want := "GET?workflow_id=wf_123 /v1/workflows/blocks,GET?workflow_id=wf_123 /v1/workflows/edges,GET?workflow_id=wf_123 /v1/workflows/tests,GET?limit=10&test_id=test_1&workflow_id=wf_123 /v1/workflows/tests/runs"
	if strings.Join(requests, ",") != want {
		t.Fatalf("requests = %s", strings.Join(requests, ","))
	}
}

func ptrString(value string) *string {
	return &value
}
