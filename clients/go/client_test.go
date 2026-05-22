package retab

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestWorkflowsCreateUpdatePublishAndDelete(t *testing.T) {
	var requests []string
	var updateBody map[string]any
	var publishBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/workflows":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "wf_123",
				"name":        "Invoice Extractor",
				"description": "",
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/workflows/wf_123":
			if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "wf_123",
				"name":        "Renamed",
				"description": "Updated",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/publish":
			if err := json.NewDecoder(r.Body).Decode(&publishBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "wf_123",
				"name":        "Renamed",
				"description": "Updated",
				"published":   map[string]any{"version_id": "ver_0123456789abcdef0123456789abcdef"},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/workflows/wf_123":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	workflow, err := client.Workflows.Create(context.Background(), CreateWorkflowRequest{Name: "Invoice Extractor"})
	if err != nil {
		t.Fatal(err)
	}
	if workflow.ID != "wf_123" {
		t.Fatalf("workflow id = %s", workflow.ID)
	}

	name := "Renamed"
	description := "Updated"
	_, err = client.Workflows.Update(context.Background(), "wf_123", UpdateWorkflowRequest{
		Name:        &name,
		Description: &description,
		EmailTrigger: &WorkflowEmailTrigger{
			AllowedSenders: []string{"ops@example.com"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	emailTrigger, ok := updateBody["email_trigger"].(map[string]any)
	if !ok {
		t.Fatalf("email_trigger = %#v", updateBody["email_trigger"])
	}
	if len(emailTrigger["allowed_senders"].([]any)) != 1 {
		t.Fatalf("allowed_senders = %#v", emailTrigger["allowed_senders"])
	}

	published, err := client.Workflows.Publish(context.Background(), "wf_123", PublishWorkflowRequest{Description: "v1"})
	if err != nil {
		t.Fatal(err)
	}
	if published.Published == nil || published.Published.VersionID != "ver_0123456789abcdef0123456789abcdef" {
		t.Fatalf("published = %#v", published.Published)
	}
	if publishBody["description"] != "v1" {
		t.Fatalf("publish body = %#v", publishBody)
	}

	if err := client.Workflows.Delete(context.Background(), "wf_123"); err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"POST /workflows",
		"PATCH /workflows/wf_123",
		"POST /workflows/wf_123/publish",
		"DELETE /workflows/wf_123",
	}
	if strings.Join(requests, ",") != strings.Join(expected, ",") {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestHandlePayloadDoesNotExposeRemovedTextField(t *testing.T) {
	if _, ok := reflect.TypeOf(HandlePayload{}).FieldByName("Text"); ok {
		t.Fatal("HandlePayload must not expose removed text handle payloads")
	}
}

func TestWorkflowsPrepareDiagnoseMatchesPythonSurface(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}

	prepared := client.Workflows.PrepareDiagnose("wf_123")
	if prepared.URL != "/workflows/wf_123/diagnose-graph" || prepared.Method != http.MethodPost {
		t.Fatalf("prepared diagnose = %#v", prepared)
	}
	body, ok := prepared.Body.(DiagnoseWorkflowRequest)
	if !ok {
		t.Fatalf("prepared diagnose body = %#v", prepared.Body)
	}
	if !body.RePropagate {
		t.Fatalf("prepared diagnose body = %#v", body)
	}

	prepared = client.Workflows.PrepareDiagnose("wf_123", false)
	body = prepared.Body.(DiagnoseWorkflowRequest)
	if body.RePropagate {
		t.Fatalf("expected re-propagate override to be false: %#v", body)
	}
}

func TestWorkflowsPrepareDiagnoseGraphKeepsInMemoryGraphSurface(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	blocks := []map[string]any{{"id": "start-1", "type": "start_document"}}
	edges := []map[string]any{{"id": "edge-1", "source": "start-1", "target": "extract-1"}}

	prepared := client.Workflows.PrepareDiagnoseGraph("wf_123", blocks, edges, false)
	if prepared.URL != "/workflows/wf_123/diagnose-graph" || prepared.Method != http.MethodPost {
		t.Fatalf("prepared diagnose graph = %#v", prepared)
	}
	body, ok := prepared.Body.(DiagnoseWorkflowGraphRequest)
	if !ok {
		t.Fatalf("prepared diagnose graph body = %#v", prepared.Body)
	}
	if body.RePropagate || body.Blocks[0]["id"] != "start-1" || body.Edges[0]["id"] != "edge-1" {
		t.Fatalf("prepared diagnose graph body = %#v", body)
	}
}

func TestWorkflowsDiagnoseGraphDecodesWarningOnlyIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/wf_123/diagnose-graph" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"is_valid": true,
			"issues": []map[string]any{
				{
					"severity": "warning",
					"code":     "MISSING_REVIEW_PREDICATE",
					"message":  "Review gate needs a predicate",
					"block_id": "extract_1",
				},
			},
			"suggestions": []string{},
			"stats": map[string]any{
				"total_blocks":          1,
				"total_edges":           0,
				"block_types":           map[string]int{"start": 1},
				"start_document_blocks": 1,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	diagnosis, err := client.Workflows.DiagnoseGraph(context.Background(), "wf_123", DiagnoseWorkflowGraphRequest{
		Blocks:      []map[string]any{{"id": "start-1", "type": "start_document"}},
		Edges:       []map[string]any{},
		RePropagate: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !diagnosis.IsValid || len(diagnosis.Issues) != 1 {
		t.Fatalf("diagnosis = %#v", diagnosis)
	}
	if diagnosis.Issues[0].Severity != "warning" || diagnosis.Issues[0].Code != "MISSING_REVIEW_PREDICATE" {
		t.Fatalf("issue = %#v", diagnosis.Issues[0])
	}
}

func TestWorkflowsDiagnosePostsDirectly(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/wf_123/diagnose-graph" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body DiagnoseWorkflowRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !body.RePropagate {
			t.Fatalf("body = %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"is_valid":    true,
			"issues":      []map[string]any{},
			"suggestions": []string{},
			"stats": map[string]any{
				"total_blocks":          1,
				"total_edges":           0,
				"block_types":           map[string]int{"start_document": 1},
				"start_document_blocks": 1,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	diagnosis, err := client.Workflows.Diagnose(context.Background(), "wf_123")
	if err != nil {
		t.Fatal(err)
	}
	if !diagnosis.IsValid {
		t.Fatalf("diagnosis = %#v", diagnosis)
	}
	if strings.Join(requests, ",") != "POST /workflows/wf_123/diagnose-graph" {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowSpecsRoutesMatchPythonAndNode(t *testing.T) {
	var requests []string
	var validateBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/workflows/spec/validate" {
			if err := json.NewDecoder(r.Body).Decode(&validateBody); err != nil {
				t.Fatal(err)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Workflows.Specs.Validate(context.Background(), "name: test"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Specs.Plan(context.Background(), "name: test"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Specs.Apply(context.Background(), "name: test"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Specs.Export(context.Background(), "wf_123"); err != nil {
		t.Fatal(err)
	}

	if validateBody["yaml_definition"] != "name: test" {
		t.Fatalf("validate body = %#v", validateBody)
	}
	expected := []string{
		"POST /workflows/spec/validate",
		"POST /workflows/spec/plan",
		"POST /workflows/spec/apply",
		"GET /workflows/wf_123/spec",
	}
	if strings.Join(requests, ",") != strings.Join(expected, ",") {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowArtifactsListAndPrepare(t *testing.T) {
	var requests []string
	var listQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/artifacts":
			listQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"operation": "extraction",
					"id":        "ext_123",
				}},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/artifacts/ext_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"operation": "extraction",
				"id":        "ext_123",
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	preparedList := client.Workflows.Artifacts.PrepareList(ListWorkflowArtifactsParams{
		RunID:     "run_123",
		Operation: "extraction",
		BlockID:   "extract-1",
	})
	if preparedList.URL != "/workflows/artifacts" || preparedList.Method != http.MethodGet || preparedList.Params.Get("run_id") != "run_123" {
		t.Fatalf("prepared artifact list = %#v", preparedList)
	}
	preparedGet := client.Workflows.Artifacts.PrepareGet("ext_123")
	if preparedGet.URL != "/workflows/artifacts/ext_123" || preparedGet.Method != http.MethodGet {
		t.Fatalf("prepared artifact get = %#v", preparedGet)
	}

	artifacts, err := client.Workflows.Artifacts.List(context.Background(), ListWorkflowArtifactsParams{
		RunID:     "run_123",
		Operation: "extraction",
		BlockID:   "extract-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts.Data) != 1 || artifacts.Data[0]["id"] != "ext_123" {
		t.Fatalf("artifacts = %#v", artifacts)
	}
	if artifacts.ListMetadata.Before != "" || artifacts.ListMetadata.After != "" {
		t.Fatalf("artifacts list_metadata = %#v", artifacts.ListMetadata)
	}
	if !strings.Contains(listQuery, "run_id=run_123") || !strings.Contains(listQuery, "operation=extraction") || !strings.Contains(listQuery, "block_id=extract-1") {
		t.Fatalf("list query = %s", listQuery)
	}
	artifact, err := client.Workflows.Artifacts.Get(context.Background(), "ext_123")
	if err != nil {
		t.Fatal(err)
	}
	if (*artifact)["id"] != "ext_123" || (*artifact)["operation"] != "extraction" {
		t.Fatalf("artifact = %#v", artifact)
	}
	if strings.Join(requests, ",") != "GET /workflows/artifacts,GET /workflows/artifacts/ext_123" {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowExperimentRunsUseRunIDFirstRoutes(t *testing.T) {
	var rawQuery string
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/experiments/runs/exprun_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "exprun_123",
				"experiment_id": "exp_123",
				"lifecycle":     map[string]any{"status": "completed"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/experiments/results":
			rawQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          []map[string]any{},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/experiments/results/expresult_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "expresult_123",
				"run_id":        "exprun_123",
				"experiment_id": "exp_123",
				"document_id":   "expdoc_123",
				"lifecycle":     map[string]any{"status": "completed"},
				"timing":        map[string]any{},
				"block_kind":    "extract",
				"handle_inputs": map[string]any{},
				"attempt":       1,
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/experiments/runs/exprun_123/cancel":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "exprun_123",
				"experiment_id": "exp_123",
				"lifecycle":     map[string]any{"status": "cancelled"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/experiments/metrics":
			_ = json.NewEncoder(w).Encode(map[string]any{"view": "summary"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Workflows.Experiments.Runs.Get(context.Background(), "exprun_123")
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "exprun_123" {
		t.Fatalf("result = %#v", result)
	}
	results, err := client.Workflows.Experiments.Runs.Results.List(context.Background(), "exprun_123", 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(results.Data) != 0 || rawQuery != "limit=25&run_id=exprun_123" {
		t.Fatalf("results = %#v rawQuery = %q", results, rawQuery)
	}
	runResult, err := client.Workflows.Experiments.Runs.Results.Get(context.Background(), "expresult_123")
	if err != nil {
		t.Fatal(err)
	}
	if runResult.ID != "expresult_123" || runResult.RunID != "exprun_123" {
		t.Fatalf("runResult = %#v", runResult)
	}
	cancelled, err := client.Workflows.Experiments.Runs.Cancel(context.Background(), "exprun_123")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.ID != "exprun_123" || cancelled.Lifecycle.Status != "cancelled" {
		t.Fatalf("cancelled = %#v", cancelled)
	}
	cancelledJSON, err := json.Marshal(cancelled)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(cancelledJSON), "experiment_id") || strings.Contains(string(cancelledJSON), "workflow") {
		t.Fatalf("cancel response should only model id/lifecycle, got %s", cancelledJSON)
	}
	metrics, err := client.Workflows.Experiments.Runs.Metrics.Get(context.Background(), "exprun_123", nil)
	if err != nil {
		t.Fatal(err)
	}
	if metrics["view"] != "summary" {
		t.Fatalf("result = %#v rawQuery = %q", result, rawQuery)
	}
	expected := []string{
		"GET /workflows/experiments/runs/exprun_123",
		"GET /workflows/experiments/results",
		"GET /workflows/experiments/results/expresult_123",
		"POST /workflows/experiments/runs/exprun_123/cancel",
		"GET /workflows/experiments/metrics",
	}
	if strings.Join(requests, ",") != strings.Join(expected, ",") {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowTestRunResultsGetUsesFlatResultIDRoute(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/tests/results/wfresult_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":      "wfresult_123",
				"run_id":  "wftestrun_123",
				"test_id": "wfnodetest_123",
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Workflows.Tests.Runs.Results.Get(context.Background(), "wfresult_123")
	if err != nil {
		t.Fatal(err)
	}
	if (*result)["id"] != "wfresult_123" || (*result)["test_id"] != "wfnodetest_123" {
		t.Fatalf("result = %#v", result)
	}
	if strings.Join(requests, ",") != "GET /workflows/tests/results/wfresult_123" {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowExperimentRunRequestsSendCanonicalBodies(t *testing.T) {
	var runBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/experiments/runs":
			if r.URL.RawQuery != "" {
				t.Fatalf("expected no query params, got %q", r.URL.RawQuery)
			}
			if err := json.NewDecoder(r.Body).Decode(&runBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                     "exprun_123",
				"experiment_id":          "exp_123",
				"lifecycle":              map[string]any{"status": "queued"},
				"definition_fingerprint": "fp",
				"document_count":         1,
				"n_consensus":            5,
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Workflows.Experiments.Runs.Create(context.Background(), "wf_123", "exp_123"); err != nil {
		t.Fatal(err)
	}
	if runBody["experiment_id"] != "exp_123" || runBody["workflow_id"] != "wf_123" {
		t.Fatalf("run body = %#v", runBody)
	}
}

func TestWorkflowTestAndExperimentRunsUseDedicatedTimingShapes(t *testing.T) {
	testRunJSON, err := json.Marshal(WorkflowTestRun{
		ID:        "wftestrun_123",
		Lifecycle: WorkflowTestRunLifecycle{Status: "completed"},
		Timing:    WorkflowTestRunTiming{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(testRunJSON), "review_waiting_started_at") ||
		strings.Contains(string(testRunJSON), "accumulated_review_waiting_ms") {
		t.Fatalf("workflow test run timing should not include workflow-run review fields: %s", testRunJSON)
	}

	experimentRunJSON, err := json.Marshal(ExperimentRun{
		ID:        "exprun_123",
		Lifecycle: ExperimentRunLifecycle{Status: "completed"},
		Timing:    ExperimentRunTiming{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(experimentRunJSON), "review_waiting_started_at") ||
		strings.Contains(string(experimentRunJSON), "accumulated_review_waiting_ms") {
		t.Fatalf("experiment run timing should not include workflow-run review fields: %s", experimentRunJSON)
	}
}

func TestWorkflowRunsListDeleteCancelRestartAndExport(t *testing.T) {
	var cancelBody map[string]any
	var restartBody map[string]any
	var exportBody map[string]any
	var listQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs":
			listQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{workflowRunResponse("run_123", "wf_123", "running")},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/workflows/runs/run_123":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/runs/run_123/cancel":
			if err := json.NewDecoder(r.Body).Decode(&cancelBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"run":                 workflowRunResponse("run_123", "wf_123", "cancelled"),
				"cancellation_status": "cancelled",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/runs":
			if err := json.NewDecoder(r.Body).Decode(&restartBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(workflowRunResponse("run_456", "wf_123", "running"))
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_123/config":
			_ = json.NewEncoder(w).Encode(map[string]any{"blocks": []string{"start-1"}})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_123/execution-order":
			_ = json.NewEncoder(w).Encode(map[string]any{"order": []string{"start-1", "extract-1"}})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_123/documents/extract-1":
			_ = json.NewEncoder(w).Encode(map[string]any{"url": "https://example.com/doc"})
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/runs/export-payload":
			if err := json.NewDecoder(r.Body).Decode(&exportBody); err != nil {
				t.Fatal(err)
			}
			_, _ = io.WriteString(w, `{"csv_data":"a,b\n1,2\n","rows":1,"columns":2}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	runs, err := client.Workflows.Runs.List(context.Background(), &ListWorkflowRunsParams{
		WorkflowID:   "wf_123",
		Statuses:     []string{"running", "awaiting_review"},
		TriggerTypes: []string{"api", "email"},
		Fields:       []string{"id", "lifecycle"},
		Limit:        5,
		Order:        "asc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(runs.Data) != 1 || runs.Data[0].WorkflowID != "wf_123" {
		t.Fatalf("runs = %#v", runs)
	}
	if !strings.Contains(listQuery, "statuses=running%2Cawaiting_review") {
		t.Fatalf("list query = %s", listQuery)
	}

	if err := client.Workflows.Runs.Delete(context.Background(), "run_123"); err != nil {
		t.Fatal(err)
	}
	cancelled, err := client.Workflows.Runs.Cancel(context.Background(), "run_123", WorkflowRunCommandRequest{CommandID: "cmd_cancel"})
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.CancellationStatus != "cancelled" || cancelBody["command_id"] != "cmd_cancel" {
		t.Fatalf("cancelled = %#v body = %#v", cancelled, cancelBody)
	}
	restarted, err := client.Workflows.Runs.Restart(context.Background(), "run_123", WorkflowRunCommandRequest{CommandID: "cmd_restart"})
	if err != nil {
		t.Fatal(err)
	}
	if restarted.ID != "run_456" || restartBody["restart_of"] != "run_123" || restartBody["command_id"] != "cmd_restart" || restartBody["config_source"] != "published" {
		t.Fatalf("restarted = %#v body = %#v", restarted, restartBody)
	}

	exported, err := client.Workflows.Runs.Export(context.Background(), ExportWorkflowRunsRequest{
		WorkflowID:       "wf_123",
		BlockID:          "extract-1",
		SelectedRunIDs:   []string{"run_123"},
		TriggerTypes:     []string{},
		PreferredColumns: []string{"total"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if exported.Rows != 1 || exportBody["export_source"] != "outputs" {
		t.Fatalf("exported = %#v body = %#v", exported, exportBody)
	}
	if _, ok := exportBody["trigger_types"]; ok {
		t.Fatalf("export body should omit empty trigger_types filter, got %#v", exportBody["trigger_types"])
	}
}

func TestWorkflowRunsExportOmitsEmptySelectedRunIDs(t *testing.T) {
	var exportBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/runs/export-payload" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&exportBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"csv_data": "filename;amount\nUnknown file;10",
			"rows":     1,
			"columns":  2,
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Workflows.Runs.Export(context.Background(), ExportWorkflowRunsRequest{
		WorkflowID:     "wf_123",
		BlockID:        "block_1",
		SelectedRunIDs: []string{},
		TriggerTypes:   []string{"manual"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := exportBody["selected_run_ids"]; ok {
		t.Fatalf("export body should omit empty selected_run_ids filter, got %#v", exportBody["selected_run_ids"])
	}
	triggerTypes, ok := exportBody["trigger_types"].([]any)
	if !ok || len(triggerTypes) != 1 || triggerTypes[0] != "manual" {
		t.Fatalf("trigger_types = %#v", exportBody["trigger_types"])
	}
}

func TestWorkflowRunsDoNotExposeWaitHelpers(t *testing.T) {
	runServiceType := reflect.TypeOf(&WorkflowRunsService{})
	for _, methodName := range []string{"WaitForCompletion", "Wait", "CreateAndWait"} {
		if _, ok := runServiceType.MethodByName(methodName); ok {
			t.Fatalf("WorkflowRunsService still exposes %s", methodName)
		}
	}
}

func TestWorkflowServicesDoNotExposeRemovedMethods(t *testing.T) {
	cases := []struct {
		name    string
		service any
		methods []string
	}{
		{
			name:    "WorkflowsService",
			service: &WorkflowsService{},
			methods: []string{"Duplicate", "GetEntities", "GetResolvedSchemas", "ListSnapshots"},
		},
		{
			name:    "WorkflowReviewsService",
			service: &WorkflowReviewsService{},
			methods: []string{"WaitFor"},
		},
		{
			name:    "WorkflowBlocksService",
			service: &WorkflowBlocksService{},
			methods: []string{"ConfigHistory", "GetResolvedSchemas", "CreateBatch", "ListBlock executions", "PrepareExecute", "Execute"},
		},
		{
			name:    "WorkflowStepsService",
			service: &WorkflowStepsService{},
			methods: []string{"ListBlock executions", "PrepareListBlock executions", "PrepareExecute", "Execute"},
		},
		{
			name:    "WorkflowEdgesService",
			service: &WorkflowEdgesService{},
			methods: []string{"CreateBatch", "DeleteAll"},
		},
		{
			name:    "WorkflowExperimentsService",
			service: &WorkflowExperimentsService{},
			methods: []string{"Duplicate", "ListEligibleBlocks"},
		},
		{
			name:    "JobsService",
			service: &JobsService{},
			methods: []string{"RetrieveFull"},
		},
	}
	for _, tc := range cases {
		serviceType := reflect.TypeOf(tc.service)
		for _, methodName := range tc.methods {
			if _, ok := serviceType.MethodByName(methodName); ok {
				t.Fatalf("%s still exposes %s", tc.name, methodName)
			}
		}
	}
}

func TestWorkflowTestsDoNotExposeWaitForCompletion(t *testing.T) {
	testServiceType := reflect.TypeOf(&WorkflowTestsService{})
	if _, ok := testServiceType.MethodByName("WaitForCompletion"); ok {
		t.Fatal("WorkflowTestsService still exposes WaitForCompletion")
	}
}

func workflowRunResponse(runID string, workflowID string, lifecycleKind string) map[string]any {
	return map[string]any{
		"id":              runID,
		"organization_id": "org_123",
		"workflow": map[string]any{
			"workflow_id":       workflowID,
			"version_id":        "ver_0123456789abcdef0123456789abcdef",
			"name_at_run_time":  "Workflow",
			"requested_version": "production",
		},
		"trigger": map[string]any{"type": "api"},
		"lifecycle": map[string]any{
			"status": lifecycleKind,
		},
		"timing": map[string]any{
			"created_at": "2026-05-10T00:00:00Z",
		},
		"inputs": map[string]any{
			"documents": map[string]any{},
			"json_data": map[string]any{},
		},
	}
}

func TestWorkflowStepsGetRequiresStepID(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Workflows.Steps.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "retab: stepID is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkflowStepsGet(t *testing.T) {
	var seenMethod string
	var seenPath string
	var seenQuery url.Values
	var seenAPIKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		seenQuery = r.URL.Query()
		seenAPIKey = r.Header.Get("Api-Key")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"step_id":     "step_123",
			"run_id":      "run_123",
			"block_id":    "extract-1",
			"block_type":  "extract",
			"block_label": "Extract",
			"lifecycle":   map[string]any{"status": "completed"},
			"handle_inputs": map[string]any{
				"input-file-document": map[string]any{"type": "file"},
			},
			"handle_outputs": map[string]any{
				"output-json-0": map[string]any{"type": "json", "data": map[string]any{"ok": true}},
				"output-json-ref": map[string]any{
					"type": "json_ref",
					"artifact_ref": map[string]any{
						"id":  "artifact_123",
						"uri": "gs://bucket/run/block/output.json",
					},
					"preview": map[string]any{"truncated": true},
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	step, err := client.Workflows.Steps.Get(context.Background(), "step_123")
	if err != nil {
		t.Fatal(err)
	}

	if seenMethod != http.MethodGet {
		t.Fatalf("method = %s", seenMethod)
	}
	if seenPath != "/workflows/steps/step_123" || seenQuery.Encode() != "" {
		t.Fatalf("path = %s?%s", seenPath, seenQuery.Encode())
	}
	if seenAPIKey != "test-key" {
		t.Fatalf("api key = %s", seenAPIKey)
	}
	if step.StepID != "step_123" || step.RunID != "run_123" {
		t.Fatalf("step ids = %#v", step)
	}
	if step.BlockID != "extract-1" {
		t.Fatalf("block id = %s", step.BlockID)
	}
	if step.Lifecycle["status"] != "completed" {
		t.Fatalf("lifecycle = %#v", step.Lifecycle)
	}
	if len(step.HandleInputs) != 1 {
		t.Fatalf("handle inputs = %#v", step.HandleInputs)
	}
	if step.HandleOutputs["output-json-0"].Type != "json" {
		t.Fatalf("output-json-0 = %#v", step.HandleOutputs["output-json-0"])
	}
	refPayload, ok := step.HandleOutputs["output-json-ref"]
	if !ok {
		t.Fatalf("json_ref payload = %#v", step.HandleOutputs["output-json-ref"])
	}
	if refPayload.Type != "json_ref" {
		t.Fatalf("json_ref type = %v", refPayload.Type)
	}
	if refPayload.ArtifactRef["id"] != "artifact_123" {
		t.Fatalf("artifact_ref = %#v", refPayload.ArtifactRef)
	}
	if refPayload.Preview["truncated"] != true {
		t.Fatalf("preview = %#v", refPayload.Preview)
	}
	encoded, err := json.Marshal(step)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"terminal"`) {
		t.Fatalf("step JSON exposed removed state fields: %s", string(encoded))
	}
}

func TestWorkflowRunStepsListNormalizesNullHandles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": [
				{
					"run_id": "run_123",
					"organization_id": "org_123",
					"block_id": "start-1",
					"step_id": "start-1",
					"block_type": "start_document",
					"block_label": "Start",
					"lifecycle": {"status": "completed"},
					"handle_inputs": null,
					"handle_outputs": null
				}
			],
			"list_metadata": {"before": null, "after": null}
		}`))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	steps, err := client.Workflows.Steps.List(context.Background(), "run_123", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps.Data) != 1 {
		t.Fatalf("steps length = %d", len(steps.Data))
	}
	if steps.Data[0].HandleInputs == nil || steps.Data[0].HandleOutputs == nil {
		t.Fatalf("handle maps should be normalized: %#v", steps.Data[0])
	}
	if steps.Data[0].Lifecycle["status"] != "completed" {
		t.Fatalf("lifecycle = %#v", steps.Data[0].Lifecycle)
	}
	if steps.ListMetadata.Before != "" || steps.ListMetadata.After != "" {
		t.Fatalf("steps list_metadata = %#v", steps.ListMetadata)
	}
	encoded, err := json.Marshal(steps.Data[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"terminal"`) {
		t.Fatalf("workflow run step JSON exposed removed state fields: %s", string(encoded))
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"detail":{"code":"HTTP_EXCEPTION","message":"An HTTP exception occurred.","details":{"error":"Invalid API Key."}}}`))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Workflows.Runs.Get(context.Background(), "run_123")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
	if apiErr.Code != "HTTP_EXCEPTION" {
		t.Fatalf("code = %s", apiErr.Code)
	}
	if apiErr.Details["error"] != "Invalid API Key." {
		t.Fatalf("details = %#v", apiErr.Details)
	}
}

// TestAPIErrorUnwrappedEnvelope pins that APIError parsing also handles the
// error envelope when it arrives at the top level instead of under a
// `detail` key. Some endpoints (e.g. POST /v1/jobs validation failures)
// return `{"code":...,"message":...,"details":...}` directly. Without
// handling that shape the SDK falls back to a generic "Request failed (N)"
// message and drops Code/Details — so the CLI renders a degraded error.
func TestAPIErrorUnwrappedEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"code":"HTTP_EXCEPTION","message":"An HTTP exception occurred.","details":{"error":"document field required"}}`))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Workflows.Runs.Get(context.Background(), "run_123")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
	if apiErr.Code != "HTTP_EXCEPTION" {
		t.Fatalf("code = %q, want HTTP_EXCEPTION", apiErr.Code)
	}
	if apiErr.Message != "An HTTP exception occurred." {
		t.Fatalf("message = %q, want the envelope message", apiErr.Message)
	}
	if apiErr.Details["error"] != "document field required" {
		t.Fatalf("details = %#v", apiErr.Details)
	}
}

// TestAPIErrorFlatValidationEnvelope pins that ParseAPIError surfaces the
// `message` field of the backend's flat 422 envelope.
//
// FastAPI request-validation failures (RequestValidationError) are returned
// by main_server's global handler as {"status_code","message","data"} —
// NOT the {"detail":{...}} shape every other error uses. ParseAPIError only
// understood the nested shape, so every 422 (the most common error class:
// bad request bodies) degraded to the generic "Request failed (422)" with
// the real validation detail buried in the raw Body field.
func TestAPIErrorFlatValidationEnvelope(t *testing.T) {
	const validationMessage = `[{"type": "missing", "loc": ["body", "document", "filename"], "msg": "Field required"}]`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status_code": 10422,
			"message":     validationMessage,
			"data":        nil,
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Workflows.Runs.Get(context.Background(), "run_123")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
	if apiErr.Message != validationMessage {
		t.Fatalf("message = %q, want the validation detail from the flat envelope", apiErr.Message)
	}
}
