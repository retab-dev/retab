//go:build !retab_oagen_cli_workflows

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var workflowsAPICallsCmd = &cobra.Command{
	Use:   "api-calls",
	Short: "Develop workflow API call blocks locally",
	Long: `Hydrate, render, and optionally execute local workflow api_call bundles.

Start by pulling an api_call block:

  retab workflows blocks config pull <workflow-id> <block-id> --out tmp/api

Bundles pulled with config pull are hydrated automatically. Re-run hydrate when
you need to repair or regenerate local support files:

  retab workflows blocks api-calls hydrate tmp/api

Render requests against local samples without making network calls:

  retab workflows blocks api-calls run tmp/api samples/*.json

Actual HTTP execution is opt-in:

  retab workflows blocks api-calls run tmp/api samples/*.json --execute`,
}

var workflowsAPICallsHydrateCmd = &cobra.Command{
	Use:   "hydrate <bundle-dir>",
	Short: "Create local runtime files for a pulled api_call block bundle",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		fillSecrets, _ := cmd.Flags().GetBool("fill-secrets")
		forceSecrets, _ := cmd.Flags().GetBool("force-secrets")
		manifest, config, err := readBlockConfigBundle(args[0])
		if err != nil {
			return err
		}
		if manifest.BlockType != "api_call" {
			return fmt.Errorf("bundle block_type is %q, expected api_call", manifest.BlockType)
		}
		result, err := hydrateAPICallBundleDetailed(args[0], config, force)
		if err != nil {
			return err
		}
		filledSecrets := []map[string]any{}
		if fillSecrets {
			filledSecrets, err = fillLocalSecretsFromRetab(cmd, args[0], config, forceSecrets)
			if err != nil {
				return err
			}
		}
		return printJSON(map[string]any{
			"ok":                  true,
			"dir":                 args[0],
			"mode":                "repair",
			"workflow_id":         manifest.WorkflowID,
			"block_id":            manifest.BlockID,
			"files":               result.Files,
			"directories":         result.Directories,
			"removed_stale_files": result.RemovedStaleFiles,
			"filled_secrets":      filledSecrets,
		})
	}),
}

var workflowsAPICallsRenderCmd = &cobra.Command{
	Use:   "render <bundle-dir> <input-json>...",
	Short: "Render API call requests locally without making HTTP requests",
	Args:  cobra.MinimumNArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		return runAPICallBundleCommand(cmd, args, false)
	}),
}

var workflowsAPICallsRunCmd = &cobra.Command{
	Use:   "run <bundle-dir> <input-json>...",
	Short: "Render API call requests locally; execute only with --execute",
	Args:  cobra.MinimumNArgs(2),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		executeHTTP, _ := cmd.Flags().GetBool("execute")
		return runAPICallBundleCommand(cmd, args, executeHTTP)
	}),
}

type localAPICallRequest struct {
	Method         string
	URL            string
	Headers        map[string]string
	TimeoutSeconds int
	Body           any
}

type localAPICallSecret struct {
	Name     string
	Env      string
	Required bool
}

type apiCallHydrateResult struct {
	Files             []string
	Directories       []string
	RemovedStaleFiles []string
}

var apiCallEnvRefPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

var apiCallSensitiveHeaderNames = map[string]bool{
	"authorization":       true,
	"proxy-authorization": true,
	"x-api-key":           true,
	"api-key":             true,
	"x-auth-token":        true,
	"cookie":              true,
	"set-cookie":          true,
}

