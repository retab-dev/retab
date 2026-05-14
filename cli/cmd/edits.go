package cmd

import (
	"fmt"

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
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		doc, err := resolveOptionalDocument(cmd)
		if err != nil {
			return err
		}
		instructions, _ := cmd.Flags().GetString("instructions")
		templateID, _ := cmd.Flags().GetString("template-id")
		model, _ := cmd.Flags().GetString("model")
		color, _ := cmd.Flags().GetString("color")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		if doc == nil && templateID == "" {
			return fmt.Errorf("either a document or --template-id is required")
		}
		req := retab.EditCreateRequest{
			Instructions: instructions,
			Document:     doc,
			TemplateID:   templateID,
			Model:        model,
			Color:        color,
			BustCache:    bustCache,
		}
		result, err := client.Edits.Create(ctx, req)
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
from a specific template. Cursor-paginate with ` + "`--before`" + ` /
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
		params := retab.ListEditsParams{ListParams: collectListParams(cmd)}
		params.TemplateID, _ = cmd.Flags().GetString("template-id")
		result, err := client.Edits.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var editsDeleteCmd = &cobra.Command{
	Use:   "delete <edit-id>",
	Short: "Delete an edit",
	Long: `Permanently delete an edit.

Destructive and irreversible. The source document and any referenced
template are not affected. Take a backup with ` + "`retab edits get`" + ` first
if you may need it.`,
	Example: `  # Back up, then delete
  retab edits get edit_xyz789 > backup.json
  retab edits delete edit_xyz789`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
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
apply it to many target documents with ` + "`retab edits templates fill`" + `
(or by passing ` + "`--template-id`" + ` to ` + "`retab edits create`" + `).
This is the right pattern for repetitive form-fill, redaction, or markup
workflows where the field shape is stable across documents.

Typical flow:
  1. ` + "`retab edits templates create`" + ` — define the template
  2. ` + "`retab edits templates fill`" + ` — apply it to each new document`,
}

var editsTemplatesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an edit template",
	Long: `Create a reusable edit template from a sample document.

` + "`--name`" + ` is the human-readable label. ` + "`--form-fields-file`" + ` is a
JSON array (or ` + "`-`" + ` for stdin) describing each field on the template —
typically name, label, type, and anchor info. The accompanying document
(supplied via the usual document flags) acts as the blueprint that future
fills are anchored against.

Once created, apply it to new documents with
` + "`retab edits templates fill`" + ` or by passing ` + "`--template-id`" + ` to
` + "`retab edits create`" + `.`,
	Example: `  # Define a template from a sample form
  retab edits templates create \
    --name "Standard Invoice Form" \
    --file ./sample-invoice.pdf \
    --form-fields-file ./fields.json

  # Inline JSON via stdin
  cat ./fields.json | retab edits templates create \
    --name "W-9" --file ./w9-sample.pdf \
    --form-fields-file -`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		name, _ := cmd.Flags().GetString("name")
		doc, err := resolveDocument(cmd)
		if err != nil {
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
		var fields []retab.FormField
		for _, item := range arr {
			obj, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("--form-fields-file: each item must be a JSON object")
			}
			fields = append(fields, retab.FormField(obj))
		}
		result, err := client.Edits.Templates.Create(ctx, retab.EditTemplateCreateRequest{
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
		result, err := client.Edits.Templates.Get(ctx, args[0])
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

Filter by name substring with ` + "`--name`" + `. Cursor-paginate with
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
		params := retab.ListEditTemplatesParams{ListParams: collectListParams(cmd)}
		params.Name, _ = cmd.Flags().GetString("name")
		result, err := client.Edits.Templates.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
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
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		var req retab.EditTemplateUpdateRequest
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			req.Name = &v
		}
		if path, _ := cmd.Flags().GetString("form-fields-file"); path != "" {
			arr, err := readJSONArray(path)
			if err != nil {
				return fmt.Errorf("--form-fields-file: %w", err)
			}
			for _, item := range arr {
				obj, ok := item.(map[string]any)
				if !ok {
					return fmt.Errorf("--form-fields-file: each item must be a JSON object")
				}
				req.FormFields = append(req.FormFields, retab.FormField(obj))
			}
		}
		result, err := client.Edits.Templates.Update(ctx, args[0], req)
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
not removed, but ` + "`retab edits templates fill --template-id`" + ` will fail
for this id afterwards. Take a backup with
` + "`retab edits templates get`" + ` first if you may need the definition.`,
	Example: `  # Back up the form fields, then delete
  retab edits templates get tmpl_abc123 > backup.json
  retab edits templates delete tmpl_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Edits.Templates.Delete(ctx, args[0]); err != nil {
			return err
		}
		confirmDeleted("edit template", args[0])
		return nil
	}),
}

var editsTemplatesFillCmd = &cobra.Command{
	Use:   "fill",
	Short: "Fill an edit template into an edit",
	Long: `Apply a saved edit template against a target document.

` + "`--template-id`" + ` selects the template (see
` + "`retab edits templates list`" + `). ` + "`--instructions`" + ` describes what the
fill pass should do — typically guidance about which fields to prioritize
or how to interpret ambiguous content. The resulting edit is persisted and
returned in full.

For one-off edits without a template, use ` + "`retab edits create`" + ` instead.`,
	Example: `  # Fill a template against a new document
  retab edits templates fill \
    --template-id tmpl_abc123 \
    --instructions "Fill all fields from the document below" \
    --model gpt-4o

  # Override the annotation color for this fill
  retab edits templates fill \
    --template-id tmpl_abc123 \
    --instructions "Fill required fields only" \
    --color "#1677ff" --model gpt-4o`,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		templateID, _ := cmd.Flags().GetString("template-id")
		instructions, _ := cmd.Flags().GetString("instructions")
		model, _ := cmd.Flags().GetString("model")
		color, _ := cmd.Flags().GetString("color")
		bustCache, _ := cmd.Flags().GetBool("bust-cache")
		result, err := client.Edits.Templates.Fill(ctx, retab.EditTemplateFillRequest{
			TemplateID:   templateID,
			Instructions: instructions,
			Model:        model,
			Color:        color,
			BustCache:    bustCache,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
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
	editsTemplatesCreateCmd.Flags().String("form-fields-file", "", "JSON array of form_fields (or - for stdin)")
	_ = editsTemplatesCreateCmd.MarkFlagRequired("name")
	_ = editsTemplatesCreateCmd.MarkFlagRequired("form-fields-file")

	addListFlags(editsTemplatesListCmd, false)
	editsTemplatesListCmd.Flags().String("name", "", "filter by template name")

	editsTemplatesUpdateCmd.Flags().String("name", "", "update template name")
	editsTemplatesUpdateCmd.Flags().String("form-fields-file", "", "JSON array of form_fields (or - for stdin)")

	editsTemplatesFillCmd.Flags().String("template-id", "", "template id (required)")
	editsTemplatesFillCmd.Flags().String("instructions", "", "instructions (required)")
	editsTemplatesFillCmd.Flags().String("model", "", "model identifier")
	editsTemplatesFillCmd.Flags().String("color", "", "edit color")
	editsTemplatesFillCmd.Flags().Bool("bust-cache", false, "bypass server-side cache")
	_ = editsTemplatesFillCmd.MarkFlagRequired("template-id")
	_ = editsTemplatesFillCmd.MarkFlagRequired("instructions")

	editsTemplatesCmd.AddCommand(editsTemplatesCreateCmd, editsTemplatesGetCmd, editsTemplatesListCmd, editsTemplatesUpdateCmd, editsTemplatesDeleteCmd, editsTemplatesFillCmd)
	editsCmd.AddCommand(editsCreateCmd, editsGetCmd, editsListCmd, editsDeleteCmd, editsTemplatesCmd)
	rootCmd.AddCommand(editsCmd)
}
