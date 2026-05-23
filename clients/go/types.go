package retab

import (
	"context"
	"encoding/json"
	"time"
)

// PaginatedList is the common Retab pagination envelope.
//
// Auto-pagination is opt-in: list methods on the SDK populate the unexported
// fetchNext closure so callers can walk every page without manually copying
// the After cursor. PaginatedList values built by hand or by direct
// json.Unmarshal still iterate correctly via AutoPaging — they just stop at
// the current page because fetchNext is nil.
type PaginatedList[T any] struct {
	Data         []T              `json:"data"`
	ListMetadata PaginationCursor `json:"list_metadata"`
	HasMore      bool             `json:"has_more,omitempty"`
	Total        int              `json:"total,omitempty"`

	// fetchNext is set by the list helpers in client.go and re-issues the
	// originating request with the After cursor advanced. Zero-value means
	// "no next page available from this PaginatedList" — AutoPaging only
	// iterates Data, and NextPage returns (nil, nil).
	fetchNext func(ctx context.Context, after string) (*PaginatedList[T], error)
}

type PaginationCursor struct {
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// HasNextPage reports whether the server returned an after cursor on this
// page. It is the canonical truncation signal across every Retab list
// endpoint — callers should branch on this rather than inspecting HasMore /
// Total, which are not universally populated.
func (p *PaginatedList[T]) HasNextPage() bool {
	return p.ListMetadata.After != ""
}

// NextPage fetches the next page using the same parameters as the originating
// request, with the After cursor advanced. Returns (nil, nil) when no further
// pages are available — either because the server reported the last page
// (After == "") or because this PaginatedList was constructed without a
// fetchNext closure (e.g. unmarshalled directly from JSON).
func (p *PaginatedList[T]) NextPage(ctx context.Context) (*PaginatedList[T], error) {
	if p.fetchNext == nil || p.ListMetadata.After == "" {
		return nil, nil
	}
	return p.fetchNext(ctx, p.ListMetadata.After)
}

// AutoPaging walks every item across every page, calling yield once per item.
// Returning a non-nil error from yield short-circuits iteration and the same
// error is returned verbatim — wrap it with errors.New / fmt.Errorf if you
// want to distinguish "stop" from "real error" at the call site.
//
// The closure-based shape is intentionally Go-1.22-compatible. Once the SDK
// moves to Go 1.23+ we can add a sibling method returning iter.Seq2[T, error]
// without breaking this API.
//
// When fetchNext is nil (PaginatedList built by hand or decoded directly),
// only the current page is iterated. When fetchNext is set, AutoPaging keeps
// fetching pages until the server reports After == "" or a page fetch fails.
func (p *PaginatedList[T]) AutoPaging(ctx context.Context, yield func(item T) error) error {
	current := p
	for {
		for _, item := range current.Data {
			if err := yield(item); err != nil {
				return err
			}
		}
		if current.fetchNext == nil || current.ListMetadata.After == "" {
			return nil
		}
		next, err := current.fetchNext(ctx, current.ListMetadata.After)
		if err != nil {
			return err
		}
		if next == nil {
			return nil
		}
		current = next
	}
}

// UnmarshalJSON normalizes a nil `data` field to an empty slice so callers can
// iterate `result.Data` without a nil check. Every list endpoint in the spec
// returns the canonical {"data": [...], "list_metadata": {...}} envelope;
// legacy bare-array responses (e.g. the old `GET /v1/workflows/{id}/edges`,
// which moved to `/v1/workflows/edges` with a proper envelope) no longer exist.
func (p *PaginatedList[T]) UnmarshalJSON(data []byte) error {
	type alias PaginatedList[T]
	aux := (*alias)(p)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if p.Data == nil {
		p.Data = []T{}
	}
	return nil
}

// MIMEData mirrors the Node SDK document shape. URL may be a data URI,
// an HTTPS URL, or a Retab storage URL. Content/MIMEType are retained for
// workflow start_document-block payloads, which use the backend's inline document form.
type MIMEData struct {
	Filename string `json:"filename,omitempty"`
	Content  string `json:"content,omitempty"`
	URL      string `json:"url,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
}

func (m MIMEData) MarshalJSON() ([]byte, error) {
	type publicMIMEData struct {
		Filename string `json:"filename,omitempty"`
		URL      string `json:"url,omitempty"`
	}
	return json.Marshal(publicMIMEData{Filename: m.Filename, URL: m.URL})
}

// FileRef references a document stored by Retab.
type FileRef struct {
	ID       string `json:"id"`
	Filename string `json:"filename,omitempty"`
	Content  string `json:"content,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
}

// HandlePayload is the typed payload attached to a workflow handle.
type HandlePayload struct {
	Type        string         `json:"type"`
	Document    *FileRef       `json:"document,omitempty"`
	Data        any            `json:"data,omitempty"`
	ArtifactRef map[string]any `json:"artifact_ref,omitempty"`
	Preview     map[string]any `json:"preview,omitempty"`
}

// StepArtifactRef points at a persisted resource produced by a workflow step.
type StepArtifactRef struct {
	Operation string `json:"operation"`
	ID        string `json:"id"`
}

// BlockExecutionLifecycle is the discriminated terminal state of a block execution.
// One of:
//
//   - {"status": "completed"}
//   - {"status": "error",   "message": "..."}
//   - {"status": "skipped", "reason":  "..."}
//
// Modelled as a map for the same reason “StepLifecycle“ is: discriminated
// payloads evolve and a typed struct here would force a Go SDK release for
// every additive field. Consumers should branch on “Status“ and read the
// variant-specific keys (“Message“ / “Reason“) directly.
type BlockExecutionLifecycle map[string]any

// StoredBlockExecution is the persisted result of replaying one workflow block
// against a prior run's inputs through POST /workflows/blocks/executions.
//
// Terminal state is carried by the discriminated “Lifecycle“ field. The
// legacy flat “Success“ / “Error“ / “Skipped“ fields were removed in
// the hard cutover; consumers must read “Lifecycle["status"]“.
type StoredBlockExecution struct {
	ID                  string                    `json:"id"`
	WorkflowID          string                    `json:"workflow_id"`
	RunID               string                    `json:"run_id"`
	BlockID             string                    `json:"block_id"`
	BlockType           string                    `json:"block_type"`
	Lifecycle           BlockExecutionLifecycle   `json:"lifecycle"`
	HandleInputs        map[string]any            `json:"handle_inputs,omitempty"`
	Artifact            *StepArtifactRef          `json:"artifact,omitempty"`
	HandleOutputs       map[string]any            `json:"handle_outputs,omitempty"`
	RoutingDecision     []string                  `json:"routing_decision,omitempty"`
	DurationMS          *float64                  `json:"duration_ms,omitempty"`
	CreatedAt           *time.Time                `json:"created_at,omitempty"`
	BlockConfig         map[string]any            `json:"block_config,omitempty"`
	StepID              string                    `json:"step_id,omitempty"`
	AvailableIterations []BlockExecutionIteration `json:"available_iterations,omitempty"`
	Raw                 json.RawMessage           `json:"-"`
}

func (s *StoredBlockExecution) UnmarshalJSON(data []byte) error {
	type alias StoredBlockExecution
	aux := (*alias)(s)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	s.Raw = append(s.Raw[:0], data...)
	return nil
}

// BlockExecutionIteration is one iteration step available for a block execution.
type BlockExecutionIteration struct {
	StepID         string `json:"step_id,omitempty"`
	IterationIndex *int   `json:"iteration_index,omitempty"`
	Label          string `json:"label,omitempty"`
}

// ContainerContextData identifies loop/container context for a step.
type ContainerContextData struct {
	ContainerID       string `json:"container_id"`
	Iteration         int    `json:"iteration"`
	IsParallel        bool   `json:"is_parallel"`
	ParallelItemIndex *int   `json:"parallel_item_index,omitempty"`
}

// StepLifecycle is intentionally permissive because lifecycle payloads evolve.
type StepLifecycle map[string]any

// WorkflowRunStep is a persisted step summary for a workflow run.
type WorkflowRunStep struct {
	RunID          string                   `json:"run_id"`
	BlockID        string                   `json:"block_id"`
	StepID         string                   `json:"step_id"`
	BlockType      string                   `json:"block_type"`
	BlockLabel     string                   `json:"block_label"`
	Lifecycle      StepLifecycle            `json:"lifecycle"`
	StartedAt      *time.Time               `json:"started_at,omitempty"`
	CompletedAt    *time.Time               `json:"completed_at,omitempty"`
	Model          string                   `json:"model,omitempty"`
	LoopContainers []ContainerContextData   `json:"loop_containers,omitempty"`
	Artifact       *StepArtifactRef         `json:"artifact,omitempty"`
	HandleInputs   map[string]HandlePayload `json:"handle_inputs"`
	HandleOutputs  map[string]HandlePayload `json:"handle_outputs"`
	RetryCount     int                      `json:"retry_count"`
	CreatedAt      *time.Time               `json:"created_at,omitempty"`
}

// UnmarshalJSON normalizes null handle maps to empty maps.
func (s *WorkflowRunStep) UnmarshalJSON(data []byte) error {
	type alias WorkflowRunStep
	aux := (*alias)(s)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if s.HandleInputs == nil {
		s.HandleInputs = map[string]HandlePayload{}
	}
	if s.HandleOutputs == nil {
		s.HandleOutputs = map[string]HandlePayload{}
	}
	if s.LoopContainers == nil {
		s.LoopContainers = []ContainerContextData{}
	}
	return nil
}

// StepExecutionResponse is the full execution record for one workflow step.
type StepExecutionResponse struct {
	BlockID        string                   `json:"block_id"`
	StepID         string                   `json:"step_id"`
	BlockType      string                   `json:"block_type"`
	BlockLabel     string                   `json:"block_label"`
	Lifecycle      StepLifecycle            `json:"lifecycle"`
	StartedAt      *time.Time               `json:"started_at,omitempty"`
	CompletedAt    *time.Time               `json:"completed_at,omitempty"`
	Model          string                   `json:"model,omitempty"`
	LoopContainers []ContainerContextData   `json:"loop_containers,omitempty"`
	Artifact       *StepArtifactRef         `json:"artifact,omitempty"`
	HandleOutputs  map[string]HandlePayload `json:"handle_outputs"`
	HandleInputs   map[string]HandlePayload `json:"handle_inputs"`
}

// UnmarshalJSON normalizes null handle maps to empty maps.
func (s *StepExecutionResponse) UnmarshalJSON(data []byte) error {
	type alias StepExecutionResponse
	aux := (*alias)(s)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if s.HandleInputs == nil {
		s.HandleInputs = map[string]HandlePayload{}
	}
	if s.HandleOutputs == nil {
		s.HandleOutputs = map[string]HandlePayload{}
	}
	if s.LoopContainers == nil {
		s.LoopContainers = []ContainerContextData{}
	}
	return nil
}

// JSONOutput returns the data for the given JSON output handle, if present.
func (s StepExecutionResponse) JSONOutput(handleID string) any {
	payload, ok := s.HandleOutputs[handleID]
	if !ok || payload.Type != "json" {
		return nil
	}
	return payload.Data
}

// ExtractedData returns the default JSON output handle, if present.
func (s StepExecutionResponse) ExtractedData() any {
	return s.JSONOutput("output-json-0")
}

type RunLifecycle struct {
	Status             string   `json:"status"`
	WaitingForBlockIDs []string `json:"waiting_for_block_ids,omitempty"`
	Message            string   `json:"message,omitempty"`
}

type RunTiming struct {
	CreatedAt                  *time.Time `json:"created_at,omitempty"`
	StartedAt                  *time.Time `json:"started_at,omitempty"`
	CompletedAt                *time.Time `json:"completed_at,omitempty"`
	DurationMs                 *int       `json:"duration_ms,omitempty"`
	ReviewWaitingStartedAt     *time.Time `json:"review_waiting_started_at,omitempty"`
	AccumulatedReviewWaitingMS int        `json:"accumulated_review_waiting_ms,omitempty"`
}

type WorkflowSnapshotRef struct {
	WorkflowID       string `json:"workflow_id"`
	VersionID        string `json:"version_id"`
	NameAtRunTime    string `json:"name_at_run_time"`
	RequestedVersion string `json:"requested_version,omitempty"`
}

type RunTrigger struct {
	Type        string `json:"type"`
	UserID      string `json:"user_id,omitempty"`
	APIKeyID    string `json:"api_key_id,omitempty"`
	ScheduleID  string `json:"schedule_id,omitempty"`
	WebhookID   string `json:"webhook_id,omitempty"`
	Sender      string `json:"sender,omitempty"`
	Subject     string `json:"subject,omitempty"`
	ParentRunID string `json:"parent_run_id,omitempty"`
}

type RunInputs struct {
	Documents map[string]FileRef `json:"documents"`
	JSONData  map[string]any     `json:"json_data"`
}

// WorkflowRun is a workflow execution.
type WorkflowRun struct {
	ID         string              `json:"id"`
	WorkflowID string              `json:"workflow_id,omitempty"`
	Workflow   WorkflowSnapshotRef `json:"workflow"`
	Trigger    RunTrigger          `json:"trigger"`
	Lifecycle  RunLifecycle        `json:"lifecycle"`
	Timing     RunTiming           `json:"timing"`
	Inputs     RunInputs           `json:"inputs"`
	Raw        json.RawMessage     `json:"-"`
}

func (r *WorkflowRun) UnmarshalJSON(data []byte) error {
	type alias WorkflowRun
	aux := (*alias)(r)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if r.WorkflowID == "" {
		r.WorkflowID = r.Workflow.WorkflowID
	}
	if r.Inputs.Documents == nil {
		r.Inputs.Documents = map[string]FileRef{}
	}
	if r.Inputs.JSONData == nil {
		r.Inputs.JSONData = map[string]any{}
	}
	r.Raw = append(r.Raw[:0], data...)
	return nil
}

// MarshalJSON re-emits the verbatim server payload captured in Raw on
// decode, so a server-side field projection (?fields=id) survives a
// decode→encode round-trip. Without it the typed struct's zero-valued
// fields re-inflate the output with empty workflow{}, trigger{}, ...
// objects the server never sent. A run constructed in code has no Raw
// and falls back to normal struct encoding.
func (r WorkflowRun) MarshalJSON() ([]byte, error) {
	if len(r.Raw) > 0 {
		return r.Raw, nil
	}
	type alias WorkflowRun
	return json.Marshal(alias(r))
}

// Completed reports whether the run reached the completed terminal state.
func (r WorkflowRun) Completed() bool {
	return r.Lifecycle.Status == "completed"
}

// Terminal reports whether the run is no longer actively executing.
func (r WorkflowRun) Terminal() bool {
	switch r.Lifecycle.Status {
	case "completed", "error", "cancelled":
		return true
	default:
		return false
	}
}

// Workflow is the top-level workflow resource.
type Workflow struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Published    *WorkflowPublished   `json:"published,omitempty"`
	EmailTrigger WorkflowEmailTrigger `json:"email_trigger"`
	CreatedAt    *time.Time           `json:"created_at,omitempty"`
	UpdatedAt    *time.Time           `json:"updated_at,omitempty"`
	Raw          json.RawMessage      `json:"-"`
}