func runAPICallBundleCommand(cmd *cobra.Command, args []string, executeHTTP bool) error {
	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}
	manifest, config, err := readBlockConfigBundle(dir)
	if err != nil {
		return err
	}
	if manifest.BlockType != "api_call" {
		return fmt.Errorf("bundle block_type is %q, expected api_call", manifest.BlockType)
	}
	outDir, _ := cmd.Flags().GetString("out")
	jobs, _ := cmd.Flags().GetString("jobs")
	timeoutRaw, _ := cmd.Flags().GetString("timeout")
	recursive, _ := cmd.Flags().GetBool("recursive")
	continueOnError, _ := cmd.Flags().GetBool("continue-on-error")
	clean, _ := cmd.Flags().GetBool("clean")
	absolutePaths, _ := cmd.Flags().GetBool("absolute-paths")
	timeout, err := parseFunctionRunTimeout(timeoutRaw)
	if err != nil {
		return err
	}
	if err := validateFunctionRunJobs(jobs); err != nil {
		return err
	}
	if err := validateFunctionRunOutDir(outDir); err != nil {
		return err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	env, err := loadLocalAPICallEnv(dir)
	if err != nil {
		return err
	}
	if err := validateLocalAPICallRequiredEnv(config, env); err != nil {
		return err
	}
	inputs, err := listLocalAPICallInputs(args[1:], recursive)
	if err != nil {
		return err
	}
	if len(inputs) == 0 {
		return fmt.Errorf("no input JSON files matched")
	}

	outputDir := filepath.Join(dir, filepath.Clean(outDir))
	traceDir := filepath.Join(dir, "traces")
	renderedDir := filepath.Join(dir, "rendered")
	if clean {
		for _, path := range []string{outputDir, traceDir, renderedDir} {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		}
	}
	for _, path := range []string{outputDir, traceDir, renderedDir} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
	}
	// Honor an explicitly-passed --jobs even without --continue-on-error:
	// previously the flag was silently ignored on the fail-fast path (only
	// --continue-on-error reached the parallel runner), so `--jobs 8` ran
	// strictly one-at-a-time. Fail-fast semantics are preserved in the
	// parallel runner (no new samples dispatched after a failure, the first
	// error is returned). The default (no --jobs) stays sequential.
	if continueOnError || (cmd.Flags().Changed("jobs") && localAPICallWorkerCount(jobs, len(inputs)) > 1) {
		return runLocalAPICallsParallel(ctx, dir, outputDir, traceDir, config, env, inputs, jobs, executeHTTP, absolutePaths, continueOnError)
	}
	return runLocalAPICallsSequential(ctx, dir, outputDir, traceDir, config, env, inputs, executeHTTP, absolutePaths)
}

func hydrateAPICallBundle(dir string, config map[string]any, force bool) error {
	_, err := hydrateAPICallBundleDetailed(dir, config, force)
	return err
}

func hydrateAPICallBundleDetailed(dir string, config map[string]any, force bool) (apiCallHydrateResult, error) {
	result := apiCallHydrateResult{
		Files:       []string{"run.sh", ".env.example", ".env.local"},
		Directories: []string{"samples/", "rendered/", "outputs/", "traces/"},
	}
	if err := writeTextFileIfAllowed(filepath.Join(dir, "run.sh"), generatedAPICallRunSh, force, 0o700); err != nil {
		return result, err
	}
	for _, rel := range []string{"run.py", "curl.sh"} {
		if err := os.Remove(filepath.Join(dir, rel)); err != nil {
			if !os.IsNotExist(err) {
				return result, err
			}
			continue
		}
		result.RemovedStaleFiles = append(result.RemovedStaleFiles, rel)
	}
	secrets := collectFunctionSecretEnvNames(config)
	if err := writeTextFileIfAllowed(filepath.Join(dir, ".env.example"), renderEnvFile(secrets, false), force, 0o600); err != nil {
		return result, err
	}
	if _, err := os.Stat(filepath.Join(dir, ".env.local")); os.IsNotExist(err) {
		if err := writeTextFileIfAllowed(filepath.Join(dir, ".env.local"), renderEnvFile(secrets, true), true, 0o600); err != nil {
			return result, err
		}
	}
	for _, rel := range []string{"samples", "rendered", "outputs", "traces"} {
		if err := os.MkdirAll(filepath.Join(dir, rel), 0o755); err != nil {
			return result, err
		}
	}
	sort.Strings(result.RemovedStaleFiles)
	return result, nil
}

