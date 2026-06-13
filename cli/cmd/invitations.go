package cmd

// `retab invitations` — list, send, and revoke invitations to the active
// WorkOS organization.
//
// These map to the internal dashboard routes under
// /internal/workos/organizations/invitations. They are NOT part of the public
// OpenAPI surface and so are absent from the generated SDK; like `members` and
// `org`, the command talks to the endpoints directly via cliJSONRequestInto.
//
// The organization scope is implicit: each route resolves the caller's
// organization from the credential. The routes are RBAC-gated (people-invite
// permission), so use an OAuth login for invitation management — an org API key
// generally lacks that grant.

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const invitationsBasePath = "/internal/workos/organizations/invitations"

// invitationRoles is the role vocabulary the backend accepts for the role an
// invitee is granted on acceptance (enum on invitationRequest.Role).
var invitationRoles = []string{"superadmin", "admin", "member"}

// cliInvitation mirrors the server's InvitationResponse. role, created_at, and
// expires_at are nullable.
type cliInvitation struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	State     string  `json:"state"`
	Role      *string `json:"role"`
	CreatedAt *string `json:"created_at"`
	ExpiresAt *string `json:"expires_at"`
}

func validateInvitationRole(role string) error {
	for _, r := range invitationRoles {
		if role == r {
			return nil
		}
	}
	return fmt.Errorf("invalid --role %q: must be one of %s", role, strings.Join(invitationRoles, ", "))
}

func invitationRowCell(row any, key string) string {
	inv, ok := row.(cliInvitation)
	if !ok {
		return ""
	}
	switch key {
	case "id":
		return inv.ID
	case "email":
		return inv.Email
	case "state":
		return inv.State
	case "role":
		if v := strings.TrimSpace(derefString(inv.Role)); v != "" {
			return v
		}
		return "-"
	case "expires_at":
		if v := strings.TrimSpace(derefString(inv.ExpiresAt)); v != "" {
			return normalizeTimestampCell(v)
		}
		return "-"
	default:
		return ""
	}
}

var invitationListColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return invitationRowCell(row, "id") }},
	{Header: "EMAIL", Extract: func(row any) string { return invitationRowCell(row, "email") }},
	{Header: "STATE", Extract: func(row any) string { return invitationRowCell(row, "state") }},
	{Header: "ROLE", Extract: func(row any) string { return invitationRowCell(row, "role") }},
	{Header: "EXPIRES_AT", Extract: func(row any) string { return invitationRowCell(row, "expires_at") }},
}

var invitationsCmd = &cobra.Command{
	Use:   "invitations",
	Short: "List, send, and revoke organization invitations",
	Long: `Manage invitations to your active organization.

Invitations target the organization the current session is scoped to (see
` + "`retab org`" + `). Sending and revoking are gated by people-management
permissions, so use an OAuth login (` + "`retab auth login`" + `) — an org API
key generally lacks those grants.`,
	Example: `  retab invitations list
  retab invitations create --email teammate@acme.com --role member
  retab invitations delete invitation_01HX...`,
}

var invitationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the organization's pending invitations",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		var invitations []cliInvitation
		if err := cliJSONRequestInto(cmd, http.MethodGet, invitationsBasePath, nil, nil, &invitations); err != nil {
			return err
		}

		format, err := ResolveOutputFormat(cmd, os.Stdout)
		if err != nil {
			return err
		}
		rows := make([]any, 0, len(invitations))
		for _, inv := range invitations {
			rows = append(rows, inv)
		}
		if format == OutputJSON {
			return printJSON(invitations)
		}
		if format == OutputCSV {
			return renderAutoCSV(os.Stdout, rows, invitationListColumns)
		}
		return renderAutoTable(os.Stdout, rows, invitationListColumns)
	}),
}

var invitationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Send an invitation to join the organization",
	Long: `Create (send) an invitation to an email address.

The invitee is granted --role on acceptance (default: member). Roles:
superadmin, admin, member.`,
	Example: `  retab invitations create --email teammate@acme.com
  retab invitations create --email lead@acme.com --role admin`,
	Args: cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		if strings.TrimSpace(email) == "" {
			return fmt.Errorf("--email is required")
		}
		body := map[string]any{"email": email}
		if role, _ := cmd.Flags().GetString("role"); strings.TrimSpace(role) != "" {
			if err := validateInvitationRole(role); err != nil {
				return err
			}
			body["role"] = role
		}
		var result cliInvitation
		if err := cliJSONRequestInto(cmd, http.MethodPost, invitationsBasePath, nil, body, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var invitationsDeleteCmd = &cobra.Command{
	Use:   "delete <invitation-id>",
	Short: "Revoke a pending invitation",
	Long: `Delete (revoke) a pending invitation.

Requires --yes when stdin is not a terminal.`,
	Example: `  retab invitations delete invitation_01HX... --yes`,
	Args:    cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "invitation", args[0]); err != nil {
			return err
		}
		path := invitationsBasePath + "/" + url.PathEscape(args[0])
		var result map[string]any
		if err := cliJSONRequestInto(cmd, http.MethodDelete, path, nil, nil, &result); err != nil {
			return err
		}
		confirmDeleted("invitation", args[0])
		return nil
	}),
}

func init() {
	invitationsCreateCmd.Flags().String("email", "", "email address to invite (required)")
	invitationsCreateCmd.Flags().String("role", "", "role granted on acceptance: superadmin, admin, or member (default member)")
	_ = invitationsCreateCmd.MarkFlagRequired("email")
	invitationsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	invitationsCmd.AddCommand(invitationsListCmd, invitationsCreateCmd, invitationsDeleteCmd)
	rootCmd.AddCommand(invitationsCmd)
}
