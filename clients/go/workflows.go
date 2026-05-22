package retab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type WorkflowsService struct {
	client      *Client
	Runs        *WorkflowRunsService
	Steps       *WorkflowStepsService
	Reviews     *WorkflowReviewsService
	Artifacts   *WorkflowArtifactsService
	Blocks      *WorkflowBlocksService
	Edges       *WorkflowEdgesService
	Specs       *WorkflowSpecsService
	Tests       *WorkflowTestsService
	Experiments *WorkflowExperimentsService
}

func newWorkflowsService(client *Client) *WorkflowsService {
	service := &WorkflowsService{client: client}
	service.Runs = &WorkflowRunsService{client: client}
	service.Steps = &WorkflowStepsService{client: client}
	service.Reviews = &WorkflowReviewsService{
		client:   client,
		Versions: &WorkflowReviewVersionsService{client: client},
	}
	service.Artifacts = &WorkflowArtifactsService{client: client}
	service.Blocks = &WorkflowBlocksService{
		client:     client,
		Executions: &WorkflowBlockExecutionsService{client: client},
	}
	service.Edges = &WorkflowEdgesService{client: client}
	service.Specs = &WorkflowSpecsService{client: client}
	service.Tests = newWorkflowTestsService(client)
	service.Experiments = newWorkflowExperimentsService(client)
	return service
}

type ListWorkflowsParams struct {
	Before string
	After  string
	Limit  int
	Order  string
	SortBy string
	Fields string
}

type CreateWorkflowRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type UpdateWorkflowRequest struct {
	Name         *string               `json:"-"`
	Description  *string               `json:"-"`
	EmailTrigger *WorkflowEmailTrigger `json:"-"`
}

type PublishWorkflowRequest struct {
	Description string `json:"description"`
}

func (s *WorkflowsService) Get(ctx context.Context, workflowID string, opts ...RequestOption) (*Workflow, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var workflow Workflow
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID), nil, nil, &workflow, opts...)
	return &workflow, err
}

