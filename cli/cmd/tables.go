//go:build !retab_oagen_cli_tables

package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

var tablesCmd = &cobra.Command{
	Use:   "tables",
	Short: "Manage CSV tables",
}

var tablesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a table",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		name, err := requireNonBlankFlag(cmd, "name")
		if err != nil {
			return err
		}
		file, err := requireNonBlankFlag(cmd, "file")
		if err != nil {
			return err
		}
		projectID, err := requireNonBlankFlag(cmd, "project-id")
		if err != nil {
			return err
		}
		fields := map[string]string{"name": name, "project_id": projectID}
		if err := addColumnSchemaOverridesField(cmd, fields); err != nil {
			return err
		}
		if err := runTableCSVUpload(cmd, http.MethodPost, "/v1/tables", file, fields); err != nil {
			return err
		}
		return nil
	}),
}

var tablesDeleteCmd = &cobra.Command{
	Use:   "delete <table-id>",
	Short: "Delete a table",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive(cmd, "table", args[0]); err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		if err := client.Tables.Delete(ctx, args[0]); err != nil {
			return err
		}
		if tableSelected(cmd) {
			return renderTableDeleteResult(args[0])
		}
		confirmDeleted("table", args[0])
		return nil
	}),
}

var tablesDownloadCmd = &cobra.Command{
	Use:   "download <table-id>",
	Short: "Download a table CSV",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		return runTableDownload(cmd, args[0])
	}),
}

var tablesGetCmd = &cobra.Command{
	Use:   "get <table-id>",
	Short: "Get a table",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Tables.Get(ctx, args[0])
		if err != nil {
			return err
		}
		// Unwrap the single-table `{table: {...}}` envelope for JSON output so
		// `tables get` matches `tables create`/`replace` (which already print the
		// table object flat). `list` keeps its plural `{tables: [...]}` shape.
		return printTableMutationResult(cmd, result)
	}),
}

var tablesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tables",
	Args:  cobra.NoArgs,
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := &retab.TablesListParams{}
		if projectID, _ := cmd.Flags().GetString("project-id"); projectID != "" {
			params.ProjectID = ptr(projectID)
		}
		result, err := client.Tables.List(ctx, params)
		if err != nil {
			return err
		}
		return printTableCommandResult(cmd, result)
	}),
}

var tablesQueryCmd = &cobra.Command{
	Use:   "query <table-id>",
	Short: "Query table rows",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		body, err := buildTableQueryBody(cmd)
		if err != nil {
			return err
		}
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		result, err := queryTableRowsForCLI(cmd, client, args[0], body)
		if err != nil {
			return err
		}
		return printTableQueryResult(cmd, result, tableQueryRenderOptionsFromFlags(cmd))
	}),
}

var tablesSchemaCmd = &cobra.Command{
	Use:   "schema <table-id>",
	Short: "Get table schema",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		var result any
		if err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/tables/"+url.PathEscape(args[0])+"/schema", nil, nil, &result); err != nil {
			return err
		}
		return printTableCommandResult(cmd, result)
	}),
}

var tablesProfileCmd = &cobra.Command{
	Use:   "profile <table-id>",
	Short: "Profile table columns",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		query := url.Values{}
		if v, _ := cmd.Flags().GetString("select"); v != "" || cmd.Flags().Changed("select") {
			if columns := splitCommaList(v); len(columns) > 0 {
				query.Set("select", strings.Join(columns, ","))
			}
		}
		var result any
		if err := cliJSONRequestInto(cmd, http.MethodGet, "/v1/tables/"+url.PathEscape(args[0])+"/profile", query, nil, &result); err != nil {
			return err
		}
		return printTableCommandResult(cmd, result)
	}),
}

var tablesValidateCmd = &cobra.Command{
	Use:   "validate <table-id>",
	Short: "Validate table shape and values",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		body, err := buildTableValidateBody(cmd)
		if err != nil {
			return err
		}
		var result any
		if err := cliJSONRequestInto(cmd, http.MethodPost, "/v1/tables/"+url.PathEscape(args[0])+"/validate", nil, body, &result); err != nil {
			return err
		}
		return printTableCommandResult(cmd, result)
	}),
}

var tablesReplaceCmd = &cobra.Command{
	Use:   "replace <table-id>",
	Short: "Replace a table CSV",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		file, err := requireNonBlankFlag(cmd, "file")
		if err != nil {
			return err
		}
		fields := map[string]string{}
		if err := addColumnSchemaOverridesField(cmd, fields); err != nil {
			return err
		}
		if err := runTableCSVUpload(cmd, http.MethodPut, "/v1/tables/"+url.PathEscape(args[0]), file, fields); err != nil {
			return err
		}
		return nil
	}),
}

// addColumnSchemaOverridesField validates the --column-schema-overrides flag (a
// JSON object mapping column name -> JSON schema) and adds it to the multipart
// fields. Validating locally gives instant feedback and avoids a wasted upload,
// matching how --filters is validated before the request is sent.
func addColumnSchemaOverridesField(cmd *cobra.Command, fields map[string]string) error {
	v, _ := cmd.Flags().GetString("column-schema-overrides")
	if v == "" && !cmd.Flags().Changed("column-schema-overrides") {
		return nil
	}
	if strings.TrimSpace(v) != "" {
		var overrides map[string]any
		if err := json.Unmarshal([]byte(v), &overrides); err != nil {
			return fmt.Errorf("--column-schema-overrides must be a JSON object: %w", err)
		}
	}
	fields["column_schema_overrides"] = v
	return nil
}

