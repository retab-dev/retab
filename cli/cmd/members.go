package cmd

// `retab members` — list and manage the people in the active WorkOS
// organization (list, change role, remove, inspect resource permissions).
//
// These map to the internal dashboard people-management routes under
// /internal/workos/organizations/members. They are NOT part of the public
// OpenAPI surface and so are absent from the generated SDK; like `org` and
// `projects`, the command talks to the endpoints directly via
// cliJSONRequestInto, which attaches the active credential (the org id rides
// on the token/api-key).
//
// The organization scope is implicit: every route resolves the caller's
// organization from the credential, so there is no --organization-id flag.
// These routes are RBAC-gated (people read/manage permissions); an org-scoped
// API key without those grants will get a 403 — use an OAuth login for member
// administration.

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const membersBasePath = "/internal/workos/organizations/members"

// memberRoles is the role vocabulary the backend accepts (enum on
// updateMemberRoleRequest.Role). Validated client-side for a clearer error
// than the server's 422.
var memberRoles = []string{"superadmin", "admin", "member"}

// cliTeamMember mirrors the server's TeamMemberResponse. Every field after
// role is nullable, matching the wire shape.
type cliTeamMember struct {
	ID                string  `json:"id"`
	Email             string  `json:"email"`
	Role              string  `json:"role"`
	MembershipID      *string `json:"membership_id"`
	FirstName         *string `json:"first_name"`
	LastName          *string `json:"last_name"`
	ProfilePictureURL *string `json:"profile_picture_url"`
	LastActiveAt      *string `json:"last_active_at"`
}

// cliMemberResourceGrant / cliMemberPermissions mirror the
// member-permissions response (per-resource role grants).
type cliMemberResourceGrant struct {
	ID    string   `json:"id"`
	Name  *string  `json:"name,omitempty"`
	Roles []string `json:"roles"`
}

type cliMemberPermissions struct {
	Projects  []cliMemberResourceGrant `json:"projects"`
	Workflows []cliMemberResourceGrant `json:"workflows"`
}

func validateMemberRole(role string) error {
	for _, r := range memberRoles {
		if role == r {
			return nil
		}
	}
	return fmt.Errorf("invalid --role %q: must be one of %s", role, strings.Join(memberRoles, ", "))
}

// resolveMemberByEmail looks up an organization member by email (case-insensitive)
// via the members list, so role/access commands can take a human email instead of
// an opaque user id. A pending invitee is NOT a member yet, so an unaccepted
// invitation will not resolve — the error says so.
func resolveMemberByEmail(cmd *cobra.Command, email string) (cliTeamMember, error) {
	var members []cliTeamMember
	if err := cliJSONRequestInto(cmd, http.MethodGet, membersBasePath, nil, nil, &members); err != nil {
		return cliTeamMember{}, err
	}
	target := strings.TrimSpace(email)
	var matches []cliTeamMember
	for _, m := range members {
		if strings.EqualFold(strings.TrimSpace(m.Email), target) {
			matches = append(matches, m)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return cliTeamMember{}, fmt.Errorf(
			"no organization member with email %q. Run `retab members list` to see members "+
				"(a person who has not yet accepted their invitation is not a member yet)",
			email,
		)
	default:
		return cliTeamMember{}, fmt.Errorf("multiple members share email %q; pass the user id instead", email)
	}
}

// resolveMemberRef maps a positional member reference to a user id: an argument
// containing "@" is treated as an email and resolved via the member list;
// anything else is used as-is (a user id). This lets `members update` / `members
// remove` accept either form.
func resolveMemberRef(cmd *cobra.Command, ref string) (string, error) {
	if strings.Contains(ref, "@") {
		m, err := resolveMemberByEmail(cmd, ref)
		if err != nil {
			return "", err
		}
		return m.ID, nil
	}
	return ref, nil
}

func memberDisplayName(m cliTeamMember) string {
	parts := make([]string, 0, 2)
	if first := strings.TrimSpace(derefString(m.FirstName)); first != "" {
		parts = append(parts, first)
	}
	if last := strings.TrimSpace(derefString(m.LastName)); last != "" {
		parts = append(parts, last)
	}
	return strings.Join(parts, " ")
}

func memberRowCell(row any, key string) string {
	m, ok := row.(cliTeamMember)
	if !ok {
		return ""
	}
	switch key {
	case "id":
		return m.ID
	case "email":
		return m.Email
	case "role":
		return m.Role
	case "name":
		if name := memberDisplayName(m); name != "" {
			return name
		}
		return "-"
	case "last_active_at":
		if v := strings.TrimSpace(derefString(m.LastActiveAt)); v != "" {
			return normalizeTimestampCell(v)
		}
		return "-"
	default:
		return ""
	}
}

var memberListColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return memberRowCell(row, "id") }},
	{Header: "EMAIL", Extract: func(row any) string { return memberRowCell(row, "email") }},
	{Header: "ROLE", Extract: func(row any) string { return memberRowCell(row, "role") }},
	{Header: "NAME", Extract: func(row any) string { return memberRowCell(row, "name") }},
	{Header: "LAST_ACTIVE_AT", Extract: func(row any) string { return memberRowCell(row, "last_active_at") }},
}

