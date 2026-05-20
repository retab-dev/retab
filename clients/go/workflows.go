package retab

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WorkflowsService struct {
	client      *Client
	Runs        *WorkflowRunsService
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
	service.Runs = &WorkflowRunsService{
		client: client,
		Steps:  &WorkflowRunStepsService{client: client},
	}
	service.Reviews = &WorkflowReviewsService{client: client}
	service.Artifacts = &WorkflowArtifactsService{client: client}
	service.Blocks = &WorkflowBlocksService{client: client}
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

func (s *WorkflowsService) Duplicate(ctx context.Context, workflowID string, opts ...RequestOption) (*Workflow, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var workflow Workflow
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(workflowID)+"/duplicate", nil, map[string]any{}, &workflow, opts...)
	return &workflow, err
}

func (s *WorkflowsService) GetEntities(ctx context.Context, workflowID string, opts ...RequestOption) (*WorkflowWithEntities, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result WorkflowWithEntities
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/entities", nil, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowsService) GetResolvedSchemas(ctx context.Context, workflowID string, opts ...RequestOption) (*WorkflowResolvedSchemasResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result WorkflowResolvedSchemasResponse
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/resolved-schemas", nil, nil, &result, opts...)
	return &result, err
}

// WorkflowSnapshot is the metadata for one published snapshot of a workflow,
// returned as the data items of (*WorkflowsService).ListSnapshots.
type WorkflowSnapshot struct {
	ID               string    `json:"id"`
	SnapshotID       string    `json:"snapshot_id"`
	WorkflowID       string    `json:"workflow_id"`
	OrganizationID   string    `json:"organization_id,omitempty"`
	Version          int       `json:"version"`
	Description      string    `json:"description"`
	BlockCount       int       `json:"block_count"`
	EdgeCount        int       `json:"edge_count"`
	PublishedBy      string    `json:"published_by,omitempty"`
	PublishedByEmail string    `json:"published_by_email,omitempty"`
	PublishedByName  string    `json:"published_by_name,omitempty"`
	PublishedAt      time.Time `json:"published_at"`
}

// ListSnapshotsParams bounds the snapshot page returned by ListSnapshots.
type ListSnapshotsParams struct {
	Limit int
}

