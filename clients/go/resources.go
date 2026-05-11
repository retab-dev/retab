package retab

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Resource map[string]any

type ListParams struct {
	Before   string
	After    string
	Limit    int
	Order    string
	Filename string
	FromDate *time.Time
	ToDate   *time.Time
}

type FilesService struct {
	client *Client
}

type File = Resource

type FileLink struct {
	DownloadURL string `json:"download_url"`
	ExpiresIn   string `json:"expires_in"`
	Filename    string `json:"filename"`
}

type PrepareUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256,omitempty"`
}

type CreateUploadResponse struct {
	FileID        string            `json:"fileId"`
	UploadURL     string            `json:"uploadUrl"`
	UploadMethod  string            `json:"uploadMethod"`
	UploadHeaders map[string]string `json:"uploadHeaders"`
	MIMEData      Resource          `json:"mimeData"`
	ExpiresAt     string            `json:"expiresAt"`
}

type ListFilesParams struct {
	ListParams
	MIMEType string
	SortBy   string
}

func (s *FilesService) PrepareUpload(filename string, contentType string, sizeBytes int64, sha256Value ...string) PreparedRequest {
	body := map[string]any{
		"filename":     filename,
		"content_type": contentType,
		"size_bytes":   sizeBytes,
	}
	if len(sha256Value) > 0 && sha256Value[0] != "" {
		body["sha256"] = sha256Value[0]
	}
	return PreparedRequest{
		URL:    "/files/upload",
		Method: http.MethodPost,
		Body:   body,
	}
}

func (s *FilesService) PrepareCompleteUpload(fileID string, sha256Value ...string) PreparedRequest {
	body := map[string]any{}
	if len(sha256Value) > 0 && sha256Value[0] != "" {
		body["sha256"] = sha256Value[0]
	}
	return PreparedRequest{
		URL:    "/files/upload/" + fileID + "/complete",
		Method: http.MethodPost,
		Body:   body,
	}
}

func (s *FilesService) CreateUpload(ctx context.Context, request PrepareUploadRequest, opts ...RequestOption) (*CreateUploadResponse, error) {
	if request.Filename == "" {
		return nil, fmt.Errorf("retab: filename is required")
	}
	if request.ContentType == "" {
		return nil, fmt.Errorf("retab: contentType is required")
	}
	if request.SizeBytes < 0 {
		return nil, fmt.Errorf("retab: sizeBytes must be >= 0")
	}
	var response CreateUploadResponse
	err := s.client.do(ctx, http.MethodPost, "/files/upload", nil, request, &response, opts...)
	return &response, err
}

func (s *FilesService) CompleteUpload(ctx context.Context, fileID string, sha256 string, opts ...RequestOption) (*MIMEData, error) {
	if fileID == "" {
		return nil, fmt.Errorf("retab: fileID is required")
	}
	body := map[string]string{}
	if sha256 != "" {
		body["sha256"] = sha256
	}
	var response MIMEData
	err := s.client.do(ctx, http.MethodPost, "/files/upload/"+url.PathEscape(fileID)+"/complete", nil, body, &response, opts...)
	return &response, err
}

func (s *FilesService) Upload(ctx context.Context, filePath string, opts ...RequestOption) (*MIMEData, error) {
	if filePath == "" {
		return nil, fmt.Errorf("retab: filePath is required")
	}
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") || strings.HasPrefix(filePath, "data:") {
		return nil, fmt.Errorf("retab: files.upload only accepts local file paths")
	}
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("retab: filePath must be a file")
	}
	sha256Hash, err := fileSHA256(filePath)
	if err != nil {
		return nil, err
	}
	filename := filepath.Base(filePath)
	contentType := contentTypeForFilename(filename)
	session, err := s.CreateUpload(ctx, PrepareUploadRequest{
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   stat.Size(),
		SHA256:      sha256Hash,
	}, opts...)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	method := session.UploadMethod
	if method == "" {
		method = http.MethodPut
	}
	req, err := http.NewRequestWithContext(ctx, method, session.UploadURL, file)
	if err != nil {
		return nil, err
	}
	for key, value := range session.UploadHeaders {
		req.Header.Set(key, value)
	}
	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("retab: direct file upload failed: %d %s", resp.StatusCode, string(body))
	}
	return s.CompleteUpload(ctx, session.FileID, sha256Hash, opts...)
}