func (s *WorkflowsService) List(ctx context.Context, params *ListWorkflowsParams, opts ...RequestOption) (*PaginatedList[Workflow], error) {
	query := url.Values{}
	query.Set("limit", "10")
	query.Set("order", "desc")
	if params != nil {
		if params.Before != "" && params.After != "" {
			return nil, fmt.Errorf("retab: Before and After are mutually exclusive")
		}
		addQuery(query, "before", params.Before)
		addQuery(query, "after", params.After)
		addQuery(query, "order", params.Order)
		addQuery(query, "sort_by", params.SortBy)
		addQuery(query, "fields", params.Fields)
		if params.Limit > 0 {
			query.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
	}
	var result PaginatedList[Workflow]
	err := s.client.do(ctx, http.MethodGet, "/workflows", query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowsService) Create(ctx context.Context, request CreateWorkflowRequest, opts ...RequestOption) (*Workflow, error) {
	name := request.Name
	if name == "" {
		name = "Untitled Workflow"
	}
	body := map[string]any{
		"name":        name,
		"description": request.Description,
	}
	var workflow Workflow
	err := s.client.do(ctx, http.MethodPost, "/workflows", nil, body, &workflow, opts...)
	return &workflow, err
}

func (s *WorkflowsService) Update(ctx context.Context, workflowID string, request UpdateWorkflowRequest, opts ...RequestOption) (*Workflow, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := map[string]any{}
	if request.Name != nil {
		body["name"] = *request.Name
	}
	if request.Description != nil {
		body["description"] = *request.Description
	}
	if request.EmailTrigger != nil {
		emailTrigger := map[string]any{}
		if request.EmailTrigger.AllowedSenders != nil {
			emailTrigger["allowed_senders"] = request.EmailTrigger.AllowedSenders
		}
		if request.EmailTrigger.AllowedDomains != nil {
			emailTrigger["allowed_domains"] = request.EmailTrigger.AllowedDomains
		}
		body["email_trigger"] = emailTrigger
	}
	var workflow Workflow
	err := s.client.do(ctx, http.MethodPatch, "/workflows/"+url.PathEscape(workflowID), nil, body, &workflow, opts...)
	return &workflow, err
}

func (s *WorkflowsService) Delete(ctx context.Context, workflowID string, opts ...RequestOption) error {
	if workflowID == "" {
		return fmt.Errorf("retab: workflowID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/"+url.PathEscape(workflowID), nil, nil, nil, opts...)
}

func (s *WorkflowsService) Publish(ctx context.Context, workflowID string, request PublishWorkflowRequest, opts ...RequestOption) (*Workflow, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := map[string]any{"description": request.Description}
	var workflow Workflow
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(workflowID)+"/publish", nil, body, &workflow, opts...)
	return &workflow, err
}

// WorkflowDiagnosisIssue is one issue found by Diagnose.
type WorkflowDiagnosisIssue struct {
	Severity string  `json:"severity"`
	Code     string  `json:"code"`
	Message  string  `json:"message"`
	BlockID  *string `json:"block_id,omitempty"`
}

// WorkflowDiagnosisStats summarises the inspected graph.
type WorkflowDiagnosisStats struct {
	TotalBlocks int            `json:"total_blocks"`
	TotalEdges  int            `json:"total_edges"`
	BlockTypes  map[string]int `json:"block_types"`
	StartBlocks int            `json:"start_document_blocks"`
}

// WorkflowDiagnosisResponse is the result of POST /workflows/{id}/diagnose-graph.
type WorkflowDiagnosisResponse struct {
	IsValid     bool                     `json:"is_valid"`
	Issues      []WorkflowDiagnosisIssue `json:"issues"`
	Suggestions []string                 `json:"suggestions"`
	Stats       WorkflowDiagnosisStats   `json:"stats"`
}

// DiagnoseWorkflowGraphRequest is the lower-level body shape if you want to
// diagnose an in-memory editor graph that hasn't been saved yet.
type DiagnoseWorkflowGraphRequest struct {
	Blocks      []map[string]any `json:"blocks"`
	Edges       []map[string]any `json:"edges"`
	RePropagate bool             `json:"re_propagate"`
}

type DiagnoseWorkflowRequest struct {
	RePropagate bool `json:"re_propagate"`
}

func prepareDiagnoseGraphRequest(workflowID string, request DiagnoseWorkflowGraphRequest) PreparedRequest {
	return PreparedRequest{
		URL:    "/workflows/" + url.PathEscape(workflowID) + "/diagnose-graph",
		Method: http.MethodPost,
		Body:   request,
	}
}

// PrepareDiagnose returns the request descriptor for diagnosing the persisted
// draft graph without fetching workflow entities client-side.
func (s *WorkflowsService) PrepareDiagnose(workflowID string, rePropagate ...bool) PreparedRequest {
	shouldRePropagate := true
	if len(rePropagate) > 0 {
		shouldRePropagate = rePropagate[0]
	}
	return PreparedRequest{
		URL:    "/workflows/" + url.PathEscape(workflowID) + "/diagnose-graph",
		Method: http.MethodPost,
		Body: DiagnoseWorkflowRequest{
			RePropagate: shouldRePropagate,
		},
	}
}

// PrepareDiagnoseGraph returns the request descriptor for diagnosing an
// arbitrary graph without executing it.
func (s *WorkflowsService) PrepareDiagnoseGraph(workflowID string, blocks []map[string]any, edges []map[string]any, rePropagate ...bool) PreparedRequest {
	shouldRePropagate := true
	if len(rePropagate) > 0 {
		shouldRePropagate = rePropagate[0]
	}
	return prepareDiagnoseGraphRequest(workflowID, DiagnoseWorkflowGraphRequest{
		Blocks:      blocks,
		Edges:       edges,
		RePropagate: shouldRePropagate,
	})
}

// Diagnose runs the structural diagnosis on the persisted draft graph.
// It POSTs directly to diagnose-graph and lets the API inspect the saved draft.
//
// To diagnose an in-memory graph that hasn't been saved, call DiagnoseGraph
// directly with your own blocks/edges payload.
func (s *WorkflowsService) Diagnose(ctx context.Context, workflowID string, opts ...RequestOption) (*WorkflowDiagnosisResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result WorkflowDiagnosisResponse
	prepared := s.PrepareDiagnose(workflowID)
	err := s.client.do(ctx, prepared.Method, prepared.URL, nil, prepared.Body, &result, opts...)
	return &result, err
}

// DiagnoseGraph posts an arbitrary graph to the diagnose-graph endpoint.
// Use this when the graph is in-memory and not yet persisted.
func (s *WorkflowsService) DiagnoseGraph(ctx context.Context, workflowID string, request DiagnoseWorkflowGraphRequest, opts ...RequestOption) (*WorkflowDiagnosisResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result WorkflowDiagnosisResponse
	prepared := prepareDiagnoseGraphRequest(workflowID, request)
	err := s.client.do(ctx, prepared.Method, prepared.URL, nil, prepared.Body, &result, opts...)
	return &result, err
}

type WorkflowSpecsService struct {
	client *Client
}

type WorkflowSpecRequest struct {
	YAMLDefinition string `json:"yaml_definition"`
}

func (s *WorkflowSpecsService) Validate(ctx context.Context, yamlDefinition string, opts ...RequestOption) (*Resource, error) {
	var result Resource
	err := s.client.do(ctx, http.MethodPost, "/workflows/spec/validate", nil, WorkflowSpecRequest{YAMLDefinition: yamlDefinition}, &result, opts...)
	return &result, err
}

func (s *WorkflowSpecsService) Plan(ctx context.Context, yamlDefinition string, opts ...RequestOption) (*Resource, error) {
	var result Resource
	err := s.client.do(ctx, http.MethodPost, "/workflows/spec/plan", nil, WorkflowSpecRequest{YAMLDefinition: yamlDefinition}, &result, opts...)
	return &result, err
}

func (s *WorkflowSpecsService) Apply(ctx context.Context, yamlDefinition string, opts ...RequestOption) (*Resource, error) {
	var result Resource
	err := s.client.do(ctx, http.MethodPost, "/workflows/spec/apply", nil, WorkflowSpecRequest{YAMLDefinition: yamlDefinition}, &result, opts...)
	return &result, err
}

func (s *WorkflowSpecsService) Export(ctx context.Context, workflowID string, opts ...RequestOption) (*Resource, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result Resource
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/spec", nil, nil, &result, opts...)
	return &result, err
}

type WorkflowArtifact = Resource

type WorkflowArtifactsService struct {
	client *Client
}

type ListWorkflowArtifactsParams struct {
	RunID     string
	Operation string
	BlockID   string
	// Cursor pagination by producing step_id. Mutually exclusive — the
	// server returns 400 when both are set.
	Before string
	After  string
	// Page size (1-200). 0 means "let the server pick the default" (100).
	Limit int
}

func (s *WorkflowArtifactsService) PrepareList(params ListWorkflowArtifactsParams) PreparedRequest {
	query := url.Values{}
	addQuery(query, "run_id", params.RunID)
	addQuery(query, "operation", params.Operation)
	addQuery(query, "block_id", params.BlockID)
	addQuery(query, "before", params.Before)
	addQuery(query, "after", params.After)
	if params.Limit > 0 {
		addQuery(query, "limit", strconv.Itoa(params.Limit))
	}
	return PreparedRequest{
		URL:    "/workflows/artifacts",
		Method: http.MethodGet,
		Params: query,
	}
}

func (s *WorkflowArtifactsService) PrepareGet(artifactID string) PreparedRequest {
	return PreparedRequest{
		URL:    "/workflows/artifacts/" + url.PathEscape(artifactID),
		Method: http.MethodGet,
	}
}

// List returns workflow artifacts produced by a single run.
//
// Response-shape tolerance: production currently returns a bare JSON
// array (`[...]`) for this endpoint, while every other list endpoint on
// the API returns the standard paginated envelope (`{"data": [...],
// "list_metadata": {...}}`). The SDK accepts BOTH shapes and always
// hands callers a `*PaginatedList[WorkflowArtifact]`, so the call site
// never has to care which response the server emits on a given deploy.
// When the server is normalised to the envelope shape later, nothing
// here or in any caller needs to change.
//
// Cursor pagination is not implemented for this endpoint regardless of
// response shape; ListMetadata is {nil, nil} when the server returns a
// bare array.
func (s *WorkflowArtifactsService) List(ctx context.Context, params ListWorkflowArtifactsParams, opts ...RequestOption) (*PaginatedList[WorkflowArtifact], error) {
	if params.RunID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var raw json.RawMessage
	prepared := s.PrepareList(params)
	if err := s.client.do(ctx, prepared.Method, prepared.URL, prepared.Params, nil, &raw, opts...); err != nil {
		return nil, err
	}
	result, err := decodeArtifactListResponse(raw)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns the single artifact identified by `artifactID`. The server
// dispatches to the right backing collection based on the id prefix (e.g.
// `extr_…` → extraction, `clss_…` → classification); an unknown prefix
// surfaces as a 404 from the API.
func (s *WorkflowArtifactsService) Get(ctx context.Context, artifactID string, opts ...RequestOption) (*WorkflowArtifact, error) {
	if artifactID == "" {
		return nil, fmt.Errorf("retab: artifactID is required")
	}
	var result WorkflowArtifact
	prepared := s.PrepareGet(artifactID)
	err := s.client.do(ctx, prepared.Method, prepared.URL, nil, nil, &result, opts...)
	return &result, err
}

// decodeArtifactListResponse normalises both response shapes into the
// PaginatedList envelope. It sniffs the first non-whitespace byte: `[`
// is a bare array, `{` is the envelope. Anything else is a protocol
// violation surfaced as a typed error rather than a silent empty list.
func decodeArtifactListResponse(raw json.RawMessage) (PaginatedList[WorkflowArtifact], error) {
	var result PaginatedList[WorkflowArtifact]
	trimmed := bytes.TrimLeft(raw, " \t\r\n")
	if len(trimmed) == 0 {
		return result, nil
	}
	switch trimmed[0] {
	case '[':
		if err := json.Unmarshal(raw, &result.Data); err != nil {
			return result, fmt.Errorf("retab: decode artifact list (array shape): %w", err)
		}
	case '{':
		if err := json.Unmarshal(raw, &result); err != nil {
			return result, fmt.Errorf("retab: decode artifact list (envelope shape): %w", err)
		}
	default:
		return result, fmt.Errorf("retab: unexpected artifact list response shape (first byte %q): %s",
			trimmed[0], string(raw))
	}
	return result, nil
}

type WorkflowBlocksService struct {
	client     *Client
	Executions *WorkflowBlockExecutionsService
}

type WorkflowBlockExecutionsService struct {
	client *Client
}

type CreateWorkflowBlockExecutionRequest struct {
	RunID            string `json:"run_id"`
	BlockID          string `json:"block_id"`
	StepID           string `json:"step_id,omitempty"`
	NConsensus       int    `json:"n_consensus,omitempty"`
	CheckEligibility *bool  `json:"check_eligibility,omitempty"`
}

type ListWorkflowBlockExecutionsParams struct {
	RunID   string
	BlockID string
	Limit   int
}

func (s *WorkflowBlockExecutionsService) Create(ctx context.Context, request CreateWorkflowBlockExecutionRequest, opts ...RequestOption) (*StoredBlockExecution, error) {
	if request.RunID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if request.BlockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var result StoredBlockExecution
	err := s.client.do(ctx, http.MethodPost, "/workflows/blocks/executions", nil, request, &result, opts...)
	return &result, err
}

func (s *WorkflowBlockExecutionsService) List(ctx context.Context, params ListWorkflowBlockExecutionsParams, opts ...RequestOption) (*PaginatedList[StoredBlockExecution], error) {
	if params.RunID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if params.BlockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	query := url.Values{}
	query.Set("run_id", params.RunID)
	query.Set("block_id", params.BlockID)
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	var result PaginatedList[StoredBlockExecution]
	err := s.client.do(ctx, http.MethodGet, "/workflows/blocks/executions", query, nil, &result, opts...)
	return &result, err
}

type WorkflowBlockCreateRequest struct {
	// ID is optional — the server generates an opaque `blk_<nanoid>` when
	// the field is absent. omitempty matters because block ids are
	// org-globally unique; sending "" trips the uniqueness check against any
	// other empty-id block (or fails when an existing one is found) instead
	// of triggering the server's default_factory.
	ID        string         `json:"id,omitempty"`
	Type      string         `json:"type"`
	Label     string         `json:"label,omitempty"`
	PositionX float64        `json:"position_x,omitempty"`
	PositionY float64        `json:"position_y,omitempty"`
	Width     *float64       `json:"width,omitempty"`
	Height    *float64       `json:"height,omitempty"`
	Config    map[string]any `json:"config,omitempty"`
	ParentID  string         `json:"parent_id,omitempty"`
}

type UpdateWorkflowBlockRequest struct {
	Label     *string        `json:"label,omitempty"`
	PositionX *float64       `json:"position_x,omitempty"`
	PositionY *float64       `json:"position_y,omitempty"`
	Width     *float64       `json:"width,omitempty"`
	Height    *float64       `json:"height,omitempty"`
	Config    map[string]any `json:"config,omitempty"`
	ParentID  *string        `json:"parent_id,omitempty"`
	// ConfigMode controls how the server folds Config into the existing
	// block config when both are sent. "merge" (default) does an RFC-7396
	// JSON Merge Patch — dicts recurse, arrays/scalars replace, null deletes
	// the key. "replace" treats Config as the full new typed config (with
	// top-level nulls pruned). Leave empty to keep the server default.
	//
	// CLI mapping: --merge-config-file sets "merge", --config-file sets
	// "replace". Pre-config_mode servers ignore the field and keep the
	// legacy shallow-merge behavior.
	ConfigMode string `json:"config_mode,omitempty"`
}

type ListWorkflowBlocksParams struct {
	// Cursor pagination by block id. Mutually exclusive — the server
	// returns 400 when both are set.
	Before string
	After  string
	// Page size (1-200). 0 means "let the server pick the default" (100).
	Limit int
}

// List returns the canonical paginated envelope
// {"data": [...], "list_metadata": {...}} for the blocks of a workflow's
// current draft. Pass “params“ to drive cursor pagination — “After“
// for the next page, “Before“ for the previous; the two are mutually
// exclusive.
func (s *WorkflowBlocksService) List(ctx context.Context, workflowID string, params *ListWorkflowBlocksParams, opts ...RequestOption) (*PaginatedList[WorkflowBlock], error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	query := url.Values{}
	if params != nil {
		addQuery(query, "before", params.Before)
		addQuery(query, "after", params.After)
		if params.Limit > 0 {
			addQuery(query, "limit", strconv.Itoa(params.Limit))
		}
	}
	// The /workflows/<id>/blocks endpoint historically returned a bare
	// JSON array; canonical responses use the paginated envelope. Sniff
	// the first non-whitespace byte and dispatch accordingly so callers
	// always receive a wrapped *PaginatedList, matching the rest of the
	// SDK. Drop the bare-array branch once the server is consistent.
	var raw json.RawMessage
	if err := s.client.do(ctx, http.MethodGet, "/workflows/blocks?workflow_id="+url.QueryEscape(workflowID), query, nil, &raw, opts...); err != nil {
		return nil, err
	}
	trimmed := bytes.TrimLeft(raw, " \t\r\n")
	var result PaginatedList[WorkflowBlock]
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var arr []WorkflowBlock
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, fmt.Errorf("retab: decode workflow blocks list (bare array): %w", err)
		}
		result.Data = arr
	} else {
		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, fmt.Errorf("retab: decode workflow blocks list (envelope): %w", err)
		}
	}
	return &result, nil
}