func runTableCSVUpload(cmd *cobra.Command, method string, path string, filePath string, fields map[string]string) error {
	var result any
	if err := cliMultipartRequestInto(cmd, method, path, nil, fields, "file", filePath, &result); err != nil {
		return err
	}
	return printTableMutationResult(cmd, result)
}

func printTableMutationResult(cmd *cobra.Command, result any) error {
	if format := tableOutputFormat(cmd); format == OutputTable || format == OutputCSV {
		return renderTableCommandResult(format, result)
	}
	table, ok, err := singleTableMutationResult(result)
	if err != nil {
		return err
	}
	if ok {
		return printJSON(table)
	}
	return printJSON(result)
}

func singleTableMutationResult(result any) (map[string]any, bool, error) {
	normalized, err := normalizeTableCommandResult(result)
	if err != nil {
		return nil, false, err
	}
	if table, ok := mapValue(normalized["table"]); ok {
		return table, true, nil
	}
	if tables, ok := mapSliceValue(normalized["tables"]); ok && len(tables) > 0 {
		return tables[0], true, nil
	}
	return nil, false, nil
}

func runTableDownload(cmd *cobra.Command, tableID string) error {
	body, err := cliRawRequestBytes(cmd, http.MethodGet, "/v1/tables/"+url.PathEscape(tableID)+"/download", nil, nil, "text/csv")
	if err != nil {
		return err
	}
	// -o/--out writes the raw CSV to a file (parity with `files download`); `-o -`
	// is an explicit stdout. Without -o, fall back to the existing behavior:
	// table-render when --output table is selected, else raw bytes to stdout.
	outPath, _ := cmd.Flags().GetString("out")
	if outPath != "" && outPath != "-" {
		if err := os.WriteFile(outPath, body, 0o600); err != nil {
			return fmt.Errorf("write table CSV to %s: %w", outPath, err)
		}
		return nil
	}
	if outPath == "" && tableSelected(cmd) {
		return renderDownloadedCSVTable(body)
	}
	_, err = os.Stdout.Write(body)
	return err
}

func printTableCommandResult(cmd *cobra.Command, result any) error {
	if format := tableOutputFormat(cmd); format == OutputTable || format == OutputCSV {
		return renderTableCommandResult(format, result)
	}
	return printJSON(tableCommandResultForPrint(result))
}

// tableOutputFormat resolves the root --output flag for the table commands,
// which share a custom multi-shape renderer rather than the generic
// RenderList path. It defaults to JSON (the historical default) on any
// resolution error so the commands degrade to machine-readable output.
func tableOutputFormat(cmd *cobra.Command) OutputFormat {
	var w io.Writer
	if cmd != nil {
		w = cmd.OutOrStdout()
	}
	format, err := ResolveOutputFormat(cmd, w)
	if err != nil {
		return OutputJSON
	}
	return format
}

func renderTableCommandResult(format OutputFormat, result any) error {
	normalized, err := normalizeTableCommandResult(result)
	if err != nil {
		return err
	}
	if tables, ok := mapSliceValue(normalized["tables"]); ok {
		return renderTableMetadataRows(format, tables)
	}
	if table, ok := mapValue(normalized["table"]); ok {
		return renderTableMetadataRows(format, []map[string]any{table})
	}
	if diagnostics, ok := mapSliceValue(normalized["diagnostics"]); ok {
		return renderTableValidationRows(format, normalized, diagnostics)
	}
	if columns, ok := mapSliceValue(normalized["columns"]); ok {
		if looksLikeProfileColumns(columns) {
			return renderTableProfileRows(format, columns)
		}
		return renderTableSchemaRows(format, columns)
	}
	if format == OutputCSV {
		return printResultCSV(result)
	}
	return printResultTable(result)
}

func normalizeTableCommandResult(result any) (map[string]any, error) {
	if result == nil {
		return map[string]any{}, nil
	}
	result = tableCommandResultForPrint(result)
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("render table result: marshal failed: %w", err)
	}
	var normalized map[string]any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		return nil, fmt.Errorf("render table result: decode failed: %w", err)
	}
	return normalized, nil
}

func tableCommandResultForPrint(result any) any {
	switch typed := result.(type) {
	case *retab.WorkflowTableListResponse:
		tables := typed.Tables
		if tables == nil {
			tables = []*retab.WorkflowTable{}
		}
		return map[string]any{"tables": tables}
	case retab.WorkflowTableListResponse:
		tables := typed.Tables
		if tables == nil {
			tables = []*retab.WorkflowTable{}
		}
		return map[string]any{"tables": tables}
	default:
		return result
	}
}

func renderTableMetadataRows(format OutputFormat, rows []map[string]any) error {
	columns := []TableColumn{
		tableMapColumn("id", "id"),
		tableMapColumn("name", "name"),
		tableMapColumn("filename", "filename"),
		tableMapColumn("rows", "row_count"),
		{
			Header: "columns",
			Extract: func(row any) string {
				table, _ := row.(map[string]any)
				return tableColumnNamesSummary(table["columns"])
			},
		},
		tableMapColumn("updated_at", "updated_at"),
	}
	return renderMapRowsTable(format, rows, columns)
}

func renderTableSchemaRows(format OutputFormat, rows []map[string]any) error {
	columns := []TableColumn{
		tableMapColumn("name", "name"),
		{
			Header: "type",
			Extract: func(row any) string {
				column, _ := row.(map[string]any)
				return tableJSONSchemaField(column, "type")
			},
		},
		{
			Header: "format",
			Extract: func(row any) string {
				column, _ := row.(map[string]any)
				return tableJSONSchemaField(column, "format")
			},
		},
		tableMapColumn("required", "required"),
		tableMapColumn("unique", "unique"),
		tableMapColumn("sample_values", "sample_values"),
	}
	return renderMapRowsTable(format, rows, columns)
}