func runLocalAPICallsSequential(ctx context.Context, bundleDir string, outDir string, traceDir string, config map[string]any, env map[string]string, inputs []string, executeHTTP bool, absolutePaths bool) error {
	for _, input := range inputs {
		result, err := runOneLocalAPICall(ctx, bundleDir, outDir, traceDir, config, env, input, executeHTTP, absolutePaths)
		if err != nil {
			return err
		}
		printLocalAPICallResult(result)
	}
	return nil
}

func runLocalAPICallsParallel(ctx context.Context, bundleDir string, outDir string, traceDir string, config map[string]any, env map[string]string, inputs []string, jobsRaw string, executeHTTP bool, absolutePaths bool, continueOnError bool) error {
	workers := localAPICallWorkerCount(jobsRaw, len(inputs))
	inputCh := make(chan string)
	var wg sync.WaitGroup
	var printMu sync.Mutex
	var stateMu sync.Mutex
	sawFailure := false
	var firstErr error
	for index := 0; index < workers; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for input := range inputCh {
				result, err := runOneLocalAPICall(ctx, bundleDir, outDir, traceDir, config, env, input, executeHTTP, absolutePaths)
				if err != nil {
					stateMu.Lock()
					sawFailure = true
					if firstErr == nil {
						firstErr = err
					}
					stateMu.Unlock()
					result = map[string]any{
						"input": input,
						"ok":    false,
						"error": err.Error(),
					}
				}
				printMu.Lock()
				printLocalAPICallResult(result)
				printMu.Unlock()
			}
		}()
	}
	for _, input := range inputs {
		// Fail-fast mode (--jobs without --continue-on-error): stop handing
		// out new samples once one has failed; in-flight samples finish.
		if !continueOnError {
			stateMu.Lock()
			failed := sawFailure
			stateMu.Unlock()
			if failed {
				break
			}
		}
		inputCh <- input
	}
	close(inputCh)
	wg.Wait()
	if !continueOnError && firstErr != nil {
		return firstErr
	}
	if sawFailure {
		return fmt.Errorf("one or more local api_call samples failed; inspect trace files under %s", traceDir)
	}
	return nil
}

func localAPICallWorkerCount(raw string, inputCount int) int {
	if inputCount <= 0 {
		return 1
	}
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "auto" {
		workers := runtime.NumCPU()
		if workers < 1 {
			workers = 1
		}
		if workers > inputCount {
			return inputCount
		}
		return workers
	}
	workers, err := strconv.Atoi(raw)
	if err != nil || workers < 1 {
		return 1
	}
	if workers > inputCount {
		return inputCount
	}
	return workers
}

func printLocalAPICallResult(result map[string]any) {
	raw, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode api_call result: %v\n", err)
		return
	}
	fmt.Fprintln(os.Stdout, string(raw))
}

