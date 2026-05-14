package retab

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestWorkflowsCreateUpdatePublishDuplicateGetEntitiesAndDelete(t *testing.T) {
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
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/duplicate":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "wf_456",
				"name":        "Renamed copy",
				"description": "Updated",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/entities":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workflow": map[string]any{"id": "wf_123", "name": "Renamed", "description": "Updated"},
				"blocks": []map[string]any{{
					"id":              "start-1",
					"workflow_id":     "wf_123",
					"organization_id": "org_123",
					"type":            "start",
				}},
				"edges": []map[string]any{},
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

	duplicate, err := client.Workflows.Duplicate(context.Background(), "wf_123")
	if err != nil {
		t.Fatal(err)
	}
	if duplicate.ID != "wf_456" {
		t.Fatalf("duplicate id = %s", duplicate.ID)
	}

	entities, err := client.Workflows.GetEntities(context.Background(), "wf_123")
	if err != nil {
		t.Fatal(err)
	}
	if len(entities.Blocks) != 1 || entities.Blocks[0].ID != "start-1" {
		t.Fatalf("entities = %#v", entities)
	}

	if err := client.Workflows.Delete(context.Background(), "wf_123"); err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"POST /workflows",
		"PATCH /workflows/wf_123",
		"POST /workflows/wf_123/publish",
		"POST /workflows/wf_123/duplicate",
		"GET /workflows/wf_123/entities",
		"DELETE /workflows/wf_123",
	}
	if strings.Join(requests, ",") != strings.Join(expected, ",") {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowsPrepareDiagnoseMatchesPythonSurface(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	blocks := []map[string]any{{"id": "start-1", "type": "start"}}
	edges := []map[string]any{{"id": "edge-1", "source": "start-1", "target": "extract-1"}}

	prepared := client.Workflows.PrepareDiagnose("wf_123", blocks, edges)
	if prepared.URL != "/workflows/wf_123/diagnose-graph" || prepared.Method != http.MethodPost {
		t.Fatalf("prepared diagnose = %#v", prepared)
	}
	body, ok := prepared.Body.(DiagnoseWorkflowGraphRequest)
	if !ok {
		t.Fatalf("prepared diagnose body = %#v", prepared.Body)
	}
	if !body.RePropagate || body.Blocks[0]["id"] != "start-1" || body.Edges[0]["id"] != "edge-1" {
		t.Fatalf("prepared diagnose body = %#v", body)
	}

	prepared = client.Workflows.PrepareDiagnose("wf_123", blocks, edges, false)
	body = prepared.Body.(DiagnoseWorkflowGraphRequest)
	if body.RePropagate {
		t.Fatalf("expected re-propagate override to be false: %#v", body)
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
		"GET /workflows/spec/wf_123",
	}
	if strings.Join(requests, ",") != strings.Join(expected, ",") {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowArtifactsGetListAndPrepare(t *testing.T) {
	var requests []string
	var listQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/artifacts/extraction/ext_123":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"operation": "extraction",
				"id":        "ext_123",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/artifacts":
			listQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"operation": "extraction",
				"id":        "ext_123",
			}})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	preparedGet := client.Workflows.Artifacts.PrepareGet("extraction", "ext_123")
	if preparedGet.URL != "/workflows/artifacts/extraction/ext_123" || preparedGet.Method != http.MethodGet {
		t.Fatalf("prepared artifact get = %#v", preparedGet)
	}
	preparedList := client.Workflows.Artifacts.PrepareList(ListWorkflowArtifactsParams{
		RunID:     "run_123",
		Operation: "extraction",
		BlockID:   "extract-1",
	})
	if preparedList.URL != "/workflows/artifacts" || preparedList.Method != http.MethodGet || preparedList.Params.Get("run_id") != "run_123" {
		t.Fatalf("prepared artifact list = %#v", preparedList)
	}

	artifact, err := client.Workflows.Artifacts.GetRef(context.Background(), StepArtifactRef{
		Operation: "extraction",
		ID:        "ext_123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if (*artifact)["id"] != "ext_123" {
		t.Fatalf("artifact = %#v", artifact)
	}

	artifacts, err := client.Workflows.Artifacts.List(context.Background(), ListWorkflowArtifactsParams{
		RunID:     "run_123",
		Operation: "extraction",
		BlockID:   "extract-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts) != 1 || artifacts[0]["id"] != "ext_123" {
		t.Fatalf("artifacts = %#v", artifacts)
	}
	if !strings.Contains(listQuery, "run_id=run_123") || !strings.Contains(listQuery, "operation=extraction") || !strings.Contains(listQuery, "block_id=extract-1") {
		t.Fatalf("list query = %s", listQuery)
	}
	if strings.Join(requests, ",") != "GET /workflows/artifacts/extraction/ext_123,GET /workflows/artifacts" {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowBlocksPrepareSimulateMatchesPythonSurface(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	checkEligibility := false

	prepared := client.Workflows.Blocks.PrepareSimulate(SimulateBlockRequest{
		RunID:            "run_123",
		BlockID:          "extract-1",
		NConsensus:       5,
		StepID:           "step_123",
		CheckEligibility: &checkEligibility,
	})
	if prepared.URL != "/workflows/runs/run_123/steps/extract-1/simulate" || prepared.Method != http.MethodPost {
		t.Fatalf("prepared simulate = %#v", prepared)
	}
	if prepared.Params.Get("n_consensus") != "5" || prepared.Params.Get("step_id") != "step_123" || prepared.Params.Get("check_eligibility") != "false" {
		t.Fatalf("prepared simulate params = %#v", prepared.Params)
	}

	prepared = client.Workflows.Blocks.PrepareSimulate(SimulateBlockRequest{
		RunID:   "run_123",
		BlockID: "extract-1",
	})
	if len(prepared.Params) != 0 {
		t.Fatalf("expected default simulate params to be empty: %#v", prepared.Params)
	}
}

func TestWorkflowExperimentRunsGetUsesContentRoute(t *testing.T) {
	var rawQuery string
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/wf_123/experiments/exp_123/documents/doc_123/cancel":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "cancelled"})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/wf_123/experiments/exp_123/content":
			rawQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"experiment_id": "exp_123",
				"run_id":        "exprun_123",
				"content":       map[string]any{"jobs": []map[string]any{}},
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

	result, err := client.Workflows.Experiments.Runs.Get(context.Background(), "wf_123", "exp_123", "exprun_123")
	if err != nil {
		t.Fatal(err)
	}
	if result.RunID != "exprun_123" || rawQuery != "run_id=exprun_123" {
		t.Fatalf("result = %#v rawQuery = %q", result, rawQuery)
	}
	cancelled, err := client.Workflows.Experiments.Runs.CancelDocument(context.Background(), "wf_123", "exp_123", "doc_123")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != "cancelled" {
		t.Fatalf("cancelled = %#v", cancelled)
	}
	expected := []string{
		"GET /workflows/wf_123/experiments/exp_123/content",
		"POST /workflows/wf_123/experiments/exp_123/documents/doc_123/cancel",
	}
	if strings.Join(requests, ",") != strings.Join(expected, ",") {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestWorkflowRunsListDeleteCancelRestartHILAndExport(t *testing.T) {
	var cancelBody map[string]any
	var restartBody map[string]any
	var hilBody map[string]any
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
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/runs/run_123/restart":
			if err := json.NewDecoder(r.Body).Decode(&restartBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(workflowRunResponse("run_456", "wf_123", "running"))
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/runs/run_123/hil-decisions":
			if err := json.NewDecoder(r.Body).Decode(&hilBody); err != nil {
				t.Fatal(err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"decision": map[string]any{
					"run_id":            "run_123",
					"block_id":          "review-1",
					"decision_received": true,
					"decision_applied":  true,
					"approved":          true,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_123/hil-decisions/review-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"run_id":            "run_123",
				"block_id":          "review-1",
				"decision_received": true,
				"decision_applied":  true,
				"approved":          true,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_123/agent-hil-reviews/review-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                       "agrev_abc",
				"organization_id":          "org_1",
				"run_id":                   "run_123",
				"block_id":                 "review-1",
				"workflow_id":              "wf_123",
				"mode":                     "pre_review",
				"status":                   "proposed",
				"managed_agent_session_id": "sesn_1",
				"managed_agent_vault_id":   "vlt_1",
				"proposed_decision": map[string]any{
					"approved":      true,
					"modified_data": map[string]any{"total": 325},
					"confidence":    0.92,
					"evidence": []any{
						map[string]any{
							"field_path": "total",
							"action":     "modified",
							"quote":      "Total $325.00",
							"source":     map[string]any{"document_index": 0, "page_number": 1},
							"from_value": 505,
							"to_value":   325,
						},
					},
					"changed_paths": []string{"total"},
					"escalate":      false,
				},
				"auto_threshold":  0.85,
				"timeout_seconds": 900,
				"created_at":      "2026-01-01T00:00:00Z",
				"updated_at":      "2026-01-01T00:00:01Z",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_123/config":
			_ = json.NewEncoder(w).Encode(map[string]any{"blocks": []string{"start-1"}})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_123/execution-order":
			_ = json.NewEncoder(w).Encode(map[string]any{"order": []string{"start-1", "extract-1"}})
		case r.Method == http.MethodGet && r.URL.Path == "/workflows/runs/run_123/documents/extract-1":
			_ = json.NewEncoder(w).Encode(map[string]any{"url": "https://example.com/doc"})
		case r.Method == http.MethodPost && r.URL.Path == "/workflows/runs/export_payload":
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
		Statuses:     []string{"running", "waiting_for_human"},
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
	if !strings.Contains(listQuery, "statuses=running%2Cwaiting_for_human") {
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
	if restarted.ID != "run_456" || restartBody["command_id"] != "cmd_restart" {
		t.Fatalf("restarted = %#v body = %#v", restarted, restartBody)
	}

	decision, err := client.Workflows.Runs.SubmitHILDecision(context.Background(), "run_123", SubmitHILDecisionRequest{
		BlockID:      "review-1",
		Approved:     true,
		ModifiedData: map[string]any{"approved_total": 12},
		CommandID:    "cmd_hil",
	})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Decision.BlockID != "review-1" || hilBody["block_id"] != "review-1" || hilBody["approved"] != true {
		t.Fatalf("decision = %#v body = %#v", decision, hilBody)
	}
	gotDecision, err := client.Workflows.Runs.GetHILDecision(context.Background(), "run_123", "review-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotDecision.BlockID != "review-1" {
		t.Fatalf("got decision = %#v", gotDecision)
	}
	agentReview, err := client.Workflows.Runs.GetAgentHILReview(context.Background(), "run_123", "review-1")
	if err != nil {
		t.Fatal(err)
	}
	if agentReview.ID != "agrev_abc" || agentReview.Mode != "pre_review" || agentReview.Status != "proposed" {
		t.Fatalf("agent review = %#v", agentReview)
	}
	if agentReview.ProposedDecision == nil || agentReview.ProposedDecision.Confidence != 0.92 {
		t.Fatalf("proposed decision = %#v", agentReview.ProposedDecision)
	}
	if len(agentReview.ProposedDecision.Evidence) != 1 || agentReview.ProposedDecision.Evidence[0].FieldPath != "total" {
		t.Fatalf("evidence = %#v", agentReview.ProposedDecision.Evidence)
	}
	config, err := client.Workflows.Runs.GetConfig(context.Background(), "run_123")
	if err != nil {
		t.Fatal(err)
	}
	if config["blocks"] == nil {
		t.Fatalf("config = %#v", config)
	}
	executionOrder, err := client.Workflows.Runs.ExecutionOrder(context.Background(), "run_123")
	if err != nil {
		t.Fatal(err)
	}
	if executionOrder["order"] == nil {
		t.Fatalf("execution order = %#v", executionOrder)
	}
	documentURL, err := client.Workflows.Runs.GetDocumentURL(context.Background(), "run_123", "extract-1")
	if err != nil {
		t.Fatal(err)
	}
	if documentURL["url"] != "https://example.com/doc" {
		t.Fatalf("document URL = %#v", documentURL)
	}

	exported, err := client.Workflows.Runs.Export(context.Background(), ExportWorkflowRunsRequest{
		WorkflowID:       "wf_123",
		BlockID:          "extract-1",
		SelectedRunIDs:   []string{"run_123"},
		PreferredColumns: []string{"total"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if exported.Rows != 1 || exportBody["export_source"] != "outputs" {
		t.Fatalf("exported = %#v body = %#v", exported, exportBody)
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

func TestWorkflowRunStepsGetRequiresBlockID(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Workflows.Runs.Steps.Get(context.Background(), "run_123", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "retab: blockID is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkflowRunStepsGet(t *testing.T) {
	var seenMethod string
	var seenPath string
	var seenAPIKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		seenAPIKey = r.Header.Get("Api-Key")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"block_id":    "extract-1",
			"step_id":     "extract-1",
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

	step, err := client.Workflows.Runs.Steps.Get(context.Background(), "run_123", "extract-1")
	if err != nil {
		t.Fatal(err)
	}

	if seenMethod != http.MethodGet {
		t.Fatalf("method = %s", seenMethod)
	}
	if seenPath != "/workflows/runs/run_123/steps/extract-1" {
		t.Fatalf("path = %s", seenPath)
	}
	if seenAPIKey != "test-key" {
		t.Fatalf("api key = %s", seenAPIKey)
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
	if step.ExtractedData() == nil {
		t.Fatal("expected extracted data")
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
		_, _ = w.Write([]byte(`[
			{
				"run_id": "run_123",
				"organization_id": "org_123",
				"block_id": "start-1",
				"step_id": "start-1",
				"block_type": "start",
				"block_label": "Start",
				"lifecycle": {"status": "completed"},
				"handle_inputs": null,
				"handle_outputs": null
			}
		]`))
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	steps, err := client.Workflows.Runs.Steps.List(context.Background(), "run_123")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 {
		t.Fatalf("steps length = %d", len(steps))
	}
	if steps[0].HandleInputs == nil || steps[0].HandleOutputs == nil {
		t.Fatalf("handle maps should be normalized: %#v", steps[0])
	}
	if steps[0].Lifecycle["status"] != "completed" {
		t.Fatalf("lifecycle = %#v", steps[0].Lifecycle)
	}
	encoded, err := json.Marshal(steps[0])
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
