package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
)

// blockSpec / edgeSpec mirror the SDK structs with the same optional
// pointer fields. The SDK now exposes `PositionX`, `PositionY`, `Width`,
// `Height`, `Label`, and the edge `SourceHandle` / `TargetHandle` fields
// as optional pointers; the ASCII renderer treats nil-or-zero as "unset",
// so each test case wraps real values with `ptr(...)` and leaves
// untouched fields nil.
type blockSpec struct {
	ID        string
	Type      retab.WorkflowBlockType
	Label     *string
	PositionX *float64
	PositionY *float64
}

func makeBlock(spec blockSpec) retab.WorkflowBlock {
	return retab.WorkflowBlock{
		ID:        spec.ID,
		Type:      spec.Type,
		Label:     spec.Label,
		PositionX: spec.PositionX,
		PositionY: spec.PositionY,
	}
}

func makeBlocks(specs ...blockSpec) []retab.WorkflowBlock {
	out := make([]retab.WorkflowBlock, len(specs))
	for i, s := range specs {
		out[i] = makeBlock(s)
	}
	return out
}

type edgeSpec struct {
	ID           string
	SourceBlock  string
	TargetBlock  string
	SourceHandle *string
	TargetHandle *string
}

func makeEdge(spec edgeSpec) retab.WorkflowEdgeDoc {
	return retab.WorkflowEdgeDoc{
		ID:           spec.ID,
		SourceBlock:  spec.SourceBlock,
		TargetBlock:  spec.TargetBlock,
		SourceHandle: spec.SourceHandle,
		TargetHandle: spec.TargetHandle,
	}
}

func makeEdges(specs ...edgeSpec) []retab.WorkflowEdgeDoc {
	out := make([]retab.WorkflowEdgeDoc, len(specs))
	for i, s := range specs {
		out[i] = makeEdge(s)
	}
	return out
}

func renderWorkflowASCIIViewString(t *testing.T, workflow *workflowGraph) string {
	t.Helper()
	var buf bytes.Buffer
	if err := renderWorkflowASCIIView(&buf, workflow); err != nil {
		t.Fatalf("render workflow ascii view: %v", err)
	}
	return buf.String()
}

func lineAndColumn(t *testing.T, out string, needle string) (int, int) {
	t.Helper()
	for lineNo, line := range strings.Split(out, "\n") {
		if col := strings.Index(line, needle); col >= 0 {
			return lineNo, col
		}
	}
	t.Fatalf("missing %q in output:\n%s", needle, out)
	return 0, 0
}

func TestRenderWorkflowASCIIViewLinearPositionedWorkflow(t *testing.T) {
	workflow := &workflowGraph{
		Workflow: retab.Workflow{ID: "wf_linear", Name: "Linear invoice flow"},
		Blocks: makeBlocks(
			blockSpec{ID: "start", Type: "start", Label: ptr("Start"), PositionX: ptr(0.0), PositionY: ptr(0.0)},
			blockSpec{ID: "extract", Type: "extract", Label: ptr("Extract totals"), PositionX: ptr(300.0), PositionY: ptr(0.0)},
		),
		Edges: makeEdges(
			edgeSpec{ID: "edge_start_extract", SourceBlock: "start", TargetBlock: "extract"},
		),
	}

	out := renderWorkflowASCIIViewString(t, workflow)
	for _, want := range []string{
		"Workflow: Linear invoice flow (wf_linear)",
		"| Start",
		"| Extract totals",
		"| start [start]  |---->| extract [extract]",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in output:\n%s", want, out)
		}
	}

	startLine, startCol := lineAndColumn(t, out, "| Start")
	extractLine, extractCol := lineAndColumn(t, out, "| Extract totals")
	if startLine != extractLine {
		t.Fatalf("same-y blocks should stay on the same terminal row, got start line %d extract line %d:\n%s", startLine, extractLine, out)
	}
	if extractCol <= startCol {
		t.Fatalf("extract should render to the right of start, got start col %d extract col %d:\n%s", startCol, extractCol, out)
	}
}

