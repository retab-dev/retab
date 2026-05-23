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
	"/v1/extractions/stream": true,
}

func TestWorkflowSDKDesignContractMatchesOpenAPI(t *testing.T) {
	openAPI := loadSDKOpenAPIContract(t)
	for _, route := range workflowSDKRouteContract() {
		assertSDKOpenAPIRoute(t, openAPI, route)
	}

	for _, path := range extractSDKWorkflowRouteStrings(readSDKSource(t)) {
		openAPIPath := path
		if _, ok := openAPI.Paths[openAPIPath]; !ok {
			t.Fatalf("Go SDK route string %s is missing from OpenAPI contract", path)
		}
	}
}

func TestSDKRouteStringsAreOpenAPIOrExplicitException(t *testing.T) {
	openAPI := loadSDKOpenAPIContract(t)
	for _, path := range extractSDKPublicRouteStrings(readSDKSource(t)) {
		openAPIPath := path
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
		`"/v1/workflows/reviews"`,
		`"/v1/workflows/reviews/versions"`,
		`"/v1/workflows/blocks/executions"`,
		`"/v1/workflows/tests"`,
		`"/v1/workflows/tests/runs"`,
		`"/v1/workflows/tests/results"`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("Go SDK source is missing canonical route string %s", required)
		}
	}

	nestedWorkflowRoute := regexp.MustCompile(`"/v1/workflows/"\s*\+\s*url\.PathEscape\([^)]*workflowID[^)]*\)\s*\+\s*"/(?:reviews|block executions|tests)(?:/|")`)
	if match := nestedWorkflowRoute.FindString(source); match != "" {
		t.Fatalf("Go SDK must use flat workflow routes for reviews/blocks/executions/tests, found nested route expression: %s", match)
	}

	for _, removed := range []string{"append_version", "append-version", "appendVersion", "AppendVersion"} {
		if strings.Contains(source, removed) {
			t.Fatalf("Go SDK source must not expose removed append-version compatibility surface %q", removed)
		}
	}

	for _, serviceType := range []reflect.Type{
		reflect.TypeOf(&WorkflowReviewService{}),
		reflect.TypeOf(&WorkflowReviewVersionService{}),
		reflect.TypeOf(&WorkflowBlockExecutionService{}),
		reflect.TypeOf(&WorkflowTestService{}),
		reflect.TypeOf(&WorkflowTestRunService{}),
		reflect.TypeOf(&WorkflowTestRunResultService{}),
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
		{method: http.MethodGet, path: "/v1/workflows/reviews"},
		{method: http.MethodGet, path: "/v1/workflows/reviews/{review_id}"},
		{method: http.MethodPost, path: "/v1/workflows/reviews/{review_id}/approve"},
		{method: http.MethodPost, path: "/v1/workflows/reviews/{review_id}/reject"},
		{method: http.MethodGet, path: "/v1/workflows/reviews/versions"},
		{method: http.MethodPost, path: "/v1/workflows/reviews/versions"},
		{method: http.MethodGet, path: "/v1/workflows/reviews/versions/{version_id}"},
		{method: http.MethodPost, path: "/v1/workflows/blocks/executions"},
		{method: http.MethodGet, path: "/v1/workflows/blocks/executions"},
		{method: http.MethodPost, path: "/v1/workflows/tests"},
		{method: http.MethodGet, path: "/v1/workflows/tests"},
		{method: http.MethodGet, path: "/v1/workflows/tests/{test_id}"},
		{method: http.MethodPatch, path: "/v1/workflows/tests/{test_id}"},
		{method: http.MethodDelete, path: "/v1/workflows/tests/{test_id}"},
		{method: http.MethodPost, path: "/v1/workflows/tests/runs"},
		{method: http.MethodGet, path: "/v1/workflows/tests/runs"},
		{method: http.MethodGet, path: "/v1/workflows/tests/runs/{run_id}"},
		{method: http.MethodPost, path: "/v1/workflows/tests/runs/{run_id}/cancel"},
		{method: http.MethodGet, path: "/v1/workflows/tests/results"},
		{method: http.MethodGet, path: "/v1/workflows/tests/results/{result_id}"},
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
	openAPIPath := route.path
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
		{prefix: "/v1/classifications/", placeholder: "{classification_id}"},
		{prefix: "/v1/edits/templates/", placeholder: "{template_id}"},
		{prefix: "/v1/edits/", placeholder: "{edit_id}"},
		{prefix: "/v1/extractions/", placeholder: "{extraction_id}"},
		{prefix: "/v1/files/upload/", placeholder: "{file_id}/complete"},
		{prefix: "/v1/files/", placeholder: "{file_id}"},
		{prefix: "/v1/jobs/", placeholder: "{job_id}"},
		{prefix: "/v1/parses/", placeholder: "{parse_id}"},
		{prefix: "/v1/partitions/", placeholder: "{partition_id}"},
		{prefix: "/v1/splits/", placeholder: "{split_id}"},
		{prefix: "/v1/workflows/runs/", placeholder: "{run_id}"},
		{prefix: "/v1/workflows/steps/", placeholder: "{step_id}"},
		{prefix: "/v1/workflows/reviews/", placeholder: "{review_id}"},
		{prefix: "/v1/workflows/reviews/versions/", placeholder: "{version_id}"},
		{prefix: "/v1/workflows/artifacts/", placeholder: "{artifact_id}"},
		{prefix: "/v1/workflows/blocks/", placeholder: "{block_id}"},
		{prefix: "/v1/workflows/edges/", placeholder: "{edge_id}"},
		{prefix: "/v1/workflows/tests/", placeholder: "{test_id}"},
		{prefix: "/v1/workflows/tests/runs/", placeholder: "{run_id}"},
		{prefix: "/v1/workflows/tests/results/", placeholder: "{result_id}"},
		{prefix: "/v1/workflows/experiments/", placeholder: "{experiment_id}"},
		{prefix: "/v1/workflows/experiments/runs/", placeholder: "{run_id}"},
		{prefix: "/v1/workflows/experiments/results/", placeholder: "{result_id}"},
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
		{prefix: "/v1/workflows/runs/", placeholder: "{run_id}"},
		{prefix: "/v1/workflows/steps/", placeholder: "{step_id}"},
		{prefix: "/v1/workflows/reviews/", placeholder: "{review_id}"},
		{prefix: "/v1/workflows/reviews/versions/", placeholder: "{version_id}"},
		{prefix: "/v1/workflows/artifacts/", placeholder: "{artifact_id}"},
		{prefix: "/v1/workflows/blocks/", placeholder: "{block_id}"},
		{prefix: "/v1/workflows/edges/", placeholder: "{edge_id}"},
		{prefix: "/v1/workflows/tests/", placeholder: "{test_id}"},
		{prefix: "/v1/workflows/tests/runs/", placeholder: "{run_id}"},
		{prefix: "/v1/workflows/tests/results/", placeholder: "{result_id}"},
		{prefix: "/v1/workflows/experiments/", placeholder: "{experiment_id}"},
		{prefix: "/v1/workflows/experiments/runs/", placeholder: "{run_id}"},
		{prefix: "/v1/workflows/experiments/results/", placeholder: "{result_id}"},
	} {
		if path == replacement.prefix {
			return path + replacement.placeholder
		}
	}
	return ""
}
