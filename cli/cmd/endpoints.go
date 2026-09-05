package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const endpointAutomationPath = "/v1/processors/automations/endpoints"

type cliEndpoint struct {
	ID              string `json:"id"`
	Object          string `json:"object"`
	Name            string `json:"name"`
	ProcessorID     string `json:"processor_id"`
	DefaultLanguage string `json:"default_language,omitempty"`
	WebhookURL      string `json:"webhook_url"`
	NeedValidation  bool   `json:"need_validation"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

var endpointsCmd = &cobra.Command{
	Use:     "endpoints",
	Aliases: []string{"endpoint"},
	Short:   "Manage document-processing endpoints",
	Long: `Create and manage endpoint automations in the selected environment.

Each endpoint accepts documents at its generated endpoint id, runs the attached
processor, and delivers results to the configured webhook.`,
}

var endpointsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List endpoints",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, _ []string) error {
		if err := validateBeforeAfterMutex(cmd); err != nil {
			return err
		}
		query := url.Values{}
		for flag, key := range map[string]string{"before": "before", "after": "after", "id": "id", "name": "name", "processor-id": "processor_id", "webhook-url": "webhook_url"} {
			if value, _ := cmd.Flags().GetString(flag); value != "" {
				query.Set(key, value)
			}
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			query.Set("limit", strconv.Itoa(limit))
		}
		if order, _ := cmd.Flags().GetString("order"); order != "" {
			query.Set("order", order)
		}
		var result cliPaginatedList[cliEndpoint]
		if err := cliJSONRequestInto(cmd, http.MethodGet, endpointAutomationPath, query, nil, &result); err != nil {
			return err
		}
		format, err := ResolveOutputFormat(cmd, os.Stdout)
		if err != nil {
			return err
		}
		if format == OutputTable || format == OutputCSV {
			return RenderList(os.Stdout, format, &result, endpointColumns)
		}
		return printJSON(&result)
	}),
}

var endpointsGetCmd = &cobra.Command{
	Use:   "get <endpoint-id>",
	Short: "Get an endpoint",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		var result cliEndpoint
		if err := cliJSONRequestInto(cmd, http.MethodGet, endpointAutomationPath+"/"+url.PathEscape(args[0]), nil, nil, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var endpointsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an endpoint",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		processorID, _ := cmd.Flags().GetString("processor-id")
		webhookURL, _ := cmd.Flags().GetString("webhook-url")
		for flag, value := range map[string]string{"--name": name, "--processor-id": processorID, "--webhook-url": webhookURL} {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("%s is required", flag)
			}
		}
		headers, err := endpointHeaders(cmd)
		if err != nil {
			return err
		}
		language, _ := cmd.Flags().GetString("default-language")
		needValidation, _ := cmd.Flags().GetBool("need-validation")
		body := map[string]any{
			"name": strings.TrimSpace(name), "processor_id": strings.TrimSpace(processorID),
			"webhook_url": strings.TrimSpace(webhookURL), "default_language": strings.TrimSpace(language),
			"webhook_headers": headers, "need_validation": needValidation,
		}
		var result cliEndpoint
		if err := cliJSONRequestInto(cmd, http.MethodPost, endpointAutomationPath, nil, body, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var endpointsUpdateCmd = &cobra.Command{
	Use:   "update <endpoint-id>",
	Short: "Update an endpoint",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		for flag, key := range map[string]string{"name": "name", "processor-id": "processor_id", "webhook-url": "webhook_url", "default-language": "default_language"} {
			if cmd.Flags().Changed(flag) {
				value, _ := cmd.Flags().GetString(flag)
				if strings.TrimSpace(value) == "" {
					return fmt.Errorf("--%s must not be empty", flag)
				}
				body[key] = strings.TrimSpace(value)
			}
		}
		if cmd.Flags().Changed("webhook-header") {
			headers, err := endpointHeaders(cmd)
			if err != nil {
				return err
			}
			body["webhook_headers"] = headers
		}
		if cmd.Flags().Changed("need-validation") {
			value, _ := cmd.Flags().GetBool("need-validation")
			body["need_validation"] = value
		}
		if len(body) == 0 {
			return fmt.Errorf("provide at least one field to update")
		}
		var result cliEndpoint
		if err := cliJSONRequestInto(cmd, http.MethodPut, endpointAutomationPath+"/"+url.PathEscape(args[0]), nil, body, &result); err != nil {
			return err
		}
		return printResult(cmd, result)
	}),
}

var endpointsDeleteCmd = &cobra.Command{
	Use:   "delete <endpoint-id>",
	Short: "Delete an endpoint",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "endpoint", args[0]); err != nil {
			return err
		}
		var result map[string]any
		if err := cliJSONRequestInto(cmd, http.MethodDelete, endpointAutomationPath+"/"+url.PathEscape(args[0]), nil, nil, &result); err != nil {
			return err
		}
		confirmDeleted("endpoint", args[0])
		return nil
	}),
}

func endpointHeaders(cmd *cobra.Command) (map[string]string, error) {
	values, _ := cmd.Flags().GetStringArray("webhook-header")
	headers := make(map[string]string, len(values))
	for _, value := range values {
		key, headerValue, ok := strings.Cut(value, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("--webhook-header must use KEY=VALUE")
		}
		headers[key] = headerValue
	}
	return headers, nil
}

var endpointColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return endpointCell(row, "id") }},
	{Header: "NAME", Extract: func(row any) string { return endpointCell(row, "name") }},
	{Header: "PROCESSOR_ID", Extract: func(row any) string { return endpointCell(row, "processor_id") }},
	{Header: "WEBHOOK_URL", Extract: func(row any) string { return endpointCell(row, "webhook_url") }},
	{Header: "UPDATED_AT", Extract: func(row any) string { return endpointCell(row, "updated_at") }, IsTimestamp: true},
}

func endpointCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok || cellIsEmpty(value) || !cellIsDisplayable(value) {
		return ""
	}
	return stringifyCell(value)
}

func addEndpointMutationFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "endpoint name")
	cmd.Flags().String("processor-id", "", "processor id")
	cmd.Flags().String("webhook-url", "", "result delivery webhook URL")
	cmd.Flags().String("default-language", "en", "default document language")
	cmd.Flags().StringArray("webhook-header", nil, "webhook header as KEY=VALUE (repeatable)")
	cmd.Flags().Bool("need-validation", false, "require validation before webhook delivery")
}

func init() {
	endpointsListCmd.Flags().String("before", "", "endpoint id: return items before this id")
	endpointsListCmd.Flags().String("after", "", "endpoint id: return items after this id")
	endpointsListCmd.Flags().Var(&nonNegativeIntFlagValue{}, "limit", "maximum number of endpoints to return")
	endpointsListCmd.Flags().String("order", "desc", "sort order: asc or desc")
	endpointsListCmd.Flags().String("id", "", "filter by endpoint id")
	endpointsListCmd.Flags().String("name", "", "filter by endpoint name")
	endpointsListCmd.Flags().String("processor-id", "", "filter by processor id")
	endpointsListCmd.Flags().String("webhook-url", "", "filter by webhook URL")
	addEndpointMutationFlags(endpointsCreateCmd)
	addEndpointMutationFlags(endpointsUpdateCmd)
	endpointsDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")
	endpointsCmd.AddCommand(endpointsListCmd, endpointsGetCmd, endpointsCreateCmd, endpointsUpdateCmd, endpointsDeleteCmd)
	rootCmd.AddCommand(endpointsCmd)
}
