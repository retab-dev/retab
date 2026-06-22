package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var primitiveListColumns = []TableColumn{
	{Header: "ID", Extract: func(row any) string { return primitiveListCell(row, "id") }},
	{Header: "STATUS", Extract: func(row any) string { return primitiveListCell(row, "status") }},
	{Header: "MODEL", Extract: func(row any) string { return primitiveListCell(row, "model") }},
	{Header: "CREATED_AT", Extract: func(row any) string {
		return normalizeTimestampCell(primitiveListCell(row, "created_at"))
	}},
}

func printPrimitiveListResult(cmd *cobra.Command, result any) error {
	raw := ""
	if cmd != nil {
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil {
			raw = f.Value.String()
		}
	}
	switch raw {
	case string(OutputTable):
		return RenderList(os.Stdout, OutputTable, result, primitiveListColumns)
	case string(OutputCSV):
		return RenderList(os.Stdout, OutputCSV, result, primitiveListColumns)
	default:
		return printJSON(result)
	}
}

func primitiveListCell(row any, key string) string {
	value, ok := rowField(row, key)
	if !ok && key == "status" {
		value, ok = rowField(row, "lifecycle.status")
	}
	if !ok {
		return ""
	}
	return stringifyCell(value)
}