func TestRenderWorkflowASCIIViewBranchingWorkflowPreservesVerticalShape(t *testing.T) {
	workflow := &workflowGraph{
		Workflow: retab.Workflow{ID: "wf_branch", Name: "Branching review flow"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start", Label: ptr("Start"), PositionX: ptr(float64(0)), PositionY: ptr(float64(220))},
			{ID: "route", Type: "conditional", Label: ptr("Route invoice"), PositionX: ptr(float64(400)), PositionY: ptr(float64(220))},
			{ID: "review", Type: "review", Label: ptr("Manual review"), PositionX: ptr(float64(900)), PositionY: ptr(float64(0))},
			{ID: "archive", Type: "api_call", Label: ptr("Archive clean invoice"), PositionX: ptr(float64(900)), PositionY: ptr(float64(440))},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "edge_start_route", SourceBlock: "start", TargetBlock: "route"},
			{ID: "edge_route_review", SourceBlock: "route", TargetBlock: "review", SourceHandle: ptr("needs_review")},
			{ID: "edge_route_archive", SourceBlock: "route", TargetBlock: "archive", SourceHandle: ptr("clean")},
		},
	}

	out := renderWorkflowASCIIViewString(t, workflow)
	reviewLine, reviewCol := lineAndColumn(t, out, "| Manual review")
	routeLine, routeCol := lineAndColumn(t, out, "| Route invoice")
	archiveLine, archiveCol := lineAndColumn(t, out, "| Archive clean")
	if reviewLine >= routeLine || routeLine >= archiveLine {
		t.Fatalf("vertical block order should follow position_y, got review=%d route=%d archive=%d:\n%s", reviewLine, routeLine, archiveLine, out)
	}
	if routeCol >= reviewCol || routeCol >= archiveCol {
		t.Fatalf("branch targets should render to the right of the router, got route=%d review=%d archive=%d:\n%s", routeCol, reviewCol, archiveCol, out)
	}
	for _, want := range []string{"needs review", "clean"} {
		if !strings.Contains(out, want) {
			t.Fatalf("edge label %q should be visible in output:\n%s", want, out)
		}
	}
}