func renderTableProfileRows(format OutputFormat, rows []map[string]any) error {
	columns := []TableColumn{
		tableMapColumn("name", "name"),
		{
			Header: "type",
			Extract: func(row any) string {
				column, _ := row.(map[string]any)
				return tableJSONSchemaField(column, "type")
			},
		},
		tableMapColumn("rows", "row_count"),
		tableMapColumn("nulls", "null_count"),
		tableMapColumn("empty", "empty_count"),
		tableMapColumn("distinct", "distinct_count"),
		tableMapColumn("min", "min"),
		tableMapColumn("max", "max"),
		tableMapColumn("samples", "sample_values"),
	}
	return renderMapRowsTable(format, rows, columns)
}

func renderTableValidationRows(format OutputFormat, result map[string]any, diagnostics []map[string]any) error {
	if len(diagnostics) == 0 {
		status := "ok"
		if hasErrors, ok := result["has_errors"].(bool); ok && hasErrors {
			status = "error"
		}
		return renderMapRowsTable(
			format,
			[]map[string]any{{
				"table_id":    result["table_id"],
				"status":      status,
				"diagnostics": 0,
			}},
			[]TableColumn{
				tableMapColumn("table_id", "table_id"),
				tableMapColumn("status", "status"),
				tableMapColumn("diagnostics", "diagnostics"),
			},
		)
	}
	columns := []TableColumn{
		tableMapColumn("severity", "severity"),
		tableMapColumn("column", "column"),
		tableMapColumn("rule", "rule"),
		tableMapColumn("message", "message"),
	}
	return renderMapRowsTable(format, diagnostics, columns)
}

func renderTableDeleteResult(tableID string) error {
	return renderMapRowsTable(
		OutputTable,
		[]map[string]any{{
			"id":     tableID,
			"status": "deleted",
		}},
		[]TableColumn{
			tableMapColumn("id", "id"),
			tableMapColumn("status", "status"),
		},
	)
}

func renderDownloadedCSVTable(body []byte) error {
	reader := csv.NewReader(strings.NewReader(string(body)))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("render downloaded CSV table: %w", err)
	}
	if len(records) == 0 {
		return renderAutoTableWithEmptyHint(os.Stdout, nil, defaultEmptyAutoColumns, os.Stdout)
	}
	headers := normalizeCSVHeaders(records[0])
	rows := make([]map[string]any, 0, len(records)-1)
	for _, record := range records[1:] {
		row := map[string]any{}
		for index, header := range headers {
			if index < len(record) {
				row[header] = record[index]
			} else {
				row[header] = ""
			}
		}
		rows = append(rows, row)
	}
	columns := make([]TableColumn, 0, len(headers))
	for _, header := range headers {
		columns = append(columns, tableMapColumn(header, header))
	}
	return renderMapRowsTable(OutputTable, rows, columns)
}

func normalizeCSVHeaders(headers []string) []string {
	used := map[string]bool{}
	normalized := make([]string, 0, len(headers))
	for index, header := range headers {
		name := strings.TrimSpace(header)
		if name == "" {
			name = fmt.Sprintf("column_%d", index+1)
		}
		// Disambiguate duplicates, but keep probing suffixes until the result
		// is genuinely unused — a one-shot `name_N` can itself collide with an
		// existing header (e.g. ["a","a","a_2"] -> "a_2" twice), and these
		// names become map keys downstream where a collision silently drops a
		// column's data.
		if used[name] {
			base := name
			for n := 2; used[name]; n++ {
				name = fmt.Sprintf("%s_%d", base, n)
			}
		}
		used[name] = true
		normalized = append(normalized, name)
	}
	if len(normalized) == 0 {
		return []string{"value"}
	}
	return normalized
}

func renderMapRowsTable(format OutputFormat, rows []map[string]any, columns []TableColumn) error {
	tabulableRows := make([]any, 0, len(rows))
	for _, row := range rows {
		tabulableRows = append(tabulableRows, row)
	}
	if format == OutputCSV {
		return renderAutoCSV(os.Stdout, tabulableRows, columns)
	}
	return renderAutoTableWithEmptyHint(os.Stdout, tabulableRows, columns, os.Stdout)
}

func tableMapColumn(header string, key string) TableColumn {
	return TableColumn{
		Header: header,
		Extract: func(row any) string {
			values, ok := row.(map[string]any)
			if !ok {
				return ""
			}
			return tableCell(values[key])
		},
	}
}

func tableCell(value any) string {
	return workflowTableCellText(value, tableQueryRenderOptions{MaxWidth: 96})
}

func tableColumnNamesSummary(value any) string {
	columns, ok := mapSliceValue(value)
	if !ok {
		return tableCell(value)
	}
	names := []string{}
	for _, column := range columns {
		if name := strings.TrimSpace(tableCell(column["name"])); name != "" && name != "-" {
			names = append(names, name)
		}
	}
	return strings.Join(names, ",")
}

func tableJSONSchemaField(column map[string]any, field string) string {
	schema, ok := mapValue(column["json_schema"])
	if !ok {
		return "-"
	}
	return tableCell(schema[field])
}

func looksLikeProfileColumns(columns []map[string]any) bool {
	for _, column := range columns {
		if _, ok := column["distinct_count"]; ok {
			return true
		}
		if _, ok := column["null_count"]; ok {
			return true
		}
	}
	return false
}

