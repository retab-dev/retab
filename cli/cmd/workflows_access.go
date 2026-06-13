package cmd

// `retab workflows access` — manage DIRECT role grants on a single workflow
// (additive grants on top of whatever the parent project already grants).
// Backed by /v1/workflow-memberships, a dashboard-internal route group absent
// from the public OpenAPI surface and the generated SDK.
//
// The membership row type, columns, and list printer are shared with
// `retab projects access` (projects_access.go).
//
// A direct workflow grant is additive on top of whatever the parent project
// already grants — use it to give someone access to one workflow without
// granting the whole project.

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

const workflowMembershipsBasePath = "/v1/workflow-memberships"

var workflowMembershipRoles = []string{"workflow-owner", "workflow-editor", "workflow-operator", "workflow-viewer"}

var workflowsAccessCmd = &cobra.Command{
	Use:   "access",
	Short: "Manage direct role grants on a workflow",
	Long: `Manage direct workflow access grants — additive role grants on a single
workflow, on top of what the parent project already grants.

Roles: workflow-owner, workflow-editor, workflow-operator, workflow-viewer.
Grants are scoped to the active organization + environment and require a
dashboard (OAuth) session.

A direct grant is additive on top of the parent project's access; for broad
access prefer granting at the project level (` + "`retab projects access grant`" + `),
which cascades to the project's workflows.`,
	Example: `  retab workflows access list --workflow-id wf_01HX...
  retab workflows access grant --workflow-id wf_01HX... --email alice@acme.com --role workflow-operator
  retab workflows access update wmem_... --role workflow-viewer
  retab workflows access revoke wmem_...`,
}

var workflowsAccessListCmd = &cobra.Command{
	Use:   "list",
	Short: "List direct workflow access grants",
	Long: `List the direct role grants on a workflow (additive grants on top of
what the parent project grants).

--workflow-id is required: the route authorizes the read against that specific
workflow. Filter further with --subject-id / --role, and include revoked grants
with --include-inactive.`,
	Example: `  retab workflows access list --workflow-id wf_01HX...`,
	Args:    cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, _ := cmd.Flags().GetString("workflow-id")
		if strings.TrimSpace(workflowID) == "" {
			return fmt.Errorf("--workflow-id is required")
		}
		query, err := membershipListQuery(cmd)
		if err != nil {
			return err
		}
		query.Set("workflow_id", workflowID)
		var result cliPaginatedList[cliMembership]
		if err := cliJSONRequestInto(cmd, http.MethodGet, workflowMembershipsBasePath, query, nil, &result); err != nil {
			return err
		}
		return printMembershipList(cmd, &result)
	}),
}

var workflowsAccessGrantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Grant a user a direct role on a workflow",
	Long: `Grant a subject a direct role on a workflow (additive on top of the
parent project's access).

Identify the person with --email (resolved against ` + "`retab members list`" + `)
or, for a service account / explicit id, --subject-id with --subject-type.
Roles: workflow-owner, workflow-editor, workflow-operator, workflow-viewer.`,
	Example: `  retab workflows access grant --workflow-id wf_01HX... --email alice@acme.com --role workflow-operator
  retab workflows access grant --workflow-id wf_01HX... --subject-id user_01HX... --role workflow-viewer`,
	Args: cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		workflowID, _ := cmd.Flags().GetString("workflow-id")
		subjectID, _ := cmd.Flags().GetString("subject-id")
		email, _ := cmd.Flags().GetString("email")
		role, _ := cmd.Flags().GetString("role")
		subjectType, _ := cmd.Flags().GetString("subject-type")
		if strings.TrimSpace(workflowID) == "" {
			return fmt.Errorf("--workflow-id is required")
		}
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("--role is required")
		}
		if err := validateRoleIn(role, workflowMembershipRoles); err != nil {
			return err
		}
		if err := validateSubjectType(subjectType); err != nil {
			return err
		}
		if strings.TrimSpace(email) != "" {
			if strings.TrimSpace(subjectID) != "" {
				return fmt.Errorf("--email and --subject-id are mutually exclusive")
			}
			if subjectType != "user" {
				return fmt.Errorf("--email resolves a user; use --subject-id with --subject-type %s", subjectType)
			}
			member, err := resolveMemberByEmail(cmd, email)
			if err != nil {
				return err
			}
			subjectID = member.ID
		}
		if strings.TrimSpace(subjectID) == "" {
			return fmt.Errorf("one of --email or --subject-id is required")
		}
		body := map[string]any{
			"workflow_id":  workflowID,
			"subject_type": subjectType,
			"subject_id":   subjectID,
			"role":         role,
		}
		var result cliMembership
		if err := cliJSONRequestInto(cmd, http.MethodPost, workflowMembershipsBasePath, nil, body, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsAccessGetCmd = &cobra.Command{
	Use:   "get <membership-id>",
	Short: "Get one direct workflow access grant",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		query := url.Values{}
		if v, _ := cmd.Flags().GetBool("include-inactive"); v {
			query.Set("include_inactive", "true")
		}
		path := workflowMembershipsBasePath + "/" + url.PathEscape(args[0])
		var result cliMembership
		if err := cliJSONRequestInto(cmd, http.MethodGet, path, query, nil, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsAccessUpdateCmd = &cobra.Command{
	Use:   "update <membership-id>",
	Short: "Change a direct workflow access grant's role",
	Long: `Change the role of an existing direct workflow access grant.

Roles: workflow-owner, workflow-editor, workflow-operator, workflow-viewer.`,
	Example: `  retab workflows access update wmem_... --role workflow-viewer`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("--role is required")
		}
		if err := validateRoleIn(role, workflowMembershipRoles); err != nil {
			return err
		}
		path := workflowMembershipsBasePath + "/" + url.PathEscape(args[0])
		body := map[string]string{"role": role}
		var result cliMembership
		if err := cliJSONRequestInto(cmd, http.MethodPatch, path, nil, body, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var workflowsAccessRevokeCmd = &cobra.Command{
	Use:   "revoke <membership-id>",
	Short: "Revoke a direct workflow access grant",
	Long: `Revoke (deactivate) a direct workflow access grant.

Requires --yes when stdin is not a terminal.`,
	Example: `  retab workflows access revoke wmem_... --yes`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "workflow access grant", args[0]); err != nil {
			return err
		}
		path := workflowMembershipsBasePath + "/" + url.PathEscape(args[0])
		var result cliMembership
		if err := cliJSONRequestInto(cmd, http.MethodDelete, path, nil, nil, &result); err != nil {
			return err
		}
		confirmDeleted("workflow access grant", args[0])
		return nil
	}),
}

func init() {
	addMembershipListFlags(workflowsAccessListCmd)
	workflowsAccessListCmd.Flags().String("workflow-id", "", "workflow whose grants to list (required)")
	_ = workflowsAccessListCmd.MarkFlagRequired("workflow-id")

	workflowsAccessGrantCmd.Flags().String("workflow-id", "", "workflow to grant access to (required)")
	workflowsAccessGrantCmd.Flags().String("email", "", "email of the user to grant (resolved via `retab members list`; alternative to --subject-id)")
	workflowsAccessGrantCmd.Flags().String("subject-id", "", "explicit subject id, e.g. a user id (alternative to --email)")
	workflowsAccessGrantCmd.Flags().String("subject-type", "user", "subject type: user, application, or organization_membership")
	workflowsAccessGrantCmd.Flags().String("role", "", "role to grant: workflow-owner, workflow-editor, workflow-operator, or workflow-viewer (required)")
	_ = workflowsAccessGrantCmd.MarkFlagRequired("workflow-id")
	_ = workflowsAccessGrantCmd.MarkFlagRequired("role")

	workflowsAccessGetCmd.Flags().Bool("include-inactive", false, "allow resolving a revoked (inactive) grant")

	workflowsAccessUpdateCmd.Flags().String("role", "", "new role: workflow-owner, workflow-editor, workflow-operator, or workflow-viewer (required)")
	_ = workflowsAccessUpdateCmd.MarkFlagRequired("role")

	workflowsAccessRevokeCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	workflowsAccessCmd.AddCommand(
		workflowsAccessListCmd,
		workflowsAccessGrantCmd,
		workflowsAccessGetCmd,
		workflowsAccessUpdateCmd,
		workflowsAccessRevokeCmd,
	)
	workflowsCmd.AddCommand(workflowsAccessCmd)
}