// ListSnapshots returns published snapshots for a workflow, newest first,
// wrapped in the canonical paginated envelope. Cursor pagination is not yet
// implemented for this endpoint; pass Limit (default 50, max 100) to bound
// the page size.
func (s *WorkflowsService) ListSnapshots(ctx context.Context, workflowID string, params *ListSnapshotsParams, opts ...RequestOption) (*PaginatedList[WorkflowSnapshot], error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	query := url.Values{}
	if params != nil && params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	var result PaginatedList[WorkflowSnapshot]
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/snapshots", query, nil, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
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

func prepareDiagnoseRequest(workflowID string, request DiagnoseWorkflowGraphRequest) PreparedRequest {
	return PreparedRequest{
		URL:    "/workflows/" + url.PathEscape(workflowID) + "/diagnose-graph",
		Method: http.MethodPost,
		Body:   request,
	}
}

// PrepareDiagnose returns the request descriptor for diagnosing an arbitrary
// workflow graph without executing it.
func (s *WorkflowsService) PrepareDiagnose(workflowID string, blocks []map[string]any, edges []map[string]any, rePropagate ...bool) PreparedRequest {
	shouldRePropagate := true
	if len(rePropagate) > 0 {
		shouldRePropagate = rePropagate[0]
	}
	return prepareDiagnoseRequest(workflowID, DiagnoseWorkflowGraphRequest{
		Blocks:      blocks,
		Edges:       edges,
		RePropagate: shouldRePropagate,
	})
}

// Diagnose runs the structural diagnosis on the persisted draft graph.
// It fetches workflow entities first, then POSTs them to diagnose-graph.
//
// To diagnose an in-memory graph that hasn't been saved, call DiagnoseGraph
// directly with your own blocks/edges payload.
func (s *WorkflowsService) Diagnose(ctx context.Context, workflowID string, opts ...RequestOption) (*WorkflowDiagnosisResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	entities, err := s.GetEntities(ctx, workflowID, opts...)
	if err != nil {
		return nil, err
	}
	blocks := make([]map[string]any, 0, len(entities.Blocks))
	for _, block := range entities.Blocks {
		blocks = append(blocks, map[string]any{
			"id":        block.ID,
			"type":      block.Type,
			"label":     block.Label,
			"config":    block.Config,
			"position":  map[string]any{"x": block.PositionX, "y": block.PositionY},
			"width":     block.Width,
			"height":    block.Height,
			"parent_id": block.ParentID,
		})
	}
	edges := make([]map[string]any, 0, len(entities.Edges))
	for _, edge := range entities.Edges {
		edges = append(edges, map[string]any{
			"id":            edge.ID,
			"source":        edge.SourceBlock,
			"target":        edge.TargetBlock,
			"source_handle": edge.SourceHandle,
			"target_handle": edge.TargetHandle,
		})
	}
	return s.DiagnoseGraph(ctx, workflowID, DiagnoseWorkflowGraphRequest{
		Blocks:      blocks,
		Edges:       edges,
		RePropagate: true,
	}, opts...)
}

// DiagnoseGraph posts an arbitrary graph to the diagnose-graph endpoint.
// Use this when the graph is in-memory and not yet persisted.
func (s *WorkflowsService) DiagnoseGraph(ctx context.Context, workflowID string, request DiagnoseWorkflowGraphRequest, opts ...RequestOption) (*WorkflowDiagnosisResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result WorkflowDiagnosisResponse
	prepared := prepareDiagnoseRequest(workflowID, request)
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
}

func (s *WorkflowArtifactsService) PrepareGet(operation string, artifactID string) PreparedRequest {
	return PreparedRequest{
		URL:    "/workflows/artifacts/" + url.PathEscape(operation) + "/" + url.PathEscape(artifactID),
		Method: http.MethodGet,
	}
}

func (s *WorkflowArtifactsService) PrepareList(params ListWorkflowArtifactsParams) PreparedRequest {
	query := url.Values{}
	addQuery(query, "run_id", params.RunID)
	addQuery(query, "operation", params.Operation)
	addQuery(query, "block_id", params.BlockID)
	return PreparedRequest{
		URL:    "/workflows/artifacts",
		Method: http.MethodGet,
		Params: query,
	}
}

func (s *WorkflowArtifactsService) Get(ctx context.Context, operation string, artifactID string, opts ...RequestOption) (*WorkflowArtifact, error) {
	if operation == "" {
		return nil, fmt.Errorf("retab: operation is required")
	}
	if artifactID == "" {
		return nil, fmt.Errorf("retab: artifactID is required")
	}
	var result WorkflowArtifact
	prepared := s.PrepareGet(operation, artifactID)
	err := s.client.do(ctx, prepared.Method, prepared.URL, nil, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowArtifactsService) GetRef(ctx context.Context, ref StepArtifactRef, opts ...RequestOption) (*WorkflowArtifact, error) {
	return s.Get(ctx, ref.Operation, ref.ID, opts...)
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
	client *Client
}

type WorkflowBlockCreateRequest struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Label     string         `json:"label,omitempty"`
	PositionX float64        `json:"position_x,omitempty"`
	PositionY float64        `json:"position_y,omitempty"`
	Width     *float64       `json:"width,omitempty"`
	Height    *float64       `json:"height,omitempty"`
	Config    map[string]any `json:"config,omitempty"`
	ParentID  string         `json:"parent_id,omitempty"`
}

type WorkflowBlockUpdateRequest struct {
	Label     *string        `json:"label,omitempty"`
	PositionX *float64       `json:"position_x,omitempty"`
	PositionY *float64       `json:"position_y,omitempty"`
	Width     *float64       `json:"width,omitempty"`
	Height    *float64       `json:"height,omitempty"`
	Config    map[string]any `json:"config,omitempty"`
	ParentID  *string        `json:"parent_id,omitempty"`
}

// List returns the canonical paginated envelope
// {"data": [...], "list_metadata": {"before": null, "after": null}} for
// the blocks of a workflow's current draft. Cursor pagination is not yet
// implemented for this endpoint; ListMetadata is always {nil, nil}.
func (s *WorkflowBlocksService) List(ctx context.Context, workflowID string, opts ...RequestOption) (*PaginatedList[WorkflowBlock], error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	// The /workflows/<id>/blocks endpoint currently returns a bare JSON
	// array; every other list endpoint returns the canonical paginated
	// envelope. Sniff the first non-whitespace byte and dispatch
	// accordingly so callers always receive a wrapped *PaginatedList,
	// matching the rest of the SDK. Drop the bare-array branch once the
	// server is consistent.
	var raw json.RawMessage
	if err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/blocks", nil, nil, &raw, opts...); err != nil {
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

// BlockConfigVersion captures one config-era for a block across workflow
// publishes. Returned as the data items of
// (*WorkflowBlocksService).ConfigHistory.
type BlockConfigVersion struct {
	ConfigFingerprint string         `json:"config_fingerprint"`
	BlockType         string         `json:"block_type"`
	BlockLabel        string         `json:"block_label"`
	ConfigSnapshot    map[string]any `json:"config_snapshot,omitempty"`
	FirstSeenAt       time.Time      `json:"first_seen_at"`
	LastSeenAt        time.Time      `json:"last_seen_at"`
	SnapshotVersions  []int          `json:"snapshot_versions"`
	RunCount          int            `json:"run_count"`
	IsCurrent         bool           `json:"is_current"`
}

// ConfigHistory returns the config-version timeline for a block, wrapped in
// the canonical paginated envelope. Each entry groups consecutive workflow
// snapshots in which the block's config did not change. Cursor pagination is
// not yet implemented for this endpoint.
func (s *WorkflowBlocksService) ConfigHistory(ctx context.Context, workflowID string, blockID string, opts ...RequestOption) (*PaginatedList[BlockConfigVersion], error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var result PaginatedList[BlockConfigVersion]
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/blocks/"+url.PathEscape(blockID)+"/config-history", nil, nil, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListSimulationsParams bounds the simulations history page returned by
// (*WorkflowBlocksService).ListSimulations.
type ListSimulationsParams struct {
	Limit int
}

// ListSimulations returns the paginated history of simulations for a block
// inside a workflow run. ID pagination is not yet implemented; pass
// Limit (default 20, max 100) to bound the page size.
func (s *WorkflowBlocksService) ListSimulations(ctx context.Context, runID string, blockID string, params *ListSimulationsParams, opts ...RequestOption) (*PaginatedList[BlockSimulation], error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	query := url.Values{}
	if params != nil && params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	var result PaginatedList[BlockSimulation]
	err := s.client.do(
		ctx,
		http.MethodGet,
		"/workflows/runs/"+url.PathEscape(runID)+"/steps/"+url.PathEscape(blockID)+"/simulations",
		query,
		nil,
		&result,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *WorkflowBlocksService) Get(ctx context.Context, workflowID string, blockID string, opts ...RequestOption) (*WorkflowBlock, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var block WorkflowBlock
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/blocks/"+url.PathEscape(blockID), nil, nil, &block, opts...)
	return &block, err
}

func (s *WorkflowBlocksService) GetResolvedSchemas(ctx context.Context, workflowID string, blockID string, opts ...RequestOption) (*BlockResolvedSchemasResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var result BlockResolvedSchemasResponse
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/blocks/"+url.PathEscape(blockID)+"/resolved-schemas", nil, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowBlocksService) Create(ctx context.Context, workflowID string, request WorkflowBlockCreateRequest, opts ...RequestOption) (*WorkflowBlock, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var block WorkflowBlock
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(workflowID)+"/blocks", nil, request, &block, opts...)
	return &block, err
}

func (s *WorkflowBlocksService) CreateBatch(ctx context.Context, workflowID string, blocks []WorkflowBlockCreateRequest, opts ...RequestOption) ([]WorkflowBlock, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result []WorkflowBlock
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(workflowID)+"/blocks/batch", nil, blocks, &result, opts...)
	return result, err
}

func (s *WorkflowBlocksService) Update(ctx context.Context, workflowID string, blockID string, request WorkflowBlockUpdateRequest, opts ...RequestOption) (*WorkflowBlock, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var block WorkflowBlock
	err := s.client.do(ctx, http.MethodPatch, "/workflows/"+url.PathEscape(workflowID)+"/blocks/"+url.PathEscape(blockID), nil, request, &block, opts...)
	return &block, err
}

func (s *WorkflowBlocksService) Delete(ctx context.Context, workflowID string, blockID string, opts ...RequestOption) error {
	if workflowID == "" {
		return fmt.Errorf("retab: workflowID is required")
	}
	if blockID == "" {
		return fmt.Errorf("retab: blockID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/"+url.PathEscape(workflowID)+"/blocks/"+url.PathEscape(blockID), nil, nil, nil, opts...)
}

// SimulateBlockRequest replays one block using inputs from a previous run
// plus the current draft block config. Note that the route is keyed by
// run_id, NOT workflow_id — the backend lives at
// /v1/workflows/runs/{run_id}/steps/{block_id}/simulate.
type SimulateBlockRequest struct {
	RunID            string `json:"-"`
	BlockID          string `json:"-"`
	NConsensus       int    `json:"-"`
	StepID           string `json:"-"`
	CheckEligibility *bool  `json:"-"`
}

// BlockSimulation is the result of replaying one block.
type BlockSimulation struct {
	ID                  string           `json:"id"`
	WorkflowID          string           `json:"workflow_id"`
	RunID               string           `json:"run_id"`
	BlockID             string           `json:"block_id"`
	BlockType           string           `json:"block_type"`
	Success             bool             `json:"success"`
	HandleInputs        map[string]any   `json:"handle_inputs,omitempty"`
	Artifact            *StepArtifactRef `json:"artifact,omitempty"`
	HandleOutputs       map[string]any   `json:"handle_outputs,omitempty"`
	RoutingDecision     []string         `json:"routing_decision,omitempty"`
	Error               string           `json:"error,omitempty"`
	DurationMs          *float64         `json:"duration_ms,omitempty"`
	Skipped             bool             `json:"skipped,omitempty"`
	CreatedAt           *time.Time       `json:"created_at,omitempty"`
	BlockConfig         map[string]any   `json:"block_config,omitempty"`
	StepID              string           `json:"step_id,omitempty"`
	AvailableIterations []map[string]any `json:"available_iterations,omitempty"`
}

func prepareSimulateBlockRequest(request SimulateBlockRequest) PreparedRequest {
	query := url.Values{}
	if request.NConsensus != 0 {
		query.Set("n_consensus", fmt.Sprintf("%d", request.NConsensus))
	}
	addQuery(query, "step_id", request.StepID)
	// Only send when overriding the default — the backend already defaults
	// to true and `?check_eligibility=true` would be redundant.
	if request.CheckEligibility != nil && !*request.CheckEligibility {
		query.Set("check_eligibility", "false")
	}
	return PreparedRequest{
		URL:    "/workflows/runs/" + url.PathEscape(request.RunID) + "/steps/" + url.PathEscape(request.BlockID) + "/simulate",
		Method: http.MethodPost,
		Params: query,
	}
}

// PrepareSimulate returns the request descriptor for replaying one block with
// inputs from a previous run.
func (s *WorkflowBlocksService) PrepareSimulate(request SimulateBlockRequest) PreparedRequest {
	return prepareSimulateBlockRequest(request)
}

// Simulate replays one block with the current draft config and returns the
// produced outputs without persisting them.
func (s *WorkflowBlocksService) Simulate(ctx context.Context, request SimulateBlockRequest, opts ...RequestOption) (*BlockSimulation, error) {
	if request.RunID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if request.BlockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var result BlockSimulation
	prepared := prepareSimulateBlockRequest(request)
	err := s.client.do(ctx, prepared.Method, prepared.URL, prepared.Params, nil, &result, opts...)
	return &result, err
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
}

// List returns the canonical paginated envelope
// {"data": [...], "list_metadata": {"before": null, "after": null}} for
// edges of a workflow's current draft. Cursor pagination is not yet
// implemented for this endpoint; ListMetadata is always {nil, nil}.
func (s *WorkflowEdgesService) List(ctx context.Context, workflowID string, params *ListWorkflowEdgesParams, opts ...RequestOption) (*PaginatedList[WorkflowEdgeDoc], error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	query := url.Values{}
	if params != nil {
		addQuery(query, "source_block", params.SourceBlock)
		addQuery(query, "target_block", params.TargetBlock)
	}
	var result PaginatedList[WorkflowEdgeDoc]
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/edges", query, nil, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *WorkflowEdgesService) Get(ctx context.Context, workflowID string, edgeID string, opts ...RequestOption) (*WorkflowEdgeDoc, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if edgeID == "" {
		return nil, fmt.Errorf("retab: edgeID is required")
	}
	var edge WorkflowEdgeDoc
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/edges/"+url.PathEscape(edgeID), nil, nil, &edge, opts...)
	return &edge, err
}

func (s *WorkflowEdgesService) Create(ctx context.Context, workflowID string, request WorkflowEdgeCreateRequest, opts ...RequestOption) (*WorkflowEdgeDoc, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var edge WorkflowEdgeDoc
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(workflowID)+"/edges", nil, request, &edge, opts...)
	return &edge, err
}

func (s *WorkflowEdgesService) CreateBatch(ctx context.Context, workflowID string, edges []WorkflowEdgeCreateRequest, opts ...RequestOption) ([]WorkflowEdgeDoc, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result []WorkflowEdgeDoc
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(workflowID)+"/edges/batch", nil, edges, &result, opts...)
	return result, err
}

