package cmd

import (
	"fmt"
	"io"
	"math"
	"slices"
	"sort"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

const (
	workflowASCIIBlockHeight = 4
	workflowASCIIXScale      = 26.0 / 300.0
	workflowASCIIMaxBoxWidth = 28
	workflowASCIIMinBoxWidth = 18
	workflowASCIIMaxEdgeText = 24
)

type workflowASCIIBox struct {
	block retab.WorkflowBlock
	x     int
	y     int
	w     int
	h     int
}

// workflowBlockX/workflowBlockY return the canvas position from a
// WorkflowBlock, dereferencing the optional *float64 fields exposed by
// the SDK. A nil pointer falls back to zero, matching the prior
// "unset position" behaviour the layout code expects.
func workflowBlockX(b retab.WorkflowBlock) float64 {
	if b.PositionX == nil {
		return 0
	}
	return *b.PositionX
}

func workflowBlockY(b retab.WorkflowBlock) float64 {
	if b.PositionY == nil {
		return 0
	}
	return *b.PositionY
}

type workflowASCIICanvas struct {
	cells [][]rune
}

var workflowsViewCmd = &cobra.Command{
	Use:   "view <workflow-id>",
	Short: "Show a workflow as an ASCII graph",
	Long: `Fetch the workflow's draft graph and render it as a compact ASCII
map. This is a human-oriented view for terminal inspection.`,
	Example: `  # Inspect the graph shape in your terminal
  retab workflows view wf_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		workflow, err := client.Workflows.Get(ctx, args[0])
		if err != nil {
			return err
		}
		blocks, err := listAllWorkflowBlocks(ctx, client, args[0])
		if err != nil {
			return err
		}
		edges, err := listAllWorkflowEdges(ctx, client, args[0])
		if err != nil {
			return err
		}
		graph := &workflowGraph{
			Workflow: *workflow,
			Blocks:   blocks,
			Edges:    edges,
		}
		if workflowViewWantsJSON(cmd) {
			return printJSON(workflowGraphJSON{
				Workflow: graph.Workflow,
				Blocks:   graph.Blocks,
				Edges:    graph.Edges,
			})
		}
		return renderWorkflowASCIIView(cmd.OutOrStdout(), graph)
	}),
}

type workflowGraph struct {
	Workflow retab.Workflow
	Blocks   []retab.WorkflowBlock
	Edges    []retab.WorkflowEdgeDoc
}

type workflowGraphJSON struct {
	Workflow retab.Workflow          `json:"workflow"`
	Blocks   []retab.WorkflowBlock   `json:"blocks"`
	Edges    []retab.WorkflowEdgeDoc `json:"edges"`
}

func workflowViewWantsJSON(cmd *cobra.Command) bool {
	var raw string
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
	}
	if raw == "" && cmd != rootCmd {
		if f := rootCmd.PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
	}
	return raw == string(OutputJSON)
}

func renderWorkflowASCIIView(w io.Writer, graph *workflowGraph) error {
	if graph == nil {
		return fmt.Errorf("workflow graph is missing")
	}

	title := graph.Workflow.ID
	if graph.Workflow.Name != "" {
		title = graph.Workflow.Name + " (" + graph.Workflow.ID + ")"
	}
	if _, err := fmt.Fprintf(w, "Workflow: %s\n\n", title); err != nil {
		return err
	}
	if len(graph.Blocks) == 0 {
		_, err := fmt.Fprintln(w, "(no blocks)")
		return err
	}

	visibleEdges := workflowASCIIVisibleEdges(graph.Edges)
	visibleBlocks, hiddenNotes := workflowASCIIVisibleBlocks(graph.Blocks, visibleEdges)
	if len(visibleBlocks) == 0 {
		_, err := fmt.Fprintln(w, "(only note blocks)")
		return err
	}

	boxes := workflowASCIIBoxes(visibleBlocks, visibleEdges)
	boxesByID := make(map[string]workflowASCIIBox, len(boxes))
	for _, box := range boxes {
		boxesByID[box.block.ID] = box
	}

	canvas := newWorkflowASCIICanvas(workflowASCIICanvasSize(boxes))
	showEdgeLabels := workflowASCIIShouldShowEdgeLabels(len(visibleEdges), len(visibleBlocks))
	// Draw boxes FIRST, then edges. drawFloatingEdgeLabel's collision
	// detection (skip if any target cell != ' ') only works against
	// content that is already on the canvas. If edges drew first, a
	// floating label landing on a box border row (startY ± 2) would see
	// only blank cells, write the whole label freely, then drawBox would
	// overwrite ONLY the cells inside the box borders — leaving the
	// label's tail (the chars past the box's right border) visible in
	// the gap between boxes. That was the source of the "ion" artifact
	// at the top of the rendered graph for edges with long handle names
	// like "output-file-booking-confirmation".
	for _, box := range boxes {
		canvas.drawBox(box)
	}
	for _, edge := range workflowASCIISortedEdges(visibleEdges) {
		source, hasSource := boxesByID[edge.SourceBlock]
		target, hasTarget := boxesByID[edge.TargetBlock]
		if !hasSource || !hasTarget {
			continue
		}
		canvas.drawEdge(source, target, edge, showEdgeLabels)
	}

	if _, err := fmt.Fprint(w, canvas.String()); err != nil {
		return err
	}
	if hiddenNotes > 0 {
		if _, err := fmt.Fprintf(w, "\nNotes: %d detached note block(s) hidden from ASCII view.\n", hiddenNotes); err != nil {
			return err
		}
	}
	isolatedBlocks := workflowASCIIIsolatedBlocks(visibleBlocks, visibleEdges)
	// Suppress the "Disconnected" line for the freshly-`workflows create`'d
	// shape: one block, which is the auto-added start_document placeholder,
	// and zero edges. That block is "isolated" only in the trivial sense
	// that nothing has been wired to it yet — surfacing it as a warning
	// misleads users still scaffolding their graph.
	if len(isolatedBlocks) > 0 && !isFreshScaffoldShape(visibleBlocks, graph.Edges) {
		if _, err := fmt.Fprintf(w, "Disconnected: %s\n", strings.Join(isolatedBlocks, ", ")); err != nil {
			return err
		}
	}
	if !showEdgeLabels {
		_, err := fmt.Fprintf(w, "Edge labels: hidden for dense graph (%d edges).\n", len(visibleEdges))
		return err
	}
	return nil
}

func workflowASCIIBoxes(blocks []retab.WorkflowBlock, edges []retab.WorkflowEdgeDoc) []workflowASCIIBox {
	sorted := append([]retab.WorkflowBlock(nil), blocks...)
	sortWorkflowASCIIBlocks(sorted)

	positions := workflowASCIIPositions(sorted, edges)
	layout := workflowASCIICompressedLayout(sorted, positions)

	boxes := make([]workflowASCIIBox, 0, len(sorted))
	for _, block := range sorted {
		position := layout[block.ID]
		box := workflowASCIIBox{
			block: block,
			x:     position.x,
			y:     position.y,
			w:     workflowASCIIBoxWidth(block),
			h:     workflowASCIIBlockHeight,
		}
		for workflowASCIIOverlapsAny(box, boxes) {
			box.y += workflowASCIIBlockHeight + 1
		}
		boxes = append(boxes, box)
	}
	sort.SliceStable(boxes, func(i, j int) bool {
		if boxes[i].y != boxes[j].y {
			return boxes[i].y < boxes[j].y
		}
		if boxes[i].x != boxes[j].x {
			return boxes[i].x < boxes[j].x
		}
		return boxes[i].block.ID < boxes[j].block.ID
	})
	return boxes
}

func workflowASCIIVisibleBlocks(blocks []retab.WorkflowBlock, edges []retab.WorkflowEdgeDoc) ([]retab.WorkflowBlock, int) {
	connected := map[string]bool{}
	for _, edge := range edges {
		connected[edge.SourceBlock] = true
		connected[edge.TargetBlock] = true
	}
	visible := make([]retab.WorkflowBlock, 0, len(blocks))
	hiddenNotes := 0
	for _, block := range blocks {
		if block.Type == "note" && !connected[block.ID] {
			hiddenNotes++
			continue
		}
		visible = append(visible, block)
	}
	return visible, hiddenNotes
}

func workflowASCIIVisibleEdges(edges []retab.WorkflowEdgeDoc) []retab.WorkflowEdgeDoc {
	visible := make([]retab.WorkflowEdgeDoc, 0, len(edges))
	for _, edge := range edges {
		if workflowASCIIIsContainerFeedbackHandle(derefString(edge.TargetHandle)) {
			continue
		}
		visible = append(visible, edge)
	}
	return visible
}

// isFreshScaffoldShape matches the canonical empty workflow: exactly one
// start_document block, zero edges. Keeps the predicate aligned with
// `isStartDocumentBlock` so both spellings + the legacy “"start"“ value
// are handled (see “isEffectivelyEmptyDraft“ for the publish-time twin).
func isFreshScaffoldShape(blocks []retab.WorkflowBlock, edges []retab.WorkflowEdgeDoc) bool {
	if len(edges) != 0 || len(blocks) != 1 {
		return false
	}
	return isStartDocumentBlock(blocks[0])
}

func workflowASCIIIsolatedBlocks(blocks []retab.WorkflowBlock, edges []retab.WorkflowEdgeDoc) []string {
	connected := map[string]bool{}
	for _, edge := range edges {
		connected[edge.SourceBlock] = true
		connected[edge.TargetBlock] = true
	}
	var isolated []string
	for _, block := range blocks {
		if connected[block.ID] {
			continue
		}
		isolated = append(isolated, workflowASCIIBlockLabel(block))
	}
	sort.Strings(isolated)
	if len(isolated) > 5 {
		isolated = append(isolated[:5], fmt.Sprintf("+%d more", len(isolated)-5))
	}
	return isolated
}

type workflowASCIIPosition struct {
	x float64
	y float64
}

func workflowASCIIPositions(blocks []retab.WorkflowBlock, edges []retab.WorkflowEdgeDoc) map[string]workflowASCIIPosition {
	out := make(map[string]workflowASCIIPosition, len(blocks))
	if workflowASCIIHasRealPositions(blocks) {
		for _, block := range blocks {
			out[block.ID] = workflowASCIIPosition{x: workflowBlockX(block), y: workflowBlockY(block)}
		}
		return out
	}

	levels := workflowASCIISyntheticLevels(blocks, edges)
	rowByLevel := map[int]int{}
	for _, block := range blocks {
		level := levels[block.ID]
		row := rowByLevel[level]
		rowByLevel[level]++
		out[block.ID] = workflowASCIIPosition{
			x: float64(level * 300),
			y: float64(row * 220),
		}
	}
	return out
}

func workflowASCIIHasRealPositions(blocks []retab.WorkflowBlock) bool {
	for _, block := range blocks {
		if workflowBlockX(block) != 0 || workflowBlockY(block) != 0 {
			return true
		}
	}
	return false
}

type workflowASCIIGridPosition struct {
	x int
	y int
}

type workflowASCIIAxisCluster struct {
	min   float64
	max   float64
	items []string
}

func workflowASCIICompressedLayout(blocks []retab.WorkflowBlock, positions map[string]workflowASCIIPosition) map[string]workflowASCIIGridPosition {
	xClusters := workflowASCIIClusters(blocks, positions, true, 180)
	yClusters := workflowASCIIClusters(blocks, positions, false, 120)

	blockWidths := map[string]int{}
	for _, block := range blocks {
		blockWidths[block.ID] = workflowASCIIBoxWidth(block)
	}

	xByCluster := make([]int, len(xClusters))
	nextX := 2
	for i, cluster := range xClusters {
		xByCluster[i] = nextX
		maxWidth := workflowASCIIMinBoxWidth
		for _, id := range cluster.items {
			if blockWidths[id] > maxWidth {
				maxWidth = blockWidths[id]
			}
		}
		nextX += maxWidth + 5
	}

	yByCluster := make([]int, len(yClusters))
	for i := range yClusters {
		yByCluster[i] = i * (workflowASCIIBlockHeight + 1)
	}

	out := map[string]workflowASCIIGridPosition{}
	for _, block := range blocks {
		out[block.ID] = workflowASCIIGridPosition{
			x: xByCluster[workflowASCIIClusterIndex(xClusters, block.ID)],
			y: yByCluster[workflowASCIIClusterIndex(yClusters, block.ID)],
		}
	}
	return out
}

func workflowASCIIClusters(blocks []retab.WorkflowBlock, positions map[string]workflowASCIIPosition, useX bool, threshold float64) []workflowASCIIAxisCluster {
	sorted := append([]retab.WorkflowBlock(nil), blocks...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := positions[sorted[i].ID]
		right := positions[sorted[j].ID]
		if useX {
			if left.x != right.x {
				return left.x < right.x
			}
			return sorted[i].ID < sorted[j].ID
		}
		if left.y != right.y {
			return left.y < right.y
		}
		return sorted[i].ID < sorted[j].ID
	})

	var clusters []workflowASCIIAxisCluster
	for _, block := range sorted {
		value := positions[block.ID].y
		if useX {
			value = positions[block.ID].x
		}
		if len(clusters) == 0 || value-clusters[len(clusters)-1].max > threshold {
			clusters = append(clusters, workflowASCIIAxisCluster{min: value, max: value, items: []string{block.ID}})
			continue
		}
		cluster := &clusters[len(clusters)-1]
		cluster.items = append(cluster.items, block.ID)
		if value < cluster.min {
			cluster.min = value
		}
		if value > cluster.max {
			cluster.max = value
		}
	}
	return clusters
}

func workflowASCIIClusterIndex(clusters []workflowASCIIAxisCluster, blockID string) int {
	for i, cluster := range clusters {
		if slices.Contains(cluster.items, blockID) {
			return i
		}
	}
	return 0
}

func workflowASCIISyntheticLevels(blocks []retab.WorkflowBlock, edges []retab.WorkflowEdgeDoc) map[string]int {
	known := map[string]bool{}
	indegree := map[string]int{}
	outgoing := map[string][]string{}
	for _, block := range blocks {
		known[block.ID] = true
		indegree[block.ID] = 0
	}
	for _, edge := range edges {
		if !known[edge.SourceBlock] || !known[edge.TargetBlock] {
			continue
		}
		outgoing[edge.SourceBlock] = append(outgoing[edge.SourceBlock], edge.TargetBlock)
		indegree[edge.TargetBlock]++
	}
	for source := range outgoing {
		sort.Strings(outgoing[source])
	}

	roots := make([]string, 0)
	for _, block := range blocks {
		if block.Type == "start" || indegree[block.ID] == 0 {
			roots = append(roots, block.ID)
		}
	}
	sort.Strings(roots)

	levels := map[string]int{}
	queue := append([]string(nil), roots...)
	// The longest acyclic path visits each block at most once, so any level
	// >= len(blocks) means we are following a cycle. Re-enqueue a target only
	// when its level genuinely increased AND stays within that bound; this
	// both guarantees termination and stops a cyclic edge set (A->B->A) from
	// re-pushing targets forever as the level keeps climbing.
	maxLevel := len(blocks)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, target := range outgoing[current] {
			nextLevel := levels[current] + 1
			if nextLevel >= maxLevel {
				continue
			}
			if levels[target] < nextLevel {
				levels[target] = nextLevel
				queue = append(queue, target)
			}
		}
	}
	return levels
}

func workflowASCIIOverlapsAny(box workflowASCIIBox, placed []workflowASCIIBox) bool {
	for _, other := range placed {
		if box.x+box.w+3 <= other.x || other.x+other.w+3 <= box.x {
			continue
		}
		if box.y+box.h+1 <= other.y || other.y+other.h+1 <= box.y {
			continue
		}
		return true
	}
	return false
}

func workflowASCIICanvasSize(boxes []workflowASCIIBox) (int, int) {
	maxX := 0
	maxY := 0
	for _, box := range boxes {
		if box.x+box.w+2 > maxX {
			maxX = box.x + box.w + 2
		}
		if box.y+box.h+2 > maxY {
			maxY = box.y + box.h + 2
		}
	}
	return maxX + 2, maxY
}

func newWorkflowASCIICanvas(width int, height int) *workflowASCIICanvas {
	cells := make([][]rune, height)
	for y := range cells {
		cells[y] = make([]rune, width)
		for x := range cells[y] {
			cells[y][x] = ' '
		}
	}
	return &workflowASCIICanvas{cells: cells}
}

func (c *workflowASCIICanvas) drawBox(box workflowASCIIBox) {
	if box.w < 2 || box.h < 3 {
		return
	}
	for x := box.x + 1; x < box.x+box.w-1; x++ {
		c.put(x, box.y, '-')
		c.put(x, box.y+box.h-1, '-')
	}
	c.put(box.x, box.y, '+')
	c.put(box.x+box.w-1, box.y, '+')
	c.put(box.x, box.y+box.h-1, '+')
	c.put(box.x+box.w-1, box.y+box.h-1, '+')
	for y := box.y + 1; y < box.y+box.h-1; y++ {
		c.put(box.x, y, '|')
		c.put(box.x+box.w-1, y, '|')
		for x := box.x + 1; x < box.x+box.w-1; x++ {
			c.put(x, y, ' ')
		}
	}

	// `range` over a string yields byte offsets; use a separate per-rune column
	// counter so multibyte labels land one cell per rune instead of skipping
	// columns by each rune's byte length.
	label := workflowASCIIFit(workflowASCIIBlockLabel(box.block), box.w-4)
	col := 0
	for _, ch := range label {
		c.put(box.x+2+col, box.y+1, ch)
		col++
	}
	meta := workflowASCIIFit(workflowASCIIBlockMeta(box.block), box.w-4)
	col = 0
	for _, ch := range meta {
		c.put(box.x+2+col, box.y+2, ch)
		col++
	}
}

func (c *workflowASCIICanvas) drawEdge(source workflowASCIIBox, target workflowASCIIBox, edge retab.WorkflowEdgeDoc, showLabel bool) {
	label := ""
	if showLabel {
		label = workflowASCIIEdgeLabel(edge)
	}
	startY := workflowASCIISourceAnchorY(source, target)
	endY := workflowASCIITargetAnchorY(source, target)
	if target.x >= source.x {
		startX := source.x + source.w
		endX := max(target.x-1, startX)
		if startY == endY {
			if !c.horizontalClear(startX, endX, startY) {
				laneY := c.edgeLaneY(source, target, startX, endX)
				c.drawVertical(startX, startY, laneY)
				c.drawHorizontal(startX, endX, laneY)
				c.drawVertical(endX, laneY, endY)
				c.put(endX, endY, '>')
				if !c.drawEdgeLabel(startX, endX, laneY, label) {
					c.drawFloatingEdgeLabel(endX, laneY, label, true)
				}
				return
			}
			c.drawHorizontal(startX, endX, startY)
			c.put(endX, endY, '>')
			if !c.drawEdgeLabel(startX, endX, startY, label) {
				c.drawFloatingEdgeLabel(endX, startY-2, label, true)
			}
			return
		}
		laneX := max(endX-3, startX)
		c.drawHorizontal(startX, laneX, startY)
		c.drawVertical(laneX, startY, endY)
		c.drawHorizontal(laneX, endX, endY)
		c.put(endX, endY, '>')
		if !c.drawEdgeLabel(startX, laneX, startY, label) {
			labelY := startY + 2
			if endY < startY {
				labelY = startY - 2
			}
			c.drawFloatingEdgeLabel(laneX, labelY, label, true)
		}
		return
	}

	startX := source.x - 1
	endX := target.x + target.w
	if startX < endX {
		startX = endX
	}
	if startY == endY {
		if !c.horizontalClear(endX, startX, startY) {
			laneY := c.edgeLaneY(source, target, endX, startX)
			c.drawVertical(startX, startY, laneY)
			c.drawHorizontal(endX, startX, laneY)
			c.drawVertical(endX, laneY, endY)
			c.put(endX, endY, '<')
			if !c.drawEdgeLabel(endX, startX, laneY, label) {
				c.drawFloatingEdgeLabel(endX, laneY, label, false)
			}
			return
		}
		c.drawHorizontal(endX, startX, startY)
		c.put(endX, endY, '<')
		if !c.drawEdgeLabel(endX, startX, startY, label) {
			c.drawFloatingEdgeLabel(endX, startY-2, label, false)
		}
		return
	}
	laneX := min(endX+3, startX)
	c.drawHorizontal(startX, laneX, startY)
	c.drawVertical(laneX, startY, endY)
	c.drawHorizontal(laneX, endX, endY)
	c.put(endX, endY, '<')
	if !c.drawEdgeLabel(laneX, startX, startY, label) {
		labelY := startY + 2
		if endY < startY {
			labelY = startY - 2
		}
		c.drawFloatingEdgeLabel(laneX, labelY, label, false)
	}
}

func workflowASCIIShouldShowEdgeLabels(edgeCount int, blockCount int) bool {
	if edgeCount > 24 {
		return false
	}
	return edgeCount <= blockCount*2
}

func (c *workflowASCIICanvas) horizontalClear(x1 int, x2 int, y int) bool {
	if y < 0 || y >= len(c.cells) {
		return false
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if x < 0 || x >= len(c.cells[y]) {
			return false
		}
		if c.cells[y][x] != ' ' {
			return false
		}
	}
	return true
}

func (c *workflowASCIICanvas) edgeLaneY(source workflowASCIIBox, target workflowASCIIBox, x1 int, x2 int) int {
	below := max(source.y+source.h, target.y+target.h)
	above := min(source.y, target.y) - 1
	candidates := []int{below, below + 1, above}
	for _, y := range candidates {
		if c.horizontalClear(x1, x2, y) {
			return y
		}
	}
	for y := range c.cells {
		if c.horizontalClear(x1, x2, y) {
			return y
		}
	}
	if below >= len(c.cells) {
		return len(c.cells) - 1
	}
	if below < 0 {
		return 0
	}
	return below
}

func workflowASCIIBoxMiddleY(box workflowASCIIBox) int {
	return box.y + box.h/2
}

func workflowASCIISourceAnchorY(source workflowASCIIBox, target workflowASCIIBox) int {
	switch {
	case workflowASCIIBoxMiddleY(target) < workflowASCIIBoxMiddleY(source):
		return source.y + 1
	case workflowASCIIBoxMiddleY(target) > workflowASCIIBoxMiddleY(source):
		return source.y + source.h - 2
	default:
		return workflowASCIIBoxMiddleY(source)
	}
}

func workflowASCIITargetAnchorY(source workflowASCIIBox, target workflowASCIIBox) int {
	switch {
	case workflowASCIIBoxMiddleY(source) < workflowASCIIBoxMiddleY(target):
		return target.y + 1
	case workflowASCIIBoxMiddleY(source) > workflowASCIIBoxMiddleY(target):
		return target.y + target.h - 2
	default:
		return workflowASCIIBoxMiddleY(target)
	}
}

func (c *workflowASCIICanvas) drawHorizontal(x1 int, x2 int, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		c.putLine(x, y, '-')
	}
}

func (c *workflowASCIICanvas) drawVertical(x int, y1 int, y2 int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		c.putLine(x, y, '|')
	}
}

func (c *workflowASCIICanvas) drawEdgeLabel(x1 int, x2 int, y int, label string) bool {
	if label == "" {
		return true
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	available := x2 - x1 - 1
	if available >= len(label)+2 {
		label = " " + label + " "
	}
	if available < len(label) {
		return false
	}
	start := x1 + (available-len(label))/2 + 1
	for i, ch := range label {
		c.put(start+i, y, ch)
	}
	return true
}

func (c *workflowASCIICanvas) drawFloatingEdgeLabel(anchorX int, y int, label string, leftOfAnchor bool) {
	if label == "" || y < 0 || y >= len(c.cells) {
		return
	}
	if len(label) > len(c.cells[y]) {
		label = workflowASCIIFit(label, len(c.cells[y]))
	}
	x := anchorX + 2
	if leftOfAnchor {
		x = anchorX - len(label) - 1
	}
	if x < 0 {
		x = 0
	}
	if x+len(label) > len(c.cells[y]) {
		x = len(c.cells[y]) - len(label)
	}
	for i := range label {
		if c.cells[y][x+i] != ' ' {
			return
		}
	}
	for i, ch := range label {
		c.put(x+i, y, ch)
	}
}

func (c *workflowASCIICanvas) putLine(x int, y int, ch rune) {
	if y < 0 || y >= len(c.cells) || x < 0 || x >= len(c.cells[y]) {
		return
	}
	current := c.cells[y][x]
	switch {
	case current == ' ':
		c.cells[y][x] = ch
	case current == ch:
		return
	case current == '-' && ch == '|':
		c.cells[y][x] = '+'
	case current == '|' && ch == '-':
		c.cells[y][x] = '+'
	case current == '+':
		return
	}
}

func (c *workflowASCIICanvas) put(x int, y int, ch rune) {
	if y < 0 || y >= len(c.cells) || x < 0 || x >= len(c.cells[y]) {
		return
	}
	c.cells[y][x] = ch
}

func (c *workflowASCIICanvas) String() string {
	var b strings.Builder
	blankRows := 0
	for _, row := range c.cells {
		line := strings.TrimRight(string(row), " ")
		if line == "" {
			blankRows++
			if blankRows > 2 {
				continue
			}
			b.WriteByte('\n')
			continue
		}
		blankRows = 0
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func workflowASCIIFit(value string, width int) string {
	if width <= 0 {
		return ""
	}
	value = strings.Join(strings.Fields(value), " ")
	// Count and slice by rune, not byte: a block label with non-ASCII runes
	// (accents, CJK) sliced at a byte boundary would emit invalid UTF-8 and
	// mis-measure the width. width is a column count, so runes are the unit.
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

func workflowASCIIBoxWidth(block retab.WorkflowBlock) int {
	width := workflowASCIIMaxInt(len(workflowASCIIBlockLabel(block)), len(workflowASCIIBlockMeta(block))) + 4
	if block.Width != nil && *block.Width > 0 {
		scaled := int(math.Round(*block.Width * workflowASCIIXScale))
		if scaled > width {
			width = scaled
		}
	}
	if width < workflowASCIIMinBoxWidth {
		return workflowASCIIMinBoxWidth
	}
	if width > workflowASCIIMaxBoxWidth {
		return workflowASCIIMaxBoxWidth
	}
	return width
}

func workflowASCIISortedEdges(edges []retab.WorkflowEdgeDoc) []retab.WorkflowEdgeDoc {
	out := append([]retab.WorkflowEdgeDoc(nil), edges...)
	sortWorkflowASCIIEdges(out)
	return out
}

func sortWorkflowASCIIBlocks(blocks []retab.WorkflowBlock) {
	sort.SliceStable(blocks, func(i, j int) bool {
		yi, yj := workflowBlockY(blocks[i]), workflowBlockY(blocks[j])
		if yi != yj {
			return yi < yj
		}
		xi, xj := workflowBlockX(blocks[i]), workflowBlockX(blocks[j])
		if xi != xj {
			return xi < xj
		}
		return blocks[i].ID < blocks[j].ID
	})
}

func sortWorkflowASCIIEdges(edges []retab.WorkflowEdgeDoc) {
	sort.SliceStable(edges, func(i, j int) bool {
		left := workflowASCIIEdgeSortKey(edges[i])
		right := workflowASCIIEdgeSortKey(edges[j])
		if left != right {
			return left < right
		}
		return edges[i].ID < edges[j].ID
	})
}

func workflowASCIIEdgeSortKey(edge retab.WorkflowEdgeDoc) string {
	return strings.Join([]string{
		edge.SourceBlock,
		derefString(edge.SourceHandle),
		derefString(edge.TargetHandle),
		edge.TargetBlock,
	}, "\x00")
}

func workflowASCIIEdgeLabel(edge retab.WorkflowEdgeDoc) string {
	source := workflowASCIIHandleLabel(derefString(edge.SourceHandle))
	target := workflowASCIIHandleLabel(derefString(edge.TargetHandle))
	if workflowASCIIIsDefaultHandle(source) {
		source = ""
	}
	if workflowASCIIIsDefaultHandle(target) {
		target = ""
	}

	var label string
	switch {
	case target != "":
		label = target
	case source != "":
		label = source
	default:
		return ""
	}
	return workflowASCIIFit(label, workflowASCIIMaxEdgeText)
}

func workflowASCIIHandleLabel(handle string) string {
	handle = strings.TrimSpace(handle)
	if handle == "" {
		return ""
	}
	for _, prefix := range []string{
		"input-json-",
		"output-json-",
		"input-file-",
		"output-file-",
		"input-",
		"output-",
	} {
		handle = strings.TrimPrefix(handle, prefix)
	}
	handle = strings.ReplaceAll(handle, "_", " ")
	handle = strings.ReplaceAll(handle, "-", " ")
	handle = strings.Join(strings.Fields(handle), " ")
	return handle
}

func workflowASCIIIsDefaultHandle(handle string) bool {
	switch strings.TrimSpace(handle) {
	case "", "0", "json 0", "file 0", "data", "document",
		"fe left in", "fe left out", "fe right in", "fe right out",
		"loop left in", "loop left out", "loop right in", "loop right out", "loop termination":
		return true
	default:
		return false
	}
}

func workflowASCIIIsContainerFeedbackHandle(handle string) bool {
	switch workflowASCIIHandleLabel(handle) {
	case "fe right in", "loop right in", "loop termination":
		return true
	default:
		return false
	}
}

func workflowASCIIBlockLabel(block retab.WorkflowBlock) string {
	label := strings.TrimSpace(derefString(block.Label))
	if label == "" {
		label = block.ID
	}
	return strings.Join(strings.Fields(label), " ")
}

func workflowASCIIBlockMeta(block retab.WorkflowBlock) string {
	// Every cell in this column is a block, so the leading `block_` prefix is
	// redundant noise. Strip legacy `blk_` too so older workflows render with
	// the same compact shape.
	displayID := workflowASCIIStripBlockPrefix(block.ID)
	if block.Type == "" {
		return workflowASCIIShortID(displayID)
	}
	typeTag := " [" + string(block.Type) + "]"
	short := workflowASCIIShortID(displayID)
	// Box content is capped at workflowASCIIMaxBoxWidth-4. When id+tag would
	// overflow, shorten the id further so the type tag survives — the type
	// is more useful than the trailing chars of an opaque id.
	maxLen := workflowASCIIMaxBoxWidth - 4
	if len(short)+len(typeTag) <= maxLen {
		return short + typeTag
	}
	idBudget := maxLen - len(typeTag)
	switch {
	case idBudget >= 10:
		// Room for prefix + "..." + 4-char suffix.
		prefix := idBudget - 7
		return displayID[:prefix] + "..." + displayID[len(displayID)-4:] + typeTag
	case idBudget >= 5:
		// Tight budget — keep the nanoid suffix (the only distinguishing
		// part) and prepend "..." to signal truncation. Drop the head
		// rather than the tail: a 4-char head of a nanoid collides
		// across blocks at the same rate as random chance, whereas the
		// suffix is what humans copy/paste to disambiguate.
		suffixLen := min(idBudget-3, len(displayID))
		return "..." + displayID[len(displayID)-suffixLen:] + typeTag
	default:
		// Type tag eats the whole budget; let the canvas truncate the tail.
		return short + typeTag
	}
}

// workflowASCIIStripBlockPrefix removes redundant block-id prefixes for display
// purposes only. The stored id is unchanged, and legacy `blk_` remains accepted.
func workflowASCIIStripBlockPrefix(id string) string {
	for _, prefix := range []string{"block_", "blk_"} {
		if rest := strings.TrimPrefix(id, prefix); rest != id && rest != "" {
			return rest
		}
	}
	return id
}

func workflowASCIIShortID(id string) string {
	if len(id) <= 24 {
		return id
	}
	return id[:12] + "..." + id[len(id)-6:]
}

func workflowASCIIMaxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func init() {
	workflowsCmd.AddCommand(workflowsViewCmd)
}