func (w *Workflow) UnmarshalJSON(data []byte) error {
	type alias Workflow
	aux := (*alias)(w)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	w.Raw = append(w.Raw[:0], data...)
	return nil
}

// MarshalJSON re-emits the verbatim server payload captured in Raw on
// decode — see WorkflowRun.MarshalJSON for the rationale. A workflow
// constructed in code has no Raw and falls back to normal struct
// encoding.
func (w Workflow) MarshalJSON() ([]byte, error) {
	if len(w.Raw) > 0 {
		return w.Raw, nil
	}
	type alias Workflow
	return json.Marshal(alias(w))
}

type WorkflowPublished struct {
	VersionID   string     `json:"version_id,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type WorkflowEmailTrigger struct {
	AllowedSenders []string `json:"allowed_senders"`
	AllowedDomains []string `json:"allowed_domains"`
}

type WorkflowBlock struct {
	ID               string            `json:"id"`
	WorkflowID       string            `json:"workflow_id"`
	DraftVersion     string            `json:"draft_version,omitempty"`
	Type             string            `json:"type"`
	Label            string            `json:"label"`
	PositionX        float64           `json:"position_x"`
	PositionY        float64           `json:"position_y"`
	Width            *float64          `json:"width,omitempty"`
	Height           *float64          `json:"height,omitempty"`
	Config           map[string]any    `json:"config,omitempty"`
	FieldRefSnapshot map[string]string `json:"field_ref_snapshot,omitempty"`
	ParentID         string            `json:"parent_id,omitempty"`
	UpdatedAt        *time.Time        `json:"updated_at,omitempty"`
}

type WorkflowEdgeDoc struct {
	ID           string     `json:"id"`
	WorkflowID   string     `json:"workflow_id"`
	DraftVersion string     `json:"draft_version,omitempty"`
	SourceBlock  string     `json:"source_block"`
	TargetBlock  string     `json:"target_block"`
	SourceHandle string     `json:"source_handle,omitempty"`
	TargetHandle string     `json:"target_handle,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type CancelWorkflowResponse struct {
	Run                WorkflowRun `json:"run"`
	CancellationStatus string      `json:"cancellation_status"`
}

