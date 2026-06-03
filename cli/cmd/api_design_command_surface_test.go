package cmd

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var apiDesignKebabNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

func TestAPICommandSurfaceUsesCanonicalResourceActionNames(t *testing.T) {
	canonicalNames := map[string]bool{
		"add": true, "api-calls": true, "apply": true, "approve": true, "artifacts": true, "auth": true,
		"blocks": true, "blueprints": true, "cancel": true, "claim": true, "classifications": true, "complete-upload": true, "env": true,
		"create": true, "create-upload": true, "delete": true, "discard-draft": true,
		"diff-config": true, "doctor-config": true, "download": true, "download-link": true, "edges": true, "edits": true,
		"experiments": true, "export": true, "extractions": true, "files": true,
		"functions": true, "generate": true, "get": true, "grep": true, "hydrate": true, "inspect": true,
		"list": true, "login": true, "logout": true, "metrics": true,
		"parse": true, "parses": true, "partitions": true, "plan": true, "profile": true, "publish": true, "pull-config": true,
		"push-config": true, "query": true, "reject": true, "remove": true, "render": true, "replace": true, "restart": true, "results": true, "retry": true, "retrieve": true,
		"reviews": true, "run": true, "runs": true, "schema": true, "schemas": true, "secrets": true,
		"set": true, "setup": true, "executions": true, "sources": true, "spec": true,
		"splits": true, "status": true, "steps": true, "stream": true,
		"switch": true, "sync": true, "tables": true, "templates": true, "tests": true, "update": true,
		"upload": true, "validate": true, "validate-config": true, "value": true, "versions": true, "version": true,
		"view": true, "wait": true, "which": true, "workflows": true,
	}

	apiDesignWalkCommands(rootCmd, func(cmd *cobra.Command) {
		if cmd == rootCmd {
			return
		}
		name := cmd.Name()
		if !apiDesignKebabNamePattern.MatchString(name) {
			t.Errorf("command %q uses non-kebab-case name %q", cmd.CommandPath(), name)
		}
		if !canonicalNames[name] {
			t.Errorf("command %q uses non-canonical API resource/action name %q", cmd.CommandPath(), name)
		}
	})
}

func TestAPICommandSurfaceFlagsUseCanonicalKebabCase(t *testing.T) {
	removedFlags := map[string]string{
		"append-version": "removed review-version compatibility flag",
		"append_version": "removed snake_case review-version compatibility flag",
		"appendVersion":  "removed camelCase review-version compatibility flag",
		"nested":         "removed nested workflow-route selector",
		"execute":        "removed execute compatibility flag",
	}

	apiDesignWalkCommands(rootCmd, func(cmd *cobra.Command) {
		checkFlags := func(kind string, flags *pflag.FlagSet) {
			flags.VisitAll(func(flag *pflag.Flag) {
				if !apiDesignKebabNamePattern.MatchString(flag.Name) {
					t.Errorf("%s flag --%s on %q is not kebab-case", kind, flag.Name, cmd.CommandPath())
				}
				if reason := removedFlags[flag.Name]; reason != "" {
					if cmd.CommandPath() == "retab workflows blocks api-calls run" && flag.Name == "execute" {
						return
					}
					t.Errorf("%s flag --%s on %q exposes %s", kind, flag.Name, cmd.CommandPath(), reason)
				}
			})
		}

		checkFlags("local", cmd.LocalFlags())
		checkFlags("persistent", cmd.PersistentFlags())
	})
}

func TestRemovedCommandSurfaceIsAbsent(t *testing.T) {
	commandPaths := apiDesignAllCommandPaths(rootCmd)
	removedPaths := []string{
		"env delete",
		"env remove",
		"files analyze",
		"files delete",
		"files doctor",
		"jobs",
		"workflows append-version",
		"workflows reviews append",
		"workflows reviews append-version",
		"workflows reviews edit",
		"workflows reviews versions append",
		"workflows reviews versions append-version",
		"workflows execute",
		"workflows api-calls",
		"workflows api-calls hydrate",
		"workflows api-calls render",
		"workflows api-calls run",
		"workflows functions",
		"workflows functions hydrate",
		"workflows functions run",
		"workflows blocks executions get",
		"workflows runs execute",
		"workflows runs block executions",
		"workflows runs block executions create",
		"workflows runs block executions list",
		"workflows runs reviews",
		"workflows runs tests",
		"workflows tests execute",
		"workflows tests runs execute",
	}

	for _, path := range removedPaths {
		if commandPaths[path] {
			t.Errorf("CLI still exposes removed command retab %s", path)
		}
	}
}