func runOneLocalAPICall(ctx context.Context, bundleDir string, outDir string, traceDir string, config map[string]any, env map[string]string, inputPath string, executeHTTP bool, absolutePaths bool) (map[string]any, error) {
	payload, err := readJSON(inputPath)
	if err != nil {
		writeLocalAPICallErrorTrace(bundleDir, traceDir, inputPath, executeHTTP, err)
		return nil, err
	}
	request := compileLocalAPICallRequest(config, env, payload)
	requestPath, requestArtifact, err := writeRenderedLocalAPICall(bundleDir, inputPath, request, absolutePaths)
	if err != nil {
		writeLocalAPICallErrorTrace(bundleDir, traceDir, inputPath, executeHTTP, err)
		return nil, err
	}
	result := map[string]any{
		"input":    localAPICallDisplayPath(bundleDir, inputPath, absolutePaths),
		"request":  requestPath,
		"executed": executeHTTP,
		"ok":       true,
	}
	trace := map[string]any{
		"input":            localAPICallDisplayPath(bundleDir, inputPath, absolutePaths),
		"request":          requestPath,
		"executed":         executeHTTP,
		"ok":               true,
		"request_artifact": requestArtifact,
	}
	if executeHTTP {
		response, err := executeLocalAPICallRequest(ctx, request)
		if err != nil {
			trace["ok"] = false
			trace["error"] = err.Error()
			_ = writeLocalAPICallTrace(bundleDir, traceDir, inputPath, trace)
			return nil, err
		}
		outPath := filepath.Join(outDir, localAPICallOutputStem(bundleDir, inputPath)+".out.json")
		if err := writeLocalAPICallJSONFile(outPath, response); err != nil {
			return nil, err
		}
		result["output"] = localAPICallDisplayPath(bundleDir, outPath, absolutePaths)
		result["ok"] = response["ok"]
		trace["output"] = localAPICallDisplayPath(bundleDir, outPath, absolutePaths)
		trace["response"] = response
		if response["ok"] != true {
			trace["ok"] = false
			trace["error"] = fmt.Sprintf("HTTP %v", response["status_code"])
		}
	}
	if err := writeLocalAPICallTrace(bundleDir, traceDir, inputPath, trace); err != nil {
		return nil, err
	}
	if executeHTTP && result["ok"] != true {
		return result, fmt.Errorf("HTTP request failed for %s: %s", inputPath, trace["error"])
	}
	return result, nil
}

func writeLocalAPICallErrorTrace(bundleDir string, traceDir string, inputPath string, executeHTTP bool, err error) {
	_ = writeLocalAPICallTrace(bundleDir, traceDir, inputPath, map[string]any{
		"input":    inputPath,
		"executed": executeHTTP,
		"ok":       false,
		"error":    err.Error(),
	})
}

func writeLocalAPICallTrace(bundleDir string, traceDir string, inputPath string, trace map[string]any) error {
	return writeLocalAPICallJSONFile(filepath.Join(traceDir, localAPICallOutputStem(bundleDir, inputPath)+".trace.json"), trace)
}

func loadLocalAPICallEnv(bundleDir string) (map[string]string, error) {
	env := map[string]string{}
	for _, pair := range os.Environ() {
		key, value, ok := strings.Cut(pair, "=")
		if ok {
			env[key] = value
		}
	}
	raw, err := os.ReadFile(filepath.Join(bundleDir, ".env.local"))
	if err != nil {
		if os.IsNotExist(err) {
			return env, nil
		}
		return nil, err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		key, value, _ := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if strings.TrimSpace(env[key]) == "" {
			env[key] = value
		}
	}
	return env, nil
}

