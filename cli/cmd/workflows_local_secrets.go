//go:build !retab_oagen_cli_workflows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type workflowSecretMount struct {
	Name     string
	Env      string
	Required bool
}

func workflowSecretMounts(config map[string]any) []workflowSecretMount {
	mounts, _ := config["mounts"].(map[string]any)
	rawSecrets, _ := mounts["secrets"].([]any)
	secrets := []workflowSecretMount{}
	for _, raw := range rawSecrets {
		secret, _ := raw.(map[string]any)
		if secret == nil {
			continue
		}
		name := strings.TrimSpace(stringFromAny(secret["name"]))
		env := strings.TrimSpace(stringFromAny(secret["env"]))
		if env == "" {
			env = name
		}
		if name == "" {
			name = env
		}
		if name == "" || env == "" {
			continue
		}
		required := true
		if rawRequired, ok := secret["required"].(bool); ok {
			required = rawRequired
		}
		secrets = append(secrets, workflowSecretMount{
			Name:     name,
			Env:      env,
			Required: required,
		})
	}
	return secrets
}

func fillLocalSecretsFromRetab(cmd *cobra.Command, bundleDir string, config map[string]any, forceSecrets bool) ([]map[string]any, error) {
	secrets := workflowSecretMounts(config)
	if len(secrets) == 0 {
		return []map[string]any{}, nil
	}
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	ctx, cancel := ctxFor(cmd)
	defer cancel()
	values := map[string]string{}
	results := make([]map[string]any, 0, len(secrets))
	for _, secret := range secrets {
		response, err := client.Secrets.GetValue(ctx, secret.Name)
		if err != nil {
			return nil, fmt.Errorf("read secret value %s: %w", secret.Name, err)
		}
		values[secret.Env] = response.Secret.Value
		results = append(results, map[string]any{
			"name":    secret.Name,
			"env":     secret.Env,
			"written": false,
		})
	}
	written, err := writeLocalSecretsEnvFile(filepath.Join(bundleDir, ".env.local"), values, forceSecrets)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		env, _ := result["env"].(string)
		result["written"] = written[env]
	}
	return results, nil
}

func writeLocalSecretsEnvFile(path string, values map[string]string, force bool) (map[string]bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	lines := []string{}
	if err == nil {
		lines = strings.Split(string(raw), "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	}
	seen := map[string]bool{}
	written := map[string]bool{}
	for idx, line := range lines {
		key, value, ok := parseLocalEnvAssignment(line)
		if !ok {
			continue
		}
		secretValue, exists := values[key]
		if !exists {
			continue
		}
		seen[key] = true
		if force || value == "" || value == "__REPLACE_ME__" {
			lines[idx] = key + "=" + secretValue
			written[key] = true
		} else {
			written[key] = false
		}
	}
	for key, value := range values {
		if seen[key] {
			continue
		}
		lines = append(lines, key+"="+value)
		written[key] = true
	}
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return nil, err
	}
	return written, nil
}

func parseLocalEnvAssignment(line string) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false
	}
	key, value, ok := strings.Cut(trimmed, "=")
	if !ok {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", "", false
	}
	return key, strings.TrimSpace(value), true
}
