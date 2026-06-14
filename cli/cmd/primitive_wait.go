package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

type primitiveWaitSpec struct {
	singular    string
	idName      string
	path        string
	commandPath string
}

var (
	classificationWaitSpec = primitiveWaitSpec{singular: "classification", idName: "classification-id", path: "/v1/classifications"}
	editWaitSpec           = primitiveWaitSpec{singular: "edit", idName: "edit-id", path: "/v1/edits"}
	extractionWaitSpec     = primitiveWaitSpec{singular: "extraction", idName: "extraction-id", path: "/v1/extractions"}
	fileBlueprintWaitSpec  = primitiveWaitSpec{singular: "file blueprint", idName: "blueprint-id", path: "/v1/files/blueprints", commandPath: "files blueprints"}
	parseWaitSpec          = primitiveWaitSpec{singular: "parse", idName: "parse-id", path: "/v1/parses"}
	partitionWaitSpec      = primitiveWaitSpec{singular: "partition", idName: "partition-id", path: "/v1/partitions"}
	splitWaitSpec          = primitiveWaitSpec{singular: "split", idName: "split-id", path: "/v1/splits"}
	// schemaGenerationWaitSpec drives `schemas generate --wait` / `schemas wait`.
	// The background primitive lives under /v1/schemas/generate (not /v1/schemas),
	// so get/cancel are /v1/schemas/generate/{id}[/cancel]; commandPath pins the
	// command group to `schemas` for the help/examples.
	schemaGenerationWaitSpec = primitiveWaitSpec{singular: "schema generation", idName: "schema-generation-id", path: "/v1/schemas/generate", commandPath: "schemas"}
)

func addPrimitiveCreateWaitFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("wait", false, "block until the primitive reaches a terminal status, then print the final record")
	addPrimitiveWaitTuningFlags(cmd, true)
}

// addPrimitiveBackgroundFlag registers the --background async-create switch on a
// primitive create command. It is kept separate from addPrimitiveCreateWaitFlags
// because `files blueprints create` registers its own --background (with a
// different default), so folding it into the shared helper would double-register
// the flag and panic at startup. Default false preserves the synchronous create
// behavior; combine with --wait to block until the queued primitive completes.
func addPrimitiveBackgroundFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("background", false, "run asynchronously: return immediately with status \"queued\" and an empty output, then poll GET until the status is terminal (combine with --wait to block until completion)")
}

// primitiveBackgroundParam reads the --background flag as a request-body pointer.
// It is always non-nil (false = synchronous) so the server receives an explicit
// background value.
func primitiveBackgroundParam(cmd *cobra.Command) *bool {
	background, _ := cmd.Flags().GetBool("background")
	return ptr(background)
}

func addPrimitiveWaitTuningFlags(cmd *cobra.Command, isCreate bool) {
	pollDescription := "poll cadence in milliseconds"
	timeoutDescription := "max seconds to wait before giving up"
	if isCreate {
		pollDescription = "poll cadence in milliseconds while --wait is set"
		timeoutDescription = "max seconds to wait while --wait is set"
	}
	cmd.Flags().Int("poll-interval-ms", 2000, pollDescription)
	cmd.Flags().Int("timeout-seconds", 600, timeoutDescription)
}

func primitiveWaitCommand(spec primitiveWaitSpec) *cobra.Command {
	commandPath := spec.commandPath
	if commandPath == "" {
		commandPath = spec.path[4:]
	}
	return &cobra.Command{
		Use:   "wait <" + spec.idName + ">",
		Short: "Poll until a " + spec.singular + " reaches a terminal status",
		Long: `Block until the primitive reaches a terminal status
(` + "`completed`" + `, ` + "`error`" + `, or ` + "`cancelled`" + `),
polling on a configurable interval. Defaults: 2-second polls, 10-minute
timeout.

Cleaner than scripting a poll loop around ` + "`get`" + ` — the CLI handles
the interval and timeout, prints the final record, and exits non-zero if
the primitive ends in ` + "`error`" + `/` + "`cancelled`" + ` or the timeout
elapses.`,
		Example: `  # Wait with defaults
  retab ` + commandPath + ` wait ` + spec.idName + `

  # Faster polls, longer ceiling
  retab ` + commandPath + ` wait ` + spec.idName + ` \
    --poll-interval-ms 1000 --timeout-seconds 1800`,
		Args: cobra.ExactArgs(1),
		RunE: runE(func(cmd *cobra.Command, args []string) error {
			pollInterval, timeout := primitiveWaitDurations(cmd)
			return waitForPrimitiveAndPrint(cmd, spec, args[0], pollInterval, timeout)
		}),
	}
}

