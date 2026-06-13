package cmd

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"unicode"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

type sdkOperation struct {
	resourcePath string
	methodName   string
	commandPath  string
}

var sdkMethodCommandAliases = map[string]string{
	"CreateBlueprint":       "blueprints create",
	"CreateBlueprintCancel": "blueprints cancel",
	"CreateStream":          "stream",
	"CreateUpload":          "create-upload",
	"CompleteUpload":        "complete-upload",
	"GetBlueprint":          "blueprints get",
	"GetDownloadLink":       "download-link",
	"WaitForCompletion":     "wait",
}

var sdkResourceCommandAliases = map[string]string{
	"EditTemplates":           "edits templates",
	"ExperimentRunMetrics":    "workflows experiments metrics",
	"ExperimentRunResults":    "workflows experiments results",
	"ExperimentRuns":          "workflows experiments runs",
	"Specs":                   "spec",
	"WorkflowArtifacts":       "workflows artifacts",
	"WorkflowBlockExecutions": "workflows blocks executions",
	"WorkflowBlocks":          "workflows blocks",
	"WorkflowEdges":           "workflows edges",
	"WorkflowExperiments":     "workflows experiments",
	"WorkflowReviewVersions":  "workflows reviews versions",
	"WorkflowReviews":         "workflows reviews",
	"WorkflowRuns":            "workflows runs",
	"WorkflowSpec":            "workflows spec",
	"WorkflowSpecs":           "workflows spec",
	"WorkflowSteps":           "workflows steps",
	"WorkflowTestRunResults":  "workflows tests results",
	"WorkflowTestRuns":        "workflows tests runs",
	"WorkflowTests":           "workflows tests",
}

var sdkOperationCommandAliases = map[string]string{
	// Secrets use one CLI-safe command that reads values from prompt/stdin/file
	// instead of exposing raw secret values as shell-history-friendly flags.
	"secrets.Create":    "secrets set",
	"secrets.ListValue": "secrets value",
	"secrets.Update":    "secrets set",
	// The table PATCH route currently edits table metadata/name only. The
	// public CLI intentionally exposes CSV lifecycle commands, not metadata-only
	// table mutation, so `retab tables update` remains absent.
	"tables.Update": "",
	// Server-side block config dry-run validation is exposed as the remote mode
	// of the local bundle validator rather than the generated block-validate-config
	// name.
	"workflows.blocks.CreateBlockValidateConfig": "workflows blocks config validate",
	// `plan`/`apply` are single SDK methods on the Workflows service that take an
	// optional workflow_id (route-merge: id-less vs existing-workflow route). The
	// CLI keeps them under the `spec` namespace; the `-to` variants pass the id
	// and are declared CLI-only below.
	"workflows.Plan":  "workflows spec plan",
	"workflows.Apply": "workflows spec apply",
	// Versioning is exposed as a `versions` subgroup on each versioned resource
	// (workflows, blocks, edges) rather than as flat verbs, so the SDK's
	// List/Get/Diff/Restore version methods map onto the grouped leaves.
	"workflows.ListVersions":                "workflows versions list",
	"workflows.GetVersion":                  "workflows versions get",
	"workflows.ListDiff":                    "workflows versions diff",
	"workflows.CreateVersionRestore":        "workflows versions restore",
	"workflows.blocks.ListVersions":         "workflows blocks versions list",
	"workflows.blocks.GetVersion":           "workflows blocks versions get",
	"workflows.blocks.ListDiff":             "workflows blocks versions diff",
	"workflows.blocks.CreateVersionRestore": "workflows blocks versions restore",
	"workflows.edges.ListVersions":          "workflows edges versions list",
	"workflows.edges.GetVersion":            "workflows edges versions get",
	"workflows.edges.ListDiff":              "workflows edges versions diff",
	"workflows.edges.CreateVersionRestore":  "workflows edges versions restore",
}