func mapSliceValue(value any) ([]map[string]any, bool) {
	raw, ok := value.([]any)
	if !ok {
		return nil, false
	}
	rows := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		row, ok := mapValue(item)
		if !ok {
			return nil, false
		}
		rows = append(rows, row)
	}
	return rows, true
}

func mapValue(value any) (map[string]any, bool) {
	row, ok := value.(map[string]any)
	return row, ok
}

type tableQueryRenderOptions struct {
	ShowRowID    bool
	ShowPosition bool
	MaxWidth     int
	NoTruncate   bool
	CSV          bool
}

func tableQueryRenderOptionsFromFlags(cmd *cobra.Command) tableQueryRenderOptions {
	maxWidth, _ := cmd.Flags().GetInt("max-width")
	csvFlag, _ := cmd.Flags().GetBool("csv")
	rawOutput := ""
	if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
		rawOutput = f.Value.String()
	}
	showRowID, _ := cmd.Flags().GetBool("show-row-id")
	showPosition, _ := cmd.Flags().GetBool("show-position")
	noTruncate, _ := cmd.Flags().GetBool("no-truncate")
	return tableQueryRenderOptions{
		ShowRowID:    showRowID,
		ShowPosition: showPosition,
		MaxWidth:     maxWidth,
		NoTruncate:   noTruncate,
		CSV:          csvFlag || rawOutput == string(OutputCSV),
	}
}

func buildTableQueryBody(cmd *cobra.Command) (map[string]any, error) {
	body := map[string]any{}
	filters := []any{}
	if v, _ := cmd.Flags().GetString("filters"); v != "" || cmd.Flags().Changed("filters") {
		if err := json.Unmarshal([]byte(v), &filters); err != nil {
			return nil, fmt.Errorf("--filters must be a JSON array: %w", err)
		}
	}
	whereFlags, _ := cmd.Flags().GetStringArray("where")
	for _, rawWhere := range whereFlags {
		filter, err := parseTableWhereFlag(rawWhere)
		if err != nil {
			return nil, err
		}
		filters = append(filters, filter)
	}
	if len(filters) > 0 {
		body["filters"] = filters
	}
	if v, _ := cmd.Flags().GetString("select"); v != "" || cmd.Flags().Changed("select") {
		body["select"] = splitCommaList(v)
	}
	if v, _ := cmd.Flags().GetString("search"); v != "" || cmd.Flags().Changed("search") {
		search := map[string]any{"query": v}
		if columns, _ := cmd.Flags().GetString("search-columns"); columns != "" || cmd.Flags().Changed("search-columns") {
			search["columns"] = splitCommaList(columns)
		}
		body["search"] = search
	} else if cmd.Flags().Changed("search-columns") {
		return nil, fmt.Errorf("--search-columns requires --search")
	}
	if v, _ := cmd.Flags().GetBool("case-sensitive"); v || cmd.Flags().Changed("case-sensitive") {
		body["case_sensitive"] = v
	}
	sortFlags, _ := cmd.Flags().GetStringArray("sort")
	if len(sortFlags) > 0 {
		if cmd.Flags().Changed("sort-column") || cmd.Flags().Changed("sort-direction") {
			return nil, fmt.Errorf("--sort cannot be combined with --sort-column or --sort-direction")
		}
		sortRules := []any{}
		for _, rawSort := range sortFlags {
			rule, err := parseTableSortFlag(rawSort)
			if err != nil {
				return nil, err
			}
			sortRules = append(sortRules, rule)
		}
		body["sort"] = sortRules
	} else {
		if v, _ := cmd.Flags().GetString("sort-column"); v != "" || cmd.Flags().Changed("sort-column") {
			body["sort_column"] = v
		}
		if v, _ := cmd.Flags().GetString("sort-direction"); v != "" || cmd.Flags().Changed("sort-direction") {
			body["sort_direction"] = v
		}
	}
	if v, _ := cmd.Flags().GetString("distinct"); v != "" || cmd.Flags().Changed("distinct") {
		body["distinct"] = map[string]any{"column": strings.TrimSpace(v)}
	}
	if v, _ := cmd.Flags().GetString("group-by"); v != "" || cmd.Flags().Changed("group-by") {
		body["group_by"] = splitCommaList(v)
	}
	aggregationFlags, _ := cmd.Flags().GetStringArray("aggregate")
	if len(aggregationFlags) > 0 {
		aggregations := []any{}
		for _, rawAggregation := range aggregationFlags {
			aggregation, err := parseTableAggregationFlag(rawAggregation)
			if err != nil {
				return nil, err
			}
			aggregations = append(aggregations, aggregation)
		}
		body["aggregations"] = aggregations
	}
	if v, _ := cmd.Flags().GetBool("count"); v || cmd.Flags().Changed("count") {
		body["count_only"] = v
	}
	if v, _ := cmd.Flags().GetBool("explain"); v || cmd.Flags().Changed("explain") {
		body["include_explain"] = v
	}
	if v, _ := cmd.Flags().GetInt("sample"); cmd.Flags().Changed("sample") {
		if v <= 0 {
			return nil, fmt.Errorf("--sample must be greater than zero")
		}
		body["sample"] = map[string]any{"size": v}
	}
	if v, _ := cmd.Flags().GetInt("tail"); cmd.Flags().Changed("tail") {
		if v <= 0 {
			return nil, fmt.Errorf("--tail must be greater than zero")
		}
		body["tail"] = map[string]any{"size": v}
	}
	if v, _ := cmd.Flags().GetInt("offset"); cmd.Flags().Changed("offset") {
		body["offset"] = v
	}
	if v, _ := cmd.Flags().GetInt("limit"); cmd.Flags().Changed("limit") {
		body["limit"] = v
	}
	if v, _ := cmd.Flags().GetString("viewer-mode"); v != "" || cmd.Flags().Changed("viewer-mode") {
		body["viewer_mode"] = v
	}
	return body, nil
}

