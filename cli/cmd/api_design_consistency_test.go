package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type cliOpenAPIContract struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

type cliRouteContract struct {
	method string
	path   string
}

// approvedWorkflowCLINonReferenceRoutes are workflow routes the CLI intentionally
// exposes ahead of the public OpenAPI reference. Drop entries here once the
// routes ship (the test below fails if a route lands in OpenAPI while still
// listed here, so the list can't go stale).
var approvedWorkflowCLINonReferenceRoutes = map[string]bool{}

// workflowRoutePathApproved reports whether any method of the given "/v1"-prefixed
// path is in approvedWorkflowCLINonReferenceRoutes, so the route-string loop below
// honors the same allowlist as the method+path loop.
func workflowRoutePathApproved(openAPIPath string) bool {
	for key := range approvedWorkflowCLINonReferenceRoutes {
		if strings.HasSuffix(key, " "+openAPIPath) {
			return true
		}
	}
	return false
}

func TestWorkflowCLIDesignContractMatchesOpenAPI(t *testing.T) {
	openAPI := loadCLIOpenAPIContract(t)
	used := map[string]bool{}
	for _, route := range workflowCLIRouteContract() {
		key := cliHTTPRouteKey(route)
		if approvedWorkflowCLINonReferenceRoutes[key] {
			used[key] = true
			if cliOpenAPIHasRoute(openAPI, route) {
				t.Fatalf("%s is now in OpenAPI; remove it from approvedWorkflowCLINonReferenceRoutes", key)
			}
			continue
		}
		assertCLIOpenAPIRoute(t, openAPI, route)
	}
	for key := range approvedWorkflowCLINonReferenceRoutes {
		if !used[key] {
			t.Fatalf("approved workflow CLI non-reference route %s is stale; no workflow route contract uses it", key)
		}
	}

	for _, path := range extractCLIWorkflowRouteStrings(readCLISource(t)) {
		openAPIPath := "/v1" + path
		if workflowRoutePathApproved(openAPIPath) {
			continue
		}
		if _, ok := openAPI.Paths[openAPIPath]; !ok {
			t.Fatalf("CLI route string %s is missing from OpenAPI contract", path)
		}
	}
}

func TestWorkflowCLIUsesOnlyCanonicalReviewBlockExecutionAndTestSurface(t *testing.T) {
	commands := collectLeafCommandPaths(rootCmd)

	for _, commandPath := range []string{
		"workflows reviews list",
		"workflows reviews get",
		"workflows reviews approve",
		"workflows reviews reject",
		"workflows reviews versions list",
		"workflows reviews versions get",
		"workflows reviews versions create",
		"workflows blocks executions create",
		"workflows blocks executions list",
		"workflows evals create",
		"workflows evals list",
		"workflows evals get",
		"workflows evals update",
		"workflows evals delete",
		"workflows evals runs create",
		"workflows evals runs list",
		"workflows evals runs get",
		"workflows evals runs cancel",
		"workflows evals results list",
		"workflows evals results get",
	} {
		if !commands[commandPath] {
			t.Fatalf("CLI is missing canonical command retab %s", commandPath)
		}
	}

	for _, removedCommandPath := range []string{
		"workflows reviews append",
		"workflows reviews edit",
		"workflows reviews versions append",
		"workflows blocks executions get",
	} {
		if commands[removedCommandPath] {
			t.Fatalf("CLI still exposes removed command retab %s", removedCommandPath)
		}
	}

	source := readCLISource(t)
	for _, required := range []string{
		`"/v1/workflows/evals/runs"`,
		`"/v1/workflows/evals/results"`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("CLI source is missing canonical route string %s", required)
		}
	}

	nestedWorkflowRoute := regexp.MustCompile(`"/workflows/"\s*\+\s*(?:url\.)?PathEscape\([^)]*workflowID[^)]*\)\s*\+\s*"/(?:reviews|block executions|tests)(?:/|")`)
	if match := nestedWorkflowRoute.FindString(source); match != "" {
		t.Fatalf("CLI must use flat workflow routes for reviews/blocks/executions/tests, found nested route expression: %s", match)
	}

	for _, removed := range []string{"append_version", "append-version", "appendVersion", "AppendVersion"} {
		if strings.Contains(source, removed) {
			t.Fatalf("CLI source must not expose removed append-version compatibility surface %q", removed)
		}
	}
}

