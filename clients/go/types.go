package retab

import (
	"bytes"
	"encoding/json"
	"time"
)

// PaginatedList is the common Retab pagination envelope.
type PaginatedList[T any] struct {
	Data         []T              `json:"data"`
	ListMetadata PaginationCursor `json:"list_metadata"`
	HasMore      bool             `json:"has_more,omitempty"`
	Total        int              `json:"total,omitempty"`
}

type PaginationCursor struct {
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

func (p *PaginatedList[T]) UnmarshalJSON(data []byte) error {
	// Most list endpoints return the {data, list_metadata} envelope, but a
	// few (e.g. GET /workflows/{id}/edges) return a bare JSON array with no
	// pagination wrapper. Decode either shape: if the first non-whitespace
	// byte is '[', unmarshal straight into Data and leave the pagination
	// fields zero-valued; otherwise decode the envelope as usual.
	//
	// Without this, `retab workflows edges list` failed with
	//   json: cannot unmarshal array into Go value of type retab.alias[...]
	// because the generic `type alias PaginatedList[T]` is a struct and the
	// wire payload was an array.
	if trimmed := bytes.TrimLeft(data, " \t\r\n"); len(trimmed) > 0 && trimmed[0] == '[' {
		if err := json.Unmarshal(data, &p.Data); err != nil {
			return err
		}
		if p.Data == nil {
			p.Data = []T{}
		}
		return nil
	}
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
// workflow start-document-block payloads, which use the backend's inline document form.
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
	Text        string         `json:"text,omitempty"`
	ArtifactRef map[string]any `json:"artifact_ref,omitempty"`
	Preview     map[string]any `json:"preview,omitempty"`
}

// StepArtifactRef points at a persisted resource produced by a workflow step.
type StepArtifactRef struct {
	Operation string `json:"operation"`
	ID        string `json:"id"`
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

type ResolvedSchemas struct {
	InputSchemas  map[string]any `json:"input_schemas"`
	OutputSchemas map[string]any `json:"output_schemas"`
	FieldRefDrift map[string]any `json:"field_ref_drift,omitempty"`
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

type WorkflowWithEntities struct {
	Workflow Workflow          `json:"workflow"`
	Blocks   []WorkflowBlock   `json:"blocks"`
	Edges    []WorkflowEdgeDoc `json:"edges"`
}

type WorkflowResolvedSchemasResponse struct {
	WorkflowID   string                     `json:"workflow_id"`
	DraftVersion string                     `json:"draft_version,omitempty"`
	Schemas      map[string]ResolvedSchemas `json:"schemas"`
}

type BlockResolvedSchemasResponse struct {
	WorkflowID   string          `json:"workflow_id"`
	BlockID      string          `json:"block_id"`
	DraftVersion string          `json:"draft_version,omitempty"`
	Schema       ResolvedSchemas `json:"schema"`
}

func (w *WorkflowWithEntities) UnmarshalJSON(data []byte) error {
	type alias WorkflowWithEntities
	aux := (*alias)(w)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if w.Blocks == nil {
		w.Blocks = []WorkflowBlock{}
	}
	if w.Edges == nil {
		w.Edges = []WorkflowEdgeDoc{}
	}
	return nil
}

type CancelWorkflowResponse struct {
	Run                WorkflowRun `json:"run"`
	CancellationStatus string      `json:"cancellation_status"`
}

// The v1 decision and managed-agent review types were removed in the hard
// cutover to the review overlay. See the Review* types below and
// WorkflowReviewsService.

type WorkflowRunExportResponse struct {
	CSVData string `json:"csv_data"`
	Rows    int    `json:"rows"`
	Columns int    `json:"columns"`
}

// --- review overlay (workflows.reviews) ----------------------------
//
// The review overlay is the versioned sidecar attached to a block run awaiting
// review. Version ids are content hashes and versions are keyed by id.
// Actor symmetry is a hard rule — model, agent, and human are one shape and
// Kind is data, never branched on.

// ReviewActor is who authored a version or made a decision.
type ReviewActor struct {
	Kind        string `json:"kind"` // model | agent | human
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ReviewOutputVersion is one immutable, full JSON snapshot of a block output.
type ReviewOutputVersion struct {
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
	DecidedBy ReviewActor `json:"decided_by"`
	DecidedAt time.Time   `json:"decided_at"`
	Reason    *string     `json:"reason"`
}

// ReviewOverlay is the full review sidecar for one reviewed block run.
type ReviewOverlay struct {
	ID                string                         `json:"_id"`
	WorkflowID        string                         `json:"workflow_id"`
	WorkflowVersionID string                         `json:"workflow_version_id"`
	WorkflowRunID     string                         `json:"workflow_run_id"`
	BlockID           string                         `json:"block_id"`
	BlockRunID        string                         `json:"block_run_id"`
	BlockType         string                         `json:"block_type"`
	TriggeredBy       map[string]any                 `json:"triggered_by"`
	AwaitingSince     time.Time                      `json:"awaiting_since"`
	Priority          int                            `json:"priority"`
	VersionsByID      map[string]ReviewOutputVersion `json:"versions_by_id"`
	Decision          *ReviewDecisionRecord          `json:"decision"`
}

// ReviewQueueItem is the lightweight review-queue projection.
type ReviewQueueItem struct {
	ID                string         `json:"_id"`
	WorkflowID        string         `json:"workflow_id"`
	WorkflowVersionID string         `json:"workflow_version_id"`
	WorkflowRunID     string         `json:"workflow_run_id"`
	BlockID           string         `json:"block_id"`
	BlockRunID        string         `json:"block_run_id"`
	BlockType         string         `json:"block_type"`
	TriggeredBy       map[string]any `json:"triggered_by"`
	AwaitingSince     time.Time      `json:"awaiting_since"`
	Priority          int            `json:"priority"`
}

// ReviewQueueResponse is one page of the review queue.
//
// Deprecated: use PaginatedList[ReviewQueueItem] returned by
// WorkflowReviewsService.List. ReviewQueueResponse is kept only as a type
// alias for source compatibility — the live backend emits the canonical
// {data, list_metadata} envelope (no has_more boolean), so paginating
// requires the cursor fields in PaginationCursor (Before/After).
type ReviewQueueResponse = PaginatedList[ReviewQueueItem]

// SubmitReviewDecisionResponse is the result of a verdict submission.
//
// SubmissionStatus reflects whether the decision write was accepted on the
// server. ResumeStatus reflects whether the downstream workflow actually
// resumed — a "failed" resume means the decision is committed but the
// workflow did not pick up the signal (e.g. Temporal returned an error); the
// associated ResumeError carries the underlying message for diagnostics. When
// the resume signal fails on a fresh write, SubmissionStatus is upgraded to
// "accepted_pending_resume" — the decision is durable but the workflow has
// not advanced past the gate yet; a reconcile loop will retry.
type SubmitReviewDecisionResponse struct {
	SubmissionStatus string        `json:"submission_status"` // accepted | already_applied | accepted_pending_resume
	Review           ReviewOverlay `json:"review"`
	ResumeStatus     string        `json:"resume_status,omitempty"` // resumed | skipped | failed
	ResumeError      *string       `json:"resume_error,omitempty"`
}

// Submission status string constants returned by Approve / Reject.
const (
	// SubmissionStatusAccepted means the decision was written and the
	// downstream workflow was successfully signalled to resume.
	SubmissionStatusAccepted = "accepted"
	// SubmissionStatusAlreadyApplied means the same (verdict, version_id) was
	// already on file — the idempotency mechanism, safe to retry.
	SubmissionStatusAlreadyApplied = "already_applied"
	// SubmissionStatusAcceptedPendingResume means the decision write succeeded
	// but the Temporal resume signal failed. The reconcile loop will retry;
	// inspect ResumeStatus and ResumeError on the response for diagnostics.
	SubmissionStatusAcceptedPendingResume = "accepted_pending_resume"
)

// Resume status string constants reported on Approve / Reject responses.
const (
	ResumeStatusResumed = "resumed"
	ResumeStatusSkipped = "skipped"
	ResumeStatusFailed  = "failed"
)
