package note

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dsb-labs/secrets/internal/cli"
	"github.com/dsb-labs/secrets/pkg/secrets"
)

func create() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name] [content]",
		Short: "Create a new note",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			note := secrets.Note{
				Name:    args[0],
				Content: args[1],
			}

			id, err := client.CreateNote(ctx, note)
			if err != nil {
				return err
			}

			fmt.Println(id)
			return nil
		},
	}
}