var workflowCLIOnlyCommands = map[string]string{
	"workflows blocks config pull":       "local block-config bundle export composed from workflows blocks get",
	"workflows blocks config push":       "local block-config bundle import composed from workflows blocks get/update",
	"workflows blocks config diff":       "local block-config bundle diff composed from workflows blocks get",
	"workflows blocks config validate":   "local block-config bundle validation with backend dry-run",
	"workflows blocks config doctor":     "local block-config bundle diagnostics",
	"workflows blocks api-calls hydrate": "local api_call runtime scaffolding for pulled api_call bundles",
	"workflows blocks api-calls render":  "local api_call request rendering for pulled api_call bundles",
	"workflows blocks api-calls run":     "local api_call dry-run renderer and opt-in executor",
	"workflows blocks functions hydrate": "local function runtime scaffolding for pulled function bundles",
	"workflows blocks functions run":     "local function runner for hydrated function bundles",
	"workflows spec plan-to":             "workflows.Plan with an explicit workflow_id (existing-workflow plan route)",
	"workflows spec apply-to":            "workflows.Apply with an explicit workflow_id (existing-workflow apply route)",
	"workflows reviews schema":           "local schema helper composed from reviews get",
	"workflows runs restart":             "local restart alias composed from runs create",
	"workflows experiments runs wait":    "local poll loop composed from experiments runs get",
	"workflows runs wait":                "local poll loop composed from runs get",
	"workflows tests runs wait":          "local poll loop composed from tests runs get",
	"workflows view":                     "terminal graph renderer composed from workflow graph reads",
	"workflows access list":              "internal /v1/workflow-memberships read (dashboard-only, not in public SDK)",
	"workflows access get":               "internal /v1/workflow-memberships read (dashboard-only, not in public SDK)",
	"workflows access update":            "internal /v1/workflow-memberships role change (dashboard-only, not in public SDK)",
	"workflows access revoke":            "internal /v1/workflow-memberships deactivate (dashboard-only, not in public SDK)",
}

func TestCLIExposesGoSDKOperationSurface(t *testing.T) {
	client, err := retab.NewClient("test-key")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	expected := collectSDKOperations(reflect.TypeOf(client).Elem())
	available := collectLeafCommandPaths(rootCmd)

	if len(expected) == 0 {
		t.Fatal("expected at least one SDK operation")
	}
	if len(available) == 0 {
		t.Fatal("expected at least one CLI command")
	}

	var missing []string
	for _, operation := range expected {
		if _, ok := available[operation.commandPath]; !ok {
			missing = append(missing, operation.resourcePath+"."+toSnakeName(operation.methodName)+" -> retab "+operation.commandPath)
		}
	}

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("CLI is missing commands for SDK operations:\n%s", strings.Join(missing, "\n"))
	}
}

func TestWorkflowCLICommandsDoNotDriftFromGoSDK(t *testing.T) {
	client, err := retab.NewClient("test-key")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	expected := map[string]bool{}
	for _, operation := range collectSDKOperations(reflect.TypeOf(client).Elem()) {
		if strings.HasPrefix(operation.commandPath, "workflows ") {
			expected[operation.commandPath] = true
		}
	}

	available := collectLeafCommandPaths(rootCmd)
	var stale []string
	for commandPath := range available {
		if !strings.HasPrefix(commandPath, "workflows ") {
			continue
		}
		if expected[commandPath] || workflowCLIOnlyCommands[commandPath] != "" {
			continue
		}
		stale = append(stale, "retab "+commandPath)
	}

	sort.Strings(stale)
	if len(stale) > 0 {
		t.Fatalf("workflow CLI exposes commands not present in Go SDK:\n%s", strings.Join(stale, "\n"))
	}
}

func collectSDKOperations(clientType reflect.Type) []sdkOperation {
	var operations []sdkOperation
	seen := map[string]bool{}

	for i := 0; i < clientType.NumField(); i++ {
		field := clientType.Field(i)
		if !field.IsExported() || !isSDKServiceField(field) {
			continue
		}
		resourcePath := []string{resourceCommandName(field.Name)}
		operations = append(operations, collectSDKServiceOperations(field.Type, resourcePath, seen)...)
	}

	sort.Slice(operations, func(i, j int) bool {
		return operations[i].commandPath < operations[j].commandPath
	})
	return operations
}

