package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var editsCmd = &cobra.Command{
	Use:   "edits",
	Short: "Modify document content while preserving formatting",
	Long: `Apply targeted edits to documents and manage reusable edit templates.

An edit takes natural-language instructions plus either a one-off document
or a saved template (see ` + "`retab edits templates`" + `) and produces a
modified document with annotations indicating what changed. Templates
codify a set of form fields once and let you fill them against many
documents — useful for repetitive form-fill or redaction workflows.`,
}

var editsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an edit",
	Long: `Edit a document using natural-language instructions.

Provide either a document (via ` + "`--file`" + `, ` + "`--url`" + `, ` + "`--file-id`" + `,
or ` + "`--document-file`" + `) for a one-off edit, OR a ` + "`--template-id`" + `
created with ` + "`retab edits templates create`" + ` to reuse a saved set of
form fields. ` + "`--instructions`" + ` is always required and describes the edit
to perform in plain English.

Use ` + "`--color`" + ` to set the visual annotation color overlaid on the
rendered output (handy when distinguishing edits from multiple passes).`,
	Example: `  # One-off edit against an ad-hoc document
  retab edits create \
    --file ./contract.pdf \
    --instructions "Redact all personal phone numbers" \
    --model gpt-4o

  # Apply a saved template to a new document
  retab edits create \
    --template-id tmpl_abc123 \
    --file ./new-invoice.pdf \
    --instructions "Fill the standard invoice fields" \
    --model gpt-4o

  # Highlight edits in a custom color for review
  retab edits create \
    --file-id file_abc123 \
    --instructions "Mark every monetary value above $10,000" \
    --color "#ff4d4f" --model gpt-4o`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		instructions, err := requireNonBlankFlag(cmd, "instructions")
		if err != nil {
			return err
		}
		doc, err := resolveOptionalDocument(cmd)
		if err != nil {
			return err
		}
		templateID, _ := cmd.Flags().GetString("template-id")
		model, _ := cmd.Flags().GetString("model")
		color, _ := cmd.Flags().GetString("color")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		if doc == nil && templateID == "" {
			return fmt.Errorf("either a document or --template-id is required")
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		req := retab.EditsCreateParams{
			Instructions: instructions,
			Document:     doc,
			TemplateID:   ptr(templateID),
			Model:        ptr(model),
			Config:       &retab.EditConfig{Color: color},
			BustCache:    ptr(bustCache),
		}
		result, err := client.Edits.Create(ctx, &req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var editsGetCmd = &cobra.Command{
	Use:   "get <edit-id>",
	Short: "Get an edit by id",
	Long: `Fetch a single edit by id.

Returns the edit record including the resolved document, applied
instructions, model, and per-annotation details.`,
	Example: `  # Fetch a known edit
  retab edits get edit_xyz789

  # Save the result before deleting
  retab edits get edit_xyz789 > backup.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Edits.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var editsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List edits",
	Long: `List edits, newest first by default.

Filter by template with ` + "`--template-id`" + ` to see only edits derived
from a specific template. Page by edit id with ` + "`--before`" + ` /
` + "`--after`" + `.`,
	Example: `  # Most recent 25 edits
  retab edits list --limit 25

  # Only edits made against a specific template
  retab edits list --template-id tmpl_abc123 --limit 50`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.EditsListParams{PaginationParams: collectListParams(cmd)}
		if templateID, _ := cmd.Flags().GetString("template-id"); templateID != "" {
			params.TemplateID = ptr(templateID)
		}
		result, err := client.Edits.List(ctx, &params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var editsDeleteCmd = &cobra.Command{
	Use:   "delete <edit-id>",
	Short: "Delete an edit",
	Long: `Permanently delete an edit.

Destructive and irreversible. The source document and any referenced
template are not affected. Take a backup with ` + "`retab edits get`" + ` first
if you may need it.

This is destructive. Pass ` + "`--yes`" + ` to skip the confirmation prompt
in scripts and CI — otherwise the command refuses to delete when stdin
is not a terminal.`,
	Example: `  # Back up, then delete
  retab edits get edit_xyz789 > backup.json
  retab edits delete edit_xyz789

  # Skip the prompt in scripts
  retab edits delete edit_xyz789 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "edit", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Edits.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("edit", args[0])
		return nil
	}),
}

// ---- templates subgroup ----

var editsTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage edit templates",
	Long: `Manage reusable edit templates ("blueprints").

A template is a named, persisted set of form fields anchored to a sample
document. Define it once with ` + "`retab edits templates create`" + `, then
apply it to many target documents by passing ` + "`--template-id`" + ` to
` + "`retab edits create`" + `. This is the right pattern for repetitive
form-fill, redaction, or markup workflows where the field shape is stable
across documents.

Typical flow:
  1. ` + "`retab edits templates create`" + ` — define the template
  2. ` + "`retab edits create --template-id`" + ` — apply it to each new document`,
}

var editsTemplatesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an edit template",
	Long: `Create a reusable edit template from a sample document.

` + "`--name`" + ` is the human-readable label. ` + "`--form-fields-file`" + ` is a
JSON array (or ` + "`-`" + ` for stdin) describing each field on the template.
Each field must include ` + "`key`" + `, ` + "`description`" + `, ` + "`type`" + ` (` + "`text`" + ` or
` + "`checkbox`" + `), and a normalized ` + "`bbox`" + ` with ` + "`left`" + `, ` + "`top`" + `, ` + "`width`" + `,
` + "`height`" + `, and 1-based ` + "`page`" + `. The accompanying document
(supplied via the usual document flags) acts as the blueprint that future
fills are anchored against.

Once created, apply it to new documents by passing ` + "`--template-id`" + `
to ` + "`retab edits create`" + `.`,
	Example: `  # Define a template from a sample form
  retab edits templates create \
    --name "Standard Invoice Form" \
    --file ./sample-invoice.pdf \
    --form-fields-file ./fields.json

  # fields.json:
  # [
  #   {
  #     "key": "name",
  #     "description": "Full name",
  #     "type": "text",
  #     "bbox": {"left": 0.1, "top": 0.1, "width": 0.3, "height": 0.05, "page": 1}
  #   }
  # ]

  # Inline JSON via stdin
  cat ./fields.json | retab edits templates create \
    --name "W-9" --file ./w9-sample.pdf \
    --form-fields-file -`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if err := validateEditTemplateName(name); err != nil {
			return err
		}
		formFieldsPath, _ := cmd.Flags().GetString("form-fields-file")
		if formFieldsPath == "" {
			return fmt.Errorf("--form-fields-file is required")
		}
		arr, err := readJSONArray(formFieldsPath)
		if err != nil {
			return fmt.Errorf("--form-fields-file: %w", err)
		}
		fields, err := formFieldsFromJSONArray(arr)
		if err != nil {
			return fmt.Errorf("--form-fields-file: %w", err)
		}
		doc, err := resolveDocument(cmd)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.EditTemplates.Create(ctx, &retab.EditTemplatesCreateParams{
			Name:       name,
			Document:   doc,
			FormFields: fields,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var editsTemplatesGetCmd = &cobra.Command{
	Use:   "get <template-id>",
	Short: "Get an edit template",
	Long: `Fetch a single edit template by id.

Returns the template name, form fields definition, and reference to the
anchor document.`,
	Example: `  # Inspect a template's form fields
  retab edits templates get tmpl_abc123 | jq '.form_fields'`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.EditTemplates.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var editsTemplatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List edit templates",
	Long: `List edit templates, newest first by default.

Filter by name substring with ` + "`--name`" + `. Page by template id with
` + "`--before`" + ` / ` + "`--after`" + `.`,
	Example: `  # All templates
  retab edits templates list

  # Find a template by name fragment
  retab edits templates list --name invoice`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.EditTemplatesListParams{PaginationParams: collectListParams(cmd)}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			params.Name = ptr(name)
		}
		result, err := client.EditTemplates.List(ctx, &params)
		if err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var editsTemplatesUpdateCmd = &cobra.Command{
	Use:   "update <template-id>",
	Short: "Update an edit template",
	Long: `Update an existing edit template in place.

Pass ` + "`--name`" + ` to rename, ` + "`--form-fields-file`" + ` to replace the
form fields definition with a new JSON array (or ` + "`-`" + ` for stdin). At
least one of the two flags must be set. Existing fills made from this
template are not retroactively re-rendered.`,
	Example: `  # Rename a template
  retab edits templates update tmpl_abc123 --name "Invoice v2"

  # Swap in a new form-fields definition
  retab edits templates update tmpl_abc123 \
    --form-fields-file ./fields-v2.json`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		// The help promises "At least one of the two flags must be set."
		// Enforce it locally so an empty invocation fails fast instead of
		// issuing a no-op PATCH that silently bumps the template's
		// updated_at timestamp.
		formFieldsPath, _ := cmd.Flags().GetString("form-fields-file")
		if !cmd.Flags().Changed("name") && formFieldsPath == "" {
			return fmt.Errorf("at least one of --name or --form-fields-file is required")
		}
		var req retab.EditTemplatesUpdateParams
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			if err := validateEditTemplateName(v); err != nil {
				return err
			}
			req.Name = &v
		}
		if path, _ := cmd.Flags().GetString("form-fields-file"); path != "" {
			arr, err := readJSONArray(path)
			if err != nil {
				return fmt.Errorf("--form-fields-file: %w", err)
			}
			fields, err := formFieldsFromJSONArray(arr)
			if err != nil {
				return fmt.Errorf("--form-fields-file: %w", err)
			}
			req.FormFields = fields
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.EditTemplates.Update(ctx, args[0], &req)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var editsTemplatesDeleteCmd = &cobra.Command{
	Use:   "delete <template-id>",
	Short: "Delete an edit template",
	Long: `Permanently delete an edit template.

Destructive and irreversible. Existing edits made from this template are
not removed, but ` + "`retab edits create --template-id`" + ` will fail for
this id afterwards. Take a backup with ` + "`retab edits templates get`" + `
first if you may need the definition.

Pass ` + "`--yes`" + ` to skip the confirmation prompt in scripts and CI —
otherwise the command refuses to delete when stdin is not a terminal.`,
	Example: `  # Back up the form fields, then delete
  retab edits templates get tmpl_abc123 > backup.json
  retab edits templates delete tmpl_abc123

  # Skip the prompt in scripts
  retab edits templates delete tmpl_abc123 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "edit template", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.EditTemplates.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("edit template", args[0])
		return nil
	}),
}

func formFieldsFromJSONArray(arr []any) ([]*retab.FormField, error) {
	fields := make([]*retab.FormField, 0, len(arr))
	for i, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("form_fields[%d] must be a JSON object", i)
		}
		if err := validateFormFieldObject(i, obj); err != nil {
			return nil, err
		}
		raw, err := json.Marshal(obj)
		if err != nil {
			return nil, fmt.Errorf("form_fields[%d]: %w", i, err)
		}
		var field retab.FormField
		if err := json.Unmarshal(raw, &field); err != nil {
			return nil, fmt.Errorf("form_fields[%d]: %w", i, err)
		}
		fields = append(fields, &field)
	}
	return fields, nil
}

func validateEditTemplateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("template name is required")
	}
	return nil
}

func validateFormFieldObject(index int, obj map[string]any) error {
	for _, key := range []string{"key", "description", "type", "bbox"} {
		if _, ok := obj[key]; !ok {
			return fmt.Errorf("form_fields[%d] missing required field %q", index, key)
		}
	}
	fieldType, ok := obj["type"].(string)
	if !ok || (fieldType != "text" && fieldType != "checkbox") {
		return fmt.Errorf("form_fields[%d].type must be \"text\" or \"checkbox\"", index)
	}
	bbox, ok := obj["bbox"].(map[string]any)
	if !ok {
		return fmt.Errorf("form_fields[%d].bbox must be a JSON object", index)
	}
	for _, key := range []string{"left", "top", "width", "height", "page"} {
		if _, ok := bbox[key]; !ok {
			return fmt.Errorf("form_fields[%d].bbox missing required field %q", index, key)
		}
	}
	return nil
}

func init() {
	addDocumentFlags(editsCreateCmd)
	editsCreateCmd.Flags().String("instructions", "", "instructions (required)")
	editsCreateCmd.Flags().String("template-id", "", "edit template id")
	editsCreateCmd.Flags().String("model", "", "model identifier")
	editsCreateCmd.Flags().String("color", "", "edit color")
	editsCreateCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	_ = editsCreateCmd.MarkFlagRequired("instructions")

	addListFlags(editsListCmd, false)
	editsListCmd.Flags().String("template-id", "", "filter by template id")

	addDocumentFlags(editsTemplatesCreateCmd)
	editsTemplatesCreateCmd.Flags().String("name", "", "template name (required)")
	editsTemplatesCreateCmd.Flags().String("form-fields-file", "", "JSON array of form_fields with key, description, type, and bbox (or - for stdin) (required)")
	_ = editsTemplatesCreateCmd.MarkFlagRequired("name")
	_ = editsTemplatesCreateCmd.MarkFlagRequired("form-fields-file")

	addListFlags(editsTemplatesListCmd, false)
	editsTemplatesListCmd.Flags().String("name", "", "filter by template name")

	editsTemplatesUpdateCmd.Flags().String("name", "", "update template name")
	editsTemplatesUpdateCmd.Flags().String("form-fields-file", "", "JSON array of form_fields with key, description, type, and bbox (or - for stdin)")

	editsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")
	editsTemplatesDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")

	editsTemplatesCmd.AddCommand(editsTemplatesCreateCmd, editsTemplatesGetCmd, editsTemplatesListCmd, editsTemplatesUpdateCmd, editsTemplatesDeleteCmd)
	editsCmd.AddCommand(editsCreateCmd, editsGetCmd, editsListCmd, editsDeleteCmd, editsTemplatesCmd)
	rootCmd.AddCommand(editsCmd)
}