func TestCoreAPIResourcesExposeExpectedCommandSurface(t *testing.T) {
	expectedChildren := map[string][]string{
		"":                              {"auth", "env", "classifications", "edits", "extractions", "files", "parses", "partitions", "schemas", "secrets", "setup", "splits", "sync", "tables", "version", "workflows"},
		"auth":                          {"login", "logout", "status"},
		"env":                           {"add", "switch", "which", "claim", "list"},
		"classifications":               {"create", "get", "list", "cancel", "delete", "wait"},
		"edits":                         {"create", "get", "list", "cancel", "delete", "wait", "templates"},
		"edits templates":               {"create", "get", "list", "update", "delete"},
		"extractions":                   {"create", "stream", "list", "get", "sources", "cancel", "delete", "wait"},
		"files":                         {"list", "get", "upload", "download-link", "download", "create-upload", "complete-upload", "blueprints", "parse", "grep", "inspect"},
		"files blueprints":              {"create", "get", "cancel", "wait"},
		"parses":                        {"create", "get", "list", "cancel", "delete", "wait"},
		"partitions":                    {"create", "get", "list", "cancel", "delete", "wait"},
		"schemas":                       {"generate"},
		"secrets":                       {"delete", "get", "list", "set", "value"},
		"splits":                        {"create", "get", "list", "cancel", "delete", "wait"},
		"tables":                        {"create", "delete", "download", "get", "list", "profile", "query", "replace", "schema", "validate"},
		"workflows":                     {"list", "get", "create", "update", "delete", "publish", "discard-draft", "view", "runs", "steps", "blocks", "edges", "artifacts", "reviews", "tests", "experiments", "spec", "versions"},
		"workflows artifacts":           {"get", "list"},
		"workflows blocks":              {"list", "get", "create", "update", "delete", "executions", "pull-config", "push-config", "diff-config", "validate-config", "doctor-config", "api-calls", "functions", "versions"},
		"workflows blocks api-calls":    {"hydrate", "render", "run"},
		"workflows blocks functions":    {"hydrate", "run"},
		"workflows blocks versions":     {"list", "get", "diff", "restore"},
		"workflows edges":               {"list", "get", "create", "delete", "versions"},
		"workflows edges versions":      {"list", "get", "diff", "restore"},
		"workflows experiments":         {"create", "list", "get", "update", "delete", "runs", "results", "metrics"},
		"workflows experiments runs":    {"create", "list", "get", "cancel", "wait"},
		"workflows experiments metrics": {"get"},
		"workflows experiments results": {"list", "get"},
		"workflows reviews":             {"list", "get", "schema", "approve", "reject", "versions"},
		"workflows reviews versions":    {"list", "get", "create"},
		"workflows runs":                {"create", "get", "list", "delete", "cancel", "restart", "export"},
		"workflows blocks executions":   {"create", "list"},
		"workflows spec":                {"validate", "plan", "apply", "get"},
		"workflows steps":               {"list", "get"},
		"workflows tests":               {"create", "get", "list", "update", "delete", "runs", "results"},
		"workflows tests runs":          {"create", "list", "get", "cancel"},
		"workflows tests results":       {"list", "get"},
		"workflows versions":            {"list", "get", "diff", "restore"},
	}

	for commandPath, expected := range expectedChildren {
		apiDesignAssertExactChildren(t, commandPath, expected)
	}
}

func apiDesignWalkCommands(root *cobra.Command, visit func(*cobra.Command)) {
	visit(root)
	for _, child := range apiDesignVisibleChildren(root) {
		apiDesignWalkCommands(child, visit)
	}
}

func apiDesignAllCommandPaths(root *cobra.Command) map[string]bool {
	paths := map[string]bool{}
	apiDesignWalkCommands(root, func(cmd *cobra.Command) {
		if cmd == root {
			return
		}
		path := strings.TrimPrefix(cmd.CommandPath(), root.Name()+" ")
		if path != "" {
			paths[path] = true
		}
	})
	return paths
}

func apiDesignAssertExactChildren(t *testing.T, commandPath string, expected []string) {
	t.Helper()
	cmd := apiDesignFindCommand(rootCmd, commandPath)
	if cmd == nil {
		t.Fatalf("missing command retab %s", commandPath)
	}

	actual := apiDesignVisibleChildNames(cmd)
	expectedSet := map[string]bool{}
	for _, name := range expected {
		expectedSet[name] = true
	}

	var missing []string
	for _, name := range expected {
		if !actual[name] {
			missing = append(missing, name)
		}
	}

	var unexpected []string
	for name := range actual {
		if !expectedSet[name] {
			unexpected = append(unexpected, name)
		}
	}

	sort.Strings(missing)
	sort.Strings(unexpected)
	if len(missing) > 0 || len(unexpected) > 0 {
		t.Fatalf("retab %s child commands mismatch\nmissing: %s\nunexpected: %s",
			commandPath,
			strings.Join(missing, ", "),
			strings.Join(unexpected, ", "),
		)
	}
}

func apiDesignFindCommand(root *cobra.Command, commandPath string) *cobra.Command {
	current := root
	for _, part := range strings.Fields(commandPath) {
		var next *cobra.Command
		for _, child := range apiDesignVisibleChildren(current) {
			if child.Name() == part {
				next = child
				break
			}
		}
		if next == nil {
			return nil
		}
		current = next
	}
	return current
}

func apiDesignVisibleChildNames(cmd *cobra.Command) map[string]bool {
	names := map[string]bool{}
	for _, child := range apiDesignVisibleChildren(cmd) {
		names[child.Name()] = true
	}
	return names
}

func apiDesignVisibleChildren(cmd *cobra.Command) []*cobra.Command {
	var children []*cobra.Command
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" {
			continue
		}
		children = append(children, child)
	}
	return children
}
