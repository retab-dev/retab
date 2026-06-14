package cmd

// `retab projects access` — manage who can access a project and at which role
// (owner / editor / operator / viewer). Backed by /v1/project-memberships, a
// dashboard-internal route group absent from the public OpenAPI surface and the
// generated SDK, so — like `members` and `org` — the commands talk to the
// endpoint directly via cliJSONRequestInto.
//
// Scope is the active organization + environment (from the session credential).
// A grant requires a dashboard session: these routes reject non-dashboard auth,
// so use an OAuth login (`retab auth login`). The subject of a grant is normally
// a user — pass the user id from `retab members list` as --subject-id.
//
// This file also owns the membership row type, columns, and list printer SHARED
// with `retab workflows access` (workflows_access.go), since the project- and
// workflow-membership wire shapes differ only in the resource id field.

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const projectMembershipsBasePath = "/v1/project-memberships"

// membershipSubjectTypes is the principal vocabulary a grant can target (server
// enum). "user" is the common case (a teammate); "organization_membership" and
// "application" are for membership-id / service-account subjects.
var membershipSubjectTypes = []string{"user", "application", "organization_membership"}

var projectMembershipRoles = []string{"project-owner", "project-editor", "project-operator", "project-viewer"}

// cliMembership decodes a project OR workflow membership response. The two wire
// shapes are identical except for the resource id field (project_id vs
// workflow_id), so a single struct with both (omitempty) decodes either, and the
// shared columns render whichever is present.
type cliMembership struct {
	ID            string  `json:"id"`
	EnvironmentID string  `json:"environment_id"`
	ProjectID     string  `json:"project_id,omitempty"`
	WorkflowID    string  `json:"workflow_id,omitempty"`
	SubjectType   string  `json:"subject_type"`
	SubjectID     string  `json:"subject_id"`
	Role          string  `json:"role"`
	IsActive      bool    `json:"is_active"`
	RemovedAt     *string `json:"removed_at"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func validateRoleIn(role string, allowed []string) error {
	for _, r := range allowed {
		if role == r {
			return nil
		}
	}
	return fmt.Errorf("invalid --role %q: must be one of %s", role, strings.Join(allowed, ", "))
}

func validateSubjectType(subjectType string) error {
	for _, t := range membershipSubjectTypes {
		if subjectType == t {
			return nil
		}
	}
	return fmt.Errorf("invalid --subject-type %q: must be one of %s", subjectType, strings.Join(membershipSubjectTypes, ", "))
}

func membershipCell(row any, key string) string {
	m, ok := row.(cliMembership)
	if !ok {
		return ""
	}
	switch key {
	case "id":
		return m.ID
	case "resource":
		if m.ProjectID != "" {
			return m.ProjectID
		}
		return m.WorkflowID
	case "subject":
		if m.SubjectID == "" {
			return "-"
		}
		return m.SubjectType + ":" + m.SubjectID
	case "role":
		return m.Role
	case "active":
		if m.IsActive {
			return "yes"
		}
		return "no"
	case "created_at":
		return normalizeTimestampCell(m.CreatedAt)
	default:
		return ""
	}
}

// membershipColumns leads with the human-readable columns and puts ID LAST.
// A facade membership id is a ~200-char base64 blob (it encodes
// resource+subject+role), so an ID-first table — the usual CLI convention —
// pushes the scannable columns off-screen. The id is still needed for
// update/revoke, so it stays in the table (and is column-1 in JSON), just at the
// end where it doesn't wreck readability.
var membershipColumns = []TableColumn{
	{Header: "RESOURCE", Extract: func(row any) string { return membershipCell(row, "resource") }},
	{Header: "SUBJECT", Extract: func(row any) string { return membershipCell(row, "subject") }},
	{Header: "ROLE", Extract: func(row any) string { return membershipCell(row, "role") }},
	{Header: "ACTIVE", Extract: func(row any) string { return membershipCell(row, "active") }},
	{Header: "CREATED_AT", Extract: func(row any) string { return membershipCell(row, "created_at") }},
	{Header: "ID", Extract: func(row any) string { return membershipCell(row, "id") }},
}

// resolveAccessResourceID returns the resource id an access-list command is
// scoped to, accepting it either positionally (preferred, matching the workflow
// subcommand convention `list <id>`) or via the named flag (--project-id /
// --workflow-id). The two forms are co-equal; passing both is allowed only when
// they agree. The id is required: these routes authorize the read against a
// specific resource, so an org-wide listing is rejected server-side anyway.
func resolveAccessResourceID(cmd *cobra.Command, args []string, flagName string) (string, error) {
	flagVal, _ := cmd.Flags().GetString(flagName)
	flagVal = strings.TrimSpace(flagVal)
	var positional string
	if len(args) > 0 {
		positional = strings.TrimSpace(args[0])
	}
	switch {
	case positional != "" && flagVal != "" && positional != flagVal:
		return "", fmt.Errorf("conflicting %s: positional %q vs --%s %q", flagName, positional, flagName, flagVal)
	case positional != "":
		return positional, nil
	case flagVal != "":
		return flagVal, nil
	default:
		return "", fmt.Errorf("a %s is required (as the first argument or via --%s)", flagName, flagName)
	}
}

// printMembershipList renders a paginated membership envelope, shared by the
// project- and workflow-access list commands.
func printMembershipList(cmd *cobra.Command, result *cliPaginatedList[cliMembership]) error {
	format, err := ResolveOutputFormat(cmd, os.Stdout)
	if err != nil {
		return err
	}
	if format == OutputTable || format == OutputCSV {
		return RenderList(os.Stdout, format, result, membershipColumns)
	}
	return printJSON(result)
}

// membershipListQuery builds the shared list query params (subject/role filters
// + pagination) from flags common to both access list commands. The caller adds
// the resource-id filter (--project-id / --workflow-id).
func membershipListQuery(cmd *cobra.Command) (url.Values, error) {
	if err := validateBeforeAfterMutex(cmd); err != nil {
		return nil, err
	}
	query := url.Values{}
	if v, _ := cmd.Flags().GetString("subject-type"); v != "" {
		query.Set("subject_type", v)
	}
	if v, _ := cmd.Flags().GetString("subject-id"); v != "" {
		query.Set("subject_id", v)
	}
	if v, _ := cmd.Flags().GetString("role"); v != "" {
		query.Set("role", v)
	}
	if v, _ := cmd.Flags().GetBool("include-inactive"); v {
		query.Set("include_inactive", "true")
	}
	if v, _ := cmd.Flags().GetString("before"); v != "" {
		query.Set("before", v)
	}
	if v, _ := cmd.Flags().GetString("after"); v != "" {
		query.Set("after", v)
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		query.Set("limit", strconv.Itoa(v))
	}
	return query, nil
}

var projectsAccessCmd = &cobra.Command{
	Use:   "access",
	Short: "Manage who can access a project",
	Long: `Manage project access grants: who can use a project and at which role.

Roles: project-owner, project-editor, project-operator, project-viewer.
A project must keep at least one owner. Grants are scoped to the active
organization + environment and require a dashboard (OAuth) session.`,
	Example: `  retab projects access list --project-id proj_01HX...
  retab projects access grant --project-id proj_01HX... --subject-id user_01HX... --role project-editor
  retab projects access update pmem_... --role project-viewer
  retab projects access revoke pmem_...`,
}

var projectsAccessListCmd = &cobra.Command{
	Use:   "list <project-id>",
	Short: "List project access grants",
	Long: `List who can access a project, and at which role.

Name the project either positionally (` + "`list <project-id>`" + `) or with the
` + "`--project-id`" + ` flag — the project id is required because the route
authorizes the read against that specific project. Filter further with
--subject-id / --role, and include revoked grants with --include-inactive.`,
	Example: `  retab projects access list proj_01HX...
  retab projects access list proj_01HX... --role project-owner`,
	Args: cobra.MaximumNArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		projectID, err := resolveAccessResourceID(cmd, args, "project-id")
		if err != nil {
			return err
		}
		query, err := membershipListQuery(cmd)
		if err != nil {
			return err
		}
		query.Set("project_id", projectID)
		var result cliPaginatedList[cliMembership]
		if err := cliJSONRequestInto(cmd, http.MethodGet, projectMembershipsBasePath, query, nil, &result); err != nil {
			return err
		}
		return printMembershipList(cmd, &result)
	}),
}

var projectsAccessGrantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Grant a user access to a project",
	Long: `Grant a subject access to a project at a role.

Identify the person with --email (resolved against ` + "`retab members list`" + `)
or, for a service account / explicit id, --subject-id with --subject-type.
Roles: project-owner, project-editor, project-operator, project-viewer.`,
	Example: `  # By email (the common case)
  retab projects access grant --project-id proj_01HX... --email alice@acme.com --role project-editor

  # By explicit subject id
  retab projects access grant --project-id proj_01HX... --subject-id user_01HX... --role project-editor`,
	Args: cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetString("project-id")
		subjectID, _ := cmd.Flags().GetString("subject-id")
		email, _ := cmd.Flags().GetString("email")
		role, _ := cmd.Flags().GetString("role")
		subjectType, _ := cmd.Flags().GetString("subject-type")
		if strings.TrimSpace(projectID) == "" {
			return fmt.Errorf("--project-id is required")
		}
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("--role is required")
		}
		if err := validateRoleIn(role, projectMembershipRoles); err != nil {
			return err
		}
		if err := validateSubjectType(subjectType); err != nil {
			return err
		}
		// Subject identification: --email (user lookup) XOR --subject-id.
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
			"project_id":   projectID,
			"subject_type": subjectType,
			"subject_id":   subjectID,
			"role":         role,
		}
		var result cliMembership
		if err := cliJSONRequestInto(cmd, http.MethodPost, projectMembershipsBasePath, nil, body, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var projectsAccessGetCmd = &cobra.Command{
	Use:   "get <membership-id>",
	Short: "Get one project access grant",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		query := url.Values{}
		if v, _ := cmd.Flags().GetBool("include-inactive"); v {
			query.Set("include_inactive", "true")
		}
		path := projectMembershipsBasePath + "/" + url.PathEscape(args[0])
		var result cliMembership
		if err := cliJSONRequestInto(cmd, http.MethodGet, path, query, nil, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var projectsAccessUpdateCmd = &cobra.Command{
	Use:   "update <membership-id>",
	Short: "Change a project access grant's role",
	Long: `Change the role of an existing project access grant.

Roles: project-owner, project-editor, project-operator, project-viewer.
Demoting the last owner is rejected.`,
	Example: `  retab projects access update pmem_... --role project-viewer`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("--role is required")
		}
		if err := validateRoleIn(role, projectMembershipRoles); err != nil {
			return err
		}
		path := projectMembershipsBasePath + "/" + url.PathEscape(args[0])
		body := map[string]string{"role": role}
		var result cliMembership
		if err := cliJSONRequestInto(cmd, http.MethodPatch, path, nil, body, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var projectsAccessRevokeCmd = &cobra.Command{
	Use:   "revoke <membership-id>",
	Short: "Revoke a project access grant",
	Long: `Revoke (deactivate) a project access grant.

Revoking the last owner is rejected. Requires --yes when stdin is not a
terminal.`,
	Example: `  retab projects access revoke pmem_... --yes`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "project access grant", args[0]); err != nil {
			return err
		}
		path := projectMembershipsBasePath + "/" + url.PathEscape(args[0])
		var result cliMembership
		if err := cliJSONRequestInto(cmd, http.MethodDelete, path, nil, nil, &result); err != nil {
			return err
		}
		confirmDeleted("project access grant", args[0])
		return nil
	}),
}

// addMembershipListFlags registers the filter + pagination flags shared by the
// project- and workflow-access list commands (the resource-id filter is added
// per-command).
func addMembershipListFlags(cmd *cobra.Command) {
	cmd.Flags().String("subject-type", "", "filter by subject type: user, application, or organization_membership")
	cmd.Flags().String("subject-id", "", "filter by subject id")
	cmd.Flags().String("role", "", "filter by role slug")
	cmd.Flags().Bool("include-inactive", false, "include revoked (inactive) grants")
	cmd.Flags().String("before", "", "membership id: return items before this id (mutually exclusive with --after)")
	cmd.Flags().String("after", "", "membership id: return items after this id (mutually exclusive with --before)")
	cmd.Flags().Int("limit", 0, "maximum number of grants to return")
}

func init() {
	addMembershipListFlags(projectsAccessListCmd)
	projectsAccessListCmd.Flags().String("project-id", "", "project whose grants to list (alternative to the positional form)")

	projectsAccessGrantCmd.Flags().String("project-id", "", "project to grant access to (required)")
	projectsAccessGrantCmd.Flags().String("email", "", "email of the user to grant (resolved via the org member list; alternative to --subject-id)")
	projectsAccessGrantCmd.Flags().String("subject-id", "", "explicit subject id, e.g. a user id (alternative to --email)")
	projectsAccessGrantCmd.Flags().String("subject-type", "user", "subject type: user, application, or organization_membership")
	projectsAccessGrantCmd.Flags().String("role", "", "role to grant: project-owner, project-editor, project-operator, or project-viewer (required)")
	_ = projectsAccessGrantCmd.MarkFlagRequired("project-id")
	_ = projectsAccessGrantCmd.MarkFlagRequired("role")

	projectsAccessGetCmd.Flags().Bool("include-inactive", false, "allow resolving a revoked (inactive) grant")

	projectsAccessUpdateCmd.Flags().String("role", "", "new role: project-owner, project-editor, project-operator, or project-viewer (required)")
	_ = projectsAccessUpdateCmd.MarkFlagRequired("role")

	projectsAccessRevokeCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	projectsAccessCmd.AddCommand(
		projectsAccessListCmd,
		projectsAccessGrantCmd,
		projectsAccessGetCmd,
		projectsAccessUpdateCmd,
		projectsAccessRevokeCmd,
	)
	projectsCmd.AddCommand(projectsAccessCmd)
}
