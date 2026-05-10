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
	client *Client
	Runs   *WorkflowRunsService
	Blocks *WorkflowBlocksService
	Edges  *WorkflowEdgesService
	Tests  *WorkflowTestsService
}

func newWorkflowsService(client *Client) *WorkflowsService {
	service := &WorkflowsService{client: client}
	service.Runs = &WorkflowRunsService{
		client: client,
		Steps:  &WorkflowRunStepsService{client: client},
	}
	service.Blocks = &WorkflowBlocksService{client: client}
	service.Edges = &WorkflowEdgesService{client: client}
	service.Tests = newWorkflowTestsService(client)
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
	Name                  *string  `json:"-"`
	Description           *string  `json:"-"`
	EmailSendersWhitelist []string `json:"-"`
	EmailDomainsWhitelist []string `json:"-"`
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
	if request.EmailSendersWhitelist != nil || request.EmailDomainsWhitelist != nil {
		emailTrigger := map[string]any{}
		if request.EmailSendersWhitelist != nil {
			emailTrigger["allowed_senders"] = request.EmailSendersWhitelist
		}
		if request.EmailDomainsWhitelist != nil {
			emailTrigger["allowed_domains"] = request.EmailDomainsWhitelist
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

type WaitForCompletionParams struct {
	PollInterval time.Duration
	Timeout      time.Duration
	OnStatus     func(*WorkflowRun)
}

func (s *WorkflowRunsService) WaitForCompletion(ctx context.Context, runID string, params *WaitForCompletionParams, opts ...RequestOption) (*WorkflowRun, error) {
	pollInterval := 2 * time.Second
	timeout := 10 * time.Minute
	var onStatus func(*WorkflowRun)
	if params != nil {
		if params.PollInterval > 0 {
			pollInterval = params.PollInterval
		}
		if params.Timeout > 0 {
			timeout = params.Timeout
		}
		onStatus = params.OnStatus
	}

	deadline := time.Now().Add(timeout)
	for {
		run, err := s.Get(ctx, runID, opts...)
		if err != nil {
			return nil, err
		}
		if onStatus != nil {
			onStatus(run)
		}
		if run.Terminal() || run.Lifecycle.Kind == "waiting_for_human" {
			return run, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("retab: workflow run %s did not complete within %s (last lifecycle.kind: %s)", runID, timeout, run.Lifecycle.Kind)
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

func (s *WorkflowRunsService) Wait(ctx context.Context, runID string, params *WaitForCompletionParams, opts ...RequestOption) (*WorkflowRun, error) {
	return s.WaitForCompletion(ctx, runID, params, opts...)
}

type CreateAndWaitWorkflowRunRequest struct {
	WorkflowID   string
	Documents    map[string]any
	JSONInputs   map[string]any
	PollInterval time.Duration
	Timeout      time.Duration
	OnStatus     func(*WorkflowRun)
}

func (s *WorkflowRunsService) CreateAndWait(ctx context.Context, request CreateAndWaitWorkflowRunRequest, opts ...RequestOption) (*WorkflowRun, error) {
	run, err := s.Create(ctx, CreateWorkflowRunRequest{
		WorkflowID: request.WorkflowID,
		Documents:  request.Documents,
		JSONInputs: request.JSONInputs,
	}, opts...)
	if err != nil {
		return nil, err
	}
	return s.WaitForCompletion(ctx, run.ID, &WaitForCompletionParams{
		PollInterval: request.PollInterval,
		Timeout:      request.Timeout,
		OnStatus:     request.OnStatus,
	}, opts...)
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

type WaitForBlockTestsParams struct {
	PollInterval time.Duration
	Timeout      time.Duration
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

func (s *WorkflowTestsService) WaitForCompletion(ctx context.Context, jobID string, params *WaitForBlockTestsParams, opts ...RequestOption) (*BlockTestBatchExecutionResult, error) {
	if jobID == "" {
		return nil, fmt.Errorf("retab: jobID is required")
	}
	pollInterval := 2 * time.Second
	timeout := 10 * time.Minute
	if params != nil {
		if params.PollInterval > 0 {
			pollInterval = params.PollInterval
		}
		if params.Timeout > 0 {
			timeout = params.Timeout
		}
	}
	deadline := time.Now().Add(timeout)
	for {
		job, err := s.client.Jobs.Retrieve(ctx, jobID, nil)
		if err != nil {
			return nil, err
		}
		status, _ := (*job)["status"].(string)
		if isTerminalJobStatus(status) {
			if status != string(JobStatusCompleted) {
				return nil, fmt.Errorf("retab: test batch job %s ended in status %q: %v", jobID, status, (*job)["error"])
			}
			response, _ := (*job)["response"].(map[string]any)
			body, _ := response["body"].(map[string]any)
			result := BlockTestBatchExecutionResult(body)
			return &result, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("retab: test batch job %s did not complete within %s", jobID, timeout)
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
