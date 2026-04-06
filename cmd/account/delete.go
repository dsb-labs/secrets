package account

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/internal/cli"
)

func delete() *cobra.Command {
	return &cobra.Command{
		Use:     "delete",
		Short:   "Delete your account",
		Aliases: []string{"rm"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			ok, err := cli.PromptYesNo("Are you sure you want to delete your account?")
			switch {
			case err != nil:
				return fmt.Errorf("failed to read from console: %w", err)
			case !ok:
				return nil
			}

			if err = client.DeleteAccount(ctx); err != nil {
				return fmt.Errorf("failed to delete account: %w", err)
			}

			return nil
		},
	}
}
