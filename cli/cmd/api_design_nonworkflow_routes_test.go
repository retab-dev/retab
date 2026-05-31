package cmd

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var approvedCLINonWorkflowNonReferenceRoutes = map[string]bool{
	// SDK conveniences used by CLI commands but not part of the generated
	// public API reference.
}

var approvedCLINonReferenceJobEndpoints = map[string]bool{
	// Legacy async job endpoints that the CLI still accepts for compatibility.
	"/v1/documents/classify":       true,
	"/v1/documents/extract":        true,
	"/v1/documents/parse":          true,
	"/v1/documents/split":          true,
	"/v1/edit/templates/generate":  true,
	"/v1/edits/templates/generate": true,
	"/v1/evals/extract/extract":    true,
	"/v1/evals/extract/process":    true,
}

func TestNonWorkflowCLIClientCallsHaveRouteContracts(t *testing.T) {
	source := readCLISource(t)
	calls := extractNonWorkflowCLIClientCalls(source)
	contracts := nonWorkflowCLIClientCallRouteContracts()

	for _, call := range calls {
		if _, ok := contracts[call]; !ok {
			t.Fatalf("CLI non-workflow client call %s has no HTTP route contract", call)
		}
	}

	callSet := map[string]bool{}
	for _, call := range calls {
		callSet[call] = true
	}
	for call := range contracts {
		if !callSet[call] {
			t.Fatalf("HTTP route contract for %s is stale; no CLI source call uses it", call)
		}
	}
}

func TestNonWorkflowCLIRoutesMatchOpenAPIOrExplicitApproval(t *testing.T) {
	openAPI := loadCLIOpenAPIContract(t)
	for _, route := range flattenNonWorkflowCLIRouteContracts(nonWorkflowCLIClientCallRouteContracts()) {
		key := cliHTTPRouteKey(route)
		if approvedCLINonWorkflowNonReferenceRoutes[key] {
			if cliOpenAPIHasRoute(openAPI, route) {
				t.Fatalf("%s is now in OpenAPI; remove it from approvedCLINonWorkflowNonReferenceRoutes", key)
			}
			continue
		}
		assertCLIOpenAPIRoute(t, openAPI, route)
	}
}

func TestNonWorkflowCLIJobEndpointsAreOpenAPIOrExplicitApproval(t *testing.T) {
	openAPI := loadCLIOpenAPIContract(t)
	for endpoint := range allowedJobEndpoints {
		if _, ok := openAPI.Paths[endpoint]; ok {
			continue
		}
		if approvedCLINonReferenceJobEndpoints[endpoint] {
			continue
		}
		t.Fatalf("CLI jobs allow non-reference endpoint %s without explicit approval", endpoint)
	}

	for endpoint := range approvedCLINonReferenceJobEndpoints {
		if !allowedJobEndpoints[endpoint] {
			t.Fatalf("approved non-reference job endpoint %s is stale; allowedJobEndpoints no longer contains it", endpoint)
		}
		if _, ok := openAPI.Paths[endpoint]; ok {
			t.Fatalf("approved non-reference job endpoint %s is now in OpenAPI; remove its approval", endpoint)
		}
	}
}

func TestNonWorkflowCLICanonicalTopLevelResourcesUseV1ResourceRoutes(t *testing.T) {
	openAPI := loadCLIOpenAPIContract(t)
	for _, endpoint := range []string{
		"/v1/classifications",
		"/v1/edits",
		"/v1/extractions",
		"/v1/parses",
		"/v1/partitions",
		"/v1/schemas/generate",
		"/v1/splits",
	} {
		if !allowedJobEndpoints[endpoint] {
			t.Fatalf("CLI jobs must allow canonical endpoint %s", endpoint)
		}
		operations, ok := openAPI.Paths[endpoint]
		if !ok {
			t.Fatalf("OpenAPI contract is missing canonical endpoint %s", endpoint)
		}
		if _, ok := operations[strings.ToLower(http.MethodPost)]; !ok {
			t.Fatalf("OpenAPI contract is missing POST %s", endpoint)
		}
	}

	for _, route := range flattenNonWorkflowCLIRouteContracts(nonWorkflowCLIClientCallRouteContracts()) {
		for _, removedPrefix := range []string{"/documents/", "/edit/", "/evals/"} {
			if strings.HasPrefix(route.path, removedPrefix) {
				t.Fatalf("CLI non-workflow route contract uses non-canonical path %s", route.path)
			}
		}
	}
}