func queryTableRowsForCLI(cmd *cobra.Command, client *retab.Client, tableID string, body map[string]any) (*retab.WorkflowTableRowsResponse, error) {
	all, _ := cmd.Flags().GetBool("all")
	if !all {
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		return client.Tables.Query(ctx, tableID, &retab.TablesQueryParams{Body: body})
	}

	pageBody := copyMap(body)
	if _, ok := pageBody["limit"]; !ok {
		pageBody["limit"] = 500
	}
	startOffset := 0
	if rawOffset, ok := pageBody["offset"].(int); ok {
		startOffset = rawOffset
	}
	pageBody["offset"] = startOffset
	var merged *retab.WorkflowTableRowsResponse
	for {
		ctx, cancel := ctxFor(cmd)
		page, err := client.Tables.Query(ctx, tableID, &retab.TablesQueryParams{Body: pageBody})
		cancel()
		if err != nil {
			return nil, err
		}
		if merged == nil {
			merged = page
		} else if page != nil {
			merged.Rows = append(merged.Rows, page.Rows...)
			merged.HasMore = page.HasMore
			merged.NextCursor = page.NextCursor
			merged.PreviousCursor = page.PreviousCursor
		}
		if page == nil || page.HasMore == nil || !*page.HasMore {
			break
		}
		nextOffset := startOffset
		if page.Offset != nil {
			nextOffset = *page.Offset
		}
		nextOffset += len(page.Rows)
		if nextOffset <= startOffset {
			return nil, fmt.Errorf("--all pagination did not advance")
		}
		startOffset = nextOffset
		pageBody["offset"] = nextOffset
	}
	if merged == nil {
		return nil, nil
	}
	hasMore := false
	merged.HasMore = &hasMore
	return merged, nil
}

func printTableQueryResult(cmd *cobra.Command, result *retab.WorkflowTableRowsResponse, options tableQueryRenderOptions) error {
	if options.CSV {
		return renderWorkflowTableRowsCSV(result, options)
	}
	if tableSelected(cmd) {
		return renderWorkflowTableRowsTable(result, options)
	}
	return printJSON(result)
}

func renderWorkflowTableRowsTable(result *retab.WorkflowTableRowsResponse, options tableQueryRenderOptions) error {
	if result == nil {
		return renderAutoTableWithEmptyHint(os.Stdout, nil, defaultEmptyAutoColumns, os.Stdout)
	}

	columnNames := workflowTableColumnNames(result)
	columns := make([]TableColumn, 0, len(columnNames)+2)
	if options.ShowRowID {
		columns = append(columns, TableColumn{
			Header: "row_id",
			Extract: func(row any) string {
				tableRow, ok := row.(*retab.WorkflowTableRow)
				if !ok || tableRow == nil {
					return ""
				}
				return tableRow.ID
			},
		})
	}
	if options.ShowPosition {
		columns = append(columns, TableColumn{
			Header: "position",
			Extract: func(row any) string {
				tableRow, ok := row.(*retab.WorkflowTableRow)
				if !ok || tableRow == nil {
					return ""
				}
				return strconv.Itoa(tableRow.Position)
			},
		})
	}
	for _, columnName := range columnNames {
		name := columnName
		columns = append(columns, TableColumn{
			Header: name,
			Extract: func(row any) string {
				tableRow, ok := row.(*retab.WorkflowTableRow)
				if !ok || tableRow == nil {
					return ""
				}
				return workflowTableCellText(tableRow.Data[name], options)
			},
		})
	}

	rows := make([]any, 0, len(result.Rows))
	for _, row := range result.Rows {
		rows = append(rows, row)
	}
	if err := renderAutoTableWithEmptyHint(os.Stdout, rows, columns, os.Stdout); err != nil {
		return err
	}
	emitWorkflowTableRowsSummary(os.Stdout, result)
	return nil
}