func contentTypeForFilename(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".txt"):
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

func (s *FilesService) List(ctx context.Context, params *ListFilesParams, opts ...RequestOption) (*PaginatedList[File], error) {
	query := listQuery(nil)
	query.Set("limit", "10")
	query.Set("order", "desc")
	if params != nil {
		applyListParams(query, &params.ListParams)
		addQuery(query, "mime_type", params.MIMEType)
		if params.SortBy != "" {
			query.Set("sort_by", params.SortBy)
		} else {
			query.Set("sort_by", "created_at")
		}
	} else {
		query.Set("sort_by", "created_at")
	}
	var result PaginatedList[File]
	err := s.client.do(ctx, http.MethodGet, "/files", query, nil, &result, opts...)
	return &result, err
}

func (s *FilesService) Get(ctx context.Context, fileID string, opts ...RequestOption) (*File, error) {
	if fileID == "" {
		return nil, fmt.Errorf("retab: fileID is required")
	}
	var file File
	err := s.client.do(ctx, http.MethodGet, "/files/"+url.PathEscape(fileID), nil, nil, &file, opts...)
	return &file, err
}

func (s *FilesService) GetDownloadLink(ctx context.Context, fileID string, opts ...RequestOption) (*FileLink, error) {
	if fileID == "" {
		return nil, fmt.Errorf("retab: fileID is required")
	}
	var link FileLink
	err := s.client.do(ctx, http.MethodGet, "/files/"+url.PathEscape(fileID)+"/download-link", nil, nil, &link, opts...)
	return &link, err
}

type SchemasService struct {
	client *Client
}

type GenerateSchemaRequest struct {
	Documents []any  `json:"documents"`
	Model     string `json:"model,omitempty"`
}

func (s *SchemasService) Generate(ctx context.Context, request GenerateSchemaRequest, opts ...RequestOption) (*Resource, error) {
	if len(request.Documents) == 0 {
		return nil, fmt.Errorf("retab: documents are required")
	}
	var result Resource
	err := s.client.do(ctx, http.MethodPost, "/schemas/generate", nil, request, &result, opts...)
	return &result, err
}

type ExtractionsService struct {
	client *Client
}

type Extraction = Resource