// approvedDirectCLINonReferenceRoutes are raw cliJSONRequest routes that the
// CLI intentionally calls but that are deliberately excluded from the public
// OpenAPI reference (see documentation_only_routes in public_api_routes.yaml).
var approvedDirectCLINonReferenceRoutes = map[string]bool{}

func TestDirectCLIJSONRequestCallsMatchOpenAPI(t *testing.T) {
	openAPI := loadCLIOpenAPIContract(t)
	routes := extractDirectCLIJSONRequestRoutes(readCLISource(t))
	if len(routes) == 0 {
		t.Fatal("no direct cliJSONRequest calls found")
	}

	used := map[string]bool{}
	for _, route := range routes {
		key := cliHTTPRouteKey(route)
		if approvedDirectCLINonReferenceRoutes[key] {
			used[key] = true
			if cliOpenAPIHasRoute(openAPI, route) {
				t.Fatalf("%s is now in OpenAPI; remove it from approvedDirectCLINonReferenceRoutes", key)
			}
			continue
		}
		assertCLIOpenAPIRoute(t, openAPI, route)
	}

	for key := range approvedDirectCLINonReferenceRoutes {
		if !used[key] {
			t.Fatalf("approved direct CLI non-reference route %s is stale; no cliJSONRequest call uses it", key)
		}
	}
}

func workflowCLIRouteContract() []cliRouteContract {
	return []cliRouteContract{
		{method: http.MethodGet, path: "/workflows/reviews"},
		{method: http.MethodGet, path: "/workflows/reviews/{review_id}"},
		{method: http.MethodPost, path: "/workflows/reviews/{review_id}/approve"},
		{method: http.MethodPost, path: "/workflows/reviews/{review_id}/reject"},
		{method: http.MethodGet, path: "/workflows/reviews/versions"},
		{method: http.MethodPost, path: "/workflows/reviews/versions"},
		{method: http.MethodGet, path: "/workflows/reviews/versions/{version_id}"},
		{method: http.MethodPost, path: "/workflows/blocks/executions"},
		{method: http.MethodGet, path: "/workflows/blocks/executions"},
		{method: http.MethodPost, path: "/workflows/evals"},
		{method: http.MethodGet, path: "/workflows/evals"},
		{method: http.MethodGet, path: "/workflows/evals/{eval_id}"},
		{method: http.MethodPatch, path: "/workflows/evals/{eval_id}"},
		{method: http.MethodDelete, path: "/workflows/evals/{eval_id}"},
		{method: http.MethodPost, path: "/workflows/evals/runs"},
		{method: http.MethodGet, path: "/workflows/evals/runs"},
		{method: http.MethodGet, path: "/workflows/evals/runs/{run_id}"},
		{method: http.MethodPost, path: "/workflows/evals/runs/{run_id}/cancel"},
		{method: http.MethodGet, path: "/workflows/evals/results"},
		{method: http.MethodGet, path: "/workflows/evals/results/{result_id}"},
	}
}

func extractDirectCLIJSONRequestRoutes(source string) []cliRouteContract {
	callPattern := regexp.MustCompile(`cliJSONRequest\(\s*cmd,\s*http\.Method([A-Za-z]+),\s*([^,\n]+),`)
	seen := map[string]cliRouteContract{}
	for _, match := range callPattern.FindAllStringSubmatch(source, -1) {
		path := normalizeDirectCLIJSONRequestPath(match[2])
		// CLI sources now pass paths with the explicit "/v1" version
		// prefix (the SDK's default baseURL no longer includes it). All
		// other route contracts in this file omit "/v1" because
		// assertCLIOpenAPIRoute prepends it before lookup — strip here
		// so the format matches.
		path = strings.TrimPrefix(path, "/v1")
		route := cliRouteContract{
			method: strings.ToUpper(match[1]),
			path:   path,
		}
		if route.path == "" {
			continue
		}
		seen[route.method+" "+route.path] = route
	}

	var keys []string
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	routes := make([]cliRouteContract, 0, len(keys))
	for _, key := range keys {
		routes = append(routes, seen[key])
	}
	return routes
}

