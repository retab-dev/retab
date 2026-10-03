package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func sourcePolicyFromFlag(cmd *cobra.Command) (map[string]string, error) {
	mode, _ := cmd.Flags().GetString("sources-mode")
	if mode == "" {
		return nil, nil
	}
	if mode != "located" && mode != "cited" {
		return nil, fmt.Errorf("--sources-mode must be located or cited")
	}
	return map[string]string{"mode": mode}, nil
}
