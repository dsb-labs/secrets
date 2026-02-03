package login

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/davidsbond/keeper/internal/cli"
)

func get() *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get a login",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := cli.ClientFromContext(ctx)

			login, err := client.GetLogin(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get login: %w", err)
			}

			return cli.Write(login)
		},
	}
}
