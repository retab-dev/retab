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

func renderWorkflowASCIIViewString(t *testing.T, workflow *retab.WorkflowWithEntities) string {
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
	workflow := &retab.WorkflowWithEntities{
		Workflow: retab.Workflow{ID: "wf_linear", Name: "Linear invoice flow"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start", Label: "Start", PositionX: 0, PositionY: 0},
			{ID: "extract", Type: "extract", Label: "Extract totals", PositionX: 300, PositionY: 0},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "edge_start_extract", SourceBlock: "start", TargetBlock: "extract"},
		},
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
	workflow := &retab.WorkflowWithEntities{
		Workflow: retab.Workflow{ID: "wf_branch", Name: "Branching review flow"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start", Label: "Start", PositionX: 0, PositionY: 220},
			{ID: "route", Type: "conditional", Label: "Route invoice", PositionX: 400, PositionY: 220},
			{ID: "review", Type: "hil", Label: "Manual review", PositionX: 900, PositionY: 0},
			{ID: "archive", Type: "api_call", Label: "Archive clean invoice", PositionX: 900, PositionY: 440},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "edge_start_route", SourceBlock: "start", TargetBlock: "route"},
			{ID: "edge_route_review", SourceBlock: "route", TargetBlock: "review", SourceHandle: "needs_review"},
			{ID: "edge_route_archive", SourceBlock: "route", TargetBlock: "archive", SourceHandle: "clean"},
		},
	}

	out := renderWorkflowASCIIViewString(t, workflow)
	reviewLine, reviewCol := lineAndColumn(t, out, "| Manual review")
	routeLine, routeCol := lineAndColumn(t, out, "| Route invoice")
	archiveLine, archiveCol := lineAndColumn(t, out, "| Archive clean")
	if !(reviewLine < routeLine && routeLine < archiveLine) {
		t.Fatalf("vertical block order should follow position_y, got review=%d route=%d archive=%d:\n%s", reviewLine, routeLine, archiveLine, out)
	}
	if !(routeCol < reviewCol && routeCol < archiveCol) {
		t.Fatalf("branch targets should render to the right of the router, got route=%d review=%d archive=%d:\n%s", routeCol, reviewCol, archiveCol, out)
	}
	for _, want := range []string{"needs review", "clean"} {
		if !strings.Contains(out, want) {
			t.Fatalf("edge label %q should be visible in output:\n%s", want, out)
		}
	}
}