func renderWorkflowTableRowsCSV(result *retab.WorkflowTableRowsResponse, options tableQueryRenderOptions) error {
	if result == nil {
		return nil
	}
	columnNames := workflowTableColumnNames(result)
	writer := csv.NewWriter(os.Stdout)
	header := []string{}
	if options.ShowRowID {
		header = append(header, "row_id")
	}
	if options.ShowPosition {
		header = append(header, "position")
	}
	header = append(header, columnNames...)
	if err := writer.Write(header); err != nil {
		return err
	}
	for _, row := range result.Rows {
		if row == nil {
			// result.Rows is []*WorkflowTableRow; a null array element decodes
			// to a nil pointer. The table renderer and workflowTableColumnNames
			// both guard this — mirror them so the CSV path can't panic.
			continue
		}
		record := []string{}
		if options.ShowRowID {
			record = append(record, row.ID)
		}
		if options.ShowPosition {
			record = append(record, strconv.Itoa(row.Position))
		}
		for _, name := range columnNames {
			record = append(record, workflowTableCellText(row.Data[name], options))
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}
	emitWorkflowTableRowsSummary(os.Stderr, result)
	return nil
}

func workflowTableColumnNames(result *retab.WorkflowTableRowsResponse) []string {
	names := []string{}
	seen := map[string]bool{}
	for _, column := range result.Columns {
		if column == nil || column.Name == "" || seen[column.Name] {
			continue
		}
		names = append(names, column.Name)
		seen[column.Name] = true
	}
	if len(names) > 0 {
		return names
	}

	for _, row := range result.Rows {
		if row == nil {
			continue
		}
		for name := range row.Data {
			if name == "" || seen[name] {
				continue
			}
			names = append(names, name)
			seen[name] = true
		}
	}
	sort.Strings(names)
	return names
}

func workflowTableCellText(value any, options tableQueryRenderOptions) string {
	// CSV output must be a faithful, re-importable table: null/empty cells render
	// as empty fields, not the human-readable "-" placeholder used in the data
	// grid (which would corrupt the value on round-trip).
	emptyCell := "-"
	if options.CSV {
		emptyCell = ""
	}
	if value == nil {
		return emptyCell
	}
	var cleaned string
	if workflowTableCellNeedsJSON(value) {
		encoded, err := json.Marshal(value)
		if err == nil {
			cleaned = cleanWorkflowTableCell(string(encoded), options)
			if cleaned == "" {
				return emptyCell
			}
			return cleaned
		}
	}
	cleaned = cleanWorkflowTableCell(stringifyCell(value), options)
	if cleaned == "" {
		return emptyCell
	}
	return cleaned
}

func workflowTableCellNeedsJSON(value any) bool {
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
		return true
	default:
		return false
	}
}

func cleanWorkflowTableCell(value string, options tableQueryRenderOptions) string {
	cleaned := strings.Join(strings.Fields(value), " ")
	// CSV output must be a faithful, re-importable table (see
	// workflowTableCellText): truncating a long cell to "<prefix>..." would
	// silently corrupt the value on round-trip, so width limits apply only to
	// the human-readable table grid.
	if options.NoTruncate || options.CSV {
		return cleaned
	}
	maxCellWidth := options.MaxWidth
	if maxCellWidth <= 0 {
		maxCellWidth = 96
	}
	if maxCellWidth < 4 {
		maxCellWidth = 4
	}
	// Truncate by runes, not bytes: a multibyte (accented / CJK) cell would
	// otherwise be cut mid-rune and emit invalid UTF-8.
	runes := []rune(cleaned)
	if len(runes) > maxCellWidth {
		return string(runes[:maxCellWidth-3]) + "..."
	}
	return cleaned
}

func emitWorkflowTableRowsSummary(w io.Writer, result *retab.WorkflowTableRowsResponse) {
	if result == nil {
		return
	}
	totalRows := result.RowCount
	if result.FilteredRowCount != nil {
		totalRows = *result.FilteredRowCount
	}
	shownRows := len(result.Rows)
	startOffset := 0
	if result.Offset != nil {
		startOffset = *result.Offset
	}
	if shownRows == 0 {
		fmt.Fprintf(w, "showing 0 of %d rows\n", totalRows)
	} else {
		fmt.Fprintf(w, "showing %d of %d rows\n", shownRows, totalRows)
	}
	if result.HasMore == nil || !*result.HasMore {
		return
	}
	nextOffset := startOffset + shownRows
	fmt.Fprintf(w, "more rows available; pass --offset %d to continue.\n", nextOffset)
}