func maybeWaitForPrimitiveCreate(cmd *cobra.Command, spec primitiveWaitSpec, result any) error {
	wait, _ := cmd.Flags().GetBool("wait")
	if !wait {
		return printJSON(result)
	}
	id, err := primitiveID(result)
	if err != nil {
		return err
	}
	pollInterval, timeout := primitiveWaitDurations(cmd)
	return waitForPrimitiveAndPrint(cmd, spec, id, pollInterval, timeout)
}

func waitForPrimitiveAndPrint(
	cmd *cobra.Command,
	spec primitiveWaitSpec,
	id string,
	pollInterval time.Duration,
	timeout time.Duration,
) error {
	ctx, cancel := ctxFor(cmd)
	defer cancel()
	final, waitErr := waitForPrimitive(ctx, cmd, spec, id, pollInterval, timeout)
	if final != nil {
		if err := printJSON(final); err != nil {
			return err
		}
	}
	if waitErr != nil {
		return waitErr
	}
	return primitiveTerminalError(spec, final)
}

func primitiveWaitDurations(cmd *cobra.Command) (time.Duration, time.Duration) {
	pollMS, _ := cmd.Flags().GetInt("poll-interval-ms")
	timeoutS, _ := cmd.Flags().GetInt("timeout-seconds")
	pollInterval := 2 * time.Second
	if pollMS > 0 {
		pollInterval = time.Duration(pollMS) * time.Millisecond
	}
	timeout := 10 * time.Minute
	if timeoutS > 0 {
		timeout = time.Duration(timeoutS) * time.Second
	}
	return pollInterval, timeout
}

func waitForPrimitive(
	ctx context.Context,
	cmd *cobra.Command,
	spec primitiveWaitSpec,
	id string,
	pollInterval time.Duration,
	timeout time.Duration,
) (map[string]any, error) {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var last map[string]any
	var lastErr error
	for {
		// A poll error is treated as transient (the server may be redeploying /
		// briefly unreachable, or a background primitive's host instance is
		// restarting) and retried until the primitive reaches a terminal status or
		// the overall timeout elapses — rather than aborting the wait on the first
		// blip. A genuinely persistent error surfaces when the timeout fires.
		result, err := cliJSONRequest(cmd, http.MethodGet, primitiveGetPath(spec, id), nil, nil)
		if err != nil {
			lastErr = err
		} else if current, perr := primitiveMap(result); perr != nil {
			lastErr = perr
		} else {
			last = current
			lastErr = nil
			if isTerminalPrimitiveStatus(primitiveStatus(current)) {
				return current, nil
			}
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			if lastErr != nil {
				return last, fmt.Errorf("gave up waiting for %s %s after repeated poll errors: %w", spec.singular, id, lastErr)
			}
			return last, fmt.Errorf("timed out waiting for %s %s: %w", spec.singular, id, ctx.Err())
		case <-timer.C:
		}
	}
}

func primitiveGetPath(spec primitiveWaitSpec, id string) string {
	return spec.path + "/" + url.PathEscape(id)
}

func primitiveID(value any) (string, error) {
	resource, err := primitiveMap(value)
	if err != nil {
		return "", err
	}
	id, ok := resource["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("created primitive response did not include an id")
	}
	return id, nil
}

func primitiveMap(value any) (map[string]any, error) {
	if resource, ok := value.(map[string]any); ok {
		return resource, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode primitive response: %w", err)
	}
	var resource map[string]any
	if err := json.Unmarshal(raw, &resource); err != nil {
		return nil, fmt.Errorf("decode primitive response: %w", err)
	}
	return resource, nil
}

func primitiveStatus(resource map[string]any) string {
	if status, ok := resource["status"].(string); ok {
		return status
	}
	lifecycle, ok := resource["lifecycle"].(map[string]any)
	if !ok {
		return ""
	}
	status, _ := lifecycle["status"].(string)
	return status
}

func isTerminalPrimitiveStatus(status string) bool {
	switch status {
	// "failed" is the real terminal-failure value: primitives expose a
	// ClassificationStatus (pending/queued/in_progress/completed/failed/
	// cancelled). "error" is kept as a defensive alias for resources that
	// surface a lifecycle-style status instead.
	case "completed", "error", "failed", "cancelled":
		return true
	default:
		return false
	}
}

func primitiveTerminalError(spec primitiveWaitSpec, resource map[string]any) error {
	if resource == nil {
		return nil
	}
	status := primitiveStatus(resource)
	switch status {
	case "", "completed":
		return nil
	case "error", "failed", "cancelled":
		id, _ := resource["id"].(string)
		if id == "" {
			return fmt.Errorf("%s ended with status %s", spec.singular, status)
		}
		return fmt.Errorf("%s %s ended with status %s", spec.singular, id, status)
	default:
		return nil
	}
}
