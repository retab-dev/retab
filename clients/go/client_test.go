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

const invoiceWorkflowYAML = `apiVersion: workflows.retab.com/v1alpha2
kind: Workflow
metadata:
  id: wf_invoice_validation
  name: Invoice Validation Workflow
spec:
  blocks:
    start:
      type: start_json
      label: Invoice JSON
      config:
        json_schema:
          type: object
          properties:
            invoice_id:
              type: string
            line_items:
              type: array
              items:
                type: object
                properties:
                  description:
                    type: string
                  amount:
                    type: number
                required:
                  - description
                  - amount
            tax_rate:
              type: number
            stated_total:
              type: number
          required:
            - invoice_id
            - line_items
            - tax_rate
            - stated_total
    validate_total:
      type: function
      label: Validate Invoice Total
      config:
        output_schema:
          type: object
          properties:
            invoice_id:
              type: string
            subtotal:
              type: number
            computed_total:
              type: number
            is_valid:
              type: boolean
          required:
            - invoice_id
            - subtotal
            - computed_total
            - is_valid
        code: |
          from models import Input, Output

          def transform(input_data: Input) -> Output:
              subtotal = sum(item.amount for item in input_data.line_items)
              computed_total = round(subtotal + subtotal * input_data.tax_rate, 2)
              return Output(
                  invoice_id=input_data.invoice_id,
                  subtotal=subtotal,
                  computed_total=computed_total,
                  is_valid=abs(computed_total - input_data.stated_total) <= 0.01,
              )
  edges:
    - from:
        block: start
        handle: output-json-0
      to:
        block: validate_total
        handle: input-json-0
`

func TestWorkflowsCreateUpdatePublishAndDelete(t *testing.T) {
	var requests []string
	var updateBody map[string]any
	var publishBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "wf_123",
				"name":        "Invoice Extractor",
				"description": "",
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/v1/workflows/wf_123":
			if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "wf_123",
				"name":        "Renamed",
				"description": "Updated",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/wf_123/publish":
			if err := json.NewDecoder(r.Body).Decode(&publishBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "wf_123",
				"name":        "Renamed",
				"description": "Updated",
				"published":   map[string]any{"version_id": "ver_0123456789abcdef0123456789abcdef"},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/workflows/wf_123":
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

	workflowName := "Invoice Extractor"
	workflow, err := client.Workflows.Create(context.Background(), &WorkflowsCreateParams{Name: &workflowName})
	if err != nil {
		t.Fatal(err)
	}
	if workflow.ID != "wf_123" {
		t.Fatalf("workflow id = %s", workflow.ID)
	}

	name := "Renamed"
	description := "Updated"
	_, err = client.Workflows.Update(context.Background(), "wf_123", &WorkflowsUpdateParams{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updateBody["name"] != "Renamed" || updateBody["description"] != "Updated" {
		t.Fatalf("update body = %#v", updateBody)
	}

	publishDescription := "v1"
	published, err := client.Workflows.Publish(context.Background(), "wf_123", &WorkflowsPublishParams{
		Body: PublishWorkflowRequest{Description: &publishDescription},
	})
	if err != nil {
		t.Fatal(err)
	}
	if published.Published == nil || published.Published.VersionID == nil || *published.Published.VersionID != "ver_0123456789abcdef0123456789abcdef" {
		t.Fatalf("published = %#v", published.Published)
	}
	if publishBody["description"] != "v1" {
		t.Fatalf("publish body = %#v", publishBody)
	}

	if err := client.Workflows.Delete(context.Background(), "wf_123"); err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"POST /v1/workflows",
		"PATCH /v1/workflows/wf_123",
		"POST /v1/workflows/wf_123/publish",
		"DELETE /v1/workflows/wf_123",
	}
	if strings.Join(requests, ",") != strings.Join(expected, ",") {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestHandlePayloadDoesNotExposeRemovedTextField(t *testing.T) {
	if _, ok := reflect.TypeOf(PublicHandlePayload{}).FieldByName("Text"); ok {
		t.Fatal("PublicHandlePayload must not expose removed text handle payloads")
	}
}

func TestWorkflowBlocksUpdateSendsExplicitEmptyConfig(t *testing.T) {
	var updateBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/workflows/blocks/block_123" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "block_123",
			"workflow_id": "wf_123",
			"type":        "function",
			"config":      map[string]any{},
			"updated_at":  "2026-06-03T00:00:00Z",
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	mode := UpdateWorkflowBlockRequestConfigModeReplace
	_, err = client.Workflows.Blocks.Update(context.Background(), "block_123", &WorkflowBlocksUpdateParams{
		Config:     Ptr(map[string]interface{}{}),
		ConfigMode: &mode,
	})
	if err != nil {
		t.Fatal(err)
	}
	config, ok := updateBody["config"].(map[string]any)
	if !ok {
		t.Fatalf("update body omitted explicit empty config: %#v", updateBody)
	}
	if len(config) != 0 {
		t.Fatalf("config = %#v, want empty object", config)
	}
	if updateBody["config_mode"] != string(UpdateWorkflowBlockRequestConfigModeReplace) {
		t.Fatalf("update body = %#v", updateBody)
	}
}

func TestWorkflowBlocksUpdateConfigPresenceStates(t *testing.T) {
	tests := []struct {
		name       string
		params     *WorkflowBlocksUpdateParams
		assertBody func(*testing.T, map[string]any)
	}{
		{
			name: "absent config omitted",
			params: &WorkflowBlocksUpdateParams{
				Label: Ptr("Renamed"),
			},
			assertBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if _, ok := body["config"]; ok {
					t.Fatalf("config should be absent: %#v", body)
				}
				if body["label"] != "Renamed" {
					t.Fatalf("label = %#v, want Renamed", body["label"])
				}
			},
		},
		{
			name: "present empty config",
			params: &WorkflowBlocksUpdateParams{
				Config: Ptr(map[string]interface{}{}),
			},
			assertBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				config, ok := body["config"].(map[string]any)
				if !ok {
					t.Fatalf("config should be present as object: %#v", body)
				}
				if len(config) != 0 {
					t.Fatalf("config = %#v, want empty object", config)
				}
			},
		},
		{
			name: "present non-empty config",
			params: &WorkflowBlocksUpdateParams{
				Config: Ptr(map[string]interface{}{"prompt": "extract totals"}),
			},
			assertBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				config, ok := body["config"].(map[string]any)
				if !ok {
					t.Fatalf("config should be present as object: %#v", body)
				}
				if config["prompt"] != "extract totals" {
					t.Fatalf("config = %#v", config)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var updateBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method != http.MethodPatch || r.URL.Path != "/v1/workflows/blocks/block_123" {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
					t.Fatal(err)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":          "block_123",
					"workflow_id": "wf_123",
					"type":        "extract",
					"config":      map[string]any{},
					"updated_at":  "2026-06-03T00:00:00Z",
				})
			}))
			defer server.Close()

			client, err := NewClient("test-key", WithBaseURL(server.URL))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.Workflows.Blocks.Update(context.Background(), "block_123", test.params); err != nil {
				t.Fatal(err)
			}
			test.assertBody(t, updateBody)
		})
	}
}