// The v1 decision and managed-agent review types were removed in the hard
// cutover to reviews. See the Review* types below and WorkflowReviewsService.

type WorkflowRunExportResponse struct {
	CSVData string `json:"csv_data"`
	Rows    int    `json:"rows"`
	Columns int    `json:"columns"`
}

// --- reviews (workflows.reviews) -----------------------------------
//
// A review is a first-class resource addressed by review id. Workflow/run/block
// identifiers are context and list filters only. Version ids are content hashes
// exposed through the flat review versions resource.
// Actor symmetry is a hard rule — model, agent, and human are one shape and
// Kind is data, never branched on.

// ReviewActor is who authored a version or made a decision.
type ReviewActor struct {
	Kind        string `json:"kind"` // model | agent | human
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ReviewVersion is one immutable, full JSON snapshot of a block output.
type ReviewVersion struct {
	ID        string         `json:"id"`
	ReviewID  string         `json:"review_id"`
	ParentID  *string        `json:"parent_id"`
	Author    ReviewActor    `json:"author"`
	Snapshot  map[string]any `json:"snapshot"`
	Note      *string        `json:"note"`
	CreatedAt time.Time      `json:"created_at"`
}

// ReviewDecisionRecord is a typed verdict cast against one version.
type ReviewDecisionRecord struct {
	Verdict   string      `json:"verdict"` // approved | rejected
	VersionID string      `json:"version_id"`
	Author    ReviewActor `json:"author"`
	DecidedAt time.Time   `json:"decided_at"`
	Reason    *string     `json:"reason"`
}

// Review is the review metadata and terminal decision for one reviewed block run.
type Review struct {
	ID                string                `json:"id"`
	WorkflowID        string                `json:"workflow_id"`
	WorkflowVersionID string                `json:"workflow_version_id"`
	WorkflowRunID     string                `json:"workflow_run_id"`
	BlockID           string                `json:"block_id"`
	StepID            string                `json:"step_id"`
	ParentStepID      *string               `json:"parent_step_id"`
	IterationKey      *string               `json:"iteration_key"`
	BlockType         string                `json:"block_type"`
	TriggeredBy       map[string]any        `json:"triggered_by"`
	CreatedAt         time.Time             `json:"created_at"`
	Decision          *ReviewDecisionRecord `json:"decision"`
}

// WorkflowReviewQueue is one page of the review queue.
//
// Deprecated: use PaginatedList[Review] returned by
// WorkflowReviewsService.List. WorkflowReviewQueue is kept only as a type
// alias for source compatibility — the live backend emits the canonical
// {data, list_metadata} envelope (no has_more boolean), so paginating
// requires the cursor fields in PaginationCursor (Before/After).
type WorkflowReviewQueue = PaginatedList[Review]

// SubmitWorkflowReviewDecisionResponse is the result of a verdict submission.
//
// SubmissionStatus reflects whether the decision write was accepted on the
// server. ResumeStatus reflects whether the downstream workflow actually
// resumed. A "pending" resume means the decision is committed but immediate
// delivery did not complete; the associated ResumeError carries diagnostic
// context while server-side reconciliation retries delivery.
type SubmitWorkflowReviewDecisionResponse struct {
	SubmissionStatus string  `json:"submission_status"` // accepted
	Review           Review  `json:"review"`
	ResumeStatus     string  `json:"resume_status,omitempty"` // pending | resumed | skipped
	ResumeError      *string `json:"resume_error,omitempty"`
}

// Submission status string constants returned by Approve / Reject.
// Kept aligned with “SubmissionStatus“ Literal in
// “backend/.../workflows/reviews/api_models.py“.
const (
	// SubmissionStatusAccepted means the decision was written.
	SubmissionStatusAccepted = "accepted"
	// SubmissionStatusAlreadyApplied means the exact same decision was
	// already recorded; the call is idempotent and successful.
	SubmissionStatusAlreadyApplied = "already_applied"
	// SubmissionStatusConflict means the review already has a *different*
	// decision (e.g. you tried to reject an already-approved review). The
	// submission was NOT applied; the existing decision stands.
	SubmissionStatusConflict = "conflict"
)

// Resume status string constants reported on Approve / Reject responses.
const (
	ResumeStatusResumed = "resumed"
	ResumeStatusSkipped = "skipped"
	ResumeStatusPending = "pending"
)