func validateLocalAPICallRequiredEnv(config map[string]any, env map[string]string) error {
	missing := []string{}
	for _, secret := range localAPICallSecrets(config) {
		value := strings.TrimSpace(env[secret.Env])
		if secret.Required && (value == "" || value == "__REPLACE_ME__") {
			missing = append(missing, fmt.Sprintf("%s -> %s", secret.Name, secret.Env))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required api_call env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}

func localAPICallSecrets(config map[string]any) []localAPICallSecret {
	mounts, ok := config["mounts"].(map[string]any)
	if !ok {
		return nil
	}
	rawSecrets, ok := mounts["secrets"].([]any)
	if !ok {
		return nil
	}
	secrets := []localAPICallSecret{}
	for _, raw := range rawSecrets {
		obj, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		env := strings.TrimSpace(apiCallStringFromAny(obj["env"]))
		name := strings.TrimSpace(apiCallStringFromAny(obj["name"]))
		if env == "" {
			env = name
		}
		if name == "" {
			name = env
		}
		if env == "" {
			continue
		}
		required := true
		if rawRequired, ok := obj["required"].(bool); ok {
			required = rawRequired
		}
		secrets = append(secrets, localAPICallSecret{Name: name, Env: env, Required: required})
	}
	return secrets
}

func listLocalAPICallInputs(rawPaths []string, recursive bool) ([]string, error) {
	inputs := []string{}
	for _, raw := range rawPaths {
		path, err := filepath.Abs(raw)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			inputs = append(inputs, path)
			continue
		}
		if recursive {
			err = filepath.WalkDir(path, func(child string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || filepath.Ext(child) != ".json" {
					return nil
				}
				inputs = append(inputs, child)
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		matches, err := filepath.Glob(filepath.Join(path, "*.json"))
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, matches...)
	}
	sort.Strings(inputs)
	return inputs, nil
}

func compileLocalAPICallRequest(config map[string]any, env map[string]string, payload any) localAPICallRequest {
	method := strings.ToUpper(strings.TrimSpace(apiCallStringFromAny(config["method"])))
	if method == "" {
		method = "POST"
	}
	body := any(nil)
	hasJSONBody := false
	if method == "POST" || method == "PUT" || method == "PATCH" {
		body = payload
		hasJSONBody = payload != nil
		if payloadMap, ok := payload.(map[string]any); ok {
			if _, hasSchema := config["request_schema"]; hasSchema || len(stringMapFromAny(config["field_mappings"])) > 0 {
				body = mapAndFilterLocalAPICallInput(payloadMap, mapFromAny(config["request_schema"]), stringMapFromAny(config["field_mappings"]))
			}
		}
	}
	headers := prepareLocalAPICallHeaders(interpolateLocalAPICallMap(stringMapFromAny(config["headers"]), env), hasJSONBody)
	// A non-positive timeout_seconds (explicit 0, or a negative/garbage
	// value) would make http.Client.Timeout == 0, i.e. NO timeout — and the
	// run's --timeout context defaults to 0 (disabled) too, so a hung
	// endpoint would block --execute forever. Fall back to the default.
	timeoutSeconds := intFromAny(config["timeout_seconds"], 180)
	if timeoutSeconds <= 0 {
		timeoutSeconds = 180
	}
	return localAPICallRequest{
		Method:         method,
		URL:            interpolateLocalAPICallString(apiCallStringFromAny(config["url"]), env),
		Headers:        headers,
		TimeoutSeconds: timeoutSeconds,
		Body:           body,
	}
}

func mapAndFilterLocalAPICallInput(payload map[string]any, requestSchema map[string]any, fieldMappings map[string]string) map[string]any {
	mapped := map[string]any{}
	if len(fieldMappings) > 0 {
		for source, target := range fieldMappings {
			if value, ok := payload[source]; ok {
				mapped[target] = value
			}
		}
		for key, value := range payload {
			if _, mappedKey := fieldMappings[key]; !mappedKey {
				mapped[key] = value
			}
		}
	} else {
		for key, value := range payload {
			mapped[key] = value
		}
	}
	properties := mapFromAny(requestSchema["properties"])
	if len(properties) == 0 {
		return mapped
	}
	filtered := map[string]any{}
	for key, value := range mapped {
		if _, ok := properties[key]; ok {
			filtered[key] = value
		}
	}
	return filtered
}

func prepareLocalAPICallHeaders(headers map[string]string, hasJSONBody bool) map[string]string {
	if !hasJSONBody {
		return headers
	}
	for key, value := range headers {
		if strings.EqualFold(key, "Content-Type") {
			mediaType := strings.ToLower(strings.TrimSpace(strings.SplitN(value, ";", 2)[0]))
			if mediaType == "application/json" || strings.HasSuffix(mediaType, "+json") {
				return headers
			}
			headers[key] = "application/json"
			return headers
		}
	}
	headers["Content-Type"] = "application/json"
	return headers
}

func interpolateLocalAPICallMap(values map[string]string, env map[string]string) map[string]string {
	resolved := map[string]string{}
	for key, value := range values {
		resolved[key] = interpolateLocalAPICallString(value, env)
	}
	return resolved
}

func interpolateLocalAPICallString(value string, env map[string]string) string {
	return apiCallEnvRefPattern.ReplaceAllStringFunc(value, func(match string) string {
		groups := apiCallEnvRefPattern.FindStringSubmatch(match)
		if len(groups) != 2 {
			return match
		}
		if resolved, ok := env[groups[1]]; ok {
			return resolved
		}
		return match
	})
}

func writeRenderedLocalAPICall(bundleDir string, inputPath string, request localAPICallRequest, absolutePaths bool) (string, map[string]any, error) {
	stem := localAPICallOutputStem(bundleDir, inputPath)
	renderedStem := filepath.Join(bundleDir, "rendered", stem)
	if err := os.MkdirAll(filepath.Dir(renderedStem), 0o755); err != nil {
		return "", nil, err
	}
	headersPath := renderedStem + ".headers.local.json"
	requestPath := renderedStem + ".request.json"
	curlPath := renderedStem + ".curl.sh"
	bodyPath := ""
	if request.Body != nil {
		bodyPath = renderedStem + ".body.json"
		if err := writeLocalAPICallJSONFile(bodyPath, request.Body); err != nil {
			return "", nil, err
		}
	}
	if err := writeLocalAPICallJSONFile(headersPath, request.Headers); err != nil {
		return "", nil, err
	}
	if err := writeRenderedLocalAPICallCurl(curlPath, request, bodyPath); err != nil {
		return "", nil, err
	}
	bodyPathValue := any(nil)
	if bodyPath != "" {
		bodyPathValue = localAPICallDisplayPath(bundleDir, bodyPath, absolutePaths)
	}
	artifact := map[string]any{
		"input":                   localAPICallDisplayPath(bundleDir, inputPath, absolutePaths),
		"method":                  request.Method,
		"url":                     request.URL,
		"headers":                 redactLocalAPICallHeaders(request.Headers),
		"headers_unredacted_path": localAPICallDisplayPath(bundleDir, headersPath, absolutePaths),
		"body_path":               bodyPathValue,
		"execute_command_path":    localAPICallDisplayPath(bundleDir, curlPath, absolutePaths),
		"ok":                      true,
	}
	if err := writeLocalAPICallJSONFile(requestPath, artifact); err != nil {
		return "", nil, err
	}
	return localAPICallDisplayPath(bundleDir, requestPath, absolutePaths), artifact, nil
}

func writeRenderedLocalAPICallCurl(path string, request localAPICallRequest, bodyPath string) error {
	parts := []string{"curl", "-sS", "-X", request.Method}
	for key, value := range request.Headers {
		parts = append(parts, "-H", fmt.Sprintf("%s: %s", key, value))
	}
	if bodyPath != "" {
		parts = append(parts, "--data-binary", "@"+bodyPath)
	}
	parts = append(parts, request.URL)
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, shellQuoteLocalAPICall(part))
	}
	content := "#!/usr/bin/env bash\nset -euo pipefail\n" + strings.Join(quoted, " ") + "\n"
	return writeTextFileIfAllowed(path, content, true, 0o700)
}

func writeLocalAPICallJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeJSONFile(path, value)
}

