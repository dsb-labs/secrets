package note

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
)

func get() *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			note, err := client.GetNote(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get note: %w", err)
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")

			return encoder.Encode(note)
		},
	}
}
