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
	"ConfigHistory":      "",
	"CreateBatch":        "create-batch",
	"CreateVersion":      "versions create",
	"CreateStream":       "stream",
	"CreateUpload":       "create-upload",
	"CompleteUpload":     "complete-upload",
	"DiagnoseGraph":      "diagnose",
	"GetConfig":          "config",
	"GetDocumentURL":     "document-url",
	"GetDownloadLink":    "download-link",
	"GetEntities":        "entities",
	"GetRef":             "",
	"GetResolvedSchemas": "resolved-schemas",
	"ListEligibleBlocks": "eligible-blocks",
	"ListSimulations":    "",
	"ListSnapshots":      "snapshots",
	"WaitFor":            "wait",
	"WaitForCompletion":  "wait",
}

var sdkResourceCommandAliases = map[string]string{
	"Specs": "spec",
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
		commandName, include := sdkMethodCommandName(method.Name)
		if !include {
			continue
		}
		commandPath := append(append([]string{}, resourcePath...), commandName)
		operations = append(operations, sdkOperation{
			resourcePath: strings.Join(resourcePath, "."),
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