func TestRenderWorkflowASCIIViewSuppressesContainerRoutingHandles(t *testing.T) {
	workflow := &workflowGraph{
		Workflow: retab.Workflow{ID: "wf_container", Name: "Container routing"},
		Blocks: []retab.WorkflowBlock{
			{ID: "score", Type: "function", Label: ptr("Score Item"), PositionX: ptr(float64(0)), PositionY: ptr(float64(0))},
			{ID: "start", Type: "start", Label: ptr("Start"), PositionX: ptr(float64(300)), PositionY: ptr(float64(0))},
			{ID: "fanout", Type: "for_each", Label: ptr("Fan Out"), PositionX: ptr(float64(600)), PositionY: ptr(float64(0))},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "edge_start_fanout", SourceBlock: "start", TargetBlock: "fanout", SourceHandle: ptr("output-json-0"), TargetHandle: ptr("fe-left-in")},
			{ID: "edge_fanout_score", SourceBlock: "fanout", TargetBlock: "score", SourceHandle: ptr("fe-right-out"), TargetHandle: ptr("input-json-0")},
			{ID: "edge_score_fanout", SourceBlock: "score", TargetBlock: "fanout", SourceHandle: ptr("output-json-0"), TargetHandle: ptr("fe-right-in")},
		},
	}

	out := renderWorkflowASCIIViewString(t, workflow)
	for _, unexpected := range []string{"fe left", "fe right"} {
		if strings.Contains(out, unexpected) {
			t.Fatalf("container routing handle %q should not be rendered:\n%s", unexpected, out)
		}
	}
	for _, want := range []string{"| Fan Out", "| Score Item", "| start [start]"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestRenderWorkflowASCIIViewMergedWorkflowKeepsVisualColumns(t *testing.T) {
	workflow := &workflowGraph{
		Workflow: retab.Workflow{ID: "wf_merge", Name: "Reconciliation flow"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start", Label: ptr("Start"), PositionX: ptr(float64(0)), PositionY: ptr(float64(220))},
			{ID: "split", Type: "split", Label: ptr("Split packet"), PositionX: ptr(float64(350)), PositionY: ptr(float64(220))},
			{ID: "invoice", Type: "extract", Label: ptr("Invoice fields"), PositionX: ptr(float64(740)), PositionY: ptr(float64(0))},
			{ID: "po", Type: "extract", Label: ptr("PO fields"), PositionX: ptr(float64(740)), PositionY: ptr(float64(440))},
			{ID: "match", Type: "function", Label: ptr("Match records"), PositionX: ptr(float64(1120)), PositionY: ptr(float64(220))},
			{ID: "approve", Type: "review", Label: ptr("Approve exception"), PositionX: ptr(float64(1500)), PositionY: ptr(float64(0))},
			{ID: "export", Type: "api_call", Label: ptr("Export result"), PositionX: ptr(float64(1500)), PositionY: ptr(float64(440))},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "edge_start_split", SourceBlock: "start", TargetBlock: "split"},
			{ID: "edge_split_invoice", SourceBlock: "split", TargetBlock: "invoice", SourceHandle: ptr("invoice")},
			{ID: "edge_split_po", SourceBlock: "split", TargetBlock: "po", SourceHandle: ptr("purchase_order")},
			{ID: "edge_invoice_match", SourceBlock: "invoice", TargetBlock: "match"},
			{ID: "edge_po_match", SourceBlock: "po", TargetBlock: "match"},
			{ID: "edge_match_approve", SourceBlock: "match", TargetBlock: "approve", SourceHandle: ptr("mismatch")},
			{ID: "edge_match_export", SourceBlock: "match", TargetBlock: "export", SourceHandle: ptr("matched")},
		},
	}

	out := renderWorkflowASCIIViewString(t, workflow)
	startLine, startCol := lineAndColumn(t, out, "| Start")
	splitLine, splitCol := lineAndColumn(t, out, "| Split packet")
	invoiceLine, invoiceCol := lineAndColumn(t, out, "| Invoice fields")
	poLine, poCol := lineAndColumn(t, out, "| PO fields")
	matchLine, matchCol := lineAndColumn(t, out, "| Match records")
	approveLine, approveCol := lineAndColumn(t, out, "| Approve exception")
	exportLine, exportCol := lineAndColumn(t, out, "| Export result")

	if startCol >= splitCol || splitCol >= invoiceCol || invoiceCol >= matchCol || matchCol >= approveCol {
		t.Fatalf("top path columns should follow position_x, got start=%d split=%d invoice=%d match=%d approve=%d:\n%s", startCol, splitCol, invoiceCol, matchCol, approveCol, out)
	}
	if splitCol >= poCol || poCol >= matchCol || matchCol >= exportCol {
		t.Fatalf("bottom path columns should follow position_x, got split=%d po=%d match=%d export=%d:\n%s", splitCol, poCol, matchCol, exportCol, out)
	}
	if invoiceLine >= splitLine || splitLine != startLine || splitLine != matchLine || poLine <= splitLine {
		t.Fatalf("merge workflow rows should follow position_y, got invoice=%d start=%d split=%d match=%d po=%d:\n%s", invoiceLine, startLine, splitLine, matchLine, poLine, out)
	}
	if approveLine >= matchLine || exportLine <= matchLine {
		t.Fatalf("post-merge branches should stay above/below match, got approve=%d match=%d export=%d:\n%s", approveLine, matchLine, exportLine, out)
	}
	for _, want := range []string{"invoice", "purchase order", "mismatch", "matched"} {
		if !strings.Contains(out, want) {
			t.Fatalf("edge label %q should be visible in output:\n%s", want, out)
		}
	}
}