func executeLocalAPICallRequest(ctx context.Context, request localAPICallRequest) (map[string]any, error) {
	var bodyReader io.Reader
	if request.Body != nil {
		raw, err := json.Marshal(request.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(raw)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, request.URL, bodyReader)
	if err != nil {
		return nil, err
	}
	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}
	started := time.Now()
	client := &http.Client{Timeout: time.Duration(request.TimeoutSeconds) * time.Second}
	response, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	elapsedMS := float64(time.Since(started).Microseconds()) / 1000
	bodyText := string(rawBody)
	bodyJSON := any(nil)
	if strings.TrimSpace(bodyText) != "" {
		var decoded any
		if err := json.Unmarshal(rawBody, &decoded); err == nil {
			bodyJSON = decoded
			bodyText = ""
		}
	}
	bodyTextValue := any(nil)
	if bodyText != "" {
		bodyTextValue = bodyText
	}
	return map[string]any{
		"status_code": response.StatusCode,
		"headers":     redactLocalAPICallHeaders(headerMapFromHTTP(response.Header)),
		"body_text":   bodyTextValue,
		"body_json":   bodyJSON,
		"elapsed_ms":  elapsedMS,
		"ok":          response.StatusCode >= 200 && response.StatusCode < 400,
	}, nil
}