func splitCommaList(raw string) []string {
	parts := strings.Split(raw, ",")
	values := []string{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func copyMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func parseTableWhereFlag(raw string) (map[string]any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("--where cannot be empty")
	}
	for _, operator := range []string{"is-not-empty", "is_not_empty", "is-not-null", "is_not_null", "is-empty", "is_empty", "is-null", "is_null"} {
		suffix := " " + operator
		if strings.HasSuffix(strings.ToLower(trimmed), suffix) {
			column := strings.TrimSpace(trimmed[:len(trimmed)-len(suffix)])
			if column == "" {
				return nil, fmt.Errorf("--where %q is missing a column", raw)
			}
			return map[string]any{
				"column":   column,
				"operator": normalizeTableOperator(operator),
			}, nil
		}
	}
	for _, operator := range []string{"not-contains", "not_contains", "starts-with", "starts_with", "ends-with", "ends_with", "not-in", "not_in", "between", "contains", "gte", "lte", "ne", "gt", "lt", "eq", "in"} {
		needle := " " + operator + " "
		index := strings.Index(strings.ToLower(trimmed), needle)
		if index < 0 {
			continue
		}
		column := strings.TrimSpace(trimmed[:index])
		value := strings.TrimSpace(trimmed[index+len(needle):])
		if column == "" || value == "" {
			return nil, fmt.Errorf("--where %q must be COLUMN %s VALUE", raw, operator)
		}
		normalizedOperator := normalizeTableOperator(operator)
		whereValue, err := tableWhereValue(normalizedOperator, value)
		if err != nil {
			return nil, fmt.Errorf("--where %q: %w", raw, err)
		}
		return map[string]any{
			"column":   column,
			"operator": normalizedOperator,
			"value":    whereValue,
		}, nil
	}
	for _, candidate := range []struct {
		token    string
		operator string
	}{
		{">=", "gte"},
		{"<=", "lte"},
		{"!=", "ne"},
		{"=", "eq"},
		{">", "gt"},
		{"<", "lt"},
	} {
		if index := strings.Index(trimmed, candidate.token); index >= 0 {
			column := strings.TrimSpace(trimmed[:index])
			value := strings.TrimSpace(trimmed[index+len(candidate.token):])
			if column == "" || value == "" {
				return nil, fmt.Errorf("--where %q must be COLUMN%sVALUE", raw, candidate.token)
			}
			return map[string]any{
				"column":   column,
				"operator": candidate.operator,
				"value":    coerceTableScalarValue(value),
			}, nil
		}
	}
	return nil, fmt.Errorf("--where %q is not understood", raw)
}

func normalizeTableOperator(operator string) string {
	return strings.ReplaceAll(operator, "-", "_")
}

func tableWhereValue(operator string, value string) (any, error) {
	switch operator {
	case "in", "not_in":
		return splitCommaList(value), nil
	case "between":
		// "between" needs exactly two non-empty bounds. Split on ".." (range
		// form) or "," (list form), then validate both branches the same way.
		// Without the arity/emptiness check, a one-sided ("100..", "..200"),
		// single ("100"), or over-long ("1..2..3", "1,2,3") input would send a
		// malformed value the server rejects with an opaque 400.
		var lo, hi string
		if strings.Contains(value, "..") {
			if parts := strings.Split(value, ".."); len(parts) == 2 {
				lo, hi = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			}
		} else if parts := splitCommaList(value); len(parts) == 2 {
			lo, hi = parts[0], parts[1]
		}
		if lo == "" || hi == "" {
			return nil, fmt.Errorf("operator \"between\" requires two values: use \"A..B\" or \"A,B\"")
		}
		return []any{coerceTableScalarValue(lo), coerceTableScalarValue(hi)}, nil
	}
	return coerceTableScalarValue(value), nil
}

func coerceTableScalarValue(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	switch lower {
	case "true":
		return true
	case "false":
		return false
	case "null":
		return nil
	}
	// Decode with UseNumber so large integers (IDs, account numbers beyond
	// 2^53) keep full precision instead of being mangled by float64. A bare
	// json.Unmarshal would silently round them, sending a different value than
	// the user typed. dec.More() rejects trailing junk so "12 34" stays a
	// string, matching json.Unmarshal's whole-input semantics.
	dec := json.NewDecoder(strings.NewReader(trimmed))
	dec.UseNumber()
	var decoded any
	if err := dec.Decode(&decoded); err == nil && !dec.More() {
		return decoded
	}
	return trimmed
}

func parseTableSortFlag(raw string) (map[string]any, error) {
	parts := strings.Split(raw, ":")
	column := strings.TrimSpace(parts[0])
	if column == "" {
		return nil, fmt.Errorf("--sort %q is missing a column", raw)
	}
	direction := "asc"
	if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
		direction = strings.ToLower(strings.TrimSpace(parts[1]))
	}
	if direction != "asc" && direction != "desc" {
		return nil, fmt.Errorf("--sort direction must be asc or desc")
	}
	return map[string]any{"column": column, "direction": direction}, nil
}

func parseTableAggregationFlag(raw string) (map[string]any, error) {
	parts := strings.Split(raw, ":")
	function := strings.TrimSpace(parts[0])
	if function == "" {
		return nil, fmt.Errorf("--aggregate %q is missing a function", raw)
	}
	aggregation := map[string]any{"function": function}
	if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" && strings.TrimSpace(parts[1]) != "*" {
		aggregation["column"] = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 && strings.TrimSpace(parts[2]) != "" {
		aggregation["alias"] = strings.TrimSpace(parts[2])
	}
	return aggregation, nil
}

func buildTableValidateBody(cmd *cobra.Command) (map[string]any, error) {
	if rawBody, _ := cmd.Flags().GetString("body"); rawBody != "" || cmd.Flags().Changed("body") {
		var body map[string]any
		if err := json.Unmarshal([]byte(rawBody), &body); err != nil {
			return nil, fmt.Errorf("--body must be a JSON object: %w", err)
		}
		return body, nil
	}
	body := map[string]any{}
	if required, _ := cmd.Flags().GetString("required"); required != "" || cmd.Flags().Changed("required") {
		body["required_columns"] = splitCommaList(required)
	}
	columnRules := map[string]map[string]any{}
	notEmptyFlags, _ := cmd.Flags().GetStringArray("not-empty")
	for _, column := range notEmptyFlags {
		name := strings.TrimSpace(column)
		if name == "" {
			continue
		}
		rule := tableValidationColumnRule(columnRules, name)
		rule["is_not_empty"] = true
	}
	typeFlags, _ := cmd.Flags().GetStringArray("type")
	for _, rawType := range typeFlags {
		column, value, err := splitAssignmentFlag("--type", rawType)
		if err != nil {
			return nil, err
		}
		rule := tableValidationColumnRule(columnRules, column)
		rule["type"] = value
	}
	formatFlags, _ := cmd.Flags().GetStringArray("format-rule")
	for _, rawFormat := range formatFlags {
		column, value, err := splitAssignmentFlag("--format-rule", rawFormat)
		if err != nil {
			return nil, err
		}
		rule := tableValidationColumnRule(columnRules, column)
		rule["format"] = value
	}
	if len(columnRules) > 0 {
		body["columns"] = columnRules
	}
	uniqueFlags, _ := cmd.Flags().GetStringArray("unique")
	if len(uniqueFlags) > 0 {
		uniqueRules := []any{}
		for _, rawUnique := range uniqueFlags {
			columns := splitCommaList(rawUnique)
			if len(columns) == 0 {
				return nil, fmt.Errorf("--unique cannot be empty")
			}
			uniqueRules = append(uniqueRules, columns)
		}
		body["unique"] = uniqueRules
	}
	return body, nil
}

func tableValidationColumnRule(rules map[string]map[string]any, column string) map[string]any {
	rule := rules[column]
	if rule == nil {
		rule = map[string]any{}
		rules[column] = rule
	}
	return rule
}