func TestRenderWorkflowASCIIViewDisconnectedSubgraphRendersOnce(t *testing.T) {
	workflow := &workflowGraph{
		Workflow: retab.Workflow{ID: "wf_detached"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start", Label: ptr("Start"), PositionX: ptr(float64(0)), PositionY: ptr(float64(0))},
			{ID: "parent", Type: "extract", Label: ptr("Parent"), PositionX: ptr(float64(0)), PositionY: ptr(float64(440))},
			{ID: "child", Type: "edit", Label: ptr("Child"), PositionX: ptr(float64(300)), PositionY: ptr(float64(440))},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "edge_detached", SourceBlock: "parent", TargetBlock: "child"},
		},
	}

	out := renderWorkflowASCIIViewString(t, workflow)
	if strings.Count(out, "| Child") != 1 {
		t.Fatalf("detached child should render exactly once, got:\n%s", out)
	}
	parentLine, parentCol := lineAndColumn(t, out, "| Parent")
	childLine, childCol := lineAndColumn(t, out, "| Child")
	if parentLine != childLine || childCol <= parentCol {
		t.Fatalf("detached subgraph should keep its own visual row and direction, parent=(%d,%d) child=(%d,%d):\n%s", parentLine, parentCol, childLine, childCol, out)
	}
	if strings.Contains(out, "\n\n\n\n") {
		t.Fatalf("ascii view should collapse excessive blank rows, got:\n%s", out)
	}
}

// Regression for CLI probing 2026-05: a freshly-`workflows create`'d
// workflow has only the auto-added start_document block and zero edges.
// The view used to render "Disconnected: Document" against that lone
// block, which is technically true but cosmetically wrong — the block
// just hasn't been wired to anything yet because the user is still
// scaffolding. Suppress the line for that exact shape.
func TestRenderWorkflowASCIIViewSuppressesDisconnectedForFreshScaffolding(t *testing.T) {
	cases := []struct {
		name string
		typ  string
	}{
		{name: "start_document underscore", typ: "start_document"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workflow := &workflowGraph{
				Workflow: retab.Workflow{ID: "wf_fresh"},
				Blocks: []retab.WorkflowBlock{
					{ID: "start", Type: retab.WorkflowBlockType(tc.typ), Label: ptr("Document"), PositionX: ptr(float64(100)), PositionY: ptr(float64(200))},
				},
				Edges: nil,
			}

			out := renderWorkflowASCIIViewString(t, workflow)
			if strings.Contains(out, "Disconnected:") {
				t.Fatalf("freshly-created scaffold workflow should not emit a Disconnected line:\n%s", out)
			}
		})
	}
}

func TestRenderWorkflowASCIIViewReportsIsolatedBlocks(t *testing.T) {
	workflow := &workflowGraph{
		Workflow: retab.Workflow{ID: "wf_isolated"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start", Label: ptr("Start"), PositionX: ptr(float64(0)), PositionY: ptr(float64(0))},
			{ID: "extract", Type: "extract", Label: ptr("Extract"), PositionX: ptr(float64(300)), PositionY: ptr(float64(0))},
			{ID: "review", Type: "review", Label: ptr("Human Review"), PositionX: ptr(float64(600)), PositionY: ptr(float64(0))},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "edge_start_extract", SourceBlock: "start", TargetBlock: "extract"},
		},
	}

	out := renderWorkflowASCIIViewString(t, workflow)
	if !strings.Contains(out, "Disconnected: Human Review") {
		t.Fatalf("isolated non-note blocks should be reported, got:\n%s", out)
	}
}