func (s *WorkflowEdgesService) Delete(ctx context.Context, workflowID string, edgeID string, opts ...RequestOption) error {
	if workflowID == "" {
		return fmt.Errorf("retab: workflowID is required")
	}
	if edgeID == "" {
		return fmt.Errorf("retab: edgeID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/"+url.PathEscape(workflowID)+"/edges/"+url.PathEscape(edgeID), nil, nil, nil, opts...)
}

func (s *WorkflowEdgesService) DeleteAll(ctx context.Context, workflowID string, opts ...RequestOption) error {
	if workflowID == "" {
		return fmt.Errorf("retab: workflowID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/"+url.PathEscape(workflowID)+"/edges", nil, nil, nil, opts...)
}

type WorkflowRunsService struct {
	client *Client
	Steps  *WorkflowRunStepsService
}

type CreateWorkflowRunRequest struct {
	WorkflowID string         `json:"-"`
	Documents  map[string]any `json:"documents,omitempty"`
	JSONInputs map[string]any `json:"json_inputs,omitempty"`
	Version    string         `json:"version,omitempty"`
}

func (s *WorkflowRunsService) Create(ctx context.Context, request CreateWorkflowRunRequest, opts ...RequestOption) (*WorkflowRun, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := map[string]any{}
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
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(request.WorkflowID)+"/run", nil, body, &run, opts...)
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
	body := map[string]any{"config_source": configSource}
	if request.CommandID != "" {
		body["command_id"] = request.CommandID
	}
	var run WorkflowRun
	err := s.client.do(ctx, http.MethodPost, "/workflows/runs/"+url.PathEscape(runID)+"/restart", nil, body, &run, opts...)
	return &run, err
}

// The v1 decision surface was removed in the hard cutover to the review
// overlay. Drive review decisions through WorkflowReviewsService instead;
// see Workflows.Reviews.

func (s *WorkflowRunsService) GetConfig(ctx context.Context, runID string, opts ...RequestOption) (map[string]any, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result map[string]any
	err := s.client.do(ctx, http.MethodGet, "/workflows/runs/"+url.PathEscape(runID)+"/config", nil, nil, &result, opts...)
	return result, err
}

func (s *WorkflowRunsService) ExecutionOrder(ctx context.Context, runID string, opts ...RequestOption) (map[string]any, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result map[string]any
	err := s.client.do(ctx, http.MethodGet, "/workflows/runs/"+url.PathEscape(runID)+"/execution-order", nil, nil, &result, opts...)
	return result, err
}

func (s *WorkflowRunsService) GetDocumentURL(ctx context.Context, runID string, blockID string, opts ...RequestOption) (map[string]any, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var result map[string]any
	err := s.client.do(ctx, http.MethodGet, "/workflows/runs/"+url.PathEscape(runID)+"/documents/"+url.PathEscape(blockID), nil, nil, &result, opts...)
	return result, err
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
	var result WorkflowRunExportResponse
	err := s.client.do(ctx, http.MethodPost, "/workflows/runs/export-payload", nil, body, &result, opts...)
	return &result, err
}

type WorkflowRunStepsService struct {
	client *Client
}

// List returns the canonical paginated envelope
// {"data": [...], "list_metadata": {"before": null, "after": null}} for
// the persisted step documents of one workflow run. Cursor pagination is
// not yet implemented for this endpoint; ListMetadata is always {nil, nil}.
func (s *WorkflowRunStepsService) List(ctx context.Context, runID string, opts ...RequestOption) (*PaginatedList[WorkflowRunStep], error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result PaginatedList[WorkflowRunStep]
	err := s.client.do(ctx, http.MethodGet, "/workflows/runs/"+url.PathEscape(runID)+"/steps", nil, nil, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *WorkflowRunStepsService) Get(ctx context.Context, runID string, blockID string, opts ...RequestOption) (*StepExecutionResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var step StepExecutionResponse
	err := s.client.do(ctx, http.MethodGet, "/workflows/runs/"+url.PathEscape(runID)+"/steps/"+url.PathEscape(blockID), nil, nil, &step, opts...)
	return &step, err
}

// WorkflowReviewsService drives the review overlay — the actor-neutral
// review loop served under /workflows/reviews. A proposal authored by a
// model, an agent, or a human flows through the SAME Approve/Edit pair;
// ReviewActor.Kind is data on the overlay, never a behavioral switch.
//
// Every mutating call carries a VersionStamp (the overlay Rev). If the
// server's Rev advanced since the caller last read it, the request fails
// with APIError{StatusCode: 409} — re-read with Get and retry.
type WorkflowReviewsService struct {
	client *Client
}

type ReviewWaitForParams struct {
	PollInterval time.Duration
	Timeout      time.Duration
}

// ListReviewsParams filters the review queue.
type ListReviewsParams struct {
	WorkflowID string // restrict to one workflow
	Status     string // awaiting_review (default) | approved | rejected
	Mine       bool   // only reviews claimed by the calling actor
	Limit      int    // page size, 1-200
}

// List returns a page of the review queue — block runs awaiting review,
// hottest first.
func (s *WorkflowReviewsService) List(ctx context.Context, params *ListReviewsParams, opts ...RequestOption) (*ReviewQueueResponse, error) {
	query := url.Values{}
	if params != nil {
		addQuery(query, "workflow_id", params.WorkflowID)
		addQuery(query, "status", params.Status)
		if params.Mine {
			query.Set("mine", "true")
		}
		if params.Limit > 0 {
			query.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
	}
	var result ReviewQueueResponse
	err := s.client.do(ctx, http.MethodGet, "/workflows/reviews", query, nil, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns the full review overlay for one reviewed (run, block): every
// version, every decision, the audit trail, and the Rev CAS token.
func (s *WorkflowReviewsService) Get(ctx context.Context, runID string, blockID string, opts ...RequestOption) (*ReviewOverlay, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var overlay ReviewOverlay
	err := s.client.do(ctx, http.MethodGet, reviewsPath(runID, blockID), nil, nil, &overlay, opts...)
	if err != nil {
		return nil, err
	}
	return &overlay, nil
}

// ApproveReviewRequest approves the reviewed output, optionally with a
// corrective edit applied as a new version before the approval lands.
type ApproveReviewRequest struct {
	VersionStamp    int            // overlay Rev last observed (CAS token)
	EditedOutput    map[string]any // optional full corrected output snapshot
	ReviewableValue map[string]any // optional primitive-specific reviewable value
	OnSeq           *int           // version the decision is made against (default: head)
	EffectiveSeq    *int           // version that ships downstream (default: OnSeq)
	CommandID       string         // optional idempotency key
}

// Approve approves a reviewed block output.
func (s *WorkflowReviewsService) Approve(ctx context.Context, runID string, blockID string, request ApproveReviewRequest, opts ...RequestOption) (*SubmitReviewDecisionResponse, error) {
	if request.EditedOutput != nil && request.ReviewableValue != nil {
		return nil, fmt.Errorf("retab: EditedOutput and ReviewableValue are mutually exclusive")
	}
	body := map[string]any{
		"verdict":       "approved",
		"version_stamp": request.VersionStamp,
	}
	if request.EditedOutput != nil {
		body["edited_output"] = request.EditedOutput
	}
	if request.ReviewableValue != nil {
		body["reviewable_value"] = request.ReviewableValue
	}
	if request.OnSeq != nil {
		body["on_seq"] = *request.OnSeq
	}
	if request.EffectiveSeq != nil {
		body["effective_seq"] = *request.EffectiveSeq
	}
	if request.CommandID != "" {
		body["command_id"] = request.CommandID
	}
	return s.submitDecision(ctx, runID, blockID, body, opts...)
}

// RejectReviewRequest rejects the reviewed output. The runtime records the
// review as rejected and marks the workflow run error/rejected.
type RejectReviewRequest struct {
	VersionStamp int    // overlay Rev last observed (CAS token)
	Reason       string // required — every rejection must be auditable
	CommandID    string // optional idempotency key
}

// Reject rejects a reviewed block output.
func (s *WorkflowReviewsService) Reject(ctx context.Context, runID string, blockID string, request RejectReviewRequest, opts ...RequestOption) (*SubmitReviewDecisionResponse, error) {
	body := map[string]any{
		"verdict":       "rejected",
		"version_stamp": request.VersionStamp,
	}
	if request.Reason != "" {
		body["reason"] = request.Reason
	}
	if request.CommandID != "" {
		body["command_id"] = request.CommandID
	}
	return s.submitDecision(ctx, runID, blockID, body, opts...)
}

// EscalateReviewRequest is retained for source compatibility. The review
// overlay API does not currently support escalation; use Edit, Approve, or
// Reject instead.
type EscalateReviewRequest struct {
	VersionStamp int    // overlay Rev last observed (CAS token)
	Reason       string // required
	EscalateTo   string // required — target queue/team id
	CommandID    string // optional idempotency key
}

// Escalate is unsupported by the review overlay API.
func (s *WorkflowReviewsService) Escalate(ctx context.Context, runID string, blockID string, request EscalateReviewRequest, opts ...RequestOption) (*SubmitReviewDecisionResponse, error) {
	return nil, fmt.Errorf("retab: review escalation is not supported by the review overlay API; use Edit, Approve, or Reject")
}

func (s *WorkflowReviewsService) submitDecision(ctx context.Context, runID string, blockID string, body map[string]any, opts ...RequestOption) (*SubmitReviewDecisionResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var result SubmitReviewDecisionResponse
	err := s.client.do(ctx, http.MethodPost, reviewsPath(runID, blockID)+"/decision", nil, body, &result, opts...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EditReviewRequest appends a corrective output version without deciding.
type EditReviewRequest struct {
	Snapshot        map[string]any // optional full corrected output snapshot
	ReviewableValue map[string]any // optional primitive-specific reviewable value
	VersionStamp    int            // overlay Rev last observed (CAS token)
	Origin          string         // human_edit (default) | agent_edit
	Note            string         // optional rationale
	CommandID       string         // optional idempotency key
}

// Edit appends a new output version to the overlay's version history. A
// proposal authored by a human or an agent uses this same call — Origin is
// descriptive provenance, not a behavioral switch.
func (s *WorkflowReviewsService) Edit(ctx context.Context, runID string, blockID string, request EditReviewRequest, opts ...RequestOption) (*ReviewOverlay, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	if request.Snapshot != nil && request.ReviewableValue != nil {
		return nil, fmt.Errorf("retab: Snapshot and ReviewableValue are mutually exclusive")
	}
	if request.Snapshot == nil && request.ReviewableValue == nil {
		return nil, fmt.Errorf("retab: Snapshot or ReviewableValue is required")
	}
	body := map[string]any{
		"version_stamp": request.VersionStamp,
	}
	if request.Snapshot != nil {
		body["snapshot"] = request.Snapshot
	}
	if request.ReviewableValue != nil {
		body["reviewable_value"] = request.ReviewableValue
	}
	if request.Origin != "" {
		body["origin"] = request.Origin
	}
	if request.Note != "" {
		body["note"] = request.Note
	}
	if request.CommandID != "" {
		body["command_id"] = request.CommandID
	}
	var overlay ReviewOverlay
	err := s.client.do(ctx, http.MethodPost, reviewsPath(runID, blockID)+"/versions", nil, body, &overlay, opts...)
	if err != nil {
		return nil, err
	}
	return &overlay, nil
}

// Claim takes the advisory review claim ("Dana is reviewing this"). A claim
// is never a lock — correctness rests on the Rev CAS.
func (s *WorkflowReviewsService) Claim(ctx context.Context, runID string, blockID string, versionStamp int, ttlSeconds int, opts ...RequestOption) (*ReviewOverlay, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	body := map[string]any{"version_stamp": versionStamp}
	if ttlSeconds > 0 {
		body["ttl_seconds"] = ttlSeconds
	}
	var overlay ReviewOverlay
	err := s.client.do(ctx, http.MethodPost, reviewsPath(runID, blockID)+"/claim", nil, body, &overlay, opts...)
	if err != nil {
		return nil, err
	}
	return &overlay, nil
}

// Release clears the advisory review claim.
func (s *WorkflowReviewsService) Release(ctx context.Context, runID string, blockID string, versionStamp int, opts ...RequestOption) (*ReviewOverlay, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	body := map[string]any{"version_stamp": versionStamp}
	var overlay ReviewOverlay
	err := s.client.do(ctx, http.MethodPost, reviewsPath(runID, blockID)+"/release", nil, body, &overlay, opts...)
	if err != nil {
		return nil, err
	}
	return &overlay, nil
}

// WaitFor polls until the block is awaiting review. A 404 means the workflow
// has not reached the review point yet, so polling continues.
func (s *WorkflowReviewsService) WaitFor(ctx context.Context, runID string, blockID string, params *ReviewWaitForParams, opts ...RequestOption) (*ReviewOverlay, error) {
	pollInterval := 2 * time.Second
	timeout := 2 * time.Minute
	if params != nil {
		if params.PollInterval > 0 {
			pollInterval = params.PollInterval
		}
		if params.Timeout > 0 {
			timeout = params.Timeout
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		overlay, err := s.Get(ctx, runID, blockID, opts...)
		if err == nil && overlay.Status == "awaiting_review" {
			return overlay, nil
		}
		if err != nil {
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
				return nil, err
			}
		}

		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func reviewsPath(runID, blockID string) string {
	return "/workflows/reviews/" + url.PathEscape(runID) + "/" + url.PathEscape(blockID)
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
type WorkflowTestRunRecord = Resource

type WorkflowTestCreateRequest struct {
	WorkflowID string   `json:"-"`
	Target     Resource `json:"target"`
	Source     Resource `json:"source"`
	Assertion  Resource `json:"assertion"`
	Name       string   `json:"name,omitempty"`
}

type WorkflowTestUpdateRequest struct {
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
// `GET /v1/workflows/{wf}/tests`. Kept as a type alias so callers can
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

type WorkflowTestRun struct {
	ID         string                `json:"id"`
	Workflow   WorkflowSnapshotRef   `json:"workflow"`
	Trigger    RunTrigger            `json:"trigger"`
	Lifecycle  RunLifecycle          `json:"lifecycle"`
	Timing     RunTiming             `json:"timing"`
	Target     *Resource             `json:"target,omitempty"`
	TestID     string                `json:"test_id,omitempty"`
	TotalTests int                   `json:"total_tests"`
	Counts     WorkflowTestRunCounts `json:"counts,omitempty"`
}

type WorkflowTestRunResultListResponse = PaginatedList[WorkflowTestRunRecord]

func (s *WorkflowTestsService) Create(ctx context.Context, request WorkflowTestCreateRequest, opts ...RequestOption) (*WorkflowTest, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := resourceFromJSON(request)
	delete(body, "WorkflowID")
	var result WorkflowTest
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(request.WorkflowID)+"/tests", nil, body, &result, opts...)
	return &result, err
}

func (s *WorkflowTestsService) Get(ctx context.Context, workflowID string, testID string, opts ...RequestOption) (*WorkflowTest, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if testID == "" {
		return nil, fmt.Errorf("retab: testID is required")
	}
	var result WorkflowTest
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/tests/"+url.PathEscape(testID), nil, nil, &result, opts...)
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
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(request.WorkflowID)+"/tests", query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowTestsService) Update(ctx context.Context, workflowID string, testID string, request WorkflowTestUpdateRequest, opts ...RequestOption) (*WorkflowTest, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if testID == "" {
		return nil, fmt.Errorf("retab: testID is required")
	}
	var result WorkflowTest
	err := s.client.do(ctx, http.MethodPatch, "/workflows/"+url.PathEscape(workflowID)+"/tests/"+url.PathEscape(testID), nil, request, &result, opts...)
	return &result, err
}

func (s *WorkflowTestsService) Delete(ctx context.Context, workflowID string, testID string, opts ...RequestOption) error {
	if workflowID == "" {
		return fmt.Errorf("retab: workflowID is required")
	}
	if testID == "" {
		return fmt.Errorf("retab: testID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/"+url.PathEscape(workflowID)+"/tests/"+url.PathEscape(testID), nil, nil, nil, opts...)
}

func (s *WorkflowTestRunsService) Create(ctx context.Context, request CreateWorkflowTestRunRequest, opts ...RequestOption) (*WorkflowTestRun, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := resourceFromJSON(request)
	delete(body, "WorkflowID")
	var result WorkflowTestRun
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(request.WorkflowID)+"/tests/runs", nil, body, &result, opts...)
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

func (s *WorkflowTestRunResultsService) List(ctx context.Context, runID string, opts ...RequestOption) (*PaginatedList[WorkflowTestRunRecord], error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result PaginatedList[WorkflowTestRunRecord]
	err := s.client.do(ctx, http.MethodGet, "/workflows/tests/runs/"+url.PathEscape(runID)+"/results", nil, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowTestRunResultsService) Get(ctx context.Context, runID string, testID string, opts ...RequestOption) (*WorkflowTestRunRecord, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if testID == "" {
		return nil, fmt.Errorf("retab: testID is required")
	}
	var result WorkflowTestRunRecord
	err := s.client.do(ctx, http.MethodGet, "/workflows/tests/runs/"+url.PathEscape(runID)+"/results/"+url.PathEscape(testID), nil, nil, &result, opts...)
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
// Mirrors /v1/workflows/{workflow_id}/experiments — see the Python SDK at
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

// ExperimentResponse is one experiment row returned by create/list/get/update/duplicate.
type ExperimentResponse struct {
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
type ExperimentRun struct {
	ID                     string              `json:"id"`
	Workflow               WorkflowSnapshotRef `json:"workflow"`
	Trigger                RunTrigger          `json:"trigger"`
	Lifecycle              RunLifecycle        `json:"lifecycle"`
	Timing                 RunTiming           `json:"timing"`
	ExperimentID           string              `json:"experiment_id"`
	BlockID                string              `json:"block_id"`
	BlockKind              string              `json:"block_kind"`
	NConsensus             int                 `json:"n_consensus"`
	DefinitionFingerprint  string              `json:"definition_fingerprint"`
	DocumentsFingerprint   string              `json:"documents_fingerprint"`
	Score                  *float64            `json:"score,omitempty"`
	TotalDocumentCount     int                 `json:"total_document_count"`
	CompletedDocumentCount int                 `json:"completed_document_count"`
	DocumentCount          int                 `json:"document_count"`
	ErrorCount             int                 `json:"error_count"`
}

// ExperimentRunListResponse is the canonical PaginatedList envelope for
// `GET /v1/workflows/experiments/runs`.
type ExperimentRunListResponse = PaginatedList[ExperimentRun]

// ExperimentResult is the per-document execution record inside a run.
type ExperimentResult struct {
	ID            string           `json:"id"`
	RunID         string           `json:"run_id"`
	ExperimentID  string           `json:"experiment_id"`
	DocumentID    string           `json:"document_id"`
	Lifecycle     RunLifecycle     `json:"lifecycle"`
	Timing        RunTiming        `json:"timing"`
	BlockKind     string           `json:"block_kind"`
	HandleInputs  map[string]any   `json:"handle_inputs"`
	Artifact      *StepArtifactRef `json:"artifact,omitempty"`
	Error         string           `json:"error,omitempty"`
	DurationMs    *int             `json:"duration_ms,omitempty"`
	CreatedAt     *time.Time       `json:"created_at,omitempty"`
	StartedAt     *time.Time       `json:"started_at,omitempty"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
	Attempt       int              `json:"attempt"`
	IsPlaceholder bool             `json:"is_placeholder"`
}

type ExperimentResultListResponse = PaginatedList[ExperimentResult]

// EligibleBlockSummary describes a block that supports experiments.
type EligibleBlockSummary struct {
	BlockID                string     `json:"block_id"`
	BlockLabel             string     `json:"block_label"`
	BlockType              string     `json:"block_type"`
	ExperimentCount        int        `json:"experiment_count"`
	DriftedExperimentCount int        `json:"drifted_experiment_count"`
	StaleExperimentCount   int        `json:"stale_experiment_count"`
	LatestRunAt            *time.Time `json:"latest_run_at,omitempty"`
	MeanScore              *float64   `json:"mean_score,omitempty"`
}

// EligibleBlockListResponse is GET /experiments/eligible-blocks.
type EligibleBlockListResponse struct {
	Blocks []EligibleBlockSummary `json:"blocks"`
}

// ExperimentMetricsResponse is the GET /workflows/experiments/runs/{run_id}/metrics payload.
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

// GetExperimentMetricsParams gathers the GET /workflows/experiments/runs/{run_id}/metrics
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
func (s *WorkflowExperimentsService) Create(ctx context.Context, request CreateExperimentRequest, opts ...RequestOption) (*ExperimentResponse, error) {
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
		"block_id": request.BlockID,
		"name":     request.Name,
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
	var result ExperimentResponse
	err := s.client.do(ctx, http.MethodPost,
		"/workflows/"+url.PathEscape(request.WorkflowID)+"/experiments",
		nil, body, &result, opts...)
	return &result, err
}

// List returns every experiment attached to the workflow, newest first.
//
// Returns a `PaginatedList[ExperimentResponse]` to match the canonical Retab
// list envelope (`{data, list_metadata}`). Iterate over `result.Data` for the
// rows.
func (s *WorkflowExperimentsService) List(ctx context.Context, workflowID string, opts ...RequestOption) (*PaginatedList[ExperimentResponse], error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result PaginatedList[ExperimentResponse]
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/experiments", nil, nil, &result, opts...)
	return &result, err
}

// Get fetches one experiment by ID.
func (s *WorkflowExperimentsService) Get(ctx context.Context, workflowID, experimentID string, opts ...RequestOption) (*ExperimentResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	var result ExperimentResponse
	err := s.client.do(ctx, http.MethodGet,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID),
		nil, nil, &result, opts...)
	return &result, err
}

// Update patches the experiment. Only fields you set are forwarded.
func (s *WorkflowExperimentsService) Update(ctx context.Context, workflowID, experimentID string, request UpdateExperimentRequest, opts ...RequestOption) (*ExperimentResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
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
	var result ExperimentResponse
	err := s.client.do(ctx, http.MethodPatch,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID),
		nil, body, &result, opts...)
	return &result, err
}

// Delete removes the experiment and its runs.
func (s *WorkflowExperimentsService) Delete(ctx context.Context, workflowID, experimentID string, opts ...RequestOption) error {
	if workflowID == "" {
		return fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return fmt.Errorf("retab: experimentID is required")
	}
	return s.client.do(ctx, http.MethodDelete,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID),
		nil, nil, nil, opts...)
}

// Duplicate copies an experiment definition. The duplicate has no runs.
func (s *WorkflowExperimentsService) Duplicate(ctx context.Context, workflowID, experimentID string, opts ...RequestOption) (*ExperimentResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	var result ExperimentResponse
	err := s.client.do(ctx, http.MethodPost,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID)+"/duplicate",
		nil, map[string]any{}, &result, opts...)
	return &result, err
}

// ListEligibleBlocks returns blocks in the workflow that support experiments,
// with rolled-up counts of how many experiments are attached / drifted / stale.
func (s *WorkflowExperimentsService) ListEligibleBlocks(ctx context.Context, workflowID string, opts ...RequestOption) (*EligibleBlockListResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result EligibleBlockListResponse
	err := s.client.do(ctx, http.MethodGet,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/eligible-blocks",
		nil, nil, &result, opts...)
	return &result, err
}

// Create triggers an experiment run with the current draft block config.
func (s *WorkflowExperimentRunsService) Create(ctx context.Context, workflowID, experimentID string, opts ...RequestOption) (*ExperimentRun, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	body := map[string]any{}
	var result ExperimentRun
	err := s.client.do(ctx, http.MethodPost,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID)+"/runs",
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

func (s *WorkflowExperimentRunsService) Cancel(ctx context.Context, runID string, opts ...RequestOption) (*ExperimentRun, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result ExperimentRun
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
	query.Set("limit", fmt.Sprintf("%d", limit))
	var result PaginatedList[ExperimentResult]
	err := s.client.do(ctx, http.MethodGet, "/workflows/experiments/runs/"+url.PathEscape(runID)+"/results", query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowExperimentRunResultsService) Get(ctx context.Context, runID string, documentID string, opts ...RequestOption) (*ExperimentResult, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if documentID == "" {
		return nil, fmt.Errorf("retab: documentID is required")
	}
	var result ExperimentResult
	err := s.client.do(ctx, http.MethodGet, "/workflows/experiments/runs/"+url.PathEscape(runID)+"/results/"+url.PathEscape(documentID), nil, nil, &result, opts...)
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
	query.Set("view", view)
	query.Set("include_prior", fmt.Sprintf("%t", includePrior))
	var result ExperimentMetricsResponse
	err := s.client.do(ctx, http.MethodGet, "/workflows/experiments/runs/"+url.PathEscape(runID)+"/metrics", query, nil, &result, opts...)
	return result, err
}
