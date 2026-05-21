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

func TestWorkflowCLIDesignContractMatchesOpenAPI(t *testing.T) {
	openAPI := loadCLIOpenAPIContract(t)
	for _, route := range workflowCLIRouteContract() {
		assertCLIOpenAPIRoute(t, openAPI, route)
	}

	for _, path := range extractCLIWorkflowRouteStrings(readCLISource(t)) {
		openAPIPath := "/v1" + path
		if _, ok := openAPI.Paths[openAPIPath]; !ok {
			t.Fatalf("CLI route string %s is missing from OpenAPI contract", path)
		}
	}
}

func TestWorkflowCLIUsesOnlyCanonicalReviewSimulationAndTestSurface(t *testing.T) {
	commands := collectLeafCommandPaths(rootCmd)

	for _, commandPath := range []string{
		"workflows reviews list",
		"workflows reviews get",
		"workflows reviews approve",
		"workflows reviews reject",
		"workflows reviews versions list",
		"workflows reviews versions get",
		"workflows reviews versions create",
		"workflows simulations create",
		"workflows simulations list",
		"workflows tests create",
		"workflows tests list",
		"workflows tests get",
		"workflows tests update",
		"workflows tests delete",
		"workflows tests runs create",
		"workflows tests runs list",
		"workflows tests runs get",
		"workflows tests runs cancel",
		"workflows tests runs results list",
		"workflows tests runs results get",
	} {
		if !commands[commandPath] {
			t.Fatalf("CLI is missing canonical command retab %s", commandPath)
		}
	}

	for _, removedCommandPath := range []string{
		"workflows reviews append",
		"workflows reviews edit",
		"workflows reviews versions append",
		"workflows simulations get",
	} {
		if commands[removedCommandPath] {
			t.Fatalf("CLI still exposes removed command retab %s", removedCommandPath)
		}
	}

	source := readCLISource(t)
	for _, required := range []string{
		`"/workflows/tests/runs"`,
		`"/workflows/tests/results"`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("CLI source is missing canonical route string %s", required)
		}
	}

	nestedWorkflowRoute := regexp.MustCompile(`"/workflows/"\s*\+\s*(?:url\.)?PathEscape\([^)]*workflowID[^)]*\)\s*\+\s*"/(?:reviews|simulations|tests)(?:/|")`)
	if match := nestedWorkflowRoute.FindString(source); match != "" {
		t.Fatalf("CLI must use flat workflow routes for reviews/simulations/tests, found nested route expression: %s", match)
	}

	for _, removed := range []string{"append_version", "append-version", "appendVersion", "AppendVersion"} {
		if strings.Contains(source, removed) {
			t.Fatalf("CLI source must not expose removed append-version compatibility surface %q", removed)
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
		{method: http.MethodPost, path: "/workflows/simulations"},
		{method: http.MethodGet, path: "/workflows/simulations"},
		{method: http.MethodPost, path: "/workflows/tests"},
		{method: http.MethodGet, path: "/workflows/tests"},
		{method: http.MethodGet, path: "/workflows/tests/{test_id}"},
		{method: http.MethodPatch, path: "/workflows/tests/{test_id}"},
		{method: http.MethodDelete, path: "/workflows/tests/{test_id}"},
		{method: http.MethodPost, path: "/workflows/tests/runs"},
		{method: http.MethodGet, path: "/workflows/tests/runs"},
		{method: http.MethodGet, path: "/workflows/tests/runs/{run_id}"},
		{method: http.MethodPost, path: "/workflows/tests/runs/{run_id}/cancel"},
		{method: http.MethodGet, path: "/workflows/tests/results"},
		{method: http.MethodGet, path: "/workflows/tests/results/{result_id}"},
	}
}

func loadCLIOpenAPIContract(t *testing.T) cliOpenAPIContract {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "api-reference", "openapi.json"))
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
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
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
		{prefix: "/workflows/tests/", placeholder: "{test_id}"},
		{prefix: "/workflows/tests/runs/", placeholder: "{run_id}"},
		{prefix: "/workflows/tests/results/", placeholder: "{result_id}"},
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
