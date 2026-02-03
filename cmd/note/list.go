package note

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
)

func list() *cobra.Command {
	var (
		query string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all notes",
		Long:    "Outputs all notes stored for the current user in JSON format.",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			notes, err := client.ListNotes(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to list notes: %w", err)
			}

			return cli.Write(notes)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&query, "query", "q", "", "filter results by search term")

	return cmd
}