func TestRenderWorkflowASCIIViewDoesNotLeakFloatingLabelTailOntoBoxBorder(t *testing.T) {
	// Regression test for the "ion" artifact: a floating edge label
	// whose source handle is too long to fit in the gap between two
	// horizontally adjacent boxes used to write the FULL label onto a
	// (then-empty) box-border row, and drawBox would then overwrite
	// only the cells that fall inside the box, leaving the label's
	// tail (the chars between the right border of the source box and
	// the left border of the target box) visible. Pin: no fragment of
	// the edge label may appear glued to a box-border `+` character on
	// the box's top or bottom row.
	workflow := &workflowGraph{
		Workflow: retab.Workflow{ID: "wf_long_handles", Name: "Long handle bleed"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start_document", Label: ptr("Document"), PositionX: ptr(float64(0)), PositionY: ptr(float64(0))},
			{ID: "split", Type: "split", Label: ptr("Split"), PositionX: ptr(float64(300)), PositionY: ptr(float64(0))},
			{ID: "extract", Type: "extract", Label: ptr("Extract booking"), PositionX: ptr(float64(600)), PositionY: ptr(float64(0))},
			{ID: "fn", Type: "function", Label: ptr("Check booking"), PositionX: ptr(float64(900)), PositionY: ptr(float64(0))},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "e1", SourceBlock: "start", TargetBlock: "split"},
			{ID: "e2", SourceBlock: "split", TargetBlock: "extract", SourceHandle: ptr("output-file-booking-confirmation")},
			{ID: "e3", SourceBlock: "extract", TargetBlock: "fn"},
		},
	}

	out := renderWorkflowASCIIViewString(t, workflow)

	// Pin: a box-border row must contain only border chars (`+`, `-`)
	// and whitespace — never letters bleeding in from an edge label
	// drawn before drawBox had a chance to claim the cells. We detect
	// this by walking every line that contains a `+` and asserting no
	// alphanumeric character appears OUTSIDE a balanced `+...+` run.
	for lineNo, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimRight(line, " ")
		if !strings.Contains(trimmed, "+") {
			continue
		}
		// Only consider lines whose only "real" content is box borders.
		// A border row's content (excluding whitespace) must be entirely
		// drawn from {+, -}.
		alphabeticOutsideBorder := false
		for _, ch := range strings.TrimSpace(trimmed) {
			if ch == '+' || ch == '-' || ch == ' ' {
				continue
			}
			alphabeticOutsideBorder = true
			break
		}
		if !alphabeticOutsideBorder {
			continue
		}
		// If we reach here, the line has letters; that's only OK when
		// the line is actually a label/content row (contains `|`). A
		// pure border row must not.
		if !strings.Contains(line, "|") {
			t.Fatalf("box-border row %d contains stray letters (likely an edge-label tail): %q\nfull output:\n%s", lineNo, line, out)
		}
	}
}

func TestRenderWorkflowASCIIViewKeepsTypeTagWhenBlockIDIsLong(t *testing.T) {
	// Long random ids must not knock the type tag out of the meta line —
	// the type is far more useful than the trailing 6 chars of an opaque id.
	workflow := &workflowGraph{
		Workflow: retab.Workflow{ID: "wf_longid", Name: "Long-id workflow"},
		Blocks: []retab.WorkflowBlock{
			{ID: "block_gOKWG4abcdefgh1a2KYG", Type: "start_document", Label: ptr("Document"), PositionX: ptr(float64(0)), PositionY: ptr(float64(0))},
			{ID: "block_BKcQijCUKCisbTQzrLDwd", Type: "extract", Label: ptr("Extract"), PositionX: ptr(float64(300)), PositionY: ptr(float64(0))},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "edge_doc_extract", SourceBlock: "block_gOKWG4abcdefgh1a2KYG", TargetBlock: "block_BKcQijCUKCisbTQzrLDwd"},
		},
	}

	out := renderWorkflowASCIIViewString(t, workflow)
	if !strings.Contains(out, "[extract]") {
		t.Fatalf("expected [extract] tag to survive truncation, got:\n%s", out)
	}
	if !strings.Contains(out, "[start_document]") {
		t.Fatalf("expected [start_document] tag to survive truncation, got:\n%s", out)
	}

	// Regression for issue #3: the start_document meta cell used to render
	// as `bloc... [start_document]` for every block whose id starts with
	// `block_`, because the tight-budget head-tail truncation only kept the
	// first 4 chars of the id — which were always the literal "bloc". The
	// fix strips the redundant `block_` prefix in the rendered string so the
	// nanoid suffix (the actually-distinguishing chars) survives. The cell
	// must NOT be the useless `bloc...`, and it must surface at least 4 chars
	// of the nanoid suffix.
	if strings.Contains(out, "bloc... [start_document]") {
		t.Fatalf("start_document meta cell must not render as 'bloc... [start_document]' (issue #3), got:\n%s", out)
	}
	// The nanoid suffix of the start block id is "a2KYG" — at least the
	// last 4 chars ("2KYG") must appear adjacent to the [start_document]
	// tag in the rendered cell.
	if !strings.Contains(out, "2KYG [start_document]") {
		t.Fatalf("expected at least 4 chars of the nanoid suffix ('2KYG') adjacent to [start_document], got:\n%s", out)
	}
}

