package login

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
)

func delete() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [id]",
		Short:   "Delete a login",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			if err := client.DeleteLogin(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to delete login: %w", err)
			}

			return nil
		},
	}
}
