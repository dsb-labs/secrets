package note

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
)

func delete() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [id]",
		Short:   "Delete a note",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			if err := client.DeleteNote(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to delete note: %w", err)
			}

			return nil
		},
	}
}