func localAPICallOutputStem(bundleDir string, inputPath string) string {
	absInput, err := filepath.Abs(inputPath)
	if err != nil {
		return strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	}
	absBundle, err := filepath.Abs(bundleDir)
	if err == nil {
		rel, relErr := filepath.Rel(absBundle, absInput)
		if relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return strings.TrimSuffix(rel, filepath.Ext(rel))
		}
	}
	return strings.TrimSuffix(filepath.Base(absInput), filepath.Ext(absInput))
}

func localAPICallDisplayPath(bundleDir string, path string, absolute bool) string {
	if absolute {
		abs, err := filepath.Abs(path)
		if err == nil {
			return abs
		}
		return path
	}
	absPath, pathErr := filepath.Abs(path)
	absBundle, bundleErr := filepath.Abs(bundleDir)
	if pathErr == nil && bundleErr == nil {
		rel, relErr := filepath.Rel(absBundle, absPath)
		if relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return filepath.ToSlash(rel)
		}
	}
	return path
}

func redactLocalAPICallHeaders(headers map[string]string) map[string]string {
	redacted := map[string]string{}
	for key, value := range headers {
		if apiCallSensitiveHeaderNames[strings.ToLower(key)] {
			redacted[key] = "[REDACTED]"
			continue
		}
		redacted[key] = value
	}
	return redacted
}

func headerMapFromHTTP(headers http.Header) map[string]string {
	out := map[string]string{}
	for key, values := range headers {
		out[key] = strings.Join(values, ", ")
	}
	return out
}

func stringMapFromAny(value any) map[string]string {
	out := map[string]string{}
	for key, raw := range mapFromAny(value) {
		out[key] = apiCallStringFromAny(raw)
	}
	return out
}

func mapFromAny(value any) map[string]any {
	obj, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return obj
}

func apiCallStringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func intFromAny(value any, fallback int) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func shellQuoteLocalAPICall(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

const generatedAPICallRunSh = `#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if ! command -v retab >/dev/null 2>&1; then
  echo "retab CLI not found on PATH; run: retab workflows blocks api-calls run \"$DIR\" <sample.json> --execute" >&2
  exit 127
fi
exec retab workflows blocks api-calls run "$DIR" "$@" --execute
`

func init() {
	workflowsAPICallsHydrateCmd.Flags().Bool("force", false, "overwrite generated runtime files if they already exist")
	workflowsAPICallsHydrateCmd.Flags().Bool("fill-secrets", false, "fill .env.local from Retab secrets when the API supports secret value reads")
	workflowsAPICallsHydrateCmd.Flags().Bool("force-secrets", false, "overwrite existing .env.local secret values when used with --fill-secrets")
	for _, command := range []*cobra.Command{workflowsAPICallsRenderCmd, workflowsAPICallsRunCmd} {
		command.Flags().String("out", "outputs", "output directory inside the bundle")
		command.Flags().String("jobs", "auto", "parallel local executions: auto or a positive integer")
		command.Flags().String("timeout", "0", "maximum wall-clock duration for the local run, e.g. 30s or 5m; 0 disables")
		command.Flags().Bool("recursive", false, "recurse into input directories")
		command.Flags().Bool("continue-on-error", false, "continue processing remaining samples after a sample fails")
		command.Flags().Bool("clean", false, "remove outputs, rendered requests, and traces before running")
		command.Flags().Bool("absolute-paths", false, "write absolute paths in CLI and JSON artifacts")
	}
	workflowsAPICallsRunCmd.Flags().Bool("execute", false, "perform HTTP requests; without this flag requests are rendered only")

	workflowsAPICallsCmd.AddCommand(
		workflowsAPICallsHydrateCmd,
		workflowsAPICallsRenderCmd,
		workflowsAPICallsRunCmd,
	)
	workflowsBlocksCmd.AddCommand(workflowsAPICallsCmd)
}