func normalizeDirectCLIJSONRequestPath(pathExpression string) string {
	tokenPattern := regexp.MustCompile(`"([^"]*)"|url\.PathEscape\([^)]*\)`)
	var builder strings.Builder
	for _, match := range tokenPattern.FindAllStringSubmatch(pathExpression, -1) {
		if match[1] != "" {
			builder.WriteString(match[1])
			continue
		}
		builder.WriteString("{}")
	}

	path := builder.String()
	for _, replacement := range []struct {
		old string
		new string
	}{
		{old: "/workflows/experiments/results/{}", new: "/workflows/experiments/results/{result_id}"},
		{old: "/workflows/experiments/runs/{}", new: "/workflows/experiments/runs/{run_id}"},
		{old: "/workflows/evals/results/{}", new: "/workflows/evals/results/{result_id}"},
		{old: "/workflows/evals/runs/{}", new: "/workflows/evals/runs/{run_id}"},
		{old: "/workflows/runs/{}", new: "/workflows/runs/{run_id}"},
	} {
		path = strings.Replace(path, replacement.old, replacement.new, 1)
	}
	if strings.Contains(path, "{}") {
		return ""
	}
	return path
}

func loadCLIOpenAPIContract(t *testing.T) cliOpenAPIContract {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "api-reference", "openapi.json"))
	if os.IsNotExist(err) {
		// The contract lives in the sibling public/docs submodule, which is
		// only checked out in the full monorepo. Standalone SDK checkouts (the
		// release worktree, a bare retab-dev/retab clone) have nothing to compare
		// against, so this monorepo-only parity check skips rather than fails.
		t.Skip("OpenAPI contract not present in this checkout; monorepo-only parity check")
	}
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	var contract cliOpenAPIContract
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatalf("decode OpenAPI contract: %v", err)
	}
	if len(contract.Paths) == 0 {
		t.Fatal("OpenAPI contract has no paths")
	}
	return contract
}

func assertCLIOpenAPIRoute(t *testing.T, contract cliOpenAPIContract, route cliRouteContract) {
	t.Helper()
	openAPIPath := "/v1" + route.path
	operations, ok := contract.Paths[openAPIPath]
	if !ok {
		t.Fatalf("OpenAPI contract is missing %s", openAPIPath)
	}
	if _, ok := operations[strings.ToLower(route.method)]; !ok {
		t.Fatalf("OpenAPI contract is missing %s %s", route.method, openAPIPath)
	}
}

func readCLISource(t *testing.T) string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read CLI source dir: %v", err)
	}
	var builder strings.Builder
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "zz_oagen_") {
			continue
		}
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		builder.Write(data)
		builder.WriteByte('\n')
	}
	return builder.String()
}

func extractCLIWorkflowRouteStrings(source string) []string {
	routeLiteral := regexp.MustCompile(`"(/workflows[^"]*)"`)
	seen := map[string]bool{}
	for _, match := range routeLiteral.FindAllStringSubmatch(source, -1) {
		path := normalizeCLIWorkflowRouteString(match[1])
		if path == "" {
			continue
		}
		seen[path] = true
	}

	var paths []string
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func normalizeCLIWorkflowRouteString(path string) string {
	if beforeQuery, _, ok := strings.Cut(path, "?"); ok {
		path = beforeQuery
	}
	if !strings.HasSuffix(path, "/") {
		return path
	}

	for _, replacement := range []struct {
		prefix      string
		placeholder string
	}{
		{prefix: "/workflows/runs/", placeholder: "{run_id}"},
		{prefix: "/workflows/steps/", placeholder: "{step_id}"},
		{prefix: "/workflows/reviews/", placeholder: "{review_id}"},
		{prefix: "/workflows/reviews/versions/", placeholder: "{version_id}"},
		{prefix: "/workflows/artifacts/", placeholder: "{artifact_id}"},
		{prefix: "/workflows/blocks/", placeholder: "{block_id}"},
		{prefix: "/workflows/edges/", placeholder: "{edge_id}"},
		{prefix: "/workflows/evals/", placeholder: "{eval_id}"},
		{prefix: "/workflows/evals/runs/", placeholder: "{run_id}"},
		{prefix: "/workflows/evals/results/", placeholder: "{result_id}"},
		{prefix: "/workflows/experiments/", placeholder: "{experiment_id}"},
		{prefix: "/workflows/experiments/runs/", placeholder: "{run_id}"},
		{prefix: "/workflows/experiments/results/", placeholder: "{result_id}"},
	} {
		if path == replacement.prefix {
			return path + replacement.placeholder
		}
	}
	return ""
}