func TestRenderWorkflowASCIIViewHidesEdgeLabelsForDenseGraphs(t *testing.T) {
	blocks := []retab.WorkflowBlock{
		{ID: "start", Type: "start", Label: ptr("Start"), PositionX: ptr(float64(0)), PositionY: ptr(float64(0))},
		{ID: "hub", Type: "merge_dicts", Label: ptr("Hub"), PositionX: ptr(float64(300)), PositionY: ptr(float64(0))},
		{ID: "end", Type: "function", Label: ptr("End"), PositionX: ptr(float64(600)), PositionY: ptr(float64(0))},
	}
	var edges []retab.WorkflowEdgeDoc
	for range 25 {
		edges = append(edges, retab.WorkflowEdgeDoc{
			ID:           "edge",
			SourceBlock:  "start",
			SourceHandle: ptr("output-json-0"),
			TargetBlock:  "hub",
			TargetHandle: ptr("input-json-noisy-label"),
		})
	}
	edges = append(edges, retab.WorkflowEdgeDoc{ID: "edge_end", SourceBlock: "hub", TargetBlock: "end"})
	workflow := &workflowGraph{
		Workflow: retab.Workflow{ID: "wf_dense"},
		Blocks:   blocks,
		Edges:    edges,
	}

	out := renderWorkflowASCIIViewString(t, workflow)
	if strings.Contains(out, "noisy label") {
		t.Fatalf("dense graph should hide noisy inline edge labels, got:\n%s", out)
	}
	if !strings.Contains(out, "Edge labels: hidden for dense graph") {
		t.Fatalf("dense graph should explain hidden edge labels, got:\n%s", out)
	}
}

