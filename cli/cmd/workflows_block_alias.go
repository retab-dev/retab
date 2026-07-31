package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
)

// A workflow block has two names. The spec alias is what the author wrote (and
// what `spec get` dumps back): the declarative block id, or its path within the
// spec. The runtime id is the opaque `block_<nanoid>` the server minted, which
// only `workflows blocks list` and the step records show.
//
// `runs create --document <alias>=file.pdf` resolves the alias, so a user who
// authored a workflow from a spec never sees the runtime id — and then reaches
// for the same alias on the commands that take `--block-id` and gets a 404 or,
// worse, a `422 Step data not found for this block in the given run` that reads
// like the run never executed the block rather than like the id is not the one
// that endpoint indexes by.
//
// Resolution is deliberately LAZY: the call goes out with the id the user typed,
// and the block list is fetched only if that call fails. A runtime id — the
// common case, and the only thing scripts hold — therefore still costs exactly
// one request. It also means resolution can only ever turn a failure into a
// success, never perturb a call that already worked.
//
// A var only so the endless-pagination regression test can shrink it, mirroring
// usageBlockLookupTimeout.
var workflowBlockAliasLookupTimeout = 15 * time.Second

// callWithBlockAliasFallback runs call with the id the user supplied and, if that
// fails, retries once with the resolved runtime id when resolution yields a
// different one.
//
// A failed retry keeps the ORIGINAL error. The user typed the alias, so an error
// naming an id they never wrote would send them looking in the wrong place.
func callWithBlockAliasFallback[T any](blockID string, call func(string) (T, error), resolve func() string) (T, error) {
	result, err := call(blockID)
	if err == nil {
		return result, nil
	}
	resolved := strings.TrimSpace(resolve())
	if resolved == "" || resolved == blockID {
		return result, err
	}
	retried, retryErr := call(resolved)
	if retryErr != nil {
		return result, err
	}
	return retried, nil
}

// resolveWorkflowBlockAlias maps a spec alias onto the runtime block id. A value
// that is already a runtime id, that matches no alias, or that matches more than
// one is returned untouched — an ambiguous hit is better served by the raw value
// than by an arbitrary pick. Every lookup failure degrades to the raw value.
func resolveWorkflowBlockAlias(ctx context.Context, client *retab.Client, workflowID string, blockID string) string {
	blockID = strings.TrimSpace(blockID)
	workflowID = strings.TrimSpace(workflowID)
	if blockID == "" || workflowID == "" || client == nil {
		return blockID
	}
	// Hard deadline, because listAllWorkflowBlocks walks every page: a blocks
	// endpoint that keeps handing back a cursor would otherwise spin forever
	// inside what is only a retry hint.
	lookupCtx, cancel := context.WithTimeout(ctx, workflowBlockAliasLookupTimeout)
	defer cancel()
	blocks, err := listAllWorkflowBlocks(lookupCtx, client, workflowID)
	if err != nil {
		return blockID
	}
	return matchWorkflowBlockAlias(blockID, blocks)
}

// matchWorkflowBlockAlias is resolveWorkflowBlockAlias's pure half, over an
// already-fetched block list.
func matchWorkflowBlockAlias(blockID string, blocks []retab.WorkflowBlock) string {
	for i := range blocks {
		if blocks[i].ID == blockID {
			return blockID
		}
	}
	matched := ""
	hits := 0
	for i := range blocks {
		decl := blocks[i].DeclarativeSourceBlockID
		path := blocks[i].DeclarativePath
		if (decl != nil && *decl == blockID) || (path != nil && *path == blockID) {
			matched = blocks[i].ID
			hits++
		}
	}
	if hits == 1 {
		return matched
	}
	return blockID
}

// rejectPublishHistoryIDAsVersion catches `--version wph_...`.
//
// `workflows versions list` returns publish-history records, whose own `id` is a
// `wph_` and whose `workflow_version_id` is the `ver_` that identifies the
// immutable version. Reaching for the record's `id` — the obvious field — earns a
// bare "Workflow version was not found", which reads like the version was
// deleted rather than like the wrong field was copied. Name the right one.
func rejectPublishHistoryIDAsVersion(version string) error {
	version = strings.TrimSpace(version)
	if !strings.HasPrefix(version, "wph_") {
		return nil
	}
	return fmt.Errorf(
		"--version %q is a publish-history id, not a workflow version id; "+
			"use the `workflow_version_id` (ver_...) field of that same "+
			"`workflows versions list` row", version)
}

// resolveWorkflowBlockAliasForRun is the same resolution for the commands scoped
// to a run rather than a workflow: it reads the run to learn which workflow to
// resolve against. Any failure degrades to the raw value.
func resolveWorkflowBlockAliasForRun(ctx context.Context, client *retab.Client, runID string, blockID string) string {
	blockID = strings.TrimSpace(blockID)
	runID = strings.TrimSpace(runID)
	if blockID == "" || runID == "" || client == nil {
		return blockID
	}
	lookupCtx, cancel := context.WithTimeout(ctx, workflowBlockAliasLookupTimeout)
	defer cancel()
	run, err := client.Workflows.Runs.Get(lookupCtx, runID, nil)
	if err != nil || run == nil {
		return blockID
	}
	return resolveWorkflowBlockAlias(ctx, client, strings.TrimSpace(run.WorkflowID), blockID)
}