func (s *WorkflowBlocksService) Get(ctx context.Context, blockID string, opts ...RequestOption) (*WorkflowBlock, error) {
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var block WorkflowBlock
	err := s.client.do(ctx, http.MethodGet, "/workflows/blocks/"+url.PathEscape(blockID), nil, nil, &block, opts...)
	return &block, err
}

func (s *WorkflowBlocksService) Create(ctx context.Context, workflowID string, request WorkflowBlockCreateRequest, opts ...RequestOption) (*WorkflowBlock, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := resourceFromJSON(request)
	body["workflow_id"] = workflowID
	var block WorkflowBlock
	err := s.client.do(ctx, http.MethodPost, "/workflows/blocks", nil, body, &block, opts...)
	return &block, err
}

func (s *WorkflowBlocksService) Update(ctx context.Context, blockID string, request UpdateWorkflowBlockRequest, opts ...RequestOption) (*WorkflowBlock, error) {
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var block WorkflowBlock
	err := s.client.do(ctx, http.MethodPatch, "/workflows/blocks/"+url.PathEscape(blockID), nil, request, &block, opts...)
	return &block, err
}

func (s *WorkflowBlocksService) Delete(ctx context.Context, blockID string, opts ...RequestOption) error {
	if blockID == "" {
		return fmt.Errorf("retab: blockID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/blocks/"+url.PathEscape(blockID), nil, nil, nil, opts...)
}

type WorkflowEdgesService struct {
	client *Client
}

type WorkflowEdgeCreateRequest struct {
	ID           string `json:"id,omitempty"`
	SourceBlock  string `json:"source_block"`
	TargetBlock  string `json:"target_block"`
	SourceHandle string `json:"source_handle,omitempty"`
	TargetHandle string `json:"target_handle,omitempty"`
}

type ListWorkflowEdgesParams struct {
	SourceBlock string
	TargetBlock string
	// Cursor pagination by edge id. Mutually exclusive — the server
	// returns 400 when both are set.
	Before string
	After  string
	// Page size (1-200). 0 means "let the server pick the default" (100).
	Limit int
}

// List returns the canonical paginated envelope
// {"data": [...], "list_metadata": {...}} for edges of a workflow's
// current draft. Pass “params“ to drive cursor pagination — “After“
// for the next page, “Before“ for the previous; the two are mutually
// exclusive.
func (s *WorkflowEdgesService) List(ctx context.Context, workflowID string, params *ListWorkflowEdgesParams, opts ...RequestOption) (*PaginatedList[WorkflowEdgeDoc], error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	query := url.Values{}
	if params != nil {
		addQuery(query, "source_block", params.SourceBlock)
		addQuery(query, "target_block", params.TargetBlock)
		addQuery(query, "before", params.Before)
		addQuery(query, "after", params.After)
		if params.Limit > 0 {
			addQuery(query, "limit", strconv.Itoa(params.Limit))
		}
	}
	var result PaginatedList[WorkflowEdgeDoc]
	err := s.client.do(ctx, http.MethodGet, "/workflows/edges?workflow_id="+url.QueryEscape(workflowID), query, nil, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *WorkflowEdgesService) Get(ctx context.Context, edgeID string, opts ...RequestOption) (*WorkflowEdgeDoc, error) {
	if edgeID == "" {
		return nil, fmt.Errorf("retab: edgeID is required")
	}
	var edge WorkflowEdgeDoc
	err := s.client.do(ctx, http.MethodGet, "/workflows/edges/"+url.PathEscape(edgeID), nil, nil, &edge, opts...)
	return &edge, err
}

func (s *WorkflowEdgesService) Create(ctx context.Context, workflowID string, request WorkflowEdgeCreateRequest, opts ...RequestOption) (*WorkflowEdgeDoc, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := resourceFromJSON(request)
	body["workflow_id"] = workflowID
	var edge WorkflowEdgeDoc
	err := s.client.do(ctx, http.MethodPost, "/workflows/edges", nil, body, &edge, opts...)
	return &edge, err
}

func (s *WorkflowEdgesService) Delete(ctx context.Context, edgeID string, opts ...RequestOption) error {
	if edgeID == "" {
		return fmt.Errorf("retab: edgeID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/edges/"+url.PathEscape(edgeID), nil, nil, nil, opts...)
}

type WorkflowRunsService struct {
	client *Client
}

type CreateWorkflowRunRequest struct {
	WorkflowID string         `json:"workflow_id"`
	Documents  map[string]any `json:"documents,omitempty"`
	JSONInputs map[string]any `json:"json_inputs,omitempty"`
	Version    string         `json:"version,omitempty"`
}

func (s *WorkflowRunsService) Create(ctx context.Context, request CreateWorkflowRunRequest, opts ...RequestOption) (*WorkflowRun, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := map[string]any{"workflow_id": request.WorkflowID}
	if request.Documents != nil {
		documents, err := workflowRunDocumentsPayload(request.Documents)
		if err != nil {
			return nil, err
		}
		body["documents"] = documents
	}
	if request.JSONInputs != nil {
		body["json_inputs"] = request.JSONInputs
	}
	if request.Version == "" {
		body["version"] = "production"
	} else {
		body["version"] = request.Version
	}
	var run WorkflowRun
	err := s.client.do(ctx, http.MethodPost, "/workflows/runs", nil, body, &run, opts...)
	return &run, err
}

func workflowRunDocumentsPayload(documents map[string]any) (map[string]map[string]string, error) {
	payload := map[string]map[string]string{}
	for blockID, document := range documents {
		if descriptor, ok, err := workflowRunDocumentDescriptor(blockID, document); ok || err != nil {
			if err != nil {
				return nil, err
			}
			payload[blockID] = descriptor
			continue
		}
		mimeData, err := InferMIMEData(document)
		if err != nil {
			return nil, err
		}
		if isWorkflowRunURLBackedDocument(mimeData) {
			payload[blockID] = map[string]string{
				"filename": mimeData.Filename,
				"url":      mimeData.URL,
			}
			if mimeData.MIMEType != "" {
				payload[blockID]["mime_type"] = mimeData.MIMEType
			}
			continue
		}
		content := mimeData.Content
		mimeType := mimeData.MIMEType
		if content == "" || mimeType == "" {
			_, inferredType, err := dataURIBytes(mimeData)
			if err != nil {
				return nil, fmt.Errorf("retab: workflow run document %s must be a local file, data URI, base64 input, bytes, reader, or inline content payload: %w", blockID, err)
			}
			parts := strings.SplitN(mimeData.URL, ",", 2)
			if len(parts) == 2 && content == "" {
				content = parts[1]
			}
			if mimeType == "" {
				mimeType = inferredType
			}
		}
		payload[blockID] = map[string]string{
			"filename":  mimeData.Filename,
			"content":   content,
			"mime_type": mimeType,
		}
	}
	return payload, nil
}

func workflowRunDocumentDescriptor(blockID string, document any) (map[string]string, bool, error) {
	switch value := document.(type) {
	case map[string]string:
		return normalizeWorkflowRunDocumentDescriptor(blockID, value), true, nil
	case map[string]any:
		descriptor := map[string]string{}
		for _, key := range []string{"filename", "url", "content", "mime_type"} {
			raw, ok := value[key]
			if !ok {
				continue
			}
			text, ok := raw.(string)
			if !ok {
				return nil, true, fmt.Errorf("retab: workflow run document %s field %q must be a string", blockID, key)
			}
			if text != "" {
				descriptor[key] = text
			}
		}
		return normalizeWorkflowRunDocumentDescriptor(blockID, descriptor), true, nil
	default:
		return nil, false, nil
	}
}

func normalizeWorkflowRunDocumentDescriptor(blockID string, descriptor map[string]string) map[string]string {
	if descriptor["filename"] == "" {
		descriptor["filename"] = blockID
	}
	return descriptor
}

func isWorkflowRunURLBackedDocument(mimeData MIMEData) bool {
	return strings.HasPrefix(mimeData.URL, "https://") || strings.HasPrefix(mimeData.URL, "http://")
}

func (s *WorkflowRunsService) Get(ctx context.Context, runID string, opts ...RequestOption) (*WorkflowRun, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var run WorkflowRun
	err := s.client.do(ctx, http.MethodGet, "/workflows/runs/"+url.PathEscape(runID), nil, nil, &run, opts...)
	return &run, err
}

type ListWorkflowRunsParams struct {
	WorkflowID    string
	Status        string
	Statuses      []string
	ExcludeStatus string
	TriggerType   string
	TriggerTypes  []string
	FromDate      string
	ToDate        string
	MinCost       *float64
	MaxCost       *float64
	MinDuration   *int
	MaxDuration   *int
	Search        string
	SortBy        string
	Fields        []string
	Before        string
	After         string
	Limit         int
	Order         string
}

func (s *WorkflowRunsService) List(ctx context.Context, params *ListWorkflowRunsParams, opts ...RequestOption) (*PaginatedList[WorkflowRun], error) {
	query := url.Values{}
	query.Set("limit", "20")
	query.Set("order", "desc")
	if params != nil {
		addQuery(query, "workflow_id", params.WorkflowID)
		addQuery(query, "status", params.Status)
		addCSVQuery(query, "statuses", params.Statuses)
		addQuery(query, "exclude_status", params.ExcludeStatus)
		addQuery(query, "trigger_type", params.TriggerType)
		addCSVQuery(query, "trigger_types", params.TriggerTypes)
		addQuery(query, "from_date", params.FromDate)
		addQuery(query, "to_date", params.ToDate)
		addQuery(query, "search", params.Search)
		addQuery(query, "sort_by", params.SortBy)
		addCSVQuery(query, "fields", params.Fields)
		addQuery(query, "before", params.Before)
		addQuery(query, "after", params.After)
		addQuery(query, "order", params.Order)
		if params.MinCost != nil {
			query.Set("min_cost", fmt.Sprintf("%g", *params.MinCost))
		}
		if params.MaxCost != nil {
			query.Set("max_cost", fmt.Sprintf("%g", *params.MaxCost))
		}
		if params.MinDuration != nil {
			query.Set("min_duration", fmt.Sprintf("%d", *params.MinDuration))
		}
		if params.MaxDuration != nil {
			query.Set("max_duration", fmt.Sprintf("%d", *params.MaxDuration))
		}
		if params.Limit > 0 {
			query.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
	}
	var result PaginatedList[WorkflowRun]
	err := s.client.do(ctx, http.MethodGet, "/workflows/runs", query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowRunsService) Delete(ctx context.Context, runID string, opts ...RequestOption) error {
	if runID == "" {
		return fmt.Errorf("retab: runID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/runs/"+url.PathEscape(runID), nil, nil, nil, opts...)
}

type WorkflowRunCommandRequest struct {
	CommandID    string `json:"command_id,omitempty"`
	ConfigSource string `json:"config_source,omitempty"`
}

func (s *WorkflowRunsService) Cancel(ctx context.Context, runID string, request WorkflowRunCommandRequest, opts ...RequestOption) (*CancelWorkflowResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	body := map[string]any{}
	if request.CommandID != "" {
		body["command_id"] = request.CommandID
	}
	var result CancelWorkflowResponse
	err := s.client.do(ctx, http.MethodPost, "/workflows/runs/"+url.PathEscape(runID)+"/cancel", nil, body, &result, opts...)
	return &result, err
}

func (s *WorkflowRunsService) Restart(ctx context.Context, runID string, request WorkflowRunCommandRequest, opts ...RequestOption) (*WorkflowRun, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	configSource := request.ConfigSource
	if configSource == "" {
		configSource = "published"
	}
	body := map[string]any{"restart_of": runID, "config_source": configSource}
	if request.CommandID != "" {
		body["command_id"] = request.CommandID
	}
	var run WorkflowRun
	err := s.client.do(ctx, http.MethodPost, "/workflows/runs", nil, body, &run, opts...)
	return &run, err
}

type ExportWorkflowRunsRequest struct {
	WorkflowID       string
	BlockID          string
	ExportSource     string
	SelectedRunIDs   []string
	Status           string
	ExcludeStatus    string
	FromDate         string
	ToDate           string
	TriggerTypes     []string
	PreferredColumns []string
	// CSV shape controls. All three default to the server's defaults
	// (``;`` / ``\n`` / ``"``) when left empty. ``Delimiter`` and
	// ``Quote`` must each be a single character if non-empty;
	// ``LineDelimiter`` must be non-empty if set.
	//
	// The server default ``;`` is convenient in EU locales (Excel uses
	// ``;`` when ``,`` is the decimal separator) but is hostile to
	// pandas / RFC-4180 consumers. Set ``Delimiter`` to "," for the
	// portable shape.
	Delimiter     string
	LineDelimiter string
	Quote         string
}

func (s *WorkflowRunsService) Export(ctx context.Context, request ExportWorkflowRunsRequest, opts ...RequestOption) (*WorkflowRunExportResponse, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if request.BlockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	exportSource := request.ExportSource
	if exportSource == "" {
		exportSource = "outputs"
	}
	body := map[string]any{
		"workflow_id":       request.WorkflowID,
		"block_id":          request.BlockID,
		"export_source":     exportSource,
		"preferred_columns": request.PreferredColumns,
	}
	if body["preferred_columns"] == nil {
		body["preferred_columns"] = []string{}
	}
	if len(request.SelectedRunIDs) > 0 {
		body["selected_run_ids"] = request.SelectedRunIDs
	}
	if request.Status != "" {
		body["status"] = request.Status
	}
	if request.ExcludeStatus != "" {
		body["exclude_status"] = request.ExcludeStatus
	}
	if request.FromDate != "" {
		body["from_date"] = request.FromDate
	}
	if request.ToDate != "" {
		body["to_date"] = request.ToDate
	}
	if len(request.TriggerTypes) > 0 {
		body["trigger_types"] = request.TriggerTypes
	}
	// CSV shape controls — forwarded only when explicitly set so the
	// server falls back to its own defaults (and so pre-this-field
	// servers don't reject an unexpected field; the route accepts
	// extras silently regardless, but omitting keeps the wire small).
	if request.Delimiter != "" {
		body["delimiter"] = request.Delimiter
	}
	if request.LineDelimiter != "" {
		body["line_delimiter"] = request.LineDelimiter
	}
	if request.Quote != "" {
		body["quote"] = request.Quote
	}
	var result WorkflowRunExportResponse
	err := s.client.do(ctx, http.MethodPost, "/workflows/runs/export-payload", nil, body, &result, opts...)
	return &result, err
}

type WorkflowStepsService struct {
	client *Client
}

type ListWorkflowStepsParams struct {
	// Cursor pagination by step_id. Mutually exclusive — the server
	// returns 400 when both are set.
	Before string
	After  string
	// Page size (1-1000). 0 means "let the server pick the default" (200).
	Limit int
}

// List returns the canonical paginated envelope
// {"data": [...], "list_metadata": {...}} for the persisted step documents
// of one workflow run. Pass “params“ to drive cursor pagination —
// “After“ for the next page, “Before“ for the previous; the two are
// mutually exclusive.
func (s *WorkflowStepsService) List(ctx context.Context, runID string, params *ListWorkflowStepsParams, opts ...RequestOption) (*PaginatedList[WorkflowRunStep], error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	query := url.Values{}
	if params != nil {
		addQuery(query, "before", params.Before)
		addQuery(query, "after", params.After)
		if params.Limit > 0 {
			addQuery(query, "limit", strconv.Itoa(params.Limit))
		}
	}
	var result PaginatedList[WorkflowRunStep]
	err := s.client.do(ctx, http.MethodGet, "/workflows/steps?run_id="+url.QueryEscape(runID), query, nil, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *WorkflowStepsService) Get(ctx context.Context, stepID string, opts ...RequestOption) (*WorkflowRunStep, error) {
	if stepID == "" {
		return nil, fmt.Errorf("retab: stepID is required")
	}
	var step WorkflowRunStep
	err := s.client.do(ctx, http.MethodGet, "/workflows/steps/"+url.PathEscape(stepID), nil, nil, &step, opts...)
	return &step, err
}

// WorkflowReviewsService drives the actor-neutral review surface served under
// /workflows/reviews.
type WorkflowReviewsService struct {
	client   *Client
	Versions *WorkflowReviewVersionsService
}

// WorkflowReviewVersionsService manages immutable review version resources
// served under /workflows/reviews/versions.
type WorkflowReviewVersionsService struct {
	client *Client
}

// ListReviewsParams filters the review queue.
//
// DecisionStatus selects which slice of the queue to return: "pending" (the
// default), "approved", "rejected", "decided", or "all".
//
// Before and After are opaque cursor strings copied verbatim from a previous
// response's list_metadata. Setting both at the same time is a client error.
type ListReviewsParams struct {
	WorkflowID     string // restrict to one workflow
	RunID          string // restrict to one workflow run
	BlockID        string // restrict to one workflow block
	StepID         string // restrict to one execution step
	IterationKey   string // restrict to one for_each iteration key
	Limit          int    // page size, 1-200
	DecisionStatus string // pending | approved | rejected | decided | all
	Before         string // cursor for the previous page (from list_metadata.before)
	After          string // cursor for the next page (from list_metadata.after)
}

// List returns a page of the review queue — block runs awaiting review,
// oldest-created first. The response is the standard cursor envelope: inspect
// result.ListMetadata.After to detect truncation and pass it back as the
// After param on the next call to fetch the following page.
func (s *WorkflowReviewsService) List(ctx context.Context, params *ListReviewsParams, opts ...RequestOption) (*PaginatedList[Review], error) {
	query := url.Values{}
	query.Set("decision_status", "pending")
	if params != nil {
		if params.Before != "" && params.After != "" {
			return nil, fmt.Errorf("retab: Before and After are mutually exclusive")
		}
		addQuery(query, "workflow_id", params.WorkflowID)
		addQuery(query, "run_id", params.RunID)
		addQuery(query, "block_id", params.BlockID)
		addQuery(query, "step_id", params.StepID)
		addQuery(query, "iteration_key", params.IterationKey)
		addQuery(query, "decision_status", params.DecisionStatus)
		addQuery(query, "before", params.Before)
		addQuery(query, "after", params.After)
		if params.Limit > 0 {
			query.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
	}
	var result PaginatedList[Review]
	err := s.client.do(ctx, http.MethodGet, "/workflows/reviews", query, nil, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns the full review by id.
func (s *WorkflowReviewsService) Get(ctx context.Context, reviewID string, opts ...RequestOption) (*Review, error) {
	if reviewID == "" {
		return nil, fmt.Errorf("retab: reviewID is required")
	}
	var review Review
	err := s.client.do(ctx, http.MethodGet, reviewPath(reviewID), nil, nil, &review, opts...)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

type ListReviewVersionsParams struct {
	ReviewID string
	Limit    int
	Before   string
	After    string
}

// List returns immutable versions for one review. ReviewID is required.
func (s *WorkflowReviewVersionsService) List(ctx context.Context, params *ListReviewVersionsParams, opts ...RequestOption) (*PaginatedList[ReviewVersion], error) {
	if params == nil || params.ReviewID == "" {
		return nil, fmt.Errorf("retab: ReviewID is required")
	}
	if params.Before != "" && params.After != "" {
		return nil, fmt.Errorf("retab: Before and After are mutually exclusive")
	}
	query := url.Values{}
	query.Set("review_id", params.ReviewID)
	addQuery(query, "before", params.Before)
	addQuery(query, "after", params.After)
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	var result PaginatedList[ReviewVersion]
	err := s.client.do(ctx, http.MethodGet, "/workflows/reviews/versions", query, nil, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns one immutable review version by id.
func (s *WorkflowReviewVersionsService) Get(ctx context.Context, versionID string, opts ...RequestOption) (*ReviewVersion, error) {
	if versionID == "" {
		return nil, fmt.Errorf("retab: versionID is required")
	}
	var version ReviewVersion
	err := s.client.do(ctx, http.MethodGet, "/workflows/reviews/versions/"+url.PathEscape(versionID), nil, nil, &version, opts...)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// CreateReviewVersionRequest creates a corrected review version.
type CreateReviewVersionRequest struct {
	ReviewID string
	ParentID string
	Snapshot map[string]any
	Note     string
}

// Create creates an immutable corrected output version for a review.
func (s *WorkflowReviewVersionsService) Create(ctx context.Context, request CreateReviewVersionRequest, opts ...RequestOption) (*ReviewVersion, error) {
	if request.ReviewID == "" {
		return nil, fmt.Errorf("retab: ReviewID is required")
	}
	if request.ParentID == "" {
		return nil, fmt.Errorf("retab: ParentID is required")
	}
	if request.Snapshot == nil {
		return nil, fmt.Errorf("retab: Snapshot is required")
	}
	body := map[string]any{
		"review_id": request.ReviewID,
		"parent_id": request.ParentID,
		"snapshot":  request.Snapshot,
	}
	if request.Note != "" {
		body["note"] = request.Note
	}
	var version ReviewVersion
	err := s.client.do(ctx, http.MethodPost, "/workflows/reviews/versions", nil, body, &version, opts...)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// ApproveReviewRequest approves one exact reviewed output version.
type ApproveReviewRequest struct {
	VersionID string // content-hash id of the version being approved
}

// Approve approves a reviewed block output.
//
// VersionID can be any id returned by Reviews.Versions.List for this review.
// To approve a correction, first call Reviews.Versions.Create to author it,
// then pass the returned version id here.
//
// Inspect Response.ResumeStatus to confirm the workflow actually resumed
// downstream — Response.SubmissionStatus only reflects the decision write.
func (s *WorkflowReviewsService) Approve(ctx context.Context, reviewID string, request ApproveReviewRequest, opts ...RequestOption) (*SubmitWorkflowReviewDecisionResponse, error) {
	if request.VersionID == "" {
		return nil, fmt.Errorf("retab: VersionID is required")
	}
	return s.submitDecision(ctx, reviewID, "approve", map[string]any{"version_id": request.VersionID}, opts...)
}

// RejectReviewRequest rejects the reviewed output. The runtime records a
// terminal rejected decision, so downstream blocks do not continue.
type RejectReviewRequest struct {
	VersionID string // content-hash id of the version being rejected
	Reason    string // required
}

// Reject rejects a reviewed block output. Reason is required and cancels the
// workflow run on the server.
//
// Inspect Response.ResumeStatus to confirm the workflow actually cancelled
// downstream — Response.SubmissionStatus only reflects the decision write.
func (s *WorkflowReviewsService) Reject(ctx context.Context, reviewID string, request RejectReviewRequest, opts ...RequestOption) (*SubmitWorkflowReviewDecisionResponse, error) {
	if request.VersionID == "" {
		return nil, fmt.Errorf("retab: VersionID is required")
	}
	if request.Reason == "" {
		return nil, fmt.Errorf("retab: Reason is required for Reject")
	}
	body := map[string]any{"version_id": request.VersionID, "reason": request.Reason}
	return s.submitDecision(ctx, reviewID, "reject", body, opts...)
}

func (s *WorkflowReviewsService) submitDecision(ctx context.Context, reviewID string, action string, body map[string]any, opts ...RequestOption) (*SubmitWorkflowReviewDecisionResponse, error) {
	if reviewID == "" {
		return nil, fmt.Errorf("retab: reviewID is required")
	}
	var result SubmitWorkflowReviewDecisionResponse
	err := s.client.do(ctx, http.MethodPost, reviewPath(reviewID)+"/"+action, nil, body, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func reviewPath(reviewID string) string {
	return "/workflows/reviews/" + url.PathEscape(reviewID)
}

type WorkflowTestsService struct {
	client *Client
	Runs   *WorkflowTestRunsService
}

type WorkflowTestRunResultsService struct {
	client *Client
}

func newWorkflowTestsService(client *Client) *WorkflowTestsService {
	service := &WorkflowTestsService{client: client}
	service.Runs = &WorkflowTestRunsService{client: client}
	service.Runs.Results = &WorkflowTestRunResultsService{client: client}
	return service
}

type WorkflowTest = Resource
type WorkflowTestResult = Resource

type WorkflowTestCreateRequest struct {
	WorkflowID string   `json:"-"`
	Target     Resource `json:"target"`
	Source     Resource `json:"source"`
	Assertion  Resource `json:"assertion"`
	Name       string   `json:"name,omitempty"`
}

type UpdateWorkflowTestRequest struct {
	Name      string   `json:"name,omitempty"`
	Assertion Resource `json:"assertion,omitempty"`
	Source    Resource `json:"source,omitempty"`
}

type ListWorkflowTestsRequest struct {
	WorkflowID    string
	TargetBlockID string
	Limit         int
}

type ListWorkflowTestRunsParams struct {
	WorkflowID    string
	TestID        string
	TargetBlockID string
	Status        string
	Statuses      []string
	ExcludeStatus string
	TriggerType   string
	TriggerTypes  []string
	FromDate      string
	ToDate        string
	SortBy        string
	Fields        []string
	Before        string
	After         string
	Limit         int
	Order         string
}

type ListWorkflowTestRunsRequest = ListWorkflowTestRunsParams

// WorkflowTestListResponse is the canonical PaginatedList envelope for
// `GET /v1/workflows/tests?workflow_id={wf}`. Kept as a type alias so callers can
// keep referring to the named type while the underlying shape rides the
// shared `{data, list_metadata}` envelope.
type WorkflowTestListResponse = PaginatedList[WorkflowTest]

type CreateWorkflowTestRunRequest struct {
	WorkflowID string   `json:"-"`
	TestID     string   `json:"test_id,omitempty"`
	Target     Resource `json:"target,omitempty"`
	NConsensus int      `json:"n_consensus,omitempty"`
}

type WorkflowTestRunCounts struct {
	Queued    int `json:"queued,omitempty"`
	Running   int `json:"running,omitempty"`
	Passed    int `json:"passed,omitempty"`
	Failed    int `json:"failed,omitempty"`
	Blocked   int `json:"blocked,omitempty"`
	Error     int `json:"error,omitempty"`
	Cancelled int `json:"cancelled,omitempty"`
}

type WorkflowTestRunLifecycle struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type WorkflowTestRunTiming struct {
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DurationMs  *int       `json:"duration_ms,omitempty"`
}

type WorkflowTestRun struct {
	ID         string                   `json:"id"`
	Workflow   WorkflowSnapshotRef      `json:"workflow"`
	Trigger    RunTrigger               `json:"trigger"`
	Lifecycle  WorkflowTestRunLifecycle `json:"lifecycle"`
	Timing     WorkflowTestRunTiming    `json:"timing"`
	Target     *Resource                `json:"target,omitempty"`
	TestID     string                   `json:"test_id,omitempty"`
	TotalTests int                      `json:"total_tests"`
	Counts     WorkflowTestRunCounts    `json:"counts,omitempty"`
}

type WorkflowTestRunResultListResponse = PaginatedList[WorkflowTestResult]

func (s *WorkflowTestsService) Create(ctx context.Context, request WorkflowTestCreateRequest, opts ...RequestOption) (*WorkflowTest, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := resourceFromJSON(request)
	delete(body, "WorkflowID")
	body["workflow_id"] = request.WorkflowID
	var result WorkflowTest
	err := s.client.do(ctx, http.MethodPost, "/workflows/tests", nil, body, &result, opts...)
	return &result, err
}

func (s *WorkflowTestsService) Get(ctx context.Context, testID string, opts ...RequestOption) (*WorkflowTest, error) {
	if testID == "" {
		return nil, fmt.Errorf("retab: testID is required")
	}
	var result WorkflowTest
	err := s.client.do(ctx, http.MethodGet, "/workflows/tests/"+url.PathEscape(testID), nil, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowTestsService) List(ctx context.Context, request ListWorkflowTestsRequest, opts ...RequestOption) (*PaginatedList[WorkflowTest], error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	query := url.Values{}
	limit := request.Limit
	if limit == 0 {
		limit = 50
	}
	query.Set("limit", fmt.Sprintf("%d", limit))
	addQuery(query, "target_block_id", request.TargetBlockID)
	var result PaginatedList[WorkflowTest]
	err := s.client.do(ctx, http.MethodGet, "/workflows/tests?workflow_id="+url.QueryEscape(request.WorkflowID), query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowTestsService) Update(ctx context.Context, testID string, request UpdateWorkflowTestRequest, opts ...RequestOption) (*WorkflowTest, error) {
	if testID == "" {
		return nil, fmt.Errorf("retab: testID is required")
	}
	var result WorkflowTest
	err := s.client.do(ctx, http.MethodPatch, "/workflows/tests/"+url.PathEscape(testID), nil, request, &result, opts...)
	return &result, err
}

func (s *WorkflowTestsService) Delete(ctx context.Context, testID string, opts ...RequestOption) error {
	if testID == "" {
		return fmt.Errorf("retab: testID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/tests/"+url.PathEscape(testID), nil, nil, nil, opts...)
}

func (s *WorkflowTestRunsService) Create(ctx context.Context, request CreateWorkflowTestRunRequest, opts ...RequestOption) (*WorkflowTestRun, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := resourceFromJSON(request)
	delete(body, "WorkflowID")
	body["workflow_id"] = request.WorkflowID
	var result WorkflowTestRun
	err := s.client.do(ctx, http.MethodPost, "/workflows/tests/runs", nil, body, &result, opts...)
	return &result, err
}

type WorkflowTestRunsService struct {
	client  *Client
	Results *WorkflowTestRunResultsService
}

type WorkflowTestRunListResponse = PaginatedList[WorkflowTestRun]

func (s *WorkflowTestRunsService) List(ctx context.Context, request ListWorkflowTestRunsParams, opts ...RequestOption) (*PaginatedList[WorkflowTestRun], error) {
	limit := request.Limit
	if limit == 0 {
		limit = 20
	}
	query := url.Values{}
	query.Set("limit", fmt.Sprintf("%d", limit))
	addQuery(query, "workflow_id", request.WorkflowID)
	addQuery(query, "test_id", request.TestID)
	addQuery(query, "target_block_id", request.TargetBlockID)
	addQuery(query, "status", request.Status)
	addCSVQuery(query, "statuses", request.Statuses)
	addQuery(query, "exclude_status", request.ExcludeStatus)
	addQuery(query, "trigger_type", request.TriggerType)
	addCSVQuery(query, "trigger_types", request.TriggerTypes)
	addQuery(query, "from_date", request.FromDate)
	addQuery(query, "to_date", request.ToDate)
	addQuery(query, "sort_by", request.SortBy)
	addCSVQuery(query, "fields", request.Fields)
	addQuery(query, "before", request.Before)
	addQuery(query, "after", request.After)
	addQuery(query, "order", request.Order)
	var result PaginatedList[WorkflowTestRun]
	err := s.client.do(ctx, http.MethodGet, "/workflows/tests/runs", query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowTestRunsService) Get(ctx context.Context, runID string, opts ...RequestOption) (*WorkflowTestRun, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result WorkflowTestRun
	err := s.client.do(ctx, http.MethodGet, "/workflows/tests/runs/"+url.PathEscape(runID), nil, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowTestRunsService) Cancel(ctx context.Context, runID string, opts ...RequestOption) (*WorkflowTestRun, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result WorkflowTestRun
	err := s.client.do(ctx, http.MethodPost, "/workflows/tests/runs/"+url.PathEscape(runID)+"/cancel", nil, map[string]any{}, &result, opts...)
	return &result, err
}

func (s *WorkflowTestRunResultsService) List(ctx context.Context, runID string, opts ...RequestOption) (*PaginatedList[WorkflowTestResult], error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	query := url.Values{}
	query.Set("run_id", runID)
	var result PaginatedList[WorkflowTestResult]
	err := s.client.do(ctx, http.MethodGet, "/workflows/tests/results", query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowTestRunResultsService) Get(ctx context.Context, resultID string, opts ...RequestOption) (*WorkflowTestResult, error) {
	if resultID == "" {
		return nil, fmt.Errorf("retab: resultID is required")
	}
	var result WorkflowTestResult
	err := s.client.do(ctx, http.MethodGet, "/workflows/tests/results/"+url.PathEscape(resultID), nil, nil, &result, opts...)
	return &result, err
}

func addQuery(values url.Values, key string, value string) {
	if value != "" {
		values.Set(key, value)
	}
}

func addCSVQuery(values url.Values, key string, value []string) {
	if len(value) > 0 {
		values.Set(key, strings.Join(value, ","))
	}
}

// ---------------------------------------------------------------------------
// Workflow experiments (consensus evaluation)
// Mirrors /v1/workflows/experiments — see the Python SDK at
// retab/resources/workflows/experiments/client.py for the canonical surface.
// ---------------------------------------------------------------------------

// WorkflowExperimentsService gives access to experiments and per-experiment runs.
type WorkflowExperimentsService struct {
	client *Client
	Runs   *WorkflowExperimentRunsService
}

// WorkflowExperimentRunsService manages experiment run lifecycle, results, and metrics.
type WorkflowExperimentRunsService struct {
	client  *Client
	Results *WorkflowExperimentRunResultsService
	Metrics *WorkflowExperimentRunMetricsService
}

type WorkflowExperimentRunResultsService struct {
	client *Client
}

type WorkflowExperimentRunMetricsService struct {
	client *Client
}

func newWorkflowExperimentsService(client *Client) *WorkflowExperimentsService {
	return &WorkflowExperimentsService{
		client: client,
		Runs: &WorkflowExperimentRunsService{
			client:  client,
			Results: &WorkflowExperimentRunResultsService{client: client},
			Metrics: &WorkflowExperimentRunMetricsService{client: client},
		},
	}
}

// ExperimentDocumentProvenance attaches workflow-execution metadata to a
// captured experiment document.
type ExperimentDocumentProvenance struct {
	WorkflowRunID string `json:"workflow_run_id,omitempty"`
	StepID        string `json:"step_id,omitempty"`
}

// ExperimentDocumentCaptureRequest captures one document from a workflow
// run's recorded inputs.
type ExperimentDocumentCaptureRequest struct {
	WorkflowRunID string `json:"workflow_run_id"`
	StepID        string `json:"step_id,omitempty"`
}

// ExplicitExperimentDocumentRequest carries inlined handle_inputs.
type ExplicitExperimentDocumentRequest struct {
	HandleInputs map[string]any                `json:"handle_inputs"`
	Provenance   *ExperimentDocumentProvenance `json:"provenance,omitempty"`
}

// WorkflowExperiment is one experiment row returned by create/list/get/update.
type WorkflowExperiment struct {
	ID                string    `json:"id"`
	WorkflowID        string    `json:"workflow_id"`
	BlockID           string    `json:"block_id"`
	NConsensus        int       `json:"n_consensus"`
	DocumentCount     int       `json:"document_count"`
	Name              string    `json:"name"`
	LastRunID         string    `json:"last_run_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Status            string    `json:"status"`
	BlockKind         string    `json:"block_kind"`
	Score             *float64  `json:"score,omitempty"`
	IsStale           bool      `json:"is_stale,omitempty"`
	SchemaDrift       string    `json:"schema_drift,omitempty"`
	SchemaDriftDetail string    `json:"schema_drift_detail,omitempty"`
}

// ExperimentRun is the public run document returned by experiment run APIs.
type ExperimentRunLifecycle struct {
	Status string `json:"status"`
}

type ExperimentRunTiming struct {
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DurationMs  *int       `json:"duration_ms,omitempty"`
}

type ExperimentRun struct {
	ID                     string                 `json:"id"`
	Workflow               WorkflowSnapshotRef    `json:"workflow"`
	Trigger                RunTrigger             `json:"trigger"`
	Lifecycle              ExperimentRunLifecycle `json:"lifecycle"`
	Timing                 ExperimentRunTiming    `json:"timing"`
	ExperimentID           string                 `json:"experiment_id"`
	BlockID                string                 `json:"block_id"`
	BlockKind              string                 `json:"block_kind"`
	NConsensus             int                    `json:"n_consensus"`
	DefinitionFingerprint  string                 `json:"definition_fingerprint"`
	DocumentsFingerprint   string                 `json:"documents_fingerprint"`
	Score                  *float64               `json:"score,omitempty"`
	TotalDocumentCount     int                    `json:"total_document_count"`
	CompletedDocumentCount int                    `json:"completed_document_count"`
	DocumentCount          int                    `json:"document_count"`
	ErrorCount             int                    `json:"error_count"`
}

// ExperimentRunListResponse is the canonical PaginatedList envelope for
// `GET /v1/workflows/experiments/runs`.
type ExperimentRunListResponse = PaginatedList[ExperimentRun]

type CancelWorkflowExperimentRunResponse struct {
	ID        string                 `json:"id"`
	Lifecycle ExperimentRunLifecycle `json:"lifecycle"`
}

// ExperimentResult is the per-document execution record inside a run.
type ExperimentResultLifecycle struct {
	Status string `json:"status"`
}

type ExperimentResultTiming struct {
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DurationMs  *int       `json:"duration_ms,omitempty"`
}

type ExperimentResult struct {
	ID            string                    `json:"id"`
	RunID         string                    `json:"run_id"`
	ExperimentID  string                    `json:"experiment_id"`
	DocumentID    string                    `json:"document_id"`
	Lifecycle     ExperimentResultLifecycle `json:"lifecycle"`
	Timing        ExperimentResultTiming    `json:"timing"`
	BlockKind     string                    `json:"block_kind"`
	HandleInputs  map[string]any            `json:"handle_inputs"`
	Artifact      *StepArtifactRef          `json:"artifact,omitempty"`
	DurationMs    *int                      `json:"duration_ms,omitempty"`
	CreatedAt     *time.Time                `json:"created_at,omitempty"`
	StartedAt     *time.Time                `json:"started_at,omitempty"`
	CompletedAt   *time.Time                `json:"completed_at,omitempty"`
	Attempt       int                       `json:"attempt"`
	IsPlaceholder bool                      `json:"is_placeholder"`
}

type ExperimentResultListResponse = PaginatedList[ExperimentResult]

// ExperimentMetricsResponse is the GET /workflows/experiments/metrics payload.
//
// Modelling the discriminated union (four "view" shapes plus two error
// envelopes) here would couple the SDK to internal structural details. We
// expose it as Resource — callers branch on the "view" or "error" key.
type ExperimentMetricsResponse = Resource

// CreateExperimentRequest is the body for POST /experiments.
type CreateExperimentRequest struct {
	WorkflowID       string                              `json:"-"`
	BlockID          string                              `json:"block_id"`
	Name             string                              `json:"name"`
	DocumentCaptures []ExperimentDocumentCaptureRequest  `json:"document_captures,omitempty"`
	Documents        []ExplicitExperimentDocumentRequest `json:"documents,omitempty"`
	NConsensus       int                                 `json:"n_consensus,omitempty"`
}

// UpdateExperimentRequest is the body for PATCH /experiments/{id}. Only
// fields you set are forwarded. Use NConsensus to disambiguate the zero value
// from "leave unchanged".
type UpdateExperimentRequest struct {
	Name             string                              `json:"-"`
	NConsensus       *int                                `json:"-"`
	DocumentCaptures []ExperimentDocumentCaptureRequest  `json:"-"`
	Documents        []ExplicitExperimentDocumentRequest `json:"-"`
}

type ListExperimentRunsParams struct {
	WorkflowID    string
	ExperimentID  string
	BlockID       string
	Status        string
	Statuses      []string
	ExcludeStatus string
	TriggerType   string
	TriggerTypes  []string
	FromDate      string
	ToDate        string
	SortBy        string
	Fields        []string
	Before        string
	After         string
	Limit         int
	Order         string
}

// GetExperimentMetricsParams gathers the GET /workflows/experiments/metrics
// query params.
type GetExperimentMetricsParams struct {
	View         string
	DocumentID   string
	TargetPath   string
	IncludePrior *bool
	PriorRunID   string
}

// Create posts a new experiment definition. Use Runs.Create afterwards to
// trigger an actual evaluation.
func (s *WorkflowExperimentsService) Create(ctx context.Context, request CreateExperimentRequest, opts ...RequestOption) (*WorkflowExperiment, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if request.BlockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	if request.Name == "" {
		return nil, fmt.Errorf("retab: name is required")
	}
	body := map[string]any{
		"workflow_id": request.WorkflowID,
		"block_id":    request.BlockID,
		"name":        request.Name,
	}
	if request.NConsensus != 0 {
		body["n_consensus"] = request.NConsensus
	}
	if request.DocumentCaptures != nil {
		body["document_captures"] = request.DocumentCaptures
	}
	if request.Documents != nil {
		body["documents"] = request.Documents
	}
	var result WorkflowExperiment
	err := s.client.do(ctx, http.MethodPost,
		"/workflows/experiments",
		nil, body, &result, opts...)
	return &result, err
}

// List returns every experiment attached to the workflow, newest first.
//
// Returns a `PaginatedList[WorkflowExperiment]` to match the canonical Retab
// list envelope (`{data, list_metadata}`). Iterate over `result.Data` for the
// rows.
func (s *WorkflowExperimentsService) List(ctx context.Context, workflowID string, opts ...RequestOption) (*PaginatedList[WorkflowExperiment], error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result PaginatedList[WorkflowExperiment]
	err := s.client.do(ctx, http.MethodGet, "/workflows/experiments?workflow_id="+url.QueryEscape(workflowID), nil, nil, &result, opts...)
	return &result, err
}

// Get fetches one experiment by ID.
func (s *WorkflowExperimentsService) Get(ctx context.Context, experimentID string, opts ...RequestOption) (*WorkflowExperiment, error) {
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	var result WorkflowExperiment
	err := s.client.do(ctx, http.MethodGet,
		"/workflows/experiments/"+url.PathEscape(experimentID),
		nil, nil, &result, opts...)
	return &result, err
}

// Update patches the experiment. Only fields you set are forwarded.
func (s *WorkflowExperimentsService) Update(ctx context.Context, experimentID string, request UpdateExperimentRequest, opts ...RequestOption) (*WorkflowExperiment, error) {
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	body := map[string]any{}
	if request.Name != "" {
		body["name"] = request.Name
	}
	if request.NConsensus != nil {
		body["n_consensus"] = *request.NConsensus
	}
	if request.DocumentCaptures != nil {
		body["document_captures"] = request.DocumentCaptures
	}
	if request.Documents != nil {
		body["documents"] = request.Documents
	}
	var result WorkflowExperiment
	err := s.client.do(ctx, http.MethodPatch,
		"/workflows/experiments/"+url.PathEscape(experimentID),
		nil, body, &result, opts...)
	return &result, err
}

// Delete removes the experiment and its runs.
func (s *WorkflowExperimentsService) Delete(ctx context.Context, experimentID string, opts ...RequestOption) error {
	if experimentID == "" {
		return fmt.Errorf("retab: experimentID is required")
	}
	return s.client.do(ctx, http.MethodDelete,
		"/workflows/experiments/"+url.PathEscape(experimentID),
		nil, nil, nil, opts...)
}

// Create triggers an experiment run with the current draft block config.
func (s *WorkflowExperimentRunsService) Create(ctx context.Context, workflowID, experimentID string, opts ...RequestOption) (*ExperimentRun, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	body := map[string]any{
		"experiment_id": experimentID,
		"workflow_id":   workflowID,
	}
	var result ExperimentRun
	err := s.client.do(ctx, http.MethodPost,
		"/workflows/experiments/runs",
		nil, body, &result, opts...)
	return &result, err
}

// List returns experiment runs, newest first by default.
func (s *WorkflowExperimentRunsService) List(ctx context.Context, params *ListExperimentRunsParams, opts ...RequestOption) (*PaginatedList[ExperimentRun], error) {
	query := url.Values{}
	limit := 20
	if params != nil {
		if params.Limit != 0 {
			limit = params.Limit
		}
		addQuery(query, "workflow_id", params.WorkflowID)
		addQuery(query, "experiment_id", params.ExperimentID)
		addQuery(query, "block_id", params.BlockID)
		addQuery(query, "status", params.Status)
		addCSVQuery(query, "statuses", params.Statuses)
		addQuery(query, "exclude_status", params.ExcludeStatus)
		addQuery(query, "trigger_type", params.TriggerType)
		addCSVQuery(query, "trigger_types", params.TriggerTypes)
		addQuery(query, "from_date", params.FromDate)
		addQuery(query, "to_date", params.ToDate)
		addQuery(query, "sort_by", params.SortBy)
		addCSVQuery(query, "fields", params.Fields)
		addQuery(query, "before", params.Before)
		addQuery(query, "after", params.After)
		addQuery(query, "order", params.Order)
	}
	query.Set("limit", fmt.Sprintf("%d", limit))
	var result PaginatedList[ExperimentRun]
	err := s.client.do(ctx, http.MethodGet, "/workflows/experiments/runs", query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowExperimentRunsService) Get(ctx context.Context, runID string, opts ...RequestOption) (*ExperimentRun, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result ExperimentRun
	err := s.client.do(ctx, http.MethodGet, "/workflows/experiments/runs/"+url.PathEscape(runID), nil, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowExperimentRunsService) Cancel(ctx context.Context, runID string, opts ...RequestOption) (*CancelWorkflowExperimentRunResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result CancelWorkflowExperimentRunResponse
	err := s.client.do(ctx, http.MethodPost, "/workflows/experiments/runs/"+url.PathEscape(runID)+"/cancel", nil, map[string]any{}, &result, opts...)
	return &result, err
}

func (s *WorkflowExperimentRunResultsService) List(ctx context.Context, runID string, limit int, opts ...RequestOption) (*PaginatedList[ExperimentResult], error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if limit == 0 {
		limit = 20
	}
	query := url.Values{}
	query.Set("run_id", runID)
	query.Set("limit", fmt.Sprintf("%d", limit))
	var result PaginatedList[ExperimentResult]
	err := s.client.do(ctx, http.MethodGet, "/workflows/experiments/results", query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowExperimentRunResultsService) Get(ctx context.Context, resultID string, opts ...RequestOption) (*ExperimentResult, error) {
	if resultID == "" {
		return nil, fmt.Errorf("retab: resultID is required")
	}
	var result ExperimentResult
	err := s.client.do(ctx, http.MethodGet, "/workflows/experiments/results/"+url.PathEscape(resultID), nil, nil, &result, opts...)
	return &result, err
}

// Get returns experiment metrics for one of the run-scoped views.
func (s *WorkflowExperimentRunMetricsService) Get(ctx context.Context, runID string, params *GetExperimentMetricsParams, opts ...RequestOption) (ExperimentMetricsResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	query := url.Values{}
	view := "summary"
	includePrior := true
	if params != nil {
		if params.View != "" {
			view = params.View
		}
		addQuery(query, "document_id", params.DocumentID)
		addQuery(query, "target_path", params.TargetPath)
		addQuery(query, "prior_run_id", params.PriorRunID)
		if params.IncludePrior != nil {
			includePrior = *params.IncludePrior
		}
	}
	query.Set("run_id", runID)
	query.Set("view", view)
	query.Set("include_prior", fmt.Sprintf("%t", includePrior))
	var result ExperimentMetricsResponse
	err := s.client.do(ctx, http.MethodGet, "/workflows/experiments/metrics", query, nil, &result, opts...)
	return result, err
}