var membersCmd = &cobra.Command{
	Use:   "members",
	Short: "List and manage organization members",
	Long: `Manage the people in your active organization.

Members belong to the organization the current session is scoped to (see
` + "`retab org`" + `). Listing and role changes are gated by people-management
permissions, so use an OAuth login (` + "`retab auth login`" + `) — an org API
key generally lacks those grants.`,
	Example: `  retab members list
  retab members update user_01HX... --role admin
  retab members remove user_01HX...
  retab members permissions user_01HX...`,
}

var membersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the organization's members",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		var members []cliTeamMember
		if err := cliJSONRequestInto(cmd, http.MethodGet, membersBasePath, nil, nil, &members); err != nil {
			return err
		}

		format, err := ResolveOutputFormat(cmd, os.Stdout)
		if err != nil {
			return err
		}
		rows := make([]any, 0, len(members))
		for _, m := range members {
			rows = append(rows, m)
		}
		if format == OutputJSON {
			return printJSON(members)
		}
		if format == OutputCSV {
			return renderAutoCSV(os.Stdout, rows, memberListColumns)
		}
		return renderAutoTable(os.Stdout, rows, memberListColumns)
	}),
}

var membersUpdateCmd = &cobra.Command{
	Use:   "update <member>",
	Short: "Change a member's organization role",
	Long: `Update a member's organization role.

The member is a user id or an email (resolved against the member list).
Roles: superadmin, admin, member. You cannot change your own role.`,
	Example: `  retab members update alice@acme.com --role admin
  retab members update user_01HX... --role admin`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("--role is required")
		}
		if err := validateMemberRole(role); err != nil {
			return err
		}
		memberID, err := resolveMemberRef(cmd, args[0])
		if err != nil {
			return err
		}
		path := membersBasePath + "/" + url.PathEscape(memberID) + "/role"
		body := map[string]string{"role": role}
		var result cliTeamMember
		if err := cliJSONRequestInto(cmd, http.MethodPatch, path, nil, body, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var membersRemoveCmd = &cobra.Command{
	Use:   "remove <member>",
	Short: "Remove a member from the organization",
	Long: `Remove a member from the organization. You cannot remove yourself.

The member is a user id or an email (resolved against the member list).
Requires --yes when stdin is not a terminal.`,
	Example: `  retab members remove alice@acme.com --yes
  retab members remove user_01HX... --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		memberID, err := resolveMemberRef(cmd, args[0])
		if err != nil {
			return err
		}
		if err := confirmDestructive(cmd, "member", memberID); err != nil {
			return err
		}
		path := membersBasePath + "/" + url.PathEscape(memberID)
		var result map[string]any
		if err := cliJSONRequestInto(cmd, http.MethodDelete, path, nil, nil, &result); err != nil {
			return err
		}
		confirmDeleted("member", memberID)
		return nil
	}),
}

var membersPermissionsCmd = &cobra.Command{
	Use:   "permissions <member-id>",
	Short: "Show a member's project and workflow permissions",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		path := membersBasePath + "/" + url.PathEscape(args[0]) + "/permissions"
		var result cliMemberPermissions
		if err := cliJSONRequestInto(cmd, http.MethodGet, path, nil, nil, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

func init() {
	membersUpdateCmd.Flags().String("role", "", "new role: superadmin, admin, or member (required)")
	_ = membersUpdateCmd.MarkFlagRequired("role")
	membersRemoveCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	membersCmd.AddCommand(membersListCmd, membersUpdateCmd, membersRemoveCmd, membersPermissionsCmd)
	rootCmd.AddCommand(membersCmd)
}
