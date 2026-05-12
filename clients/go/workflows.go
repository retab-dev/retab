package retab

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WorkflowsService struct {
	client      *Client
	Runs        *WorkflowRunsService
	Artifacts   *WorkflowArtifactsService
	Blocks      *WorkflowBlocksService
	Edges       *WorkflowEdgesService
	Tests       *WorkflowTestsService
	Experiments *WorkflowExperimentsService
}

func newWorkflowsService(client *Client) *WorkflowsService {
	service := &WorkflowsService{client: client}
	service.Runs = &WorkflowRunsService{
		client: client,
		Steps:  &WorkflowRunStepsService{client: client},
	}
	service.Artifacts = &WorkflowArtifactsService{client: client}
	service.Blocks = &WorkflowBlocksService{client: client}
	service.Edges = &WorkflowEdgesService{client: client}
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
	StartBlocks int            `json:"start_blocks"`
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

func (s *WorkflowArtifactsService) List(ctx context.Context, params ListWorkflowArtifactsParams, opts ...RequestOption) ([]WorkflowArtifact, error) {
	if params.RunID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result []WorkflowArtifact
	prepared := s.PrepareList(params)
	err := s.client.do(ctx, prepared.Method, prepared.URL, prepared.Params, nil, &result, opts...)
	return result, err
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

func (s *WorkflowBlocksService) List(ctx context.Context, workflowID string, opts ...RequestOption) ([]WorkflowBlock, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var blocks []WorkflowBlock
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/blocks", nil, nil, &blocks, opts...)
	return blocks, err
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

func (s *WorkflowEdgesService) List(ctx context.Context, workflowID string, params *ListWorkflowEdgesParams, opts ...RequestOption) ([]WorkflowEdgeDoc, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	query := url.Values{}
	if params != nil {
		addQuery(query, "source_block", params.SourceBlock)
		addQuery(query, "target_block", params.TargetBlock)
	}
	var edges []WorkflowEdgeDoc
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/edges", query, nil, &edges, opts...)
	return edges, err
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
		mimeData, err := InferMIMEData(document)
		if err != nil {
			return nil, err
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
	CommandID string `json:"command_id,omitempty"`
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
	body := map[string]any{}
	if request.CommandID != "" {
		body["command_id"] = request.CommandID
	}
	var run WorkflowRun
	err := s.client.do(ctx, http.MethodPost, "/workflows/runs/"+url.PathEscape(runID)+"/restart", nil, body, &run, opts...)
	return &run, err
}

type SubmitHILDecisionRequest struct {
	BlockID      string         `json:"-"`
	Approved     bool           `json:"-"`
	ModifiedData map[string]any `json:"-"`
	CommandID    string         `json:"-"`
}

func (s *WorkflowRunsService) SubmitHILDecision(ctx context.Context, runID string, request SubmitHILDecisionRequest, opts ...RequestOption) (*SubmitHILDecisionResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if request.BlockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	body := map[string]any{
		"block_id": request.BlockID,
		"approved": request.Approved,
	}
	if request.ModifiedData != nil {
		body["modified_data"] = request.ModifiedData
	}
	if request.CommandID != "" {
		body["command_id"] = request.CommandID
	}
	var result SubmitHILDecisionResponse
	err := s.client.do(ctx, http.MethodPost, "/workflows/runs/"+url.PathEscape(runID)+"/hil-decisions", nil, body, &result, opts...)
	return &result, err
}

func (s *WorkflowRunsService) GetHILDecision(ctx context.Context, runID string, blockID string, opts ...RequestOption) (*HILDecisionResource, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	if blockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	var result HILDecisionResource
	err := s.client.do(ctx, http.MethodGet, "/workflows/runs/"+url.PathEscape(runID)+"/hil-decisions/"+url.PathEscape(blockID), nil, nil, &result, opts...)
	return &result, err
}

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
	if request.SelectedRunIDs != nil {
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
	if request.TriggerTypes != nil {
		body["trigger_types"] = request.TriggerTypes
	}
	var result WorkflowRunExportResponse
	err := s.client.do(ctx, http.MethodPost, "/workflows/runs/export_payload", nil, body, &result, opts...)
	return &result, err
}

type WorkflowRunStepsService struct {
	client *Client
}

func (s *WorkflowRunStepsService) List(ctx context.Context, runID string, opts ...RequestOption) ([]WorkflowRunStep, error) {
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var steps []WorkflowRunStep
	err := s.client.do(ctx, http.MethodGet, "/workflows/runs/"+url.PathEscape(runID)+"/steps", nil, nil, &steps, opts...)
	return steps, err
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

type WorkflowTestsService struct {
	client *Client
	Runs   *WorkflowTestRunsService
}

func newWorkflowTestsService(client *Client) *WorkflowTestsService {
	service := &WorkflowTestsService{client: client}
	service.Runs = &WorkflowTestRunsService{client: client}
	return service
}

type WorkflowTest = Resource
type WorkflowTestRunRecord = Resource
type BlockTestBatchExecutionResult = Resource

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

type BlockTestListResponse struct {
	Data []WorkflowTest `json:"data"`
}

type ExecuteBlockTestsRequest struct {
	WorkflowID string   `json:"-"`
	TestID     string   `json:"test_id,omitempty"`
	Target     Resource `json:"target,omitempty"`
	NConsensus int      `json:"n_consensus,omitempty"`
}

type ExecuteBlockTestsResponse struct {
	BatchID string `json:"batch_id"`
	JobID   string `json:"job_id"`
}

func (s *WorkflowTestsService) Create(ctx context.Context, request WorkflowTestCreateRequest, opts ...RequestOption) (*WorkflowTest, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := resourceFromJSON(request)
	delete(body, "WorkflowID")
	var result WorkflowTest
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(request.WorkflowID)+"/block-tests", nil, body, &result, opts...)
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
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/block-tests/"+url.PathEscape(testID), nil, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowTestsService) List(ctx context.Context, request ListWorkflowTestsRequest, opts ...RequestOption) (*BlockTestListResponse, error) {
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
	var result BlockTestListResponse
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(request.WorkflowID)+"/block-tests", query, nil, &result, opts...)
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
	err := s.client.do(ctx, http.MethodPatch, "/workflows/"+url.PathEscape(workflowID)+"/block-tests/"+url.PathEscape(testID), nil, request, &result, opts...)
	return &result, err
}

func (s *WorkflowTestsService) Delete(ctx context.Context, workflowID string, testID string, opts ...RequestOption) error {
	if workflowID == "" {
		return fmt.Errorf("retab: workflowID is required")
	}
	if testID == "" {
		return fmt.Errorf("retab: testID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/workflows/"+url.PathEscape(workflowID)+"/block-tests/"+url.PathEscape(testID), nil, nil, nil, opts...)
}

func (s *WorkflowTestsService) Execute(ctx context.Context, request ExecuteBlockTestsRequest, opts ...RequestOption) (*ExecuteBlockTestsResponse, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	body := resourceFromJSON(request)
	delete(body, "WorkflowID")
	var result ExecuteBlockTestsResponse
	err := s.client.do(ctx, http.MethodPost, "/workflows/"+url.PathEscape(request.WorkflowID)+"/block-tests/execute", nil, body, &result, opts...)
	return &result, err
}

type WorkflowTestRunsService struct {
	client *Client
}

type BlockTestRunListResponse struct {
	Data []WorkflowTestRunRecord `json:"data"`
}

func (s *WorkflowTestRunsService) List(ctx context.Context, workflowID string, testID string, limit int, opts ...RequestOption) (*BlockTestRunListResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if testID == "" {
		return nil, fmt.Errorf("retab: testID is required")
	}
	if limit == 0 {
		limit = 20
	}
	query := url.Values{}
	query.Set("limit", fmt.Sprintf("%d", limit))
	var result BlockTestRunListResponse
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/block-tests/"+url.PathEscape(testID)+"/runs", query, nil, &result, opts...)
	return &result, err
}

func (s *WorkflowTestRunsService) Get(ctx context.Context, workflowID string, testID string, runID string, opts ...RequestOption) (*WorkflowTestRunRecord, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if testID == "" {
		return nil, fmt.Errorf("retab: testID is required")
	}
	if runID == "" {
		return nil, fmt.Errorf("retab: runID is required")
	}
	var result WorkflowTestRunRecord
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/block-tests/"+url.PathEscape(testID)+"/runs/"+url.PathEscape(runID), nil, nil, &result, opts...)
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

// WorkflowExperimentRunsService lists per-experiment runs and run content.
type WorkflowExperimentRunsService struct {
	client *Client
}

func newWorkflowExperimentsService(client *Client) *WorkflowExperimentsService {
	return &WorkflowExperimentsService{
		client: client,
		Runs:   &WorkflowExperimentRunsService{client: client},
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
	JobID             string    `json:"job_id,omitempty"`
	IsStale           bool      `json:"is_stale,omitempty"`
	SchemaDrift       string    `json:"schema_drift,omitempty"`
	SchemaDriftDetail string    `json:"schema_drift_detail,omitempty"`
}

// PreviousRunSummary is a slim summary of the previous completed run.
type PreviousRunSummary struct {
	RunID                 string   `json:"run_id"`
	DefinitionFingerprint string   `json:"definition_fingerprint,omitempty"`
	Score                 *float64 `json:"score,omitempty"`
}

// RunExperimentResponse is returned by POST /experiments/{id}/run.
type RunExperimentResponse struct {
	ExperimentID          string              `json:"experiment_id"`
	RunID                 string              `json:"run_id"`
	JobID                 string              `json:"job_id"`
	Status                string              `json:"status"`
	DefinitionFingerprint string              `json:"definition_fingerprint"`
	DocumentCount         int                 `json:"document_count"`
	NConsensus            int                 `json:"n_consensus"`
	PreviousRun           *PreviousRunSummary `json:"previous_run,omitempty"`
}

// CancelExperimentResponse is returned by POST /experiments/{id}/cancel.
type CancelExperimentResponse struct {
	Status string `json:"status"`
	RunID  string `json:"run_id"`
}

// ExperimentRunSummary is one row of GET /experiments/{id}/runs.
type ExperimentRunSummary struct {
	ID                    string         `json:"id"`
	ParentRunID           string         `json:"parent_run_id,omitempty"`
	BlockConfig           map[string]any `json:"block_config,omitempty"`
	DefinitionFingerprint string         `json:"definition_fingerprint"`
	DocumentsFingerprint  string         `json:"documents_fingerprint"`
	Status                string         `json:"status"`
	BlockKind             string         `json:"block_kind"`
	Score                 *float64       `json:"score,omitempty"`
	DocumentCount         int            `json:"document_count"`
	ErrorCount            int            `json:"error_count"`
	NConsensus            int            `json:"n_consensus"`
	CreatedAt             time.Time      `json:"created_at"`
	CompletedAt           *time.Time     `json:"completed_at,omitempty"`
	DurationMs            *int           `json:"duration_ms,omitempty"`
	JobID                 string         `json:"job_id,omitempty"`
}

// ExperimentRunListResponse is the GET /experiments/{id}/runs envelope.
type ExperimentRunListResponse struct {
	Runs []ExperimentRunSummary `json:"runs"`
}

// ExperimentJobResponse is the per-document execution record inside a run.
type ExperimentJobResponse struct {
	ID            string           `json:"id"`
	RunID         string           `json:"run_id"`
	ExperimentID  string           `json:"experiment_id"`
	DocumentID    string           `json:"document_id"`
	Status        string           `json:"status"`
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

// ExperimentContent wraps the per-job execution payload for one run.
type ExperimentContent struct {
	Jobs []ExperimentJobResponse `json:"jobs"`
}

// ExperimentContentResponse is the GET /experiments/{id}/content envelope.
type ExperimentContentResponse struct {
	ExperimentID string            `json:"experiment_id"`
	RunID        string            `json:"run_id"`
	Content      ExperimentContent `json:"content"`
}

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

// RunBatchResponse is POST /experiments/run-batch.
type RunBatchResponse struct {
	BlockID         string                  `json:"block_id"`
	ExperimentCount int                     `json:"experiment_count"`
	Runs            []RunExperimentResponse `json:"runs"`
}

// ExperimentMetricsResponse is the GET /experiments/{id}/metrics payload.
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
// fields you set are forwarded. Use NConsensusPtr to disambiguate the zero
// value from "leave unchanged".
type UpdateExperimentRequest struct {
	Name             string                              `json:"-"`
	NConsensus       *int                                `json:"-"`
	DocumentCaptures []ExperimentDocumentCaptureRequest  `json:"-"`
	Documents        []ExplicitExperimentDocumentRequest `json:"-"`
}

// RunExperimentOptions tunes POST /experiments/{id}/run.
type RunExperimentOptions struct {
	NConsensus      int
	RetryFailedOnly bool
}

// RunBatchExperimentsRequest is the body for POST /experiments/run-batch.
type RunBatchExperimentsRequest struct {
	WorkflowID string `json:"-"`
	BlockID    string `json:"block_id"`
	NConsensus int    `json:"n_consensus,omitempty"`
}

// GetExperimentMetricsParams gathers the GET /experiments/{id}/metrics
// query params.
type GetExperimentMetricsParams struct {
	View         string
	RunID        string
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
func (s *WorkflowExperimentsService) List(ctx context.Context, workflowID string, opts ...RequestOption) ([]ExperimentResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	var result []ExperimentResponse
	err := s.client.do(ctx, http.MethodGet, "/workflows/"+url.PathEscape(workflowID)+"/experiments", nil, nil, &result, opts...)
	return result, err
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

// Cancel stops the latest pending or running run for this experiment.
// Returns 404 if no in-flight run exists.
func (s *WorkflowExperimentsService) Cancel(ctx context.Context, workflowID, experimentID string, opts ...RequestOption) (*CancelExperimentResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	var result CancelExperimentResponse
	err := s.client.do(ctx, http.MethodPost,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID)+"/cancel",
		nil, map[string]any{}, &result, opts...)
	return &result, err
}

// GetMetrics returns experiment metrics for one of the four views.
//
// The response is shape-agnostic: callers branch on result["view"] or
// result["error"] to interpret the payload.
func (s *WorkflowExperimentsService) GetMetrics(ctx context.Context, workflowID, experimentID string, params *GetExperimentMetricsParams, opts ...RequestOption) (ExperimentMetricsResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	query := url.Values{}
	view := "summary"
	includePrior := true
	if params != nil {
		if params.View != "" {
			view = params.View
		}
		addQuery(query, "run_id", params.RunID)
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
	err := s.client.do(ctx, http.MethodGet,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID)+"/metrics",
		query, nil, &result, opts...)
	return result, err
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

// RunBatch triggers one new run for every experiment attached to a block.
// Returns 404 if no experiments are attached to that block.
func (s *WorkflowExperimentsService) RunBatch(ctx context.Context, request RunBatchExperimentsRequest, opts ...RequestOption) (*RunBatchResponse, error) {
	if request.WorkflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if request.BlockID == "" {
		return nil, fmt.Errorf("retab: blockID is required")
	}
	body := map[string]any{"block_id": request.BlockID}
	if request.NConsensus != 0 {
		body["n_consensus"] = request.NConsensus
	}
	var result RunBatchResponse
	err := s.client.do(ctx, http.MethodPost,
		"/workflows/"+url.PathEscape(request.WorkflowID)+"/experiments/run-batch",
		nil, body, &result, opts...)
	return &result, err
}

// Create triggers an experiment run with the current draft block config.
// Async — returns a job_id; poll with client.Jobs.
func (s *WorkflowExperimentRunsService) Create(ctx context.Context, workflowID, experimentID string, params *RunExperimentOptions, opts ...RequestOption) (*RunExperimentResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	body := map[string]any{}
	if params != nil {
		if params.NConsensus != 0 {
			body["n_consensus"] = params.NConsensus
		}
		if params.RetryFailedOnly {
			body["retry_failed_only"] = true
		}
	}
	var result RunExperimentResponse
	err := s.client.do(ctx, http.MethodPost,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID)+"/run",
		nil, body, &result, opts...)
	return &result, err
}

// List returns the run history for one experiment, newest first.
func (s *WorkflowExperimentRunsService) List(ctx context.Context, workflowID, experimentID string, opts ...RequestOption) (*ExperimentRunListResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	var result ExperimentRunListResponse
	err := s.client.do(ctx, http.MethodGet,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID)+"/runs",
		nil, nil, &result, opts...)
	return &result, err
}

// Get returns per-document execution content for one run.
// runID defaults to the latest run when empty.
func (s *WorkflowExperimentRunsService) Get(ctx context.Context, workflowID, experimentID, runID string, opts ...RequestOption) (*ExperimentContentResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("retab: workflowID is required")
	}
	if experimentID == "" {
		return nil, fmt.Errorf("retab: experimentID is required")
	}
	query := url.Values{}
	addQuery(query, "run_id", runID)
	var result ExperimentContentResponse
	err := s.client.do(ctx, http.MethodGet,
		"/workflows/"+url.PathEscape(workflowID)+"/experiments/"+url.PathEscape(experimentID)+"/content",
		query, nil, &result, opts...)
	return &result, err
}