func splitAssignmentFlag(flagName string, raw string) (string, string, error) {
	column, value, ok := strings.Cut(raw, "=")
	if !ok {
		return "", "", fmt.Errorf("%s must be COLUMN=VALUE", flagName)
	}
	column = strings.TrimSpace(column)
	value = strings.TrimSpace(value)
	if column == "" || value == "" {
		return "", "", fmt.Errorf("%s must be COLUMN=VALUE", flagName)
	}
	return column, value, nil
}

func init() {
	tablesCreateCmd.Flags().String("name", "", "table name (required)")
	_ = tablesCreateCmd.MarkFlagRequired("name")
	tablesCreateCmd.Flags().String("file", "", "CSV file path (required)")
	_ = tablesCreateCmd.MarkFlagRequired("file")
	tablesCreateCmd.Flags().String("project-id", "", "project that will own the table (required)")
	_ = tablesCreateCmd.MarkFlagRequired("project-id")
	tablesCreateCmd.Flags().String("column-schema-overrides", "", "JSON column schema overrides")
	tablesDeleteCmd.Flags().BoolP("yes", "y", false, "skip the confirmation prompt (required when stdin is not a TTY)")
	tablesQueryCmd.Flags().String("filters", "", "JSON array of filter rules")
	tablesQueryCmd.Flags().StringArray("where", []string{}, "filter rule, e.g. countrycode=GB or holidaystartdate>=2026-01-01 (repeatable)")
	tablesQueryCmd.Flags().String("select", "", "comma-separated columns to return")
	tablesQueryCmd.Flags().String("search", "", "search text across table columns")
	tablesQueryCmd.Flags().String("search-columns", "", "comma-separated columns for --search")
	tablesQueryCmd.Flags().Bool("case-sensitive", false, "make string filters and search case-sensitive")
	tablesQueryCmd.Flags().String("sort-column", "", "column to sort by")
	tablesQueryCmd.Flags().Var(newEnumStringFlagValue("--sort-direction", "asc", "desc"), "sort-direction", "sort direction: asc | desc")
	tablesQueryCmd.Flags().StringArray("sort", []string{}, "sort rule column:asc or column:desc (repeatable)")
	tablesQueryCmd.Flags().String("distinct", "", "return distinct values for a column with counts")
	tablesQueryCmd.Flags().String("group-by", "", "comma-separated group-by columns for aggregations")
	tablesQueryCmd.Flags().StringArray("aggregate", []string{}, "aggregation function[:column[:alias]], e.g. count or sum:amount (repeatable)")
	tablesQueryCmd.Flags().Bool("count", false, "return only the filtered row count")
	tablesQueryCmd.Flags().Bool("all", false, "fetch every page until all rows are returned")
	tablesQueryCmd.Flags().Bool("csv", false, "render query rows as CSV")
	tablesQueryCmd.Flags().Bool("show-row-id", false, "include the stable row id in table or CSV output")
	tablesQueryCmd.Flags().Bool("show-position", false, "include the zero-based CSV row position in table or CSV output")
	tablesQueryCmd.Flags().Int("max-width", 96, "maximum displayed cell width for table output")
	tablesQueryCmd.Flags().Bool("no-truncate", false, "do not truncate table cells")
	tablesQueryCmd.Flags().Int("sample", 0, "return a deterministic sample of N rows")
	tablesQueryCmd.Flags().Int("tail", 0, "return the last N matching rows")
	tablesQueryCmd.Flags().Bool("explain", false, "include query explanation metadata")
	tablesQueryCmd.Flags().Var(&nonNegativeIntFlagValue{}, "offset", "zero-based row offset")
	tablesQueryCmd.Flags().Var(&nonNegativeIntFlagValue{}, "limit", "max rows to return")
	tablesQueryCmd.Flags().String("viewer-mode", "", "viewer mode; use windowed for large tables")
	tablesDownloadCmd.Flags().StringP("out", "o", "", "write the CSV to this path, - for stdout (default: stdout)")
	tablesProfileCmd.Flags().String("select", "", "comma-separated columns to profile")
	tablesValidateCmd.Flags().String("body", "", "raw JSON validation request")
	tablesValidateCmd.Flags().String("required", "", "comma-separated required columns")
	tablesValidateCmd.Flags().StringArray("not-empty", []string{}, "column that must not contain null or empty cells (repeatable)")
	tablesValidateCmd.Flags().StringArray("unique", []string{}, "comma-separated unique key columns (repeatable)")
	tablesValidateCmd.Flags().StringArray("type", []string{}, "column type rule COLUMN=TYPE (repeatable)")
	tablesValidateCmd.Flags().StringArray("format-rule", []string{}, "column format rule COLUMN=FORMAT (repeatable)")
	tablesReplaceCmd.Flags().String("file", "", "CSV file path (required)")
	_ = tablesReplaceCmd.MarkFlagRequired("file")
	tablesReplaceCmd.Flags().String("column-schema-overrides", "", "JSON column schema overrides")
	tablesListCmd.Flags().String("project-id", "", "project whose tables to list (required)")
	_ = tablesListCmd.MarkFlagRequired("project-id")

	tablesCmd.AddCommand(tablesCreateCmd, tablesDeleteCmd, tablesDownloadCmd, tablesGetCmd, tablesListCmd, tablesProfileCmd, tablesQueryCmd, tablesReplaceCmd, tablesSchemaCmd, tablesValidateCmd)
	rootCmd.AddCommand(tablesCmd)
}