type ExtractionCreateRequest struct {
	Document           any               `json:"document"`
	JSONSchema         any               `json:"json_schema"`
	Model              string            `json:"model"`
	ImageResolutionDPI int               `json:"image_resolution_dpi,omitempty"`
	NConsensus         int               `json:"n_consensus,omitempty"`
	Instructions       string            `json:"instructions,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	AdditionalMessages []Resource        `json:"additional_messages,omitempty"`
	BustCache          bool              `json:"bust_cache,omitempty"`
}

type ListExtractionsParams struct {
	ListParams
	OriginType string
	OriginID   string
	Metadata   map[string]string
}

func (s *ExtractionsService) Create(ctx context.Context, request ExtractionCreateRequest, opts ...RequestOption) (*Extraction, error) {
	if request.Document == nil {
		return nil, fmt.Errorf("retab: document is required")
	}
	if request.JSONSchema == nil {
		return nil, fmt.Errorf("retab: jsonSchema is required")
	}
	if request.Model == "" {
		return nil, fmt.Errorf("retab: model is required")
	}
	var result Extraction
	err := s.client.do(ctx, http.MethodPost, "/extractions", nil, request, &result, opts...)
	return &result, err
}

func (s *ExtractionsService) List(ctx context.Context, params *ListExtractionsParams, opts ...RequestOption) (*PaginatedList[Extraction], error) {
	query := listQuery(nil)
	if params != nil {
		applyListParams(query, &params.ListParams)
		addQuery(query, "origin_type", params.OriginType)
		addQuery(query, "origin_id", params.OriginID)
		addJSONQuery(query, "metadata", params.Metadata)
	}
	var result PaginatedList[Extraction]
	err := s.client.do(ctx, http.MethodGet, "/extractions", query, nil, &result, opts...)
	return &result, err
}

func (s *ExtractionsService) Get(ctx context.Context, extractionID string, opts ...RequestOption) (*Extraction, error) {
	if extractionID == "" {
		return nil, fmt.Errorf("retab: extractionID is required")
	}
	var result Extraction
	err := s.client.do(ctx, http.MethodGet, "/extractions/"+url.PathEscape(extractionID), nil, nil, &result, opts...)
	return &result, err
}

func (s *ExtractionsService) Sources(ctx context.Context, extractionID string, opts ...RequestOption) (*Resource, error) {
	if extractionID == "" {
		return nil, fmt.Errorf("retab: extractionID is required")
	}
	var result Resource
	err := s.client.do(ctx, http.MethodGet, "/extractions/"+url.PathEscape(extractionID)+"/sources", nil, nil, &result, opts...)
	return &result, err
}

func (s *ExtractionsService) Delete(ctx context.Context, extractionID string, opts ...RequestOption) error {
	if extractionID == "" {
		return fmt.Errorf("retab: extractionID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/extractions/"+url.PathEscape(extractionID), nil, nil, nil, opts...)
}

type SplitsService struct {
	client *Client
}

type Split = Resource

type SplitSubdocument struct {
	Name                   string `json:"name"`
	Description            string `json:"description"`
	AllowMultipleInstances bool   `json:"allow_multiple_instances"`
}

type SplitCreateRequest struct {
	Document     any                `json:"document"`
	Subdocuments []SplitSubdocument `json:"subdocuments"`
	Model        string             `json:"model"`
	NConsensus   int                `json:"n_consensus,omitempty"`
	BustCache    bool               `json:"bust_cache,omitempty"`
	Instructions string             `json:"instructions,omitempty"`
}

func (s *SplitsService) Create(ctx context.Context, request SplitCreateRequest, opts ...RequestOption) (*Split, error) {
	if request.Document == nil {
		return nil, fmt.Errorf("retab: document is required")
	}
	if len(request.Subdocuments) == 0 {
		return nil, fmt.Errorf("retab: subdocuments are required")
	}
	if request.Model == "" {
		return nil, fmt.Errorf("retab: model is required")
	}
	var result Split
	err := s.client.do(ctx, http.MethodPost, "/splits", nil, request, &result, opts...)
	return &result, err
}

func (s *SplitsService) Get(ctx context.Context, splitID string, opts ...RequestOption) (*Split, error) {
	if splitID == "" {
		return nil, fmt.Errorf("retab: splitID is required")
	}
	var result Split
	err := s.client.do(ctx, http.MethodGet, "/splits/"+url.PathEscape(splitID), nil, nil, &result, opts...)
	return &result, err
}

func (s *SplitsService) List(ctx context.Context, params *ListParams, opts ...RequestOption) (*PaginatedList[Split], error) {
	query := listQuery(params)
	var result PaginatedList[Split]
	err := s.client.do(ctx, http.MethodGet, "/splits", query, nil, &result, opts...)
	return &result, err
}

func (s *SplitsService) Delete(ctx context.Context, splitID string, opts ...RequestOption) error {
	if splitID == "" {
		return fmt.Errorf("retab: splitID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/splits/"+url.PathEscape(splitID), nil, nil, nil, opts...)
}

type ClassificationsService struct {
	client *Client
}

type Classification = Resource

type ClassificationCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ClassificationCreateRequest struct {
	Document     any                      `json:"document"`
	Categories   []ClassificationCategory `json:"categories"`
	Model        string                   `json:"model"`
	NConsensus   int                      `json:"n_consensus,omitempty"`
	BustCache    bool                     `json:"bust_cache,omitempty"`
	FirstNPages  int                      `json:"first_n_pages,omitempty"`
	Instructions string                   `json:"instructions,omitempty"`
}

func (s *ClassificationsService) Create(ctx context.Context, request ClassificationCreateRequest, opts ...RequestOption) (*Classification, error) {
	if request.Document == nil {
		return nil, fmt.Errorf("retab: document is required")
	}
	if len(request.Categories) == 0 {
		return nil, fmt.Errorf("retab: categories are required")
	}
	if request.Model == "" {
		return nil, fmt.Errorf("retab: model is required")
	}
	var result Classification
	err := s.client.do(ctx, http.MethodPost, "/classifications", nil, request, &result, opts...)
	return &result, err
}

func (s *ClassificationsService) Get(ctx context.Context, classificationID string, opts ...RequestOption) (*Classification, error) {
	if classificationID == "" {
		return nil, fmt.Errorf("retab: classificationID is required")
	}
	var result Classification
	err := s.client.do(ctx, http.MethodGet, "/classifications/"+url.PathEscape(classificationID), nil, nil, &result, opts...)
	return &result, err
}

func (s *ClassificationsService) List(ctx context.Context, params *ListParams, opts ...RequestOption) (*PaginatedList[Classification], error) {
	query := listQuery(params)
	var result PaginatedList[Classification]
	err := s.client.do(ctx, http.MethodGet, "/classifications", query, nil, &result, opts...)
	return &result, err
}

func (s *ClassificationsService) Delete(ctx context.Context, classificationID string, opts ...RequestOption) error {
	if classificationID == "" {
		return fmt.Errorf("retab: classificationID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/classifications/"+url.PathEscape(classificationID), nil, nil, nil, opts...)
}

type ParsesService struct {
	client *Client
}

type Parse = Resource

type ParseCreateRequest struct {
	Document           any    `json:"document"`
	Model              string `json:"model"`
	TableParsingFormat string `json:"table_parsing_format,omitempty"`
	ImageResolutionDPI int    `json:"image_resolution_dpi,omitempty"`
	Instructions       string `json:"instructions,omitempty"`
	BustCache          bool   `json:"bust_cache,omitempty"`
}

func (s *ParsesService) Create(ctx context.Context, request ParseCreateRequest, opts ...RequestOption) (*Parse, error) {
	if request.Document == nil {
		return nil, fmt.Errorf("retab: document is required")
	}
	if request.Model == "" {
		return nil, fmt.Errorf("retab: model is required")
	}
	var result Parse
	err := s.client.do(ctx, http.MethodPost, "/parses", nil, request, &result, opts...)
	return &result, err
}

func (s *ParsesService) Get(ctx context.Context, parseID string, opts ...RequestOption) (*Parse, error) {
	if parseID == "" {
		return nil, fmt.Errorf("retab: parseID is required")
	}
	var result Parse
	err := s.client.do(ctx, http.MethodGet, "/parses/"+url.PathEscape(parseID), nil, nil, &result, opts...)
	return &result, err
}

func (s *ParsesService) List(ctx context.Context, params *ListParams, opts ...RequestOption) (*PaginatedList[Parse], error) {
	query := listQuery(params)
	var result PaginatedList[Parse]
	err := s.client.do(ctx, http.MethodGet, "/parses", query, nil, &result, opts...)
	return &result, err
}

func (s *ParsesService) Delete(ctx context.Context, parseID string, opts ...RequestOption) error {
	if parseID == "" {
		return fmt.Errorf("retab: parseID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/parses/"+url.PathEscape(parseID), nil, nil, nil, opts...)
}

type EditsService struct {
	client    *Client
	Templates *EditTemplatesService
}

type Edit = Resource

func newEditsService(client *Client) *EditsService {
	service := &EditsService{client: client}
	service.Templates = &EditTemplatesService{client: client}
	return service
}

type EditCreateRequest struct {
	Instructions string `json:"instructions"`
	Document     any    `json:"document,omitempty"`
	TemplateID   string `json:"template_id,omitempty"`
	Model        string `json:"model,omitempty"`
	Color        string `json:"-"`
	BustCache    bool   `json:"bust_cache,omitempty"`
}

type ListEditsParams struct {
	ListParams
	TemplateID string
}

func (s *EditsService) Create(ctx context.Context, request EditCreateRequest, opts ...RequestOption) (*Edit, error) {
	if request.Instructions == "" {
		return nil, fmt.Errorf("retab: instructions are required")
	}
	body := resourceFromJSON(request)
	if request.Color != "" {
		body["config"] = Resource{"color": request.Color}
	}
	var result Edit
	err := s.client.do(ctx, http.MethodPost, "/edits", nil, body, &result, opts...)
	return &result, err
}

func (s *EditsService) Get(ctx context.Context, editID string, opts ...RequestOption) (*Edit, error) {
	if editID == "" {
		return nil, fmt.Errorf("retab: editID is required")
	}
	var result Edit
	err := s.client.do(ctx, http.MethodGet, "/edits/"+url.PathEscape(editID), nil, nil, &result, opts...)
	return &result, err
}

func (s *EditsService) List(ctx context.Context, params *ListEditsParams, opts ...RequestOption) (*PaginatedList[Edit], error) {
	query := listQuery(nil)
	if params != nil {
		applyListParams(query, &params.ListParams)
		addQuery(query, "template_id", params.TemplateID)
	}
	var result PaginatedList[Edit]
	err := s.client.do(ctx, http.MethodGet, "/edits", query, nil, &result, opts...)
	return &result, err
}

func (s *EditsService) Delete(ctx context.Context, editID string, opts ...RequestOption) error {
	if editID == "" {
		return fmt.Errorf("retab: editID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/edits/"+url.PathEscape(editID), nil, nil, nil, opts...)
}

type JobsService struct {
	client *Client
}

type JobStatus string

const (
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
	JobStatusExpired   JobStatus = "expired"
)

type Job = Resource

type JobCreateRequest struct {
	Endpoint string            `json:"endpoint"`
	Request  Resource          `json:"request"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type JobRetrieveParams struct {
	IncludeRequest  bool
	IncludeResponse bool
}

type ListJobsParams struct {
	Before           string
	After            string
	Limit            int
	Order            string
	ID               string
	Status           string
	Endpoint         string
	Source           string
	ProjectID        string
	WorkflowID       string
	WorkflowBlockID  string
	Model            string
	FilenameRegex    string
	FilenameContains string
	DocumentType     []string
	FromDate         string
	ToDate           string
	Metadata         map[string]string
	IncludeRequest   *bool
	IncludeResponse  *bool
}

type JobWaitForCompletionParams struct {
	PollInterval    time.Duration
	Timeout         time.Duration
	IncludeRequest  bool
	IncludeResponse bool
}

func (s *JobsService) Create(ctx context.Context, request JobCreateRequest, opts ...RequestOption) (*Job, error) {
	if request.Endpoint == "" {
		return nil, fmt.Errorf("retab: endpoint is required")
	}
	if request.Request == nil {
		return nil, fmt.Errorf("retab: request is required")
	}
	var result Job
	err := s.client.do(ctx, http.MethodPost, "/jobs", nil, request, &result, opts...)
	return &result, err
}

func (s *JobsService) Retrieve(ctx context.Context, jobID string, params *JobRetrieveParams, opts ...RequestOption) (*Job, error) {
	if jobID == "" {
		return nil, fmt.Errorf("retab: jobID is required")
	}
	query := url.Values{}
	query.Set("include_request", "false")
	query.Set("include_response", "false")
	if params != nil {
		query.Set("include_request", strconv.FormatBool(params.IncludeRequest))
		query.Set("include_response", strconv.FormatBool(params.IncludeResponse))
	}
	var result Job
	err := s.client.do(ctx, http.MethodGet, "/jobs/"+url.PathEscape(jobID), query, nil, &result, opts...)
	return &result, err
}

func (s *JobsService) RetrieveFull(ctx context.Context, jobID string, opts ...RequestOption) (*Job, error) {
	return s.Retrieve(ctx, jobID, &JobRetrieveParams{IncludeRequest: true, IncludeResponse: true}, opts...)
}

func (s *JobsService) WaitForCompletion(ctx context.Context, jobID string, params *JobWaitForCompletionParams, opts ...RequestOption) (*Job, error) {
	pollInterval := 2 * time.Second
	timeout := 10 * time.Minute
	includeRequest := false
	includeResponse := true
	if params != nil {
		if params.PollInterval > 0 {
			pollInterval = params.PollInterval
		}
		if params.Timeout > 0 {
			timeout = params.Timeout
		}
		includeRequest = params.IncludeRequest
		includeResponse = params.IncludeResponse
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		job, err := s.Retrieve(ctx, jobID, &JobRetrieveParams{}, opts...)
		if err != nil {
			return nil, err
		}
		if isTerminalJobStatus((*job)["status"]) {
			if includeRequest || includeResponse {
				return s.Retrieve(ctx, jobID, &JobRetrieveParams{
					IncludeRequest:  includeRequest,
					IncludeResponse: includeResponse,
				}, opts...)
			}
			return job, nil
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

func (s *JobsService) Cancel(ctx context.Context, jobID string, opts ...RequestOption) (*Job, error) {
	if jobID == "" {
		return nil, fmt.Errorf("retab: jobID is required")
	}
	var result Job
	err := s.client.do(ctx, http.MethodPost, "/jobs/"+url.PathEscape(jobID)+"/cancel", nil, nil, &result, opts...)
	return &result, err
}

func (s *JobsService) Retry(ctx context.Context, jobID string, opts ...RequestOption) (*Job, error) {
	if jobID == "" {
		return nil, fmt.Errorf("retab: jobID is required")
	}
	var result Job
	err := s.client.do(ctx, http.MethodPost, "/jobs/"+url.PathEscape(jobID)+"/retry", nil, nil, &result, opts...)
	return &result, err
}

func (s *JobsService) List(ctx context.Context, params *ListJobsParams, opts ...RequestOption) (*PaginatedList[Job], error) {
	query := url.Values{}
	query.Set("limit", "20")
	query.Set("order", "desc")
	if params != nil {
		addQuery(query, "before", params.Before)
		addQuery(query, "after", params.After)
		if params.Limit > 0 {
			query.Set("limit", strconv.Itoa(params.Limit))
		}
		addQuery(query, "order", params.Order)
		addQuery(query, "id", params.ID)
		addQuery(query, "status", params.Status)
		addQuery(query, "endpoint", params.Endpoint)
		addQuery(query, "source", params.Source)
		addQuery(query, "project_id", params.ProjectID)
		addQuery(query, "workflow_id", params.WorkflowID)
		addQuery(query, "workflow_block_id", params.WorkflowBlockID)
		addQuery(query, "model", params.Model)
		addQuery(query, "filename_regex", params.FilenameRegex)
		addQuery(query, "filename_contains", params.FilenameContains)
		for _, documentType := range params.DocumentType {
			query.Add("document_type", documentType)
		}
		addQuery(query, "from_date", params.FromDate)
		addQuery(query, "to_date", params.ToDate)
		addJSONQuery(query, "metadata", params.Metadata)
		if params.IncludeRequest != nil {
			query.Set("include_request", strconv.FormatBool(*params.IncludeRequest))
		}
		if params.IncludeResponse != nil {
			query.Set("include_response", strconv.FormatBool(*params.IncludeResponse))
		}
	}
	var result PaginatedList[Job]
	err := s.client.do(ctx, http.MethodGet, "/jobs", query, nil, &result, opts...)
	return &result, err
}

func listQuery(params *ListParams) url.Values {
	query := url.Values{}
	query.Set("limit", "10")
	query.Set("order", "desc")
	if params != nil {
		applyListParams(query, params)
	}
	return query
}

func applyListParams(query url.Values, params *ListParams) {
	addQuery(query, "before", params.Before)
	addQuery(query, "after", params.After)
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	addQuery(query, "order", params.Order)
	addQuery(query, "filename", params.Filename)
	if params.FromDate != nil {
		query.Set("from_date", params.FromDate.Format(time.RFC3339))
	}
	if params.ToDate != nil {
		query.Set("to_date", params.ToDate.Format(time.RFC3339))
	}
}

func addJSONQuery(query url.Values, key string, value any) {
	if value == nil {
		return
	}
	encoded, err := json.Marshal(value)
	if err == nil {
		query.Set(key, string(encoded))
	}
}

func resourceFromJSON(value any) Resource {
	bytes, err := json.Marshal(value)
	if err != nil {
		return Resource{}
	}
	var resource Resource
	if err := json.Unmarshal(bytes, &resource); err != nil {
		return Resource{}
	}
	return resource
}

func isTerminalJobStatus(status any) bool {
	switch status {
	case string(JobStatusCompleted), string(JobStatusFailed), string(JobStatusCancelled), string(JobStatusExpired):
		return true
	default:
		return false
	}
}

func fileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