func TestWorkflowSpecRoutesMatchPythonAndNode(t *testing.T) {
	var requests []string
	var validateBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/spec/validate" {
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

	if _, err := client.Workflows.Spec.Validate(context.Background(), &WorkflowSpecValidateParams{YamlDefinition: invoiceWorkflowYAML}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Plan(context.Background(), &WorkflowsPlanParams{YamlDefinition: invoiceWorkflowYAML}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Apply(context.Background(), &WorkflowsApplyParams{YamlDefinition: invoiceWorkflowYAML}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Workflows.Spec.Get(context.Background(), "wf_123"); err != nil {
		t.Fatal(err)
	}

	if validateBody["yaml_definition"] != invoiceWorkflowYAML {
		t.Fatalf("validate body = %#v", validateBody)
	}
	expected := []string{
		"POST /v1/workflows/spec/validate",
		"POST /v1/workflows/spec/plan",
		"POST /v1/workflows/spec/apply",
		"GET /v1/workflows/wf_123/spec",
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
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/artifacts":
			listQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"operation": "extraction",
					"id":        "ext_123",
				}},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/artifacts/ext_123":
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

	runID := "run_123"
	operation := WorkflowArtifactsOperation(StepArtifactRefOperationExtraction)
	blockID := "extract-1"
	artifacts, err := client.Workflows.Artifacts.List(context.Background(), &WorkflowArtifactsListParams{
		RunID:     &runID,
		Operation: &operation,
		BlockID:   &blockID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts.Data) != 1 || artifacts.Data[0].ID != "ext_123" {
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
	// Get returns the polymorphic artifact union losslessly as interface{}
	// (the union's only natural name, WorkflowArtifact, is taken by the flat
	// list-summary model). A JSON object decodes to map[string]any intact.
	artifactMap, ok := artifact.(map[string]any)
	if !ok || artifactMap["id"] != "ext_123" || artifactMap["operation"] != "extraction" {
		t.Fatalf("artifact = %#v", artifact)
	}
	if strings.Join(requests, ",") != "GET /v1/workflows/artifacts,GET /v1/workflows/artifacts/ext_123" {
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
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/experiments/runs/exprun_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "exprun_123",
				"experiment_id": "exp_123",
				"lifecycle":     map[string]any{"status": "completed"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/experiments/results":
			rawQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          []map[string]any{},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/experiments/results/expresult_123":
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
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/experiments/runs/exprun_123/cancel":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "exprun_123",
				"experiment_id": "exp_123",
				"lifecycle":     map[string]any{"status": "cancelled"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/experiments/metrics":
			_ = json.NewEncoder(w).Encode(map[string]any{"kind": "summary", "view": "summary"})
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
	limit := 25
	results, err := client.Workflows.Experiments.Results.List(context.Background(), &ExperimentRunResultsListParams{
		RunID:            "exprun_123",
		PaginationParams: PaginationParams{Limit: &limit},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results.Data) != 0 || rawQuery != "limit=25&run_id=exprun_123" {
		t.Fatalf("results = %#v rawQuery = %q", results, rawQuery)
	}
	runResult, err := client.Workflows.Experiments.Results.Get(context.Background(), "expresult_123")
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
	if cancelled.ID != "exprun_123" || cancelled.Lifecycle.Status() != "cancelled" {
		t.Fatalf("cancelled = %#v", cancelled)
	}
	cancelledJSON, err := json.Marshal(cancelled)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(cancelledJSON), "experiment_id") || strings.Contains(string(cancelledJSON), "workflow") {
		t.Fatalf("cancel response should only model id/lifecycle, got %s", cancelledJSON)
	}
	metrics, err := client.Workflows.Experiments.Metrics.Get(context.Background(), &ExperimentRunMetricsGetParams{RunID: "exprun_123"})
	if err != nil {
		t.Fatal(err)
	}
	if metrics.Kind() != "summary" {
		t.Fatalf("metrics kind = %q rawQuery = %q", metrics.Kind(), rawQuery)
	}
	if _, err := metrics.AsExperimentSummaryMetricsResponse(); err != nil {
		t.Fatalf("metrics as summary: %v", err)
	}
	expected := []string{
		"GET /v1/workflows/experiments/runs/exprun_123",
		"GET /v1/workflows/experiments/results",
		"GET /v1/workflows/experiments/results/expresult_123",
		"POST /v1/workflows/experiments/runs/exprun_123/cancel",
		"GET /v1/workflows/experiments/metrics",
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
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/tests/results/wfresult_123":
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

	result, err := client.Workflows.Tests.Results.Get(context.Background(), "wfresult_123")
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "wfresult_123" || result.TestID != "wfnodetest_123" {
		t.Fatalf("result = %#v", result)
	}
	if strings.Join(requests, ",") != "GET /v1/workflows/tests/results/wfresult_123" {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowExperimentRunRequestsSendCanonicalBodies(t *testing.T) {
	var runBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/experiments/runs":
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

	workflowID := "wf_123"
	if _, err := client.Workflows.Experiments.Runs.Create(context.Background(), &ExperimentRunsCreateParams{
		WorkflowID:   &workflowID,
		ExperimentID: "exp_123",
	}); err != nil {
		t.Fatal(err)
	}
	if runBody["experiment_id"] != "exp_123" || runBody["workflow_id"] != "wf_123" {
		t.Fatalf("run body = %#v", runBody)
	}
}

func TestWorkflowTestAndExperimentRunsUseDedicatedTimingShapes(t *testing.T) {
	completedStatus := "completed"
	testRunJSON, err := json.Marshal(WorkflowTestRun{
		ID:        "wftestrun_123",
		Lifecycle: WorkflowTestRunStatusFromPendingWorkflowTestRun(PendingWorkflowTestRun{Status: &completedStatus}),
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
		Lifecycle: WorkflowExperimentRunFromPendingWorkflowExperimentRun(PendingWorkflowExperimentRun{Status: &completedStatus}),
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

func TestWorkflowRunsListDeleteCancelCreateAndExport(t *testing.T) {
	var cancelBody map[string]any
	var createBody map[string]any
	var exportBody map[string]any
	var listQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/runs":
			listQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{workflowRunResponse("run_123", "wf_123", "running")},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/workflows/runs/run_123":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs/run_123/cancel":
			if err := json.NewDecoder(r.Body).Decode(&cancelBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"run":                 workflowRunResponse("run_123", "wf_123", "cancelled"),
				"cancellation_status": "cancelled",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs":
			if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(workflowRunResponse("run_456", "wf_123", "running"))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/runs/run_123/config":
			_ = json.NewEncoder(w).Encode(map[string]any{"blocks": []string{"start-1"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/runs/run_123/execution-order":
			_ = json.NewEncoder(w).Encode(map[string]any{"order": []string{"start-1", "extract-1"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/workflows/runs/run_123/documents/extract-1":
			_ = json.NewEncoder(w).Encode(map[string]any{"url": "https://example.com/doc"})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/workflows/runs/export":
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

	workflowID := "wf_123"
	status := WorkflowRunsStatus("running")
	triggerType := WorkflowRunsTriggerType("api")
	limit := 5
	order := "asc"
	runs, err := client.Workflows.Runs.List(context.Background(), &WorkflowRunsListParams{
		WorkflowID:  &workflowID,
		Status:      &status,
		TriggerType: &triggerType,
		PaginationParams: PaginationParams{
			Limit: &limit,
			Order: &order,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(runs.Data) != 1 || runs.Data[0].WorkflowID != "wf_123" {
		t.Fatalf("runs = %#v", runs)
	}
	if !strings.Contains(listQuery, "status=running") {
		t.Fatalf("list query = %s", listQuery)
	}

	if err := client.Workflows.Runs.Delete(context.Background(), "run_123"); err != nil {
		t.Fatal(err)
	}
	cancelCommandID := "cmd_cancel"
	cancelled, err := client.Workflows.Runs.Cancel(context.Background(), "run_123", &WorkflowRunsCancelParams{
		Body: CancelWorkflowRequest{CommandID: &cancelCommandID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.CancellationStatus == nil || *cancelled.CancellationStatus != "cancelled" || cancelBody["command_id"] != "cmd_cancel" {
		t.Fatalf("cancelled = %#v body = %#v", cancelled, cancelBody)
	}
	version := "production"
	jsonInputs := map[string]interface{}{"start": map[string]interface{}{"value": float64(1)}}
	created, err := client.Workflows.Runs.Create(context.Background(), &WorkflowRunsCreateParams{
		WorkflowID: "wf_123",
		Version:    &version,
		JSONInputs: &jsonInputs,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "run_456" || createBody["workflow_id"] != "wf_123" || createBody["version"] != "production" {
		t.Fatalf("created = %#v body = %#v", created, createBody)
	}
	if _, ok := createBody["restart_of"]; ok {
		t.Fatalf("restart_of leaked into create body: %#v", createBody)
	}

	exportSource := WorkflowExportPayloadRequestExportSourceOutputs
	exported, err := client.Workflows.Runs.Export(context.Background(), &WorkflowRunsExportParams{
		WorkflowID:       "wf_123",
		BlockID:          "extract-1",
		ExportSource:     &exportSource,
		SelectedRunIDs:   []string{"run_123"},
		PreferredColumns: []string{"total"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if exported.Rows != 1 || exportBody["export_source"] != "outputs" {
		t.Fatalf("exported = %#v body = %#v", exported, exportBody)
	}
	if _, ok := exportBody["trigger_type"]; ok {
		t.Fatalf("export body should omit empty trigger_type filter, got %#v", exportBody["trigger_type"])
	}
}

func TestWorkflowRunsCreateCoercesDocuments(t *testing.T) {
	var createBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(workflowRunResponse("run_456", "wf_123", "running"))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	documentsInput := map[string]interface{}{
		"remote":   "https://example.com/invoice.pdf",
		"uploaded": FileRef{ID: "file_123", Filename: "stored.pdf", MIMEType: "application/pdf"},
		"raw_ref":  map[string]any{"id": "file_456", "filename": "raw.pdf"},
	}
	_, err = client.Workflows.Runs.Create(context.Background(), &WorkflowRunsCreateParams{
		WorkflowID: "wf_123",
		Documents:  &documentsInput,
	})
	if err != nil {
		t.Fatal(err)
	}

	documents, ok := createBody["documents"].(map[string]any)
	if !ok {
		t.Fatalf("documents = %#v", createBody["documents"])
	}
	remote, ok := documents["remote"].(map[string]any)
	if !ok {
		t.Fatalf("documents.remote = %#v", documents["remote"])
	}
	if remote["filename"] != "invoice.pdf" || remote["url"] != "https://example.com/invoice.pdf" {
		t.Fatalf("documents.remote = %#v", remote)
	}
	uploaded, ok := documents["uploaded"].(map[string]any)
	if !ok {
		t.Fatalf("documents.uploaded = %#v", documents["uploaded"])
	}
	if uploaded["id"] != "file_123" || uploaded["filename"] != "stored.pdf" || uploaded["mime_type"] != "application/pdf" {
		t.Fatalf("documents.uploaded = %#v", uploaded)
	}
	rawRef, ok := documents["raw_ref"].(map[string]any)
	if !ok {
		t.Fatalf("documents.raw_ref = %#v", documents["raw_ref"])
	}
	if rawRef["id"] != "file_456" || rawRef["filename"] != "raw.pdf" {
		t.Fatalf("documents.raw_ref = %#v", rawRef)
	}
}

func TestWorkflowRunDocumentsSetCoercesDocuments(t *testing.T) {
	params := &WorkflowRunsCreateParams{WorkflowID: "wf_123"}
	if err := params.SetDocument("remote", "https://example.com/invoice.pdf"); err != nil {
		t.Fatal(err)
	}
	if err := params.SetDocument("uploaded", FileRef{ID: "file_123", Filename: "stored.pdf", MIMEType: "application/pdf"}); err != nil {
		t.Fatal(err)
	}

	remote, ok := (*params.Documents)["remote"].(MIMEData)
	if !ok {
		t.Fatalf("remote = %#v", (*params.Documents)["remote"])
	}
	if remote.Filename != "invoice.pdf" || remote.URL != "https://example.com/invoice.pdf" {
		t.Fatalf("remote = %#v", remote)
	}
	uploaded, ok := (*params.Documents)["uploaded"].(FileRef)
	if !ok {
		t.Fatalf("uploaded = %#v", (*params.Documents)["uploaded"])
	}
	if uploaded.ID != "file_123" {
		t.Fatalf("uploaded = %#v", uploaded)
	}
}

func TestWorkflowRunsExportOmitsEmptySelectedRunIDs(t *testing.T) {
	var exportBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/runs/export" {
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
	manualTrigger := WorkflowExportPayloadRequestTriggerTypeManual
	_, err = client.Workflows.Runs.Export(context.Background(), &WorkflowRunsExportParams{
		WorkflowID:     "wf_123",
		BlockID:        "block_1",
		SelectedRunIDs: []string{},
		TriggerType:    &manualTrigger,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := exportBody["selected_run_ids"]; ok {
		t.Fatalf("export body should omit empty selected_run_ids filter, got %#v", exportBody["selected_run_ids"])
	}
	triggerType, ok := exportBody["trigger_type"].(string)
	if !ok || triggerType != "manual" {
		t.Fatalf("trigger_type = %#v", exportBody["trigger_type"])
	}
}

func TestWorkflowTestRunsCreateSendsTypedScopeBody(t *testing.T) {
	var createBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workflows/tests/runs" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"wftestrun_123"}`)
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	testID := "wfnodetest_123"
	created, err := client.Workflows.Tests.Runs.Create(context.Background(), &WorkflowTestRunsCreateParams{
		WorkflowID: "wf_123",
		Scope: &WorkflowTestRunScope{
			Type:   WorkflowTestRunScopeTypeSingle,
			TestID: &testID,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "wftestrun_123" {
		t.Fatalf("created = %#v", created)
	}
	scope, ok := createBody["scope"].(map[string]any)
	if !ok {
		t.Fatalf("scope = %#v", createBody["scope"])
	}
	if createBody["workflow_id"] != "wf_123" || scope["type"] != "single" || scope["test_id"] != "wfnodetest_123" {
		t.Fatalf("body = %#v", createBody)
	}
	if _, ok := createBody["n_consensus"]; ok {
		t.Fatalf("n_consensus leaked into create body: %#v", createBody)
	}
}

func TestWorkflowRunsDoNotExposeWaitHelpers(t *testing.T) {
	runServiceType := reflect.TypeOf(&WorkflowRunService{})
	for _, methodName := range []string{"WaitForCompletion", "Wait", "CreateAndWait"} {
		if _, ok := runServiceType.MethodByName(methodName); ok {
			t.Fatalf("WorkflowRunService still exposes %s", methodName)
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
			name:    "WorkflowService",
			service: &WorkflowService{},
			methods: []string{"Duplicate", "GetEntities", "GetResolvedSchemas", "ListSnapshots"},
		},
		{
			name:    "WorkflowReviewService",
			service: &WorkflowReviewService{},
			methods: []string{"WaitFor"},
		},
		{
			name:    "WorkflowBlockService",
			service: &WorkflowBlockService{},
			methods: []string{"ConfigHistory", "GetResolvedSchemas", "CreateBatch", "ListBlock executions", "PrepareExecute", "Execute"},
		},
		{
			name:    "WorkflowStepService",
			service: &WorkflowStepService{},
			methods: []string{"ListBlock executions", "PrepareListBlock executions", "PrepareExecute", "Execute"},
		},
		{
			name:    "WorkflowEdgeService",
			service: &WorkflowEdgeService{},
			methods: []string{"CreateBatch", "DeleteAll"},
		},
		{
			name:    "WorkflowExperimentService",
			service: &WorkflowExperimentService{},
			methods: []string{"Duplicate", "ListEligibleBlocks"},
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
	testServiceType := reflect.TypeOf(&WorkflowTestService{})
	if _, ok := testServiceType.MethodByName("WaitForCompletion"); ok {
		t.Fatal("WorkflowTestService still exposes WaitForCompletion")
	}
}

func workflowRunResponse(runID string, workflowID string, lifecycleKind string) map[string]any {
	return map[string]any{
		"id":                  runID,
		"organization_id":     "org_123",
		"workflow_id":         workflowID,
		"workflow_version_id": "ver_0123456789abcdef0123456789abcdef",
		"trigger":             map[string]any{"type": "api"},
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

	_, err = client.Workflows.Steps.Get(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "retab: step_id is required" {
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
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	step, err := client.Workflows.Steps.Get(context.Background(), "step_123", nil)
	if err != nil {
		t.Fatal(err)
	}

	if seenMethod != http.MethodGet {
		t.Fatalf("method = %s", seenMethod)
	}
	if seenPath != "/v1/workflows/steps/step_123" || seenQuery.Encode() != "" {
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
	if step.Lifecycle.Status() != "completed" {
		t.Fatalf("lifecycle = %#v", step.Lifecycle)
	}
	if len(step.HandleInputs) != 1 {
		t.Fatalf("handle inputs = %#v", step.HandleInputs)
	}
	if step.HandleOutputs["output-json-0"].Type != "json" {
		t.Fatalf("output-json-0 = %#v", step.HandleOutputs["output-json-0"])
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

	runID := "run_123"
	steps, err := client.Workflows.Steps.List(context.Background(), &WorkflowStepsListParams{RunID: &runID})
	if err != nil {
		t.Fatal(err)
	}
	if len(steps.Data) != 1 {
		t.Fatalf("steps length = %d", len(steps.Data))
	}
	if steps.Data[0].HandleInputs != nil || steps.Data[0].HandleOutputs != nil {
		t.Fatalf("null handle maps should remain nil: %#v", steps.Data[0])
	}
	if steps.Data[0].Lifecycle.Status() != "completed" {
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

// TestAPIErrorEmptyDetailFallsBackToSiblings pins that an empty-string
// `detail` no longer wipes the message: parsing must fall back to the
// top-level message/code rather than surfacing a blank error.
func TestAPIErrorEmptyDetailFallsBackToSiblings(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://x.test/v1/foo", nil)
	resp := &http.Response{StatusCode: 400, Header: http.Header{}, Request: req}
	apiErr := ParseAPIError(resp, []byte(`{"detail":"","message":"real message","code":"ERR_X"}`))
	if apiErr.Message != "real message" {
		t.Errorf("Message = %q, want 'real message' (fell back to sibling)", apiErr.Message)
	}
	if apiErr.Code != "ERR_X" {
		t.Errorf("Code = %q, want ERR_X", apiErr.Code)
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

// FastAPI's *native* validation shape is {"detail": [ {loc, msg, type}, ... ]}
// (not wrapped by main_server). Before the []any case in parseAPIError this
// degraded to the generic "Request failed (422)" because the switch only
// handled string/map details; now the array is surfaced as the message.
func TestAPIErrorFastAPIDetailArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"detail": []any{
				map[string]any{"type": "missing", "loc": []any{"body", "document"}, "msg": "Field required"},
			},
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
	if !strings.Contains(apiErr.Message, "Field required") {
		t.Fatalf("message = %q, want the FastAPI validation detail surfaced", apiErr.Message)
	}
	if strings.Contains(apiErr.Message, "Request failed") {
		t.Fatalf("message degraded to generic fallback: %q", apiErr.Message)
	}
}