func nonWorkflowCLIClientCallRouteContracts() map[string][]cliRouteContract {
	return map[string][]cliRouteContract{
		"Classifications.Create": {
			{method: http.MethodPost, path: "/classifications"},
		},
		"Classifications.Delete": {
			{method: http.MethodDelete, path: "/classifications/{classification_id}"},
		},
		"Classifications.Get": {
			{method: http.MethodGet, path: "/classifications/{classification_id}"},
		},
		"Classifications.List": {
			{method: http.MethodGet, path: "/classifications"},
		},
		"Edits.Create": {
			{method: http.MethodPost, path: "/edits"},
		},
		"Edits.Delete": {
			{method: http.MethodDelete, path: "/edits/{edit_id}"},
		},
		"Edits.Get": {
			{method: http.MethodGet, path: "/edits/{edit_id}"},
		},
		"Edits.List": {
			{method: http.MethodGet, path: "/edits"},
		},
		"Edits.Templates.Create": {
			{method: http.MethodPost, path: "/edits/templates"},
		},
		"Edits.Templates.Delete": {
			{method: http.MethodDelete, path: "/edits/templates/{template_id}"},
		},
		"Edits.Templates.Get": {
			{method: http.MethodGet, path: "/edits/templates/{template_id}"},
		},
		"Edits.Templates.List": {
			{method: http.MethodGet, path: "/edits/templates"},
		},
		"Edits.Templates.Update": {
			{method: http.MethodPatch, path: "/edits/templates/{template_id}"},
		},
		"Extractions.Create": {
			{method: http.MethodPost, path: "/extractions"},
		},
		"Extractions.Delete": {
			{method: http.MethodDelete, path: "/extractions/{extraction_id}"},
		},
		"Extractions.Get": {
			{method: http.MethodGet, path: "/extractions/{extraction_id}"},
		},
		"Extractions.List": {
			{method: http.MethodGet, path: "/extractions"},
		},
		"Extractions.Sources": {
			{method: http.MethodGet, path: "/extractions/{extraction_id}/sources"},
		},
		"Files.CompleteUpload": {
			{method: http.MethodPost, path: "/files/upload/{file_id}/complete"},
		},
		"Files.CreateUpload": {
			{method: http.MethodPost, path: "/files/upload"},
		},
		"Files.Get": {
			{method: http.MethodGet, path: "/files/{file_id}"},
		},
		"Files.GetDownloadLink": {
			{method: http.MethodGet, path: "/files/{file_id}/download-link"},
		},
		"Files.List": {
			{method: http.MethodGet, path: "/files"},
		},
		"Jobs.Cancel": {
			{method: http.MethodPost, path: "/jobs/{job_id}/cancel"},
		},
		"Jobs.Create": {
			{method: http.MethodPost, path: "/jobs"},
		},
		"Jobs.List": {
			{method: http.MethodGet, path: "/jobs"},
		},
		"Jobs.Retry": {
			{method: http.MethodPost, path: "/jobs/{job_id}/retry"},
		},
		"Jobs.Get": {
			{method: http.MethodGet, path: "/jobs/{job_id}"},
		},
		"Parses.Create": {
			{method: http.MethodPost, path: "/parses"},
		},
		"Parses.Delete": {
			{method: http.MethodDelete, path: "/parses/{parse_id}"},
		},
		"Parses.Get": {
			{method: http.MethodGet, path: "/parses/{parse_id}"},
		},
		"Parses.List": {
			{method: http.MethodGet, path: "/parses"},
		},
		"Partitions.Create": {
			{method: http.MethodPost, path: "/partitions"},
		},
		"Partitions.Delete": {
			{method: http.MethodDelete, path: "/partitions/{partition_id}"},
		},
		"Partitions.Get": {
			{method: http.MethodGet, path: "/partitions/{partition_id}"},
		},
		"Partitions.List": {
			{method: http.MethodGet, path: "/partitions"},
		},
		"Schemas.Generate": {
			{method: http.MethodPost, path: "/schemas/generate"},
		},
		"Splits.Create": {
			{method: http.MethodPost, path: "/splits"},
		},
		"Splits.Delete": {
			{method: http.MethodDelete, path: "/splits/{split_id}"},
		},
		"Splits.Get": {
			{method: http.MethodGet, path: "/splits/{split_id}"},
		},
		"Splits.List": {
			{method: http.MethodGet, path: "/splits"},
		},
	}
}

func extractNonWorkflowCLIClientCalls(source string) []string {
	callPattern := regexp.MustCompile(`client\.(Classifications|Edits|EditTemplates|Extractions|Files|Jobs|Parses|Partitions|Schemas|Splits)((?:\.[A-Za-z]+)+)\(`)
	seen := map[string]bool{}
	for _, match := range callPattern.FindAllStringSubmatch(source, -1) {
		seen[match[1]+match[2]] = true
	}

	var calls []string
	for call := range seen {
		calls = append(calls, call)
	}
	sort.Strings(calls)
	return calls
}

func flattenNonWorkflowCLIRouteContracts(contracts map[string][]cliRouteContract) []cliRouteContract {
	seen := map[string]cliRouteContract{}
	for _, routes := range contracts {
		for _, route := range routes {
			seen[cliHTTPRouteKey(route)] = route
		}
	}

	var keys []string
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	flattened := make([]cliRouteContract, 0, len(keys))
	for _, key := range keys {
		flattened = append(flattened, seen[key])
	}
	return flattened
}

func cliHTTPRouteKey(route cliRouteContract) string {
	return route.method + " /v1" + route.path
}

func cliOpenAPIHasRoute(contract cliOpenAPIContract, route cliRouteContract) bool {
	operations, ok := contract.Paths["/v1"+route.path]
	if !ok {
		return false
	}
	_, ok = operations[strings.ToLower(route.method)]
	return ok
}