func TestRenderWorkflowASCIIViewMergedWorkflowKeepsVisualColumns(t *testing.T) {
	workflow := &retab.WorkflowWithEntities{
		Workflow: retab.Workflow{ID: "wf_merge", Name: "Reconciliation flow"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start", Label: "Start", PositionX: 0, PositionY: 220},
			{ID: "split", Type: "split", Label: "Split packet", PositionX: 350, PositionY: 220},
			{ID: "invoice", Type: "extract", Label: "Invoice fields", PositionX: 740, PositionY: 0},
			{ID: "po", Type: "extract", Label: "PO fields", PositionX: 740, PositionY: 440},
			{ID: "match", Type: "function", Label: "Match records", PositionX: 1120, PositionY: 220},
			{ID: "approve", Type: "hil", Label: "Approve exception", PositionX: 1500, PositionY: 0},
			{ID: "export", Type: "api_call", Label: "Export result", PositionX: 1500, PositionY: 440},
		},
		Edges: []retab.WorkflowEdgeDoc{
			{ID: "edge_start_split", SourceBlock: "start", TargetBlock: "split"},
			{ID: "edge_split_invoice", SourceBlock: "split", TargetBlock: "invoice", SourceHandle: "invoice"},
			{ID: "edge_split_po", SourceBlock: "split", TargetBlock: "po", SourceHandle: "purchase_order"},
			{ID: "edge_invoice_match", SourceBlock: "invoice", TargetBlock: "match"},
			{ID: "edge_po_match", SourceBlock: "po", TargetBlock: "match"},
			{ID: "edge_match_approve", SourceBlock: "match", TargetBlock: "approve", SourceHandle: "mismatch"},
			{ID: "edge_match_export", SourceBlock: "match", TargetBlock: "export", SourceHandle: "matched"},
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

	if !(startCol < splitCol && splitCol < invoiceCol && invoiceCol < matchCol && matchCol < approveCol) {
		t.Fatalf("top path columns should follow position_x, got start=%d split=%d invoice=%d match=%d approve=%d:\n%s", startCol, splitCol, invoiceCol, matchCol, approveCol, out)
	}
	if !(splitCol < poCol && poCol < matchCol && matchCol < exportCol) {
		t.Fatalf("bottom path columns should follow position_x, got split=%d po=%d match=%d export=%d:\n%s", splitCol, poCol, matchCol, exportCol, out)
	}
	if !(invoiceLine < splitLine && splitLine == startLine && splitLine == matchLine && poLine > splitLine) {
		t.Fatalf("merge workflow rows should follow position_y, got invoice=%d start=%d split=%d match=%d po=%d:\n%s", invoiceLine, startLine, splitLine, matchLine, poLine, out)
	}
	if !(approveLine < matchLine && exportLine > matchLine) {
		t.Fatalf("post-merge branches should stay above/below match, got approve=%d match=%d export=%d:\n%s", approveLine, matchLine, exportLine, out)
	}
	for _, want := range []string{"invoice", "purchase order", "mismatch", "matched"} {
		if !strings.Contains(out, want) {
			t.Fatalf("edge label %q should be visible in output:\n%s", want, out)
		}
	}
}

func TestRenderWorkflowASCIIViewDisconnectedSubgraphRendersOnce(t *testing.T) {
	workflow := &retab.WorkflowWithEntities{
		Workflow: retab.Workflow{ID: "wf_detached"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start", Label: "Start", PositionX: 0, PositionY: 0},
			{ID: "parent", Type: "extract", Label: "Parent", PositionX: 0, PositionY: 440},
			{ID: "child", Type: "edit", Label: "Child", PositionX: 300, PositionY: 440},
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

func TestRenderWorkflowASCIIViewReportsIsolatedBlocks(t *testing.T) {
	workflow := &retab.WorkflowWithEntities{
		Workflow: retab.Workflow{ID: "wf_isolated"},
		Blocks: []retab.WorkflowBlock{
			{ID: "start", Type: "start", Label: "Start", PositionX: 0, PositionY: 0},
			{ID: "extract", Type: "extract", Label: "Extract", PositionX: 300, PositionY: 0},
			{ID: "review", Type: "hil", Label: "Human Review", PositionX: 600, PositionY: 0},
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

func TestRenderWorkflowASCIIViewHidesEdgeLabelsForDenseGraphs(t *testing.T) {
	blocks := []retab.WorkflowBlock{
		{ID: "start", Type: "start", Label: "Start", PositionX: 0, PositionY: 0},
		{ID: "hub", Type: "merge_dicts", Label: "Hub", PositionX: 300, PositionY: 0},
		{ID: "end", Type: "function", Label: "End", PositionX: 600, PositionY: 0},
	}
	var edges []retab.WorkflowEdgeDoc
	for i := 0; i < 25; i++ {
		edges = append(edges, retab.WorkflowEdgeDoc{
			ID:           "edge",
			SourceBlock:  "start",
			SourceHandle: "output-json-0",
			TargetBlock:  "hub",
			TargetHandle: "input-json-noisy-label",
		})
	}
	edges = append(edges, retab.WorkflowEdgeDoc{ID: "edge_end", SourceBlock: "hub", TargetBlock: "end"})
	workflow := &retab.WorkflowWithEntities{
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

func TestWorkflowsViewCommandFetchesEntitiesAndPrintsASCII(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "test-key")
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/workflows/wf_graph/entities" {
			t.Fatalf("path = %s, want entities", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workflow": map[string]any{"id": "wf_graph", "name": "Invoice flow"},
			"blocks": []map[string]any{
				{"id": "start", "type": "start", "label": "Start", "position_x": 0, "position_y": 0},
				{"id": "extract", "type": "extract", "label": "Extract totals", "position_x": 300, "position_y": 0},
			},
			"edges": []map[string]any{
				{"id": "edge_1", "source_block": "start", "target_block": "extract"},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RETAB_BASE_URL", server.URL)

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
		"| start [start]",
		"| extract [extract]",
		"--->",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in stdout, got:\n%s", want, stdout)
		}
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