func TestWorkflowsViewCommandFetchesGraphPartsAndPrintsASCII(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/wf_graph":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "wf_graph", "name": "Invoice flow"})
		case "/v1/workflows/blocks":
			if r.URL.Query().Get("workflow_id") != "wf_graph" {
				t.Fatalf("blocks workflow_id = %q", r.URL.Query().Get("workflow_id"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "start", "type": "start_document", "label": "Start", "position_x": 0, "position_y": 0},
					{"id": "extract", "type": "extract", "label": "Extract totals", "position_x": 300, "position_y": 0},
				},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		case "/v1/workflows/edges":
			if r.URL.Query().Get("workflow_id") != "wf_graph" {
				t.Fatalf("edges workflow_id = %q", r.URL.Query().Get("workflow_id"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "edge_1", "source_block": "start", "target_block": "extract"},
				},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		default:
			t.Fatalf("unexpected path = %s", r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsViewCmd.RunE(workflowsViewCmd, []string{"wf_graph"}); err != nil {
			t.Fatalf("workflows view: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	for _, want := range []string{
		"Workflow: Invoice flow (wf_graph)",
		"| Start",
		"| Extract totals",
		"| start [start_document]",
		"| extract [extract]",
		"--->",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in stdout, got:\n%s", want, stdout)
		}
	}
}

// TestWorkflowsViewCommandWalksAllBlockAndEdgePages guards the graph view
// against first-page-only scans: blocks and edges are split across two pages
// (page 1 returns an `after` cursor). A single-page scan would drop the
// page-2 "load" block entirely and, lacking the page-2 edge that wires
// extract->load, would falsely flag both as Disconnected. Walking all pages
// must render the full graph with no disconnected warning.
func TestWorkflowsViewCommandWalksAllBlockAndEdgePages(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/wf_graph":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "wf_graph", "name": "Invoice flow"})
		case "/v1/workflows/blocks":
			if r.URL.Query().Get("after") == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{
						{"id": "start", "type": "start_document", "label": "Start", "position_x": 0, "position_y": 0},
						{"id": "extract", "type": "extract", "label": "Extract totals", "position_x": 300, "position_y": 0},
					},
					"list_metadata": map[string]any{"after": "blkcursor"},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "load", "type": "api_call", "label": "Load to ERP", "position_x": 600, "position_y": 0},
				},
				"list_metadata": map[string]any{"after": nil},
			})
		case "/v1/workflows/edges":
			if r.URL.Query().Get("after") == "" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data":          []map[string]any{{"id": "edge_1", "source_block": "start", "target_block": "extract"}},
					"list_metadata": map[string]any{"after": "edgecursor"},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          []map[string]any{{"id": "edge_2", "source_block": "extract", "target_block": "load"}},
				"list_metadata": map[string]any{"after": nil},
			})
		default:
			t.Fatalf("unexpected path = %s", r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	stdout, stderr := captureStd(t, func() {
		if err := workflowsViewCmd.RunE(workflowsViewCmd, []string{"wf_graph"}); err != nil {
			t.Fatalf("workflows view: %v", err)
		}
	})
	if strings.Contains(stdout, "Disconnected") {
		t.Fatalf("page-2 edge ignored: graph falsely reports Disconnected:\n%s", stdout)
	}
	if !strings.Contains(stdout, "| load [api_call]") {
		t.Fatalf("page-2 block missing from graph; got:\n%s%s", stdout, stderr)
	}
}

func TestWorkflowsViewCommandHonorsExplicitOutputJSON(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/wf_graph":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "wf_graph", "name": "Invoice flow"})
		case "/v1/workflows/blocks":
			if r.URL.Query().Get("workflow_id") != "wf_graph" {
				t.Fatalf("blocks workflow_id = %q", r.URL.Query().Get("workflow_id"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "start", "type": "start_document", "label": "Start", "position_x": 0, "position_y": 0},
					{"id": "extract", "type": "extract", "label": "Extract totals", "position_x": 300, "position_y": 0},
				},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		case "/v1/workflows/edges":
			if r.URL.Query().Get("workflow_id") != "wf_graph" {
				t.Fatalf("edges workflow_id = %q", r.URL.Query().Get("workflow_id"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "edge_1", "source_block": "start", "target_block": "extract"},
				},
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		default:
			t.Fatalf("unexpected path = %s", r.URL.Path)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)
	if err := rootCmd.PersistentFlags().Set("output", "json"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := workflowsViewCmd.RunE(workflowsViewCmd, []string{"wf_graph"}); err != nil {
			t.Fatalf("workflows view: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if strings.Contains(stdout, "Workflow: Invoice flow") {
		t.Fatalf("expected JSON output, got ASCII:\n%s", stdout)
	}
	var got struct {
		Workflow retab.Workflow          `json:"workflow"`
		Blocks   []retab.WorkflowBlock   `json:"blocks"`
		Edges    []retab.WorkflowEdgeDoc `json:"edges"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode workflows view JSON: %v\n%s", err, stdout)
	}
	if got.Workflow.ID != "wf_graph" || got.Workflow.Name != "Invoice flow" {
		t.Fatalf("workflow = %#v", got.Workflow)
	}
	if len(got.Blocks) != 2 || len(got.Edges) != 1 {
		t.Fatalf("blocks=%d edges=%d output=%s", len(got.Blocks), len(got.Edges), stdout)
	}
}

func TestWorkflowsViewCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"workflows", "view", "wf_abc"})
	if err != nil {
		t.Fatalf("workflows view not registered: %v", err)
	}
	if cmd.Name() != "view" {
		t.Fatalf("resolved command = %q, want view", cmd.Name())
	}
}
