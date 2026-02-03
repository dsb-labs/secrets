package account

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
)

func info() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Display the current user's account information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			account, err := client.GetAccount(ctx)
			if err != nil {
				return fmt.Errorf("failed to get account: %w", err)
			}

			return cli.Write(account)
		},
	}
}
