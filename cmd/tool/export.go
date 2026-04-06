package tool

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/internal/cli"
)

func export() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export your entire secrets database",
		Long:  "Outputs all data stored for the current user in JSON format.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			exp, err := client.Export(ctx)
			if err != nil {
				return fmt.Errorf("failed to export database: %w", err)
			}

			return cli.Write(exp)
		},
	}
}
