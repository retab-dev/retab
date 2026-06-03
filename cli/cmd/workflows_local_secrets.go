//go:build !retab_oagen_cli_workflows

package cmd

import (
	"fmt"
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

func fillLocalSecretsFromRetab(cmd *cobra.Command, config map[string]any, forceSecrets bool) ([]map[string]any, error) {
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
	checked := make([]map[string]any, 0, len(secrets))
	for _, secret := range secrets {
		if _, err := client.Secrets.Get(ctx, secret.Name); err != nil {
			return nil, fmt.Errorf("read secret metadata %s: %w", secret.Name, err)
		}
		checked = append(checked, map[string]any{
			"name":    secret.Name,
			"env":     secret.Env,
			"written": false,
		})
	}
	return checked, fmt.Errorf("--fill-secrets requested, but the current Retab secrets API exposes metadata only and does not return secret values; leave .env.local placeholders or export the env vars locally")
}