func collectSDKServiceOperations(serviceType reflect.Type, resourcePath []string, seen map[string]bool) []sdkOperation {
	if serviceType.Kind() != reflect.Pointer {
		return nil
	}
	elemType := serviceType.Elem()
	visitKey := elemType.PkgPath() + "." + elemType.Name() + ":" + strings.Join(resourcePath, ".")
	if seen[visitKey] {
		return nil
	}
	seen[visitKey] = true

	var operations []sdkOperation
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		if !method.IsExported() || strings.HasPrefix(method.Name, "Prepare") {
			continue
		}
		resourceKey := strings.Join(resourcePath, ".")
		if alias, ok := sdkOperationCommandAliases[resourceKey+"."+method.Name]; ok {
			if alias == "" {
				continue
			}
			operations = append(operations, sdkOperation{
				resourcePath: resourceKey,
				methodName:   method.Name,
				commandPath:  alias,
			})
			continue
		}
		commandName, include := sdkMethodCommandName(method.Name)
		if !include {
			continue
		}
		commandPath := append(append([]string{}, resourcePath...), commandName)
		operations = append(operations, sdkOperation{
			resourcePath: resourceKey,
			methodName:   method.Name,
			commandPath:  strings.Join(commandPath, " "),
		})
	}

	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if !field.IsExported() || !isSDKServiceField(field) {
			continue
		}
		childPath := append(append([]string{}, resourcePath...), resourceCommandName(field.Name))
		operations = append(operations, collectSDKServiceOperations(field.Type, childPath, seen)...)
	}

	return operations
}

func isSDKServiceField(field reflect.StructField) bool {
	return field.Type.Kind() == reflect.Pointer &&
		field.Type.Elem().PkgPath() == "github.com/retab-dev/retab/clients/go" &&
		strings.HasSuffix(field.Type.Elem().Name(), "Service")
}

func sdkMethodCommandName(methodName string) (string, bool) {
	if alias, ok := sdkMethodCommandAliases[methodName]; ok {
		return alias, alias != ""
	}
	return methodCommandName(methodName), true
}

func methodCommandName(methodName string) string {
	return kebabName(stripVerbPrefix(methodName))
}

func stripVerbPrefix(methodName string) string {
	for _, prefix := range []string{"Create", "Get", "List"} {
		if methodName == prefix {
			return methodName
		}
		if strings.HasPrefix(methodName, prefix) && len(methodName) > len(prefix) {
			return methodName[len(prefix):]
		}
	}
	return methodName
}

func resourceCommandName(fieldName string) string {
	if alias, ok := sdkResourceCommandAliases[fieldName]; ok {
		return alias
	}
	return kebabName(fieldName)
}

func toSnakeName(value string) string {
	return strings.ReplaceAll(kebabName(value), "-", "_")
}

func kebabName(value string) string {
	normalized := strings.NewReplacer(
		"API", "Api",
		"CSV", "Csv",
		"REVIEW", "Review",
		"HTTP", "Http",
		"ID", "Id",
		"JSON", "Json",
		"MIME", "Mime",
		"URL", "Url",
		"YAML", "Yaml",
	).Replace(value)

	var out strings.Builder
	for index, r := range normalized {
		if r == '_' {
			out.WriteByte('-')
			continue
		}
		if unicode.IsUpper(r) {
			if index > 0 {
				out.WriteByte('-')
			}
			out.WriteRune(unicode.ToLower(r))
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func collectLeafCommandPaths(root *cobra.Command) map[string]bool {
	paths := map[string]bool{}
	collectLeafCommandPathsInto(root, nil, paths)
	return paths
}

func collectLeafCommandPathsInto(cmd *cobra.Command, prefix []string, paths map[string]bool) {
	name := cmd.Name()
	nextPrefix := prefix
	if name != "" && name != "retab" {
		nextPrefix = append(append([]string{}, prefix...), name)
	}

	children := cmd.Commands()
	if len(children) == 0 && len(nextPrefix) > 0 {
		paths[strings.Join(nextPrefix, " ")] = true
		return
	}
	for _, child := range children {
		collectLeafCommandPathsInto(child, nextPrefix, paths)
	}
}
