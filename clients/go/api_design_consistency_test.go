package retab

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type openAPIContract struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

type apiRouteContract struct {
	method string
	path   string
}

var approvedSDKNonReferenceRoutes = map[string]bool{
	// Existing SDK conveniences that are not part of the generated public API
	// reference. Keep explicit so new non-reference routes cannot appear silently.
	"/v1/edits/templates/fill": true,
	"/v1/extractions/stream":   true,
}

func TestWorkflowSDKDesignContractMatchesOpenAPI(t *testing.T) {
	openAPI := loadSDKOpenAPIContract(t)
	for _, route := range workflowSDKRouteContract() {
		assertSDKOpenAPIRoute(t, openAPI, route)
	}

	for _, path := range extractSDKWorkflowRouteStrings(readSDKSource(t)) {
		openAPIPath := "/v1" + path
		if _, ok := openAPI.Paths[openAPIPath]; !ok {
			t.Fatalf("Go SDK route string %s is missing from OpenAPI contract", path)
		}
	}
}

func TestSDKRouteStringsAreOpenAPIOrExplicitException(t *testing.T) {
	openAPI := loadSDKOpenAPIContract(t)
	for _, path := range extractSDKPublicRouteStrings(readSDKSource(t)) {
		openAPIPath := "/v1" + path
		if _, ok := openAPI.Paths[openAPIPath]; ok {
			continue
		}
		if approvedSDKNonReferenceRoutes[openAPIPath] {
			continue
		}
		t.Fatalf("Go SDK route string %s is missing from OpenAPI contract", path)
	}
}

func TestWorkflowSDKUsesOnlyCanonicalReviewBlockExecutionAndTestSurface(t *testing.T) {
	source := readSDKSource(t)

	for _, required := range []string{
		`"/workflows/reviews"`,
		`"/workflows/reviews/versions"`,
		`"/workflows/blocks/executions"`,
		`"/workflows/tests"`,
		`"/workflows/tests/runs"`,
		`"/workflows/tests/results"`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("Go SDK source is missing canonical route string %s", required)
		}
	}

	nestedWorkflowRoute := regexp.MustCompile(`"/workflows/"\s*\+\s*url\.PathEscape\([^)]*workflowID[^)]*\)\s*\+\s*"/(?:reviews|block executions|tests)(?:/|")`)
	if match := nestedWorkflowRoute.FindString(source); match != "" {
		t.Fatalf("Go SDK must use flat workflow routes for reviews/blocks/executions/tests, found nested route expression: %s", match)
	}

	for _, removed := range []string{"append_version", "append-version", "appendVersion", "AppendVersion"} {
		if strings.Contains(source, removed) {
			t.Fatalf("Go SDK source must not expose removed append-version compatibility surface %q", removed)
		}
	}

	for _, serviceType := range []reflect.Type{
		reflect.TypeOf(&WorkflowReviewsService{}),
		reflect.TypeOf(&WorkflowReviewVersionsService{}),
		reflect.TypeOf(&WorkflowBlockExecutionsService{}),
		reflect.TypeOf(&WorkflowTestsService{}),
		reflect.TypeOf(&WorkflowTestRunsService{}),
		reflect.TypeOf(&WorkflowTestRunResultsService{}),
	} {
		for i := 0; i < serviceType.NumMethod(); i++ {
			method := serviceType.Method(i)
			if strings.Contains(method.Name, "Append") {
				t.Fatalf("%s exposes removed append-version method %s", serviceType, method.Name)
			}
		}
	}
}

func workflowSDKRouteContract() []apiRouteContract {
	return []apiRouteContract{
		{method: http.MethodGet, path: "/workflows/reviews"},
		{method: http.MethodGet, path: "/workflows/reviews/{review_id}"},
		{method: http.MethodPost, path: "/workflows/reviews/{review_id}/approve"},
		{method: http.MethodPost, path: "/workflows/reviews/{review_id}/reject"},
		{method: http.MethodGet, path: "/workflows/reviews/versions"},
		{method: http.MethodPost, path: "/workflows/reviews/versions"},
		{method: http.MethodGet, path: "/workflows/reviews/versions/{version_id}"},
		{method: http.MethodPost, path: "/workflows/blocks/executions"},
		{method: http.MethodGet, path: "/workflows/blocks/executions"},
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

func loadSDKOpenAPIContract(t *testing.T) openAPIContract {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "api-reference", "openapi.json"))
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	var contract openAPIContract
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatalf("decode OpenAPI contract: %v", err)
	}
	if len(contract.Paths) == 0 {
		t.Fatal("OpenAPI contract has no paths")
	}
	return contract
}

func assertSDKOpenAPIRoute(t *testing.T, contract openAPIContract, route apiRouteContract) {
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

func readSDKSource(t *testing.T) string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read SDK source dir: %v", err)
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

func extractSDKWorkflowRouteStrings(source string) []string {
	routeLiteral := regexp.MustCompile(`"(/workflows[^"]*)"`)
	seen := map[string]bool{}
	for _, match := range routeLiteral.FindAllStringSubmatch(source, -1) {
		path := normalizeSDKWorkflowRouteString(match[1])
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

func extractSDKPublicRouteStrings(source string) []string {
	routeLiteral := regexp.MustCompile(`"(/(?:classifications|edits|extractions|files|jobs|parses|partitions|schemas|splits|workflows)[^"]*)"`)
	seen := map[string]bool{}
	for _, match := range routeLiteral.FindAllStringSubmatch(source, -1) {
		path := normalizeSDKPublicRouteString(match[1])
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

func normalizeSDKPublicRouteString(path string) string {
	if beforeQuery, _, ok := strings.Cut(path, "?"); ok {
		path = beforeQuery
	}
	if !strings.HasSuffix(path, "/") {
		return strings.TrimSuffix(path, "/")
	}

	for _, replacement := range []struct {
		prefix      string
		placeholder string
	}{
		{prefix: "/classifications/", placeholder: "{classification_id}"},
		{prefix: "/edits/templates/", placeholder: "{template_id}"},
		{prefix: "/edits/", placeholder: "{edit_id}"},
		{prefix: "/extractions/", placeholder: "{extraction_id}"},
		{prefix: "/files/upload/", placeholder: "{file_id}/complete"},
		{prefix: "/files/", placeholder: "{file_id}"},
		{prefix: "/jobs/", placeholder: "{job_id}"},
		{prefix: "/parses/", placeholder: "{parse_id}"},
		{prefix: "/partitions/", placeholder: "{partition_id}"},
		{prefix: "/splits/", placeholder: "{split_id}"},
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
	return strings.TrimSuffix(path, "/")
}

func normalizeSDKWorkflowRouteString(path string) string {
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
