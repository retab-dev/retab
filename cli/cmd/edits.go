package cmd

import (
	"fmt"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var editsCmd = &cobra.Command{
	Use:   "edits",
	Short: "Run and manage document edits",
}

var editsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an edit",
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
	Args:  cobra.ExactArgs(1),
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
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Edits.Delete(ctx, args[0])
	}),
}

// ---- templates subgroup ----

var editsTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage edit templates",
}

var editsTemplatesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an edit template",
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
	Args:  cobra.ExactArgs(1),
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
	Args:  cobra.ExactArgs(1),
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
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Edits.Templates.Delete(ctx, args[0])
	}),
}

var editsTemplatesFillCmd = &cobra.Command{
	Use:   "fill",
	Short: "Fill an edit template into an edit",
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
